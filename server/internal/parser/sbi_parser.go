package parser

import (
	"bytes"
	"errors"
	"expenses/internal/models"
	"expenses/pkg/logger"
	"expenses/pkg/utils"
	"fmt"
	"regexp"
	"strings"

	"github.com/xuri/excelize/v2"
)

type SBIParser struct{}

func (p *SBIParser) Parse(fileBytes []byte, metadata string, fileName string) ([]models.CreateTransactionInput, error) {
	f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to open XLSX file: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, errors.New("no sheets found in XLSX file")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("failed to read rows from sheet: %w", err)
	}

	headerRowIndex := -1
	for i, row := range rows {
		if len(row) < 2 {
			continue
		}
		rowStr := strings.ToLower(row[0]) + " " + strings.ToLower(row[1])
		if strings.Contains(rowStr, "date") && strings.Contains(rowStr, "details") {
			headerRowIndex = i
			break
		}
	}

	if headerRowIndex == -1 {
		return nil, errors.New("transaction header row not found")
	}

	var transactions []models.CreateTransactionInput
	for i := headerRowIndex + 1; i < len(rows); i++ {
		row := rows[i]

		if len(row) < 6 {
			logger.Debugf("Skipping row %d: expected at least 6 columns, got %d", i+1, len(row))
			continue
		}

		transaction, err := p.parseTransactionRow(row)
		if err != nil {
			logger.Warnf("Failed to parse row %d: %v\n", i+1, err)
			continue
		}

		if transaction != nil {
			transactions = append(transactions, *transaction)
		}
	}

	return transactions, nil
}

func (p *SBIParser) parseTransactionRow(fields []string) (*models.CreateTransactionInput, error) {
	if len(fields) < 6 {
		return nil, errors.New("insufficient columns in row")
	}

	txnDateStr := strings.TrimSpace(fields[0])
	description := strings.TrimSpace(fields[1])
	refNo := strings.TrimSpace(fields[2])
	debitStr := strings.TrimSpace(fields[3])
	creditStr := strings.TrimSpace(fields[4])

	txnDate, err := utils.ParseDate(txnDateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction date '%s': %w", txnDateStr, err)
	}

	var amount float64
	var isCredit bool

	if debitStr != "" && debitStr != " " {
		amount, err = utils.ParseFloat(debitStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse debit amount '%s': %w", debitStr, err)
		}
		isCredit = false
	} else if creditStr != "" && creditStr != " " {
		amount, err = utils.ParseFloat(creditStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse credit amount '%s': %w", creditStr, err)
		}
		amount = -amount
		isCredit = true
	} else {
		return nil, errors.New("both debit and credit amounts are empty")
	}

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

func (p *SBIParser) generateTransactionName(description string, isCredit bool) string {
	desc := strings.TrimSpace(description)

	prefixes := []string{
		"DEP TFR",
		"WDL TFR",
		"TO TRANSFER-",
		"BY TRANSFER-",
		"DEBIT-",
		"CREDIT-",
	}

	for _, prefix := range prefixes {
		if newDesc, ok := strings.CutPrefix(desc, prefix); ok {
			desc = strings.TrimSpace(newDesc)
			break
		}
	}

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
