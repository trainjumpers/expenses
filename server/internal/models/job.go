package models

import (
	"encoding/json"
	"time"
)

type JobType string
type JobStatus string

const (
	JobTypeRuleExecution       JobType = "rule_execution"
	JobTypeStatementProcessing JobType = "statement_processing"
)

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

type CreateJobInput struct {
	JobType     JobType         `json:"job_type" binding:"required"`
	ReferenceId *int64          `json:"reference_id,omitempty"`
	CreatedBy   int64           `json:"created_by" binding:"required"`
	Status      JobStatus       `json:"status" binding:"required,oneof=pending processing completed failed"`
	Message     *string         `json:"message,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
}

type UpdateJobStatusInput struct {
	Status      JobStatus  `json:"status" binding:"required,oneof=pending processing completed failed"`
	Message     *string    `json:"message,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type JobResponse struct {
	Id          int64           `json:"id"`
	JobType     JobType         `json:"job_type"`
	ReferenceId *int64          `json:"reference_id,omitempty"`
	CreatedBy   int64           `json:"created_by"`
	Status      JobStatus       `json:"status"`
	Message     *string         `json:"message,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
	StartedAt   *time.Time      `json:"started_at,omitempty"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type JobListQuery struct {
	JobType   *JobType   `form:"job_type,omitempty"`
	Status    *JobStatus `form:"status,omitempty"`
	CreatedBy *int64     `form:"created_by,omitempty"`
	Page      int        `form:"page,default=1"`
	PageSize  int        `form:"page_size,default=20"`
}

type PaginatedJobResponse struct {
	Jobs     []JobResponse `json:"jobs"`
	Total    int           `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// Rule execution specific metadata
type RuleExecutionMetadata struct {
	RuleIds        *[]int64 `json:"rule_ids,omitempty"`
	TransactionIds *[]int64 `json:"transaction_ids,omitempty"`
	PageSize       int      `json:"page_size,omitempty"`
}
