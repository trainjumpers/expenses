package service

import (
	"context"
	"encoding/csv"
	"errors"
	"expenses/internal/models"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

type CustomParserServiceInterface interface {
	PreviewStatement(ctx context.Context, fileBytes []byte, fileType string) (*models.StatementPreview, error)
	ParseWithMapping(ctx context.Context, input models.ParseStatementInput) (*models.ParsedStatementResult, error)
}

type CustomParserService struct{}

func NewCustomParserService() CustomParserServiceInterface {
	return &CustomParserService{}
}

func (s *CustomParserService) PreviewStatement(ctx context.Context, fileBytes []byte, fileType string) (*models.StatementPreview, error) {
	switch strings.ToLower(fileType) {
	case "csv":
		return s.previewCSV(fileBytes)
	case "excel", "xls", "xlsx":
		return s.previewExcel(fileBytes)
	default:
		return nil, errors.New("unsupported file type")
	}
}

func (s *CustomParserService) previewCSV(fileBytes []byte) (*models.StatementPreview, error) {
	reader := csv.NewReader(strings.NewReader(string(fileBytes)))
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	var rows [][]string
	rowCount := 0
	maxPreviewRows := 10

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV: %v", err)
		}

		rows = append(rows, record)
		rowCount++

		if rowCount >= maxPreviewRows {
			break
		}
	}

	if len(rows) == 0 {
		return nil, errors.New("empty file")
	}

	// Detect headers (first row)
	headers := rows[0]
	var sampleData [][]string
	if len(rows) > 1 {
		sampleData = rows[1:]
	}

	return &models.StatementPreview{
		Headers:    headers,
		SampleData: sampleData,
		TotalRows:  rowCount,
		FileType:   "csv",
	}, nil
}

func (s *CustomParserService) previewExcel(fileBytes []byte) (*models.StatementPreview, error) {
	// Create a temporary reader from bytes
	reader := strings.NewReader(string(fileBytes))
	
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, fmt.Errorf("error opening Excel file: %v", err)
	}
	defer f.Close()

	// Get the first sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, errors.New("no sheets found in Excel file")
	}

	sheetName := sheets[0]
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("error reading Excel sheet: %v", err)
	}

	if len(rows) == 0 {
		return nil, errors.New("empty sheet")
	}

	maxPreviewRows := 10
	if len(rows) > maxPreviewRows {
		rows = rows[:maxPreviewRows]
	}

	// Detect headers (first row)
	headers := rows[0]
	var sampleData [][]string
	if len(rows) > 1 {
		sampleData = rows[1:]
	}

	return &models.StatementPreview{
		Headers:    headers,
		SampleData: sampleData,
		TotalRows:  len(rows),
		FileType:   "excel",
	}, nil
}

func (s *CustomParserService) ParseWithMapping(ctx context.Context, input models.ParseStatementInput) (*models.ParsedStatementResult, error) {
	var rows [][]string
	var err error

	switch strings.ToLower(input.FileType) {
	case "csv":
		rows, err = s.parseCSVRows(input.FileBytes)
	case "excel", "xls", "xlsx":
		rows, err = s.parseExcelRows(input.FileBytes)
	default:
		return nil, errors.New("unsupported file type")
	}

	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, errors.New("no data found in file")
	}

	// Skip header row if specified
	dataRows := rows
	if input.HasHeaders {
		if len(rows) <= 1 {
			return nil, errors.New("file only contains headers")
		}
		dataRows = rows[1:]
	}

	var transactions []models.ParsedTransaction
	var errors []string

	for i, row := range dataRows {
		transaction, parseErr := s.parseRowToTransaction(row, input.ColumnMapping, i+1)
		if parseErr != nil {
			errors = append(errors, fmt.Sprintf("Row %d: %s", i+1, parseErr.Error()))
			continue
		}
		transactions = append(transactions, *transaction)
	}

	return &models.ParsedStatementResult{
		Transactions:   transactions,
		TotalRows:      len(dataRows),
		SuccessfulRows: len(transactions),
		FailedRows:     len(errors),
		Errors:         errors,
	}, nil
}

func (s *CustomParserService) parseCSVRows(fileBytes []byte) ([][]string, error) {
	reader := csv.NewReader(strings.NewReader(string(fileBytes)))
	reader.FieldsPerRecord = -1

	return reader.ReadAll()
}

func (s *CustomParserService) parseExcelRows(fileBytes []byte) ([][]string, error) {
	reader := strings.NewReader(string(fileBytes))
	
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, fmt.Errorf("error opening Excel file: %v", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, errors.New("no sheets found")
	}

	return f.GetRows(sheets[0])
}

func (s *CustomParserService) parseRowToTransaction(row []string, mapping models.ColumnMapping, rowNum int) (*models.ParsedTransaction, error) {
	transaction := &models.ParsedTransaction{}

	// Parse date
	if mapping.DateColumn >= 0 && mapping.DateColumn < len(row) {
		dateStr := strings.TrimSpace(row[mapping.DateColumn])
		if dateStr == "" {
			return nil, errors.New("date is required")
		}

		parsedDate, err := s.parseDate(dateStr, mapping.DateFormat)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %v", err)
		}
		transaction.Date = parsedDate
	} else {
		return nil, errors.New("date column not found")
	}

	// Parse description
	if mapping.DescriptionColumn >= 0 && mapping.DescriptionColumn < len(row) {
		transaction.Description = strings.TrimSpace(row[mapping.DescriptionColumn])
	}

	// Parse amount
	if mapping.AmountColumn >= 0 && mapping.AmountColumn < len(row) {
		amountStr := strings.TrimSpace(row[mapping.AmountColumn])
		if amountStr == "" {
			return nil, errors.New("amount is required")
		}

		amount, err := s.parseAmount(amountStr)
		if err != nil {
			return nil, fmt.Errorf("invalid amount format: %v", err)
		}
		transaction.Amount = amount
	} else {
		return nil, errors.New("amount column not found")
	}

	// Parse reference (optional)
	if mapping.ReferenceColumn >= 0 && mapping.ReferenceColumn < len(row) {
		transaction.Reference = strings.TrimSpace(row[mapping.ReferenceColumn])
	}

	return transaction, nil
}

func (s *CustomParserService) parseDate(dateStr, format string) (time.Time, error) {
	// Common date formats
	formats := []string{
		"2006-01-02",
		"02/01/2006",
		"01/02/2006",
		"2006/01/02",
		"02-01-2006",
		"01-02-2006",
		"2006-01-02 15:04:05",
		"02/01/2006 15:04:05",
	}

	// If custom format is provided, try it first
	if format != "" {
		if date, err := time.Parse(format, dateStr); err == nil {
			return date, nil
		}
	}

	// Try common formats
	for _, fmt := range formats {
		if date, err := time.Parse(fmt, dateStr); err == nil {
			return date, nil
		}
	}

	return time.Time{}, errors.New("unable to parse date")
}

func (s *CustomParserService) parseAmount(amountStr string) (float64, error) {
	// Clean the amount string
	cleaned := strings.ReplaceAll(amountStr, ",", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.TrimSpace(cleaned)

	// Handle currency symbols
	cleaned = strings.TrimPrefix(cleaned, "₹")
	cleaned = strings.TrimPrefix(cleaned, "$")
	cleaned = strings.TrimPrefix(cleaned, "€")
	cleaned = strings.TrimPrefix(cleaned, "£")

	// Handle parentheses for negative amounts
	isNegative := false
	if strings.HasPrefix(cleaned, "(") && strings.HasSuffix(cleaned, ")") {
		isNegative = true
		cleaned = strings.Trim(cleaned, "()")
	}

	amount, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0, err
	}

	if isNegative {
		amount = -amount
	}

	return amount, nil
}
