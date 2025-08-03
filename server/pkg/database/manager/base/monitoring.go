package base

import (
	"time"
)

// DatabaseStats provides database connection pool statistics
type DatabaseStats struct {
	TotalConnections        int32
	IdleConnections         int32
	AcquiredConnections     int32
	ConstructingConnections int32
	MaxConnections          int32
	AcquireCount            int64
	AcquireDuration         time.Duration
	EmptyAcquireCount       int64
	CanceledAcquireCount    int64
}

// TransactionInfo provides information about the current transaction
type TransactionInfo struct {
	ID             string
	StartTime      time.Time
	Duration       time.Duration
	IsReadOnly     bool
	IsNested       bool
	LockKey        *int64
	IsolationLevel string
}

// TransactionMetrics tracks transaction performance
type TransactionMetrics struct {
	// Counters
	TotalTransactions      int64
	CommittedTransactions  int64
	RolledBackTransactions int64
	FailedTransactions     int64

	// Timing
	TotalDuration   time.Duration
	AverageDuration time.Duration
	MaxDuration     time.Duration
	MinDuration     time.Duration

	// Current state
	ActiveTransactions int64

	// Error tracking
	ErrorsByType map[string]int64
}

// BatchOperation represents a single operation in a batch
type BatchOperation struct {
	Query string
	Args  []any
}

// BatchResult contains results from batch operations
type BatchResult struct {
	RowsAffected []int64
	Errors       []error
}
