package parser

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"expenses/internal/models"
	"expenses/pkg/logger"
	"expenses/pkg/utils"
	"fmt"
	"path/filepath"
	"strings"
)

type CustomParser struct{}

// StatementMetadata defines the structure for custom parser configuration.
type StatementMetadata struct {
	SkipRows      int               `json:"skip_rows"`
	ColumnMapping map[string]string `json:"column_mapping"`
}

// trimRecords iterates through a 2D string slice and trims whitespace from each element.
func trimRecords(records [][]string) [][]string {
	trimmed := make([][]string, len(records))
	for i, row := range records {
		trimmedRow := make([]string, len(row))
		for j, field := range row {
			trimmedRow[j] = strings.TrimSpace(field)
		}
		trimmed[i] = trimmedRow
	}
	return trimmed
}

func (p *CustomParser) Preview(fileBytes []byte, fileName string, skipRows int, rowSize int) (*models.StatementPreview, error) {
	logger.Infof("Parser: Processing file '%s'", fileName)
	lowerFileName := strings.ToLower(fileName)
	extension := filepath.Ext(lowerFileName)

	var records [][]string
	var err error

	switch extension {
	case ".csv":
		r := csv.NewReader(bytes.NewReader(fileBytes))
		r.FieldsPerRecord = -1 // To handle "wrong number of fields" error
		records, err = r.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("failed to read csv: %w", err)
		}
	case ".xls":
		r := csv.NewReader(bytes.NewReader(fileBytes))
		r.Comma = '\t'
		r.FieldsPerRecord = -1 // To handle "wrong number of fields" error
		records, err = r.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("failed to read tsv from xls file: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported file type for preview: %s. Only .csv and .xls are supported", fileName)
	}

	// Trim whitespace from all records at once to ensure clean data.
	records = trimRecords(records)

	if len(records) <= skipRows {
		return &models.StatementPreview{Headers: []string{}, Rows: [][]string{}}, nil
	}

	// Skip rows to drop the metadata or header rows
	records = records[skipRows:]

	if len(records) == 0 {
		return &models.StatementPreview{Headers: []string{}, Rows: [][]string{}}, nil
	}

	headers := records[0]
	dataRows := records[1:]

	if rowSize != -1 && len(dataRows) > rowSize {
		dataRows = dataRows[:rowSize]
	}

	preview := &models.StatementPreview{
		Headers: headers,
		Rows:    dataRows,
	}

	return preview, nil
}

// Parse processes a file using metadata to map columns and create transactions.
func (p *CustomParser) Parse(fileBytes []byte, metadata string, fileName string) ([]models.CreateTransactionInput, error) {
	if metadata == "" {
		return nil, errors.New("metadata is required for custom parser")
	}

	var meta StatementMetadata
	if err := json.Unmarshal([]byte(metadata), &meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	preview, err := p.Preview(fileBytes, fileName, meta.SkipRows, -1)
	if err != nil {
		return nil, fmt.Errorf("failed to preview file for parsing: %w", err)
	}

	// Create a map of header names to their column index. Data is already trimmed by Preview.
	headerIndex := make(map[string]int, len(preview.Headers))
	for i, h := range preview.Headers {
		headerIndex[h] = i
	}

	columnIndex := make(map[string]int)
	for field, columnName := range meta.ColumnMapping {
		if idx, ok := headerIndex[columnName]; ok {
			columnIndex[field] = idx
		} else {
			if field != "description" {
				return nil, fmt.Errorf("mapped column '%s' not found in statement header", columnName)
			}
		}
	}

	err = p.validateMappings(columnIndex)
	if err != nil {
		return nil, err
	}

	var transactions []models.CreateTransactionInput
	for i, row := range preview.Rows {
		transaction, err := p.parseRow(row, columnIndex)
		if err != nil {
			logger.Warnf("skipping row %d due to parsing error: %v", i+meta.SkipRows+2, err)
			continue
		}
		transactions = append(transactions, *transaction)
	}

	return transactions, nil
}

func (p *CustomParser) validateMappings(columnIndex map[string]int) error {
	// Validate that all required fields are present in the column mapping.
	requiredFields := []string{"txn_date", "name"}
	for _, f := range requiredFields {
		if _, ok := columnIndex[f]; !ok {
			return fmt.Errorf("required field '%s' is not mapped in metadata", f)
		}
	}
	_, amountOk := columnIndex["amount"]
	_, creditOk := columnIndex["credit"]
	_, debitOk := columnIndex["debit"]
	if !amountOk && !(creditOk && debitOk) {
		return errors.New("insufficient amount information in metadata: map either 'amount' or both 'credit' and 'debit'")
	}
	return nil
}

// parseRow parses a single row into a transaction based on the column index map.
func (p *CustomParser) parseRow(row []string, columnIndex map[string]int) (*models.CreateTransactionInput, error) {
	dateStr, err := p.getRequiredField(row, columnIndex, "txn_date")
	if err != nil {
		return nil, err
	}
	name, err := p.getRequiredField(row, columnIndex, "name")
	if err != nil {
		return nil, err
	}
	description := p.getOptionalField(row, columnIndex, "description")

	date, err := utils.ParseDate(dateStr)
	if err != nil {
		return nil, err
	}

	amount, err := p.getAmount(row, columnIndex)
	if err != nil {
		return nil, err
	}

	tx := &models.CreateTransactionInput{
		CreateBaseTransactionInput: models.CreateBaseTransactionInput{
			Name:        name,
			Description: description,
			Amount:      amount,
			Date:        date,
		},
		CategoryIds: []int64{},
	}

	return tx, nil
}

func (p *CustomParser) getAmount(row []string, columnIndex map[string]int) (*float64, error) {
	var amount float64
	amountStr, amountOk := p.getOptionalFieldOk(row, columnIndex, "amount")
	creditStr, creditOk := p.getOptionalFieldOk(row, columnIndex, "credit")
	debitStr, debitOk := p.getOptionalFieldOk(row, columnIndex, "debit")

	if amountOk && amountStr != "" {
		val, err := utils.ParseFloat(amountStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse amount: %w", err)
		}
		amount = val
	} else if creditOk && debitOk {
		credit, err := utils.ParseFloat(creditStr)
		if err != nil && creditStr != "" {
			return nil, fmt.Errorf("failed to parse credit: %w", err)
		}
		debit, err := utils.ParseFloat(debitStr)
		if err != nil && debitStr != "" {
			return nil, fmt.Errorf("failed to parse debit: %w", err)
		}

		if credit > 0 {
			amount = credit
		} else {
			amount = -debit
		}
	} else {
		return nil, errors.New("insufficient amount information: map either 'amount' or both 'credit' and 'debit'")
	}
	return &amount, nil
}

// getRequiredField safely extracts a required field from a row.
func (p *CustomParser) getRequiredField(row []string, columnIndex map[string]int, fieldName string) (string, error) {
	idx, ok := columnIndex[fieldName]
	if !ok {
		return "", fmt.Errorf("field '%s' is not mapped", fieldName)
	}
	if idx >= len(row) {
		return "", fmt.Errorf("column index %d for '%s' is out of bounds", idx, fieldName)
	}
	return row[idx], nil
}

// getOptionalField safely extracts an optional field from a row.
func (p *CustomParser) getOptionalField(row []string, columnIndex map[string]int, fieldName string) string {
	val, _ := p.getOptionalFieldOk(row, columnIndex, fieldName)
	return val
}

// getOptionalFieldOk safely extracts an optional field and indicates if the mapping exists.
func (p *CustomParser) getOptionalFieldOk(row []string, columnIndex map[string]int, fieldName string) (string, bool) {
	idx, ok := columnIndex[fieldName]
	if !ok || idx >= len(row) {
		return "", false
	}
	return row[idx], true
}

func init() {
	RegisterParser(models.BankTypeOthers, &CustomParser{})
}
