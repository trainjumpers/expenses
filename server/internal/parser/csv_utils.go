package parser

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"expenses/internal/errors"
	"expenses/pkg/logger"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// CSVParseResult represents the result of parsing a CSV/Excel file
type CSVParseResult struct {
	Columns []string
	Rows    [][]string
	Total   int
}

// ParseCSVFile parses CSV or Excel files and returns structured data
func ParseCSVFile(fileBytes []byte, filename string) (*CSVParseResult, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	
	switch ext {
	case ".csv":
		return parseCSV(fileBytes)
	case ".xls", ".xlsx":
		return parseExcelAsText(fileBytes)
	default:
		return nil, errors.NewInvalidCSVFormatError(fmt.Errorf("unsupported file extension: %s", ext))
	}
}

// parseCSV handles CSV file parsing
func parseCSV(fileBytes []byte) (*CSVParseResult, error) {
	reader := csv.NewReader(strings.NewReader(string(fileBytes)))
	reader.FieldsPerRecord = -1 // Allow variable number of fields
	reader.TrimLeadingSpace = true

	var allRows [][]string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.NewInvalidCSVFormatError(err)
		}
		
		// Skip completely empty rows
		if len(record) == 0 || (len(record) == 1 && strings.TrimSpace(record[0]) == "") {
			continue
		}
		
		allRows = append(allRows, record)
	}

	if len(allRows) == 0 {
		return nil, errors.NewInsufficientColumnsError()
	}

	// Don't assume headers yet - return all rows and let ApplySkipRows handle header determination
	logger.Infof("parseCSV: Final result - %d total rows (headers will be determined after row skipping)", len(allRows))
	
	// For now, treat first row as temporary headers, but this will be corrected by ApplySkipRows
	if len(allRows) == 0 {
		return &CSVParseResult{
			Columns: []string{},
			Rows:    [][]string{},
			Total:   0,
		}, nil
	}

	// Temporary assignment - the real headers will be determined after skipping
	tempColumns := allRows[0]
	tempRows := allRows[1:]
	
	// Clean up temporary column names
	for i, col := range tempColumns {
		tempColumns[i] = strings.TrimSpace(col)
	}

	logger.Debugf("parseCSV: Temporary columns (before skipping): %v", tempColumns)

	return &CSVParseResult{
		Columns: tempColumns,
		Rows:    tempRows,
		Total:   len(allRows),
	}, nil
}

// parseExcelAsText handles Excel files as tab-delimited text (similar to SBI parser approach)
// Note: This is a basic implementation that may not work well with binary Excel files
func parseExcelAsText(fileBytes []byte) (*CSVParseResult, error) {
	logger.Debugf("parseExcelAsText: Starting to parse Excel file with %d bytes", len(fileBytes))
	logger.Warnf("parseExcelAsText: Excel file parsing is limited. For best results, please save your Excel file as CSV format before uploading.")
	
	// Check if this looks like a binary Excel file
	if len(fileBytes) > 8 {
		// Check for Excel file signatures
		header := string(fileBytes[:8])
		if strings.Contains(header, "PK") || strings.Contains(header, "\xd0\xcf\x11\xe0") {
			logger.Errorf("parseExcelAsText: This appears to be a binary Excel file which cannot be parsed as text")
			return nil, errors.NewInvalidCSVFormatError(fmt.Errorf("binary Excel files are not supported. Please save your file as CSV format and try again"))
		}
	}
	
	scanner := bufio.NewScanner(bytes.NewReader(fileBytes))
	
	var allRows [][]string
	lineNum := 0
	validLines := 0
	
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		
		// Log first few lines for debugging
		if lineNum <= 10 {
			logger.Debugf("parseExcelAsText: Line %d: %q", lineNum, line)
		}
		
		// Skip empty lines
		if line == "" {
			continue
		}
		
		// Skip lines that look like binary data
		if strings.ContainsAny(line, "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x0b\x0c\x0e\x0f") {
			logger.Debugf("parseExcelAsText: Skipping line %d (contains binary data)", lineNum)
			continue
		}

		// Try different delimiters to split the line
		var fields []string
		var delimiter string
		
		// Try tab first (most common for Excel)
		if strings.Contains(line, "\t") {
			fields = strings.Split(line, "\t")
			delimiter = "tab"
		} else if strings.Contains(line, ",") {
			// Try comma
			fields = strings.Split(line, ",")
			delimiter = "comma"
		} else if strings.Contains(line, "|") {
			// Try pipe
			fields = strings.Split(line, "|")
			delimiter = "pipe"
		} else if strings.Contains(line, ";") {
			// Try semicolon
			fields = strings.Split(line, ";")
			delimiter = "semicolon"
		} else {
			// Try splitting by multiple spaces (common in fixed-width formats)
			fields = strings.Fields(line)
			if len(fields) > 1 {
				delimiter = "spaces"
			} else {
				// Single field or unrecognized format
				fields = []string{line}
				delimiter = "none"
			}
		}
		
		logger.Debugf("parseExcelAsText: Line %d split by %s into %d fields", lineNum, delimiter, len(fields))
		if lineNum <= 5 {
			logger.Debugf("parseExcelAsText: Line %d fields: %v", lineNum, fields)
		}

		// Clean up fields
		for i, field := range fields {
			fields[i] = strings.TrimSpace(field)
		}

		// Skip rows with no meaningful content
		hasContent := false
		for _, field := range fields {
			if field != "" && len(field) > 0 {
				hasContent = true
				break
			}
		}

		if hasContent {
			allRows = append(allRows, fields)
			validLines++
			// Log first few meaningful rows
			if len(allRows) <= 10 {
				logger.Debugf("parseExcelAsText: Valid row %d fields: %v", len(allRows), fields)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Errorf("parseExcelAsText: Scanner error: %v", err)
		return nil, errors.NewInvalidCSVFormatError(err)
	}

	logger.Infof("parseExcelAsText: Processed %d total lines, found %d valid lines, created %d meaningful rows", lineNum, validLines, len(allRows))

	if len(allRows) == 0 {
		logger.Errorf("parseExcelAsText: No meaningful rows found - this may be a binary Excel file")
		return nil, errors.NewInvalidCSVFormatError(fmt.Errorf("no readable data found. If this is an Excel file, please save it as CSV format and try again"))
	}

	// Log column count variations for debugging (but don't fail on inconsistency)
	if len(allRows) > 1 {
		columnCounts := make(map[int]int)
		for i, row := range allRows {
			colCount := len(row)
			columnCounts[colCount]++
			if i < 10 { // Log first 10 rows for debugging
				logger.Debugf("parseExcelAsText: Row %d has %d columns", i+1, colCount)
			}
		}
		logger.Infof("parseExcelAsText: Column count distribution: %v", columnCounts)
		
		// Only warn if there's extreme inconsistency (more than 5 different column counts)
		if len(columnCounts) > 5 {
			logger.Warnf("parseExcelAsText: High column count variation detected (%d different counts). This may indicate parsing issues.", len(columnCounts))
		}
	}

	// Don't assume headers yet - return all rows and let ApplySkipRows handle header determination
	logger.Infof("parseExcelAsText: Final result - %d total rows (headers will be determined after row skipping)", len(allRows))
	
	// For now, treat first row as temporary headers, but this will be corrected by ApplySkipRows
	if len(allRows) == 0 {
		return &CSVParseResult{
			Columns: []string{},
			Rows:    [][]string{},
			Total:   0,
		}, nil
	}

	// Temporary assignment - the real headers will be determined after skipping
	tempColumns := allRows[0]
	tempRows := allRows[1:]
	
	// Clean up temporary column names
	for i, col := range tempColumns {
		tempColumns[i] = strings.TrimSpace(col)
	}

	logger.Debugf("parseExcelAsText: Temporary columns (before skipping): %v", tempColumns)
	logger.Debugf("parseExcelAsText: First row has %d columns: %v", len(tempColumns), tempColumns)
	
	// Log first few data rows for comparison
	for i, row := range tempRows {
		if i < 5 {
			logger.Debugf("parseExcelAsText: Data row %d has %d columns: %v", i+1, len(row), row)
		}
	}

	return &CSVParseResult{
		Columns: tempColumns,
		Rows:    tempRows,
		Total:   len(allRows),
	}, nil
}

// GetPreviewRows returns the first N rows for preview
func (r *CSVParseResult) GetPreviewRows(count int) [][]string {
	if count >= len(r.Rows) {
		return r.Rows
	}
	return r.Rows[:count]
}

// ApplySkipRows returns a new result with the specified number of rows skipped from the beginning
// This will skip rows including headers and treat the next available row as new headers
func (r *CSVParseResult) ApplySkipRows(skipRows int) *CSVParseResult {
	if skipRows <= 0 {
		return r
	}

	// Reconstruct all rows (headers + data rows)
	allRows := make([][]string, 0, len(r.Rows)+1)
	allRows = append(allRows, r.Columns)
	allRows = append(allRows, r.Rows...)

	// Skip the specified number of rows from the beginning
	if skipRows >= len(allRows) {
		return &CSVParseResult{
			Columns: []string{},
			Rows:    [][]string{},
			Total:   0,
		}
	}

	remainingRows := allRows[skipRows:]
	if len(remainingRows) == 0 {
		return &CSVParseResult{
			Columns: []string{},
			Rows:    [][]string{},
			Total:   0,
		}
	}

	// First remaining row becomes the new headers
	newColumns := remainingRows[0]
	newRows := remainingRows[1:]

	// Clean up new column names
	for i, col := range newColumns {
		newColumns[i] = strings.TrimSpace(col)
	}

	return &CSVParseResult{
		Columns: newColumns,
		Rows:    newRows,
		Total:   len(remainingRows),
	}
}