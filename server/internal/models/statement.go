package models

import (
	"mime/multipart"
	"time"
)

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

type ParseStatementInput struct {
	FileBytes        []byte `json:"file_bytes" binding:"required"`
	FileName         string `json:"file_name" binding:"required"`
	AccountId        int64  `json:"account_id" binding:"required"`
	OriginalFilename string `json:"original_filename" binding:"required"`
	BankType         string `json:"bank_type,omitempty" binding:"optional"`
	Metadata         string `json:"metadata,omitempty" binding:"optional"`
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
	CreatedAt        time.Time       `json:"created_at"`
}

type PaginatedStatementResponse struct {
	Statements []StatementResponse `json:"statements"`
	Total      int                 `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
}

type StatementPreview struct {
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
}

// Form parsing
type ParseStatementForm struct {
	AccountId int64                 `form:"account_id" binding:"required"`
	BankType  string                `form:"bank_type"`
	Metadata  string                `form:"metadata"`
	File      *multipart.FileHeader `form:"file" binding:"required"`
}

type PreviewStatementForm struct {
	SkipRows int                   `form:"skip_rows"`
	RowSize  int                   `form:"row_size"`
	File     *multipart.FileHeader `form:"file" binding:"required"`
}
