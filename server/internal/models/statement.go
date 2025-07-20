package models

type StatementStatus string

const (
	StatementStatusPending    StatementStatus = "pending"
	StatementStatusProcessing StatementStatus = "processing"
	StatementStatusDone       StatementStatus = "done"
	StatementStatusError      StatementStatus = "error"
)

type CreateStatementInput struct {
	AccountId        int64           `json:"account_id" binding:"required"`
	CreatedBy        int64           `json:"created_by" binding:"required"`
	OriginalFilename string          `json:"original_filename" binding:"required"`
	FileType         string          `json:"file_type" binding:"required"`
	Status           StatementStatus `json:"status" binding:"required,oneof=pending processing done error"`
	Message          *string         `json:"message,omitempty"`
}

type UpdateStatementStatusInput struct {
	Status  StatementStatus `json:"status" binding:"required,oneof=pending processing done error"`
	Message *string         `json:"message,omitempty"`
}

type StatementResponse struct {
	Id               int64           `json:"id"`
	AccountId        int64           `json:"account_id"`
	CreatedBy        int64           `json:"created_by"`
	OriginalFilename string          `json:"original_filename"`
	FileType         string          `json:"file_type"`
	Status           StatementStatus `json:"status"`
	Message          *string         `json:"message,omitempty"`
}

type PaginatedStatementResponse struct {
	Statements []StatementResponse `json:"statements"`
	Total      int                 `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
}

// Custom CSV Import Models
type CSVPreviewInput struct {
	AccountId int64
	FileBytes []byte
	Filename  string
	SkipRows  int
}

type CSVPreviewResult struct {
	Columns []string   `json:"columns"`
	Rows    [][]string `json:"rows"`
	Total   int        `json:"total"`
}

type ColumnMapping struct {
	SourceColumn string `json:"source_column" binding:"required"`
	TargetField  string `json:"target_field" binding:"required,oneof=name amount description date credit debit"`
}

type CustomImportInput struct {
	AccountId int64           `json:"account_id" binding:"required"`
	SkipRows  int             `json:"skip_rows" binding:"min=0"`
	Mappings  []ColumnMapping `json:"mappings" binding:"required,min=1"`
}

type CustomImportResult struct {
	Statement          StatementResponse `json:"statement"`
	TransactionsCreated int              `json:"transactions_created"`
	Message            string           `json:"message"`
}

// Note: ParseOptions and related types are now defined in parser.go
// StatementImportMetadata is kept here for backward compatibility if needed
type StatementImportMetadata struct {
	SkipRows int             `json:"skip_rows"`
	Mappings []ColumnMapping `json:"mappings"`
}

// ToParseOptions converts StatementImportMetadata to ParseOptions
func (m StatementImportMetadata) ToParseOptions() ParseOptions {
	return ParseOptions{
		SkipRows: m.SkipRows,
		Mappings: m.Mappings,
	}
}
