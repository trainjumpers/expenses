package models

import "time"

// StatementPreview represents a preview of the statement file
type StatementPreview struct {
	Headers    []string   `json:"headers"`
	SampleData [][]string `json:"sample_data"`
	TotalRows  int        `json:"total_rows"`
	FileType   string     `json:"file_type"`
}

// ColumnMapping defines how columns in the file map to our transaction fields
type ColumnMapping struct {
	DateColumn        int    `json:"date_column"`
	DescriptionColumn int    `json:"description_column"`
	AmountColumn      int    `json:"amount_column"`
	ReferenceColumn   int    `json:"reference_column"`
	DateFormat        string `json:"date_format,omitempty"`
}

// ParseStatementInput contains the data needed to parse a statement with custom mapping
type ParseStatementInput struct {
	FileBytes     []byte        `json:"-"`
	FileType      string        `json:"file_type"`
	HasHeaders    bool          `json:"has_headers"`
	ColumnMapping ColumnMapping `json:"column_mapping"`
}

// ParsedTransaction represents a transaction parsed from the statement
type ParsedTransaction struct {
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Reference   string    `json:"reference,omitempty"`
}

// ParsedStatementResult contains the result of parsing a statement
type ParsedStatementResult struct {
	Transactions   []ParsedTransaction `json:"transactions"`
	TotalRows      int                 `json:"total_rows"`
	SuccessfulRows int                 `json:"successful_rows"`
	FailedRows     int                 `json:"failed_rows"`
	Errors         []string            `json:"errors,omitempty"`
}

// PreviewStatementRequest for API endpoint
type PreviewStatementRequest struct {
	FileType string `json:"file_type" binding:"required"`
}

// ParseStatementRequest for API endpoint
type ParseStatementRequest struct {
	StatementID   int64         `json:"statement_id" binding:"required"`
	HasHeaders    bool          `json:"has_headers"`
	ColumnMapping ColumnMapping `json:"column_mapping" binding:"required"`
}
