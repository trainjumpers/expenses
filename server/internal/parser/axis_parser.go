package parser

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"expenses/internal/models"
	"expenses/pkg/logger"
	"expenses/pkg/utils"
)

// AxisParser parses Axis bank account CSV statements
type AxisParser struct{}

// Precompiled regex patterns for Axis transaction description parsing.
// Compiling once improves performance when parsing many rows.
var axisPatterns = []struct {
	regex      *regexp.Regexp
	creditName string
	debitName  string
}{
	{regexp.MustCompile(`(?i)UPI/P2[AM]/\d+/([^/]+)/`), "UPI from $1", "UPI to $1"},
	{regexp.MustCompile(`(?i)IMPS/P2A/\d+/([^/]+)/`), "IMPS from $1", "IMPS to $1"},
	{regexp.MustCompile(`(?i)NEFT/([^/]+)`), "NEFT from $1", "NEFT to $1"},
	{regexp.MustCompile(`(?i)RTGS/[^/]+/([^/]+)/`), "RTGS from $1", "RTGS to $1"},
	{regexp.MustCompile(`(?i)INT\.PD|Int\.Pd`), "Interest", "Interest"},
}

func (p *AxisParser) Parse(fileBytes []byte, metadata string, fileName string, password string) ([]models.CreateTransactionInput, error) {
	r := csv.NewReader(bytes.NewReader(fileBytes))
	r.FieldsPerRecord = -1
	// Bank CSVs can contain malformed quoting; be permissive
	r.LazyQuotes = true
	recs, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read csv: %w", err)
	}

	// find header row
	headerIdx := -1
	for i, row := range recs {
		if len(row) < 3 {
			continue
		}
		joined := strings.ToLower(strings.Join(row, " "))
		if strings.Contains(joined, "tran date") || (strings.Contains(joined, "particulars") && (strings.Contains(joined, "dr") || strings.Contains(joined, "cr"))) {
			headerIdx = i
			break
		}
	}

	if headerIdx == -1 {
		return nil, errors.New("transaction header row not found")
	}

	var transactions []models.CreateTransactionInput
	for i := headerIdx + 1; i < len(recs); i++ {
		row := recs[i]
		// trim fields
		for j := range row {
			row[j] = strings.TrimSpace(row[j])
		}

		if len(row) < 6 {
			logger.Debugf("Skipping row %d: expected at least 6 columns, got %d", i+1, len(row))
			continue
		}

		txn, err := p.parseTransactionRow(row)
		if err != nil {
			logger.Warnf("Failed to parse row %d: %v", i+1, err)
			continue
		}
		if txn != nil {
			transactions = append(transactions, *txn)
		}
	}

	return transactions, nil
}

func (p *AxisParser) parseTransactionRow(fields []string) (*models.CreateTransactionInput, error) {
	if len(fields) < 6 {
		return nil, errors.New("insufficient columns in row")
	}

	txnDateStr := strings.TrimSpace(fields[0])
	description := strings.TrimSpace(fields[2])
	debitStr := strings.TrimSpace(fields[3])
	creditStr := strings.TrimSpace(fields[4])

	if txnDateStr == "" {
		return nil, errors.New("empty transaction date")
	}

	// Normalize common date separators (31-03-2025 -> 31/03/2025) so utils.ParseDate can handle it
	txnDateStr = strings.ReplaceAll(txnDateStr, "-", "/")

	txnDate, err := utils.ParseDate(txnDateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction date '%s': %w", txnDateStr, err)
	}

	var amount float64
	var isCredit bool

	if debitStr != "" {
		val, err := utils.ParseFloat(debitStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse debit amount '%s': %w", debitStr, err)
		}
		amount = val
		isCredit = false
	} else if creditStr != "" {
		val, err := utils.ParseFloat(creditStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse credit amount '%s': %w", creditStr, err)
		}
		amount = -val // credits (incoming) are represented as negative amounts
		isCredit = true
	} else {
		return nil, errors.New("both debit and credit amounts are empty")
	}

	name := p.generateTransactionName(description, isCredit)

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

func (p *AxisParser) generateTransactionName(description string, isCredit bool) string {
	desc := strings.TrimSpace(description)

	// Use precompiled patterns for performance
	for _, pattern := range axisPatterns {
		if pattern.regex.MatchString(desc) {
			if matches := pattern.regex.FindStringSubmatch(desc); len(matches) > 1 {
				if isCredit {
					return strings.Replace(pattern.creditName, "$1", strings.TrimSpace(matches[1]), 1)
				}
				return strings.Replace(pattern.debitName, "$1", strings.TrimSpace(matches[1]), 1)
			}
			// For patterns with no capture group (Interest), fallthrough
			return pattern.creditName
		}
	} 

	prefix := ""
	if isCredit {
		prefix = "Credit: "
	} else {
		prefix = "Debit: "
	}
	n := prefix + desc
	// Truncate safely for unicode by slicing runes
	runes := []rune(n)
	if len(runes) > 40 {
		return strings.TrimSpace(string(runes[:37])) + "..."
	}
	return n
}

func init() {
	RegisterParser(models.BankTypeAxis, &AxisParser{})
}
