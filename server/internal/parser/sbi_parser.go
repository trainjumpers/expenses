package parser

import (
	"bufio"
	"bytes"
	"errors"
	"expenses/internal/models"
	"expenses/pkg/logger"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type SBIParser struct{}

func (p *SBIParser) Parse(fileBytes []byte, metadata string, fileName string) ([]models.CreateTransactionInput, error) {
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
		if strings.Contains(strings.ToLower(line), "txn date") &&
			strings.Contains(strings.ToLower(line), "description") {
			headerRowIndex = i
			break
		}
	}

	if headerRowIndex == -1 {
		return nil, errors.New("transaction header row not found")
	}

	var transactions []models.CreateTransactionInput
	for i := headerRowIndex + 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		if line == "" {
			continue
		}

		if strings.Contains(strings.ToLower(line), "computer generated") {
			logger.Debugf("Breaking at line %d: found 'computer generated' text. End of statement reached", i+1)
			break
		}

		fields := strings.Split(line, "\t")

		if len(fields) < 7 {
			logger.Debugf("Skipping line %d: expected at least 7 columns, got %d", i+1, len(fields))
			continue
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

func (p *SBIParser) parseTransactionRow(fields []string) (*models.CreateTransactionInput, error) {
	if len(fields) < 7 {
		return nil, errors.New("insufficient columns in row")
	}

	txnDateStr := strings.TrimSpace(fields[0])
	description := strings.TrimSpace(fields[2])
	refNo := strings.TrimSpace(fields[3])
	debitStr := strings.TrimSpace(fields[4])
	creditStr := strings.TrimSpace(fields[5])

	txnDate, err := p.parseDate(txnDateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction date '%s': %w", txnDateStr, err)
	}

	var amount float64
	var isCredit bool

	if debitStr != "" && debitStr != " " {
		amount, err = p.parseAmount(debitStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse debit amount '%s': %w", debitStr, err)
		}
		isCredit = false
	} else if creditStr != "" && creditStr != " " {
		amount, err = p.parseAmount(creditStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse credit amount '%s': %w", creditStr, err)
		}
		amount = -amount
		isCredit = true
	} else {
		return nil, errors.New("both debit and credit amounts are empty")
	}

	// Generate transaction name from description
	name := p.generateTransactionName(description, isCredit)

	fullDescription := description
	if refNo != "" {
		fullDescription = fmt.Sprintf("%s (Ref: %s)", description, refNo)
	}

	transaction := &models.CreateTransactionInput{
		CreateBaseTransactionInput: models.CreateBaseTransactionInput{
			Name:        name,
			Description: fullDescription,
			Amount:      &amount,
			Date:        txnDate,
		},
		CategoryIds: []int64{},
	}

	return transaction, nil
}

func (p *SBIParser) parseDate(dateStr string) (time.Time, error) {
	layouts := []string{
		"2 Jan 2006",
		"02 Jan 2006",
		"2 January 2006",
		"02 January 2006",
	}

	for _, layout := range layouts {
		if date, err := time.Parse(layout, dateStr); err == nil {
			return date, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// parseAmount parses amount string and removes commas
func (p *SBIParser) parseAmount(amountStr string) (float64, error) {
	// Remove commas and extra spaces
	cleanAmount := strings.ReplaceAll(amountStr, ",", "")
	cleanAmount = strings.TrimSpace(cleanAmount)

	if cleanAmount == "" {
		return 0, errors.New("empty amount string")
	}

	amount, err := strconv.ParseFloat(cleanAmount, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount format: %s", amountStr)
	}

	return amount, nil
}

// generateTransactionName creates a readable transaction name from description
func (p *SBIParser) generateTransactionName(description string, isCredit bool) string {
	// Clean up the description
	desc := strings.TrimSpace(description)

	// Remove common prefixes
	prefixes := []string{
		"TO TRANSFER-",
		"BY TRANSFER-",
		"DEBIT-",
		"CREDIT-",
	}

	for _, prefix := range prefixes {
		if newDesc, ok := strings.CutPrefix(desc, prefix); ok {
			desc = newDesc
			break
		}
	}

	// Extract meaningful parts using regex patterns
	patterns := []struct {
		regex      *regexp.Regexp
		creditName string
		debitName  string
	}{
		{regexp.MustCompile(`UPI/[CD]R/\d+/([^/]+)/`), "UPI from $1", "UPI to $1"},
		{regexp.MustCompile(`NEFT\*([^*]+)\*`), "NEFT from $1", "NEFT to $1"},
		{regexp.MustCompile(`ATMCard\s+AMC\s+(\d+)`), "ATM Card AMC $1 (Credit)", "ATM Card AMC $1 (Debit)"},
		{regexp.MustCompile(`e-TDR/e-STDR\s+(\d+)`), "Term Deposit $1 (Credit)", "Term Deposit $1 (Debit)"},
	}

	for _, pattern := range patterns {
		if pattern.regex.MatchString(desc) {
			if matches := pattern.regex.FindStringSubmatch(desc); len(matches) > 1 {
				if isCredit {
					return strings.Replace(pattern.creditName, "$1", matches[1], 1)
				} else {
					return strings.Replace(pattern.debitName, "$1", matches[1], 1)
				}
			}
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
	RegisterParser(models.BankTypeSBI, &SBIParser{})
}
