package parser

import (
	"bufio"
	"bytes"
	"errors"
	"expenses/internal/models"
	"expenses/pkg/logger"
	"expenses/pkg/utils"
	"fmt"
	"regexp"
	"strings"
)

// HDFCParser parses HDFC bank statements exported as CSV-like text
type HDFCParser struct{}

func (p *HDFCParser) Parse(fileBytes []byte, metadata string, fileName string, password string) ([]models.CreateTransactionInput, error) {
	scanner := bufio.NewScanner(bytes.NewReader(fileBytes))

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	headerRowIndex := -1
	for i, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "date") && strings.Contains(lower, "narration") {
			headerRowIndex = i
			break
		}
	}

	if headerRowIndex == -1 {
		return nil, errors.New("transaction header row not found")
	}

	var transactions []models.CreateTransactionInput
	for i := headerRowIndex + 1; i < len(lines); i++ {
		raw := strings.TrimSpace(lines[i])
		if raw == "" {
			continue
		}

		// The files are comma-separated with fixed-width spacing; split and trim per-field
		fields := strings.Split(raw, ",")
		// Ensure at least the required columns exist
		if len(fields) < 6 {
			logger.Debugf("Skipping line %d: expected at least 6 columns, got %d", i+1, len(fields))
			continue
		}

		// Trim all fields
		for idx := range fields {
			fields[idx] = strings.TrimSpace(fields[idx])
		}

		transaction, err := p.parseTransactionRow(fields)
		if err != nil {
			logger.Warnf("Failed to parse line %d: %v\n", i+1, err)
			continue
		}

		if transaction != nil {
			transactions = append(transactions, *transaction)
		}
	}

	return transactions, nil
}

func (p *HDFCParser) parseTransactionRow(fields []string) (*models.CreateTransactionInput, error) {
	// Expected columns (index-based):
	// 0: Date, 1: Narration, 2: Value Date, 3: Debit Amount, 4: Credit Amount, 5: Chq/Ref Number, 6: Closing Balance
	if len(fields) < 6 {
		return nil, errors.New("insufficient columns in row")
	}

	dateStr := strings.TrimSpace(fields[0])
	narration := strings.TrimSpace(fields[1])
	debitStr := strings.TrimSpace(fields[3])
	creditStr := strings.TrimSpace(fields[4])
	refNo := strings.TrimSpace(fields[5])

	txnDate, err := utils.ParseDate(dateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction date '%s': %w", dateStr, err)
	}

	var amount float64
	var isCredit bool

	if debitStr != "" && debitStr != "0" && strings.TrimSpace(debitStr) != "0.00" {
		amount, err = utils.ParseFloat(debitStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse debit amount '%s': %w", debitStr, err)
		}
		isCredit = false
	} else if creditStr != "" && creditStr != "0" && strings.TrimSpace(creditStr) != "0.00" {
		amount, err = utils.ParseFloat(creditStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse credit amount '%s': %w", creditStr, err)
		}
		amount = -amount
		isCredit = true
	} else {
		return nil, errors.New("both debit and credit amounts are empty or zero")
	}

	name := p.generateTransactionName(narration, isCredit)

	description := narration
	if refNo != "" && refNo != "0" && refNo != "000000000000000" { // common zero placeholders
		description = fmt.Sprintf("%s (Ref: %s)", narration, refNo)
	}

	transaction := &models.CreateTransactionInput{
		CreateBaseTransactionInput: models.CreateBaseTransactionInput{
			Name:        name,
			Description: description,
			Amount:      &amount,
			Date:        txnDate,
		},
		CategoryIds: []int64{},
	}

	return transaction, nil
}

func (p *HDFCParser) generateTransactionName(narration string, isCredit bool) string {
	desc := strings.TrimSpace(narration)

	// Known patterns
	patterns := []struct {
		regex      *regexp.Regexp
		creditName string
		debitName  string
	}{
		// UPI-<NAME>-<...>
		{regexp.MustCompile(`(?i)UPI-\s*([^\-]+?)\s*-`), "UPI from $1", "UPI to $1"},
		// NEFT CR-<CODE>-<...>
		{regexp.MustCompile(`(?i)NEFT\s*CR-\s*([^\-]+)`), "NEFT from $1", "NEFT to $1"},
		// RTGS CR-<CODE>-<...>
		{regexp.MustCompile(`(?i)RTGS\s*CR-\s*([^\-]+)`), "RTGS from $1", "RTGS to $1"},
		// IMPS or POS simple labels
		{regexp.MustCompile(`(?i)IMPS`), "IMPS (Credit)", "IMPS (Debit)"},
		{regexp.MustCompile(`(?i)POS`), "Card POS (Credit)", "Card POS (Debit)"},
		{regexp.MustCompile(`(?i)INTEREST PAID`), "Interest (Credit)", "Interest (Debit)"},
	}

	for _, pattern := range patterns {
		if pattern.regex.MatchString(desc) {
			if matches := pattern.regex.FindStringSubmatch(desc); len(matches) > 1 {
				name := pattern.debitName
				if isCredit {
					name = pattern.creditName
				}
				cleaned := strings.TrimSpace(matches[1])
				cleaned = strings.Join(strings.Fields(cleaned), " ")
				return strings.Replace(name, "$1", cleaned, 1)
			}
			// If no capture group, just return label
			if isCredit {
				return pattern.creditName
			}
			return pattern.debitName
		}
	}

	if isCredit {
		desc = "Credit: " + desc
	} else {
		desc = "Debit: " + desc
	}

	if len(desc) > 25 {
		return strings.TrimSpace(desc[:22]) + "..."
	}
	return desc
}

func init() {
	RegisterParser(models.BankTypeHDFC, &HDFCParser{})
}
