package parser

import (
	"expenses/internal/errors"
	"expenses/internal/models"
	"expenses/pkg/logger"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type CustomCSVParser struct{}

// NewCustomCSVParser creates a new custom CSV parser
func NewCustomCSVParser() *CustomCSVParser {
	return &CustomCSVParser{}
}

// Parse implements the BankStatementParser interface
func (p *CustomCSVParser) Parse(fileBytes []byte, metadata models.CreateStatementMetadata) ([]models.CreateTransactionInput, error) {
	logger.Debugf("CustomCSVParser.Parse: Starting to parse file with %d bytes", len(fileBytes))
	logger.Debugf("CustomCSVParser.Parse: Mappings: %+v, SkipRows: %d", metadata.Mappings, metadata.SkipRows)
	
	// Parse the CSV file
	parseResult, err := ParseCSVFile(fileBytes, "custom.csv")
	if err != nil {
		logger.Errorf("CustomCSVParser.Parse: Failed to parse CSV file: %v", err)
		return nil, err
	}
	logger.Infof("CustomCSVParser.Parse: CSV parsed - %d columns, %d total rows", len(parseResult.Columns), parseResult.Total)
	logger.Debugf("CustomCSVParser.Parse: Original columns: %v", parseResult.Columns)

	// Apply row skipping
	if metadata.SkipRows > 0 {
		logger.Debugf("CustomCSVParser.Parse: Applying row skipping: %d rows", metadata.SkipRows)
		parseResult = parseResult.ApplySkipRows(metadata.SkipRows)
		logger.Infof("CustomCSVParser.Parse: After skipping - %d columns, %d data rows", len(parseResult.Columns), len(parseResult.Rows))
		logger.Debugf("CustomCSVParser.Parse: New columns after skipping: %v", parseResult.Columns)
	}

	// Convert rows to transactions
	var transactions []models.CreateTransactionInput
	var parseErrors []string
	
	logger.Infof("CustomCSVParser.Parse: Processing %d data rows", len(parseResult.Rows))
	for i, row := range parseResult.Rows {
		logger.Debugf("CustomCSVParser.Parse: Processing row %d: %v", i+1, row)
		transaction, err := p.parseRow(parseResult.Columns, row, i+1, metadata.Mappings)
		if err != nil {
			logger.Warnf("CustomCSVParser.Parse: Failed to parse row %d: %v", i+1, err)
			parseErrors = append(parseErrors, fmt.Sprintf("Row %d: %v", i+1, err))
			continue
		}
		if transaction != nil {
			logger.Debugf("CustomCSVParser.Parse: Successfully parsed transaction from row %d: Name=%s, Amount=%v, Date=%v", 
				i+1, transaction.Name, transaction.Amount, transaction.Date)
			transactions = append(transactions, *transaction)
		} else {
			logger.Warnf("CustomCSVParser.Parse: Row %d resulted in nil transaction", i+1)
		}
	}

	logger.Infof("CustomCSVParser.Parse: Completed parsing - %d transactions created, %d errors", len(transactions), len(parseErrors))
	if len(parseErrors) > 0 {
		logger.Debugf("CustomCSVParser.Parse: Parse errors: %v", parseErrors)
	}

	return transactions, nil
}

// parseRow converts a single CSV row to a transaction
func (p *CustomCSVParser) parseRow(columns []string, row []string, rowNum int, mappings []models.ColumnMapping) (*models.CreateTransactionInput, error) {
	if len(row) != len(columns) {
		logger.Debugf("CustomCSVParser.parseRow: Row %d column count mismatch - expected %d columns, got %d columns", 
			rowNum, len(columns), len(row))
		logger.Debugf("CustomCSVParser.parseRow: Expected columns: %v", columns)
		logger.Debugf("CustomCSVParser.parseRow: Actual row data: %v", row)
		return nil, fmt.Errorf("row %d: column count mismatch", rowNum)
	}

	// Create a map of column values
	rowData := make(map[string]string)
	for i, col := range columns {
		if i < len(row) {
			rowData[col] = strings.TrimSpace(row[i])
		}
	}
	
	logger.Debugf("CustomCSVParser.parseRow: Row %d data map: %+v", rowNum, rowData)

	// Extract required fields using mappings
	name, err := p.extractName(rowData, mappings)
	if err != nil {
		return nil, fmt.Errorf("row %d: %w", rowNum, err)
	}

	amount, err := p.extractAmount(rowData, mappings)
	if err != nil {
		return nil, fmt.Errorf("row %d: %w", rowNum, err)
	}

	date, err := p.extractDate(rowData, mappings)
	if err != nil {
		return nil, fmt.Errorf("row %d: %w", rowNum, err)
	}

	description := p.extractDescription(rowData, mappings)

	transaction := &models.CreateTransactionInput{
		CreateBaseTransactionInput: models.CreateBaseTransactionInput{
			Name:        name,
			Description: description,
			Amount:      &amount,
			Date:        date,
		},
		CategoryIds: []int64{}, // Empty categories for now
	}

	return transaction, nil
}

// extractName extracts the transaction name from row data
func (p *CustomCSVParser) extractName(rowData map[string]string, mappings []models.ColumnMapping) (string, error) {
	for _, mapping := range mappings {
		if mapping.TargetField == "name" {
			if value, exists := rowData[mapping.SourceColumn]; exists && value != "" {
				return value, nil
			}
		}
	}
	return "", errors.NewMissingRequiredFieldError("name")
}

// extractAmount extracts and calculates the transaction amount
func (p *CustomCSVParser) extractAmount(rowData map[string]string, mappings []models.ColumnMapping) (float64, error) {
	// Check if amount field is directly mapped
	for _, mapping := range mappings {
		if mapping.TargetField == "amount" {
			if value, exists := rowData[mapping.SourceColumn]; exists && value != "" {
				amount, err := p.parseAmount(value)
				if err != nil {
					return 0, errors.NewInvalidAmountFormatError(err)
				}
				return amount, nil
			}
		}
	}

	// Check for separate credit/debit fields
	var creditAmount, debitAmount float64
	var hasCredit, hasDebit bool

	for _, mapping := range mappings {
		if mapping.TargetField == "credit" {
			if value, exists := rowData[mapping.SourceColumn]; exists && value != "" {
				amount, err := p.parseAmount(value)
				if err != nil {
					return 0, errors.NewInvalidAmountFormatError(err)
				}
				creditAmount = amount
				hasCredit = true
			}
		} else if mapping.TargetField == "debit" {
			if value, exists := rowData[mapping.SourceColumn]; exists && value != "" {
				amount, err := p.parseAmount(value)
				if err != nil {
					return 0, errors.NewInvalidAmountFormatError(err)
				}
				debitAmount = amount
				hasDebit = true
			}
		}
	}

	// Calculate final amount based on credit/debit
	if hasCredit && hasDebit {
		// If both exist, use debit as positive and credit as negative
		if debitAmount > 0 {
			return debitAmount, nil
		} else if creditAmount > 0 {
			return -creditAmount, nil
		}
	} else if hasCredit {
		return -creditAmount, nil
	} else if hasDebit {
		return debitAmount, nil
	}

	return 0, errors.NewMissingRequiredFieldError("amount")
}

// extractDate extracts and parses the transaction date
func (p *CustomCSVParser) extractDate(rowData map[string]string, mappings []models.ColumnMapping) (time.Time, error) {
	for _, mapping := range mappings {
		if mapping.TargetField == "date" {
			if value, exists := rowData[mapping.SourceColumn]; exists && value != "" {
				date, err := p.parseDate(value)
				if err != nil {
					return time.Time{}, errors.NewInvalidDateFormatError(err)
				}
				return date, nil
			}
		}
	}
	return time.Time{}, errors.NewMissingRequiredFieldError("date")
}

// extractDescription extracts the transaction description (optional)
func (p *CustomCSVParser) extractDescription(rowData map[string]string, mappings []models.ColumnMapping) string {
	for _, mapping := range mappings {
		if mapping.TargetField == "description" {
			if value, exists := rowData[mapping.SourceColumn]; exists {
				return value
			}
		}
	}
	return ""
}

// parseAmount parses amount string and handles various formats
func (p *CustomCSVParser) parseAmount(amountStr string) (float64, error) {
	// Clean the amount string
	cleanAmount := strings.TrimSpace(amountStr)
	cleanAmount = strings.ReplaceAll(cleanAmount, ",", "")
	cleanAmount = strings.ReplaceAll(cleanAmount, " ", "")

	// Handle currency symbols
	cleanAmount = strings.TrimPrefix(cleanAmount, "$")
	cleanAmount = strings.TrimPrefix(cleanAmount, "₹")
	cleanAmount = strings.TrimPrefix(cleanAmount, "Rs.")
	cleanAmount = strings.TrimPrefix(cleanAmount, "Rs")

	// Handle parentheses (negative amounts)
	isNegative := false
	if strings.HasPrefix(cleanAmount, "(") && strings.HasSuffix(cleanAmount, ")") {
		isNegative = true
		cleanAmount = strings.Trim(cleanAmount, "()")
	}

	if cleanAmount == "" {
		return 0, fmt.Errorf("empty amount string")
	}

	amount, err := strconv.ParseFloat(cleanAmount, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount format: %s", amountStr)
	}

	if isNegative {
		amount = -amount
	}

	return amount, nil
}

// parseDate parses date string in various formats
func (p *CustomCSVParser) parseDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	
	// Common date formats
	layouts := []string{
		"2006-01-02",           // YYYY-MM-DD
		"02/01/2006",           // DD/MM/YYYY
		"01/02/2006",           // MM/DD/YYYY
		"2006/01/02",           // YYYY/MM/DD
		"02-01-2006",           // DD-MM-YYYY
		"01-02-2006",           // MM-DD-YYYY
		"2006-01-02 15:04:05",  // YYYY-MM-DD HH:MM:SS
		"02/01/2006 15:04:05",  // DD/MM/YYYY HH:MM:SS
		"2 Jan 2006",           // D MMM YYYY
		"02 Jan 2006",          // DD MMM YYYY
		"2 January 2006",       // D MMMM YYYY
		"02 January 2006",      // DD MMMM YYYY
		"Jan 2, 2006",          // MMM D, YYYY
		"January 2, 2006",      // MMMM D, YYYY
	}

	for _, layout := range layouts {
		if date, err := time.Parse(layout, dateStr); err == nil {
			return date, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}