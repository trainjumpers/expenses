package models

import "encoding/json"

type JobStatus struct {
	ID       int64           `json:"id"`
	Name     string          `json:"name"`
	Metadata json.RawMessage `json:"metadata"`
	Status   string          `json:"status"`
}

type JobStatusMapping struct {
	ID     int64 `json:"id"`
	JobID  int64 `json:"job_id"`
	UserID int64 `json:"user_id"`
}

const (
	JobStatusCreated   = "created"
	JobStatusRunning   = "running"
	JobStatusComplete  = "complete"
	JobStatusFailed    = "failed"
	JobStatusCancelled = "cancelled"
)
