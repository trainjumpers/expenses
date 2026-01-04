package parser

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"expenses/internal/models"
	"expenses/pkg/logger"
	"expenses/pkg/utils"

	"github.com/xuri/excelize/v2"
)

// AxisCreditParser parses Axis credit card statements exported as XLSX
type AxisCreditParser struct{}

func (p *AxisCreditParser) Parse(fileBytes []byte, metadata string, fileName string, password string) ([]models.CreateTransactionInput, error) {
	f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to open xlsx: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	for _, sheet := range f.GetSheetList() {
		rows, err := f.GetRows(sheet)
		if err != nil {
			continue
		}

		headerIndex := -1
		var dateIdx, descIdx, amountIdx, signIdx int
		for i, row := range rows {
			if len(row) == 0 {
				continue
			}
			joined := strings.ToLower(strings.Join(row, " "))
			if strings.Contains(joined, "date") && strings.Contains(joined, "amount") && (strings.Contains(joined, "debit") || strings.Contains(joined, "credit") || strings.Contains(joined, "debit/credit")) {
				headerIndex = i
				// determine column indexes
				dateIdx, descIdx, amountIdx, signIdx = -1, -1, -1, -1
				for idx, cell := range row {
					lc := strings.ToLower(strings.TrimSpace(cell))
					if dateIdx == -1 && strings.Contains(lc, "date") {
						dateIdx = idx
					}
					if descIdx == -1 && (strings.Contains(lc, "transaction") || strings.Contains(lc, "transaction details") || strings.Contains(lc, "details")) {
						descIdx = idx
					}
					if amountIdx == -1 && strings.Contains(lc, "amount") {
						amountIdx = idx
					}
					if signIdx == -1 && (strings.Contains(lc, "debit") || strings.Contains(lc, "credit")) {
						signIdx = idx
					}
				}
				break
			}
		}

		if headerIndex == -1 {
			// try next sheet
			continue
		}

		var transactions []models.CreateTransactionInput
		for i := headerIndex + 1; i < len(rows); i++ {
			row := rows[i]
			// Skip empty rows
			if len(row) == 0 {
				continue
			}

			// Normalize cell slice to avoid index issues
			for len(row) <= descIdx && len(row) < 5 {
				row = append(row, "")
			}

			txn, err := p.parseTransactionRow(row, dateIdx, descIdx, amountIdx, signIdx)
			if err != nil {
				logger.Warnf("Failed to parse row %d in sheet %s: %v", i+1, sheet, err)
				continue
			}
			if txn != nil {
				transactions = append(transactions, *txn)
			}
		}

		if len(transactions) > 0 {
			return transactions, nil
		}
	}

	return nil, errors.New("transaction header row not found in Axis credit statement")
}

func (p *AxisCreditParser) parseTransactionRow(row []string, dateIdx int, descIdx int, amountIdx int, signIdx int) (*models.CreateTransactionInput, error) {
	// Validate indexes
	if dateIdx < 0 || descIdx < 0 {
		return nil, nil
	}

	get := func(idx int) string {
		if idx >= 0 && idx < len(row) {
			return strings.TrimSpace(row[idx])
		}
		return ""
	}

	dateStr := get(dateIdx)
	if dateStr == "" || strings.EqualFold(dateStr, "date") {
		return nil, nil
	}

	// Normalize two-digit year like 14 Nov '25 -> 14 Nov 2025
	re := regexp.MustCompile(`'(?P<yy>\d{2})`)
	dateStr = re.ReplaceAllString(dateStr, "20$1")

	txnDate, err := utils.ParseDate(dateStr)
	if err != nil {
		// Not a valid transaction row
		return nil, nil
	}

	desc := get(descIdx)
	// If description is empty, try next available column
	if desc == "" {
		if len(row) > descIdx+1 {
			desc = strings.TrimSpace(row[descIdx+1])
		}
	}

	amountStr := get(amountIdx)
	var inferredSign string
	var amount float64
	// If amount not found in configured column, try to find numeric cell after description
	if amountStr == "" {
		for j := descIdx + 1; j < len(row); j++ {
			cell := strings.TrimSpace(row[j])
			if cell == "" {
				continue
			}
			if v, sgn, err := p.parseAmount(cell); err == nil {
				amountStr = cell
				inferredSign = sgn
				amount = v
				break
			}
		}
	}

	if amountStr == "" {
		return nil, fmt.Errorf("amount not found for row: %v", row)
	}

	// Try parse amount from amountStr if not already parsed
	if amount == 0 {
		v, sgn, err := p.parseAmount(amountStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse amount '%s': %w", amountStr, err)
		}
		amount = v
		if inferredSign == "" {
			inferredSign = sgn
		}
	}

	// sign handling: prefer explicit sign column; fall back to inferred sign
	sign := strings.ToLower(get(signIdx))
	if sign != "" {
		if strings.Contains(sign, "debit") || strings.Contains(sign, "dr") {
			amount = math.Abs(amount)
		} else if strings.Contains(sign, "credit") || strings.Contains(sign, "cr") {
			amount = -math.Abs(amount)
		}
	} else {
		switch inferredSign {
		case "debit":
			amount = math.Abs(amount)
		case "credit":
			amount = -math.Abs(amount)
		}
	}

	name := strings.TrimSpace(desc)
	if len(name) > 40 {
		name = name[:37] + "..."
	}

	txn := &models.CreateTransactionInput{
		CreateBaseTransactionInput: models.CreateBaseTransactionInput{
			Name:        name,
			Description: desc,
			Amount:      &amount,
			Date:        txnDate,
		},
		CategoryIds: []int64{},
	}
	return txn, nil
}

// parseAmount cleans the raw amount cell and returns numeric value and inferred sign ("debit" or "credit") if detectable.
func (p *AxisCreditParser) parseAmount(raw string) (float64, string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, "", fmt.Errorf("empty amount")
	}

	// Replace non-breaking spaces
	s = strings.ReplaceAll(s, "\u00A0", " ")
	lower := strings.ToLower(s)
	inferred := ""

	// Parentheses indicate negative amount (treat as credit/payment)
	if strings.Contains(s, "(") && strings.Contains(s, ")") {
		// don't make the numeric value negative here; infer sign as credit
		inferred = "credit"
	}

	// Trailing CR/DR indicators
	if strings.HasSuffix(strings.TrimSpace(lower), "cr") {
		inferred = "credit"
		s = strings.TrimSpace(s[:len(s)-2])
	} else if strings.HasSuffix(strings.TrimSpace(lower), "dr") {
		inferred = "debit"
		s = strings.TrimSpace(s[:len(s)-2])
	}
	// Remove currency symbols and words
	replacements := []string{"₹", "rs.", "rs", "inr"}
	for _, r := range replacements {
		s = strings.ReplaceAll(strings.ToLower(s), r, "")
	}

	// Remove commas, spaces and parentheses
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "(", "")
	s = strings.ReplaceAll(s, ")", "")

	if s == "" {
		return 0, inferred, fmt.Errorf("empty amount after cleaning")
	}

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, inferred, err
	}

	// If parentheses were present we already set inferred to credit; keep value positive
	return v, inferred, nil
}

func init() {
	RegisterParser(models.BankTypeAxisCredit, &AxisCreditParser{})
}
