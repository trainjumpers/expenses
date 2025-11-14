package parser

import (
	"bytes"
	"errors"
	"expenses/internal/models"
	"expenses/pkg/logger"
	"expenses/pkg/utils"
	"fmt"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// AxisCreditParser parses Axis bank credit card statements from Excel files.
type AxisCreditParser struct{}

func (p *AxisCreditParser) Parse(fileBytes []byte, metadata string, fileName string) ([]models.CreateTransactionInput, error) {
	f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to open excel file: %w", err)
	}

	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, errors.New("no sheets found in the excel file")
	}


rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows from sheet: %w", err)
	}

	headerRowIndex := -1
	for i, row := range rows {
		if len(row) >= 4 && strings.Contains(row[0], "Date") && strings.Contains(row[1], "Transaction Details") {
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
		if len(row) == 0 || (len(row) > 0 && strings.Contains(row[0], "** End of Statement **")) {
			break
		}

		if len(row) < 4 {
			logger.Debugf("Skipping row %d: expected at least 4 columns, got %d", i+1, len(row))
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

func (p *AxisCreditParser) parseTransactionRow(row []string) (*models.CreateTransactionInput, error) {
	dateStr := strings.TrimSpace(row[0])
	details := strings.TrimSpace(row[1])
	amountStr := strings.TrimSpace(row[2])
	typeStr := strings.TrimSpace(row[3])

	txnDate, err := time.Parse("02 Jan '06", dateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date '%s': %w", dateStr, err)
	}

	amountStr = strings.ReplaceAll(amountStr, "₹", "")
	amountStr = strings.ReplaceAll(amountStr, ",", "")
	amount, err := utils.ParseFloat(amountStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse amount '%s': %w", amountStr, err)
	}

	if strings.ToLower(typeStr) == "debit" {
		// amount is already positive
	} else if strings.ToLower(typeStr) == "credit" {
		amount = -amount
	} else {
		return nil, fmt.Errorf("unknown transaction type: %s", typeStr)
	}

	transaction := &models.CreateTransactionInput{
		CreateBaseTransactionInput: models.CreateBaseTransactionInput{
			Name:        details,
			Description: details,
			Amount:      &amount,
			Date:        txnDate,
		},
		CategoryIds: []int64{},
	}

	return transaction, nil
}

func init() {
	RegisterParser(models.BankTypeAxisCredit, &AxisCreditParser{})
}
