package parser

import (
	"encoding/csv"
	"errors"
	"expenses/internal/models"
	"expenses/pkg/logger"
	"expenses/pkg/utils"
	"fmt"
	"io"
	"strings"
)

// ICICICreditParser is a parser for ICICI credit card statements in CSV format.
type ICICICreditParser struct{}

// Parse extracts transactions from an ICICI credit card statement.
func (p *ICICICreditParser) Parse(fileBytes []byte, metadata string, fileName string) ([]models.CreateTransactionInput, error) {
	// The CSV is not standard; it has metadata at the top. We need to find the header row first.
	// A simple way is to convert to string and find the start of the actual CSV data.
	content := string(fileBytes)
	headerIndex := strings.Index(content, `"Date","Sr.No.","Transaction Details"`)
	if headerIndex == -1 {
		return nil, errors.New("transaction header row not found in ICICI credit statement")
	}

	// Read from the header onwards
	csvContent := content[headerIndex:]
	reader := csv.NewReader(strings.NewReader(csvContent))
	reader.LazyQuotes = true // Handles potential quote inconsistencies

	// The first line is now the header
	_, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header row: %w", err)
	}

	var transactions []models.CreateTransactionInput
	lineNum := 1 // Start counting from after the header
	for {
		record, err := reader.Read()
		lineNum++
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Warnf("Skipping malformed CSV record at line %d: %v", lineNum, err)
			continue
		}

		transaction, err := p.parseTransactionRow(record)
		if err != nil {
			logger.Warnf("Failed to parse transaction row at line %d: %v", lineNum, err)
			continue
		}

		if transaction != nil {
			transactions = append(transactions, *transaction)
		}
	}

	return transactions, nil
}

// parseTransactionRow parses a single row from the CSV into a transaction.
func (p *ICICICreditParser) parseTransactionRow(fields []string) (*models.CreateTransactionInput, error) {
	// A valid transaction row should have at least 7 columns.
	// This filters out lines with just card numbers (e.g., "5241XXXXXXXX5008").
	if len(fields) < 7 {
		return nil, nil // Not an error, just a row to skip.
	}

	dateStr := strings.TrimSpace(fields[0])
	description := strings.TrimSpace(fields[2])
	amountStr := strings.TrimSpace(fields[5])
	billingSign := strings.TrimSpace(fields[6])

	// Skip what might be empty rows or header-like rows that were not filtered out.
	if dateStr == "" || dateStr == "Date" {
		return nil, nil
	}

	txnDate, err := utils.ParseDate(dateStr)
	if err != nil {
		// If date parsing fails, it's likely not a valid transaction row.
		return nil, nil
	}

	amount, err := utils.ParseFloat(amountStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse amount '%s': %w", amountStr, err)
	}

	// In credit card statements, "CR" means a credit to the account (payment/refund),
	// which reduces the amount owed. We'll represent this as a negative transaction.
	if billingSign == "CR" {
		amount = -amount
	}

	name := p.generateTransactionName(description)

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

// generateTransactionName creates a transaction name from the description.
// For now, it's a simple cleanup, but can be expanded with regex like in SBIParser.
func (p *ICICICreditParser) generateTransactionName(description string) string {
	name := strings.TrimSpace(description)
	return name
}

func init() {
	RegisterParser(models.BankTypeICICICredit, &ICICICreditParser{})
}
