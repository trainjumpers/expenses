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

	"github.com/xuri/excelize/v2"
)

type CustomParser struct{}

// StatementMetadata defines the structure for custom parser configuration.
type StatementMetadata struct {
	SkipRows      int               `json:"skip_rows"`
	ColumnMapping map[string]string `json:"column_mapping"`
}

var ErrWorkbookPasswordRequired = errors.New("workbook password required")

func openWorkbook(fileBytes []byte, password string) (*excelize.File, error) {
	f, err := excelize.OpenReader(bytes.NewReader(fileBytes), excelize.Options{Password: password})
	if err != nil {
		return nil, err
	}
	return f, nil
}

var zipHeader = []byte{0x50, 0x4B, 0x03, 0x04}

// IsExcelPasswordProtected checks if an XLSX file is likely encrypted
func IsExcelPasswordProtectedBytes(data []byte) bool {
	if len(data) < 4 {
		return false
	}

	for i := 0; i < 4; i++ {
		if data[i] != zipHeader[i] {
			return true
		}
	}
	
	return false
}

func ValidateWorkbookPassword(fileBytes []byte, password string) error {
	f, err := openWorkbook(fileBytes, password)
	if err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}

// trimRecords iterates through a 2D string slice and trims whitespace from each element.
func trimRecords(records [][]string) [][]string {
	logger.Debugf("CustomParser.trimRecords: Trimming %d records", len(records))

	trimmed := make([][]string, len(records))
	for i, row := range records {
		trimmedRow := make([]string, len(row))
		for j, field := range row {
			original := field
			trimmedRow[j] = strings.TrimSpace(field)
			if original != trimmedRow[j] {
				logger.Debugf("CustomParser.trimRecords: Trimmed field [%d][%d]: '%s' -> '%s'", i, j, original, trimmedRow[j])
			}
		}
		trimmed[i] = trimmedRow
	}

	logger.Debugf("CustomParser.trimRecords: Completed trimming records")
	return trimmed
}

func (p *CustomParser) Preview(fileBytes []byte, fileName string, skipRows int, rowSize int, password string) (*models.StatementPreview, error) {
	logger.Debugf("CustomParser.Preview: Starting preview for file '%s', skipRows=%d, rowSize=%d", fileName, skipRows, rowSize)
	logger.Debugf("CustomParser.Preview: File size: %d bytes", len(fileBytes))

	lowerFileName := strings.ToLower(fileName)
	extension := filepath.Ext(lowerFileName)
	logger.Debugf("CustomParser.Preview: Detected file extension: %s", extension)

	var records [][]string
	var err error

	switch extension {
	case ".csv":
		logger.Debugf("CustomParser.Preview: Processing as CSV file")
		r := csv.NewReader(bytes.NewReader(fileBytes))
		r.FieldsPerRecord = -1 // To handle "wrong number of fields" error
		records, err = r.ReadAll()
		if err != nil {
			logger.Debugf("CustomParser.Preview: Failed to read CSV: %v", err)
			return nil, fmt.Errorf("failed to read csv: %w", err)
		}
		logger.Debugf("CustomParser.Preview: Successfully read %d CSV records", len(records))
	case ".xls":
		logger.Debugf("CustomParser.Preview: Processing as XLS file (tab-separated)")
		r := csv.NewReader(bytes.NewReader(fileBytes))
		r.Comma = '\t'
		r.FieldsPerRecord = -1 // To handle "wrong number of fields" error
		records, err = r.ReadAll()
		if err != nil {
			logger.Debugf("CustomParser.Preview: Failed to read TSV from XLS file: %v", err)
			return nil, fmt.Errorf("failed to read tsv from xls file: %w", err)
		}
		logger.Debugf("CustomParser.Preview: Successfully read %d TSV records from XLS", len(records))
	case ".xlsx":
		logger.Debugf("CustomParser.Preview: Processing as XLSX file")
		f, err := openWorkbook(fileBytes, password)
		if err != nil {
			logger.Debugf("CustomParser.Preview: Failed to open XLSX file: %v", err)
			return nil, fmt.Errorf("failed to open xlsx file: %w", err)
		}
		defer f.Close()
		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			return nil, errors.New("no sheets found in XLSX file")
		}
		records, err = f.GetRows(sheets[0])
		if err != nil {
			logger.Debugf("CustomParser.Preview: Failed to read rows from sheet: %v", err)
			return nil, fmt.Errorf("failed to read rows from sheet: %w", err)
		}
		logger.Debugf("CustomParser.Preview: Successfully read %d XLSX records", len(records))
	default:
		logger.Debugf("CustomParser.Preview: Unsupported file extension: %s", extension)
		return nil, fmt.Errorf("unsupported file type for preview: %s. Only .csv and .xls are supported", fileName)
	}

	// Trim whitespace from all records at once to ensure clean data.
	logger.Debugf("CustomParser.Preview: Trimming whitespace from %d records", len(records))
	records = trimRecords(records)

	if len(records) <= skipRows {
		logger.Debugf("CustomParser.Preview: Not enough records (%d) to skip %d rows, returning empty preview", len(records), skipRows)
		return &models.StatementPreview{Headers: []string{}, Rows: [][]string{}}, nil
	}

	// Skip rows to drop the metadata or header rows
	logger.Debugf("CustomParser.Preview: Skipping first %d rows, remaining records: %d", skipRows, len(records)-skipRows)
	records = records[skipRows:]

	if len(records) == 0 {
		logger.Debugf("CustomParser.Preview: No records remaining after skipping rows")
		return &models.StatementPreview{Headers: []string{}, Rows: [][]string{}}, nil
	}

	headers := records[0]
	dataRows := records[1:]
	logger.Debugf("CustomParser.Preview: Found %d headers: %v", len(headers), headers)
	logger.Debugf("CustomParser.Preview: Found %d data rows", len(dataRows))

	if rowSize != -1 && len(dataRows) > rowSize {
		logger.Debugf("CustomParser.Preview: Limiting data rows from %d to %d", len(dataRows), rowSize)
		dataRows = dataRows[:rowSize]
	}

	preview := &models.StatementPreview{
		Headers: headers,
		Rows:    dataRows,
	}

	logger.Debugf("CustomParser.Preview: Preview generated successfully with %d headers and %d rows", len(preview.Headers), len(preview.Rows))
	return preview, nil
}

// Parse processes a file using metadata to map columns and create transactions.
func (p *CustomParser) Parse(fileBytes []byte, metadata string, fileName string, password string) ([]models.CreateTransactionInput, error) {
	logger.Debugf("CustomParser.Parse: Starting parse for file '%s'", fileName)
	logger.Debugf("CustomParser.Parse: File size: %d bytes", len(fileBytes))
	logger.Debugf("CustomParser.Parse: Metadata: %s", metadata)

	if metadata == "" {
		logger.Debugf("CustomParser.Parse: No metadata provided")
		return nil, errors.New("metadata is required for custom parser")
	}

	var meta StatementMetadata
	if err := json.Unmarshal([]byte(metadata), &meta); err != nil {
		logger.Debugf("CustomParser.Parse: Failed to unmarshal metadata: %v", err)
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}
	logger.Debugf("CustomParser.Parse: Parsed metadata - SkipRows: %d, ColumnMapping: %v", meta.SkipRows, meta.ColumnMapping)

	preview, err := p.Preview(fileBytes, fileName, meta.SkipRows, -1, password)
	if err != nil {
		logger.Debugf("CustomParser.Parse: Failed to preview file: %v", err)
		return nil, fmt.Errorf("failed to preview file for parsing: %w", err)
	}
	logger.Debugf("CustomParser.Parse: Preview generated with %d headers and %d rows", len(preview.Headers), len(preview.Rows))

	// Create a map of header names to their column index. Data is already trimmed by Preview.
	headerIndex := make(map[string]int, len(preview.Headers))
	for i, h := range preview.Headers {
		headerIndex[h] = i
	}
	logger.Debugf("CustomParser.Parse: Created header index map: %v", headerIndex)

	columnIndex := make(map[string]int)
	for field, columnName := range meta.ColumnMapping {
		if idx, ok := headerIndex[columnName]; ok {
			columnIndex[field] = idx
			logger.Debugf("CustomParser.Parse: Mapped field '%s' to column '%s' at index %d", field, columnName, idx)
		} else {
			if field != "description" {
				logger.Debugf("CustomParser.Parse: Required column '%s' not found in headers", columnName)
				return nil, fmt.Errorf("mapped column '%s' not found in statement header", columnName)
			}
			logger.Debugf("CustomParser.Parse: Optional field '%s' column '%s' not found, skipping", field, columnName)
		}
	}
	logger.Debugf("CustomParser.Parse: Final column index mapping: %v", columnIndex)

	err = p.validateMappings(columnIndex)
	if err != nil {
		logger.Debugf("CustomParser.Parse: Validation failed: %v", err)
		return nil, err
	}
	logger.Debugf("CustomParser.Parse: Column mapping validation passed")

	var transactions []models.CreateTransactionInput
	logger.Debugf("CustomParser.Parse: Starting to parse %d data rows", len(preview.Rows))

	for i, row := range preview.Rows {
		logger.Debugf("CustomParser.Parse: Processing row %d: %v", i+1, row)
		transaction, err := p.parseRow(row, columnIndex)
		if err != nil {
			logger.Debugf("CustomParser.Parse: Failed to parse row %d: %v", i+meta.SkipRows+2, err)
			logger.Warnf("skipping row %d due to parsing error: %v", i+meta.SkipRows+2, err)
			continue
		}
		logger.Debugf("CustomParser.Parse: Successfully parsed row %d into transaction: Name='%s', Amount=%v, Date=%v",
			i+1, transaction.Name, transaction.Amount, transaction.Date)
		transactions = append(transactions, *transaction)
	}

	logger.Debugf("CustomParser.Parse: Parse completed successfully with %d transactions", len(transactions))
	return transactions, nil
}

func (p *CustomParser) validateMappings(columnIndex map[string]int) error {
	logger.Debugf("CustomParser.validateMappings: Validating column mappings: %v", columnIndex)

	// Validate that all required fields are present in the column mapping.
	requiredFields := []string{"txn_date", "name"}
	for _, f := range requiredFields {
		if _, ok := columnIndex[f]; !ok {
			logger.Debugf("CustomParser.validateMappings: Required field '%s' is missing", f)
			return fmt.Errorf("required field '%s' is not mapped in metadata", f)
		}
		logger.Debugf("CustomParser.validateMappings: Required field '%s' is mapped", f)
	}

	_, amountOk := columnIndex["amount"]
	_, creditOk := columnIndex["credit"]
	_, debitOk := columnIndex["debit"]

	logger.Debugf("CustomParser.validateMappings: Amount field mappings - amount: %v, credit: %v, debit: %v", amountOk, creditOk, debitOk)

	if !amountOk && !(creditOk && debitOk) {
		logger.Debugf("CustomParser.validateMappings: Insufficient amount information")
		return errors.New("insufficient amount information in metadata: map either 'amount' or both 'credit' and 'debit'")
	}

	logger.Debugf("CustomParser.validateMappings: All validations passed")
	return nil
}

// parseRow parses a single row into a transaction based on the column index map.
func (p *CustomParser) parseRow(row []string, columnIndex map[string]int) (*models.CreateTransactionInput, error) {
	logger.Debugf("CustomParser.parseRow: Parsing row with %d columns: %v", len(row), row)

	dateStr, err := p.getRequiredField(row, columnIndex, "txn_date")
	if err != nil {
		logger.Debugf("CustomParser.parseRow: Failed to get txn_date: %v", err)
		return nil, err
	}
	logger.Debugf("CustomParser.parseRow: Extracted date string: '%s'", dateStr)

	name, err := p.getRequiredField(row, columnIndex, "name")
	if err != nil {
		logger.Debugf("CustomParser.parseRow: Failed to get name: %v", err)
		return nil, err
	}
	logger.Debugf("CustomParser.parseRow: Extracted name: '%s'", name)

	description := p.getOptionalField(row, columnIndex, "description")
	logger.Debugf("CustomParser.parseRow: Extracted description: '%s'", description)

	date, err := utils.ParseDate(dateStr)
	if err != nil {
		logger.Debugf("CustomParser.parseRow: Failed to parse date '%s': %v", dateStr, err)
		return nil, err
	}
	logger.Debugf("CustomParser.parseRow: Parsed date: %v", date)

	amount, err := p.getAmount(row, columnIndex)
	if err != nil {
		logger.Debugf("CustomParser.parseRow: Failed to get amount: %v", err)
		return nil, err
	}
	logger.Debugf("CustomParser.parseRow: Extracted amount: %v", amount)

	tx := &models.CreateTransactionInput{
		CreateBaseTransactionInput: models.CreateBaseTransactionInput{
			Name:        name,
			Description: description,
			Amount:      amount,
			Date:        date,
		},
		CategoryIds: []int64{},
	}

	logger.Debugf("CustomParser.parseRow: Created transaction successfully")
	return tx, nil
}

func (p *CustomParser) getAmount(row []string, columnIndex map[string]int) (*float64, error) {
	logger.Debugf("CustomParser.getAmount: Extracting amount from row")

	var amount float64
	amountStr, amountOk := p.getOptionalFieldOk(row, columnIndex, "amount")
	creditStr, creditOk := p.getOptionalFieldOk(row, columnIndex, "credit")
	debitStr, debitOk := p.getOptionalFieldOk(row, columnIndex, "debit")

	logger.Debugf("CustomParser.getAmount: Field values - amount: '%s' (ok=%v), credit: '%s' (ok=%v), debit: '%s' (ok=%v)",
		amountStr, amountOk, creditStr, creditOk, debitStr, debitOk)

	if amountOk && amountStr != "" {
		logger.Debugf("CustomParser.getAmount: Using amount field: '%s'", amountStr)
		val, err := utils.ParseFloat(amountStr)
		if err != nil {
			logger.Debugf("CustomParser.getAmount: Failed to parse amount '%s': %v", amountStr, err)
			return nil, fmt.Errorf("failed to parse amount: %w", err)
		}
		amount = val
		logger.Debugf("CustomParser.getAmount: Parsed amount: %f", amount)
	} else if creditOk && debitOk {
		logger.Debugf("CustomParser.getAmount: Using credit/debit fields - credit: '%s', debit: '%s'", creditStr, debitStr)

		credit, err := utils.ParseFloat(creditStr)
		if err != nil && creditStr != "" {
			logger.Debugf("CustomParser.getAmount: Failed to parse credit '%s': %v", creditStr, err)
			return nil, fmt.Errorf("failed to parse credit: %w", err)
		}
		debit, err := utils.ParseFloat(debitStr)
		if err != nil && debitStr != "" {
			logger.Debugf("CustomParser.getAmount: Failed to parse debit '%s': %v", debitStr, err)
			return nil, fmt.Errorf("failed to parse debit: %w", err)
		}

		logger.Debugf("CustomParser.getAmount: Parsed values - credit: %f, debit: %f", credit, debit)

		if credit > 0 {
			amount = credit
			logger.Debugf("CustomParser.getAmount: Using credit amount: %f", amount)
		} else {
			amount = -debit
			logger.Debugf("CustomParser.getAmount: Using negative debit amount: %f", amount)
		}
	} else {
		logger.Debugf("CustomParser.getAmount: Insufficient amount information available")
		return nil, errors.New("insufficient amount information: map either 'amount' or both 'credit' and 'debit'")
	}

	logger.Debugf("CustomParser.getAmount: Final amount: %f", amount)
	return &amount, nil
}

// getRequiredField safely extracts a required field from a row.
func (p *CustomParser) getRequiredField(row []string, columnIndex map[string]int, fieldName string) (string, error) {
	logger.Debugf("CustomParser.getRequiredField: Getting required field '%s'", fieldName)

	idx, ok := columnIndex[fieldName]
	if !ok {
		logger.Debugf("CustomParser.getRequiredField: Field '%s' is not mapped in columnIndex", fieldName)
		return "", fmt.Errorf("field '%s' is not mapped", fieldName)
	}

	if idx >= len(row) {
		logger.Debugf("CustomParser.getRequiredField: Column index %d for '%s' is out of bounds (row length: %d)", idx, fieldName, len(row))
		return "", fmt.Errorf("column index %d for '%s' is out of bounds", idx, fieldName)
	}

	value := row[idx]
	logger.Debugf("CustomParser.getRequiredField: Retrieved value '%s' for field '%s' at index %d", value, fieldName, idx)
	return value, nil
}

// getOptionalField safely extracts an optional field from a row.
func (p *CustomParser) getOptionalField(row []string, columnIndex map[string]int, fieldName string) string {
	val, ok := p.getOptionalFieldOk(row, columnIndex, fieldName)
	logger.Debugf("CustomParser.getOptionalField: Retrieved optional field '%s': '%s' (found=%v)", fieldName, val, ok)
	return val
}

// getOptionalFieldOk safely extracts an optional field and indicates if the mapping exists.
func (p *CustomParser) getOptionalFieldOk(row []string, columnIndex map[string]int, fieldName string) (string, bool) {
	logger.Debugf("CustomParser.getOptionalFieldOk: Getting optional field '%s'", fieldName)

	idx, ok := columnIndex[fieldName]
	if !ok {
		logger.Debugf("CustomParser.getOptionalFieldOk: Field '%s' is not mapped", fieldName)
		return "", false
	}

	if idx >= len(row) {
		logger.Debugf("CustomParser.getOptionalFieldOk: Column index %d for '%s' is out of bounds (row length: %d)", idx, fieldName, len(row))
		return "", false
	}

	value := row[idx]
	logger.Debugf("CustomParser.getOptionalFieldOk: Retrieved value '%s' for field '%s' at index %d", value, fieldName, idx)
	return value, true
}

func init() {
	logger.Debugf("CustomParser.init: Registering CustomParser for BankTypeOthers")
	RegisterParser(models.BankTypeOthers, &CustomParser{})
}
