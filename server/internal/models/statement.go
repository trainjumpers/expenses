package models

import "time"

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
	CreatedAt        time.Time       `json:"created_at"`
}

type PaginatedStatementResponse struct {
	Statements []StatementResponse `json:"statements"`
	Total      int                 `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
}
