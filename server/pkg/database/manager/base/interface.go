package base

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// TransactionFunc is a function type for transaction operations
// The function receives a context that contains the transaction
type TransactionFunc func(ctx context.Context) error

// LockFunc is a function type for lock operations with transaction
// The function receives a context that contains the transaction
type LockFunc func(ctx context.Context) error

// DatabaseManager defines the unified interface for all database operations
// This interface combines basic, enhanced, and monitoring features
type DatabaseManager interface {
	// === CORE OPERATIONS (always available) ===
	
	// ExecuteQuery executes a query that doesn't return rows (INSERT, UPDATE, DELETE)
	// Returns the number of rows affected and any error
	// If called within a transaction context, uses the transaction; otherwise uses the pool
	ExecuteQuery(ctx context.Context, query string, args ...any) (rowsAffected int64, err error)

	// FetchOne executes a query and returns a single row
	// Returns error if no rows found or multiple rows returned
	// If called within a transaction context, uses the transaction; otherwise uses the pool
	FetchOne(ctx context.Context, query string, args ...any) pgx.Row

	// FetchAll executes a query and returns multiple rows
	// If called within a transaction context, uses the transaction; otherwise uses the pool
	FetchAll(ctx context.Context, query string, args ...any) (pgx.Rows, error)

	// WithTxn executes a function within a transaction
	// Automatically commits on success, rolls back on error
	// The function receives a context that contains the transaction
	WithTxn(ctx context.Context, fn TransactionFunc) error

	// WithLock executes a function with a table lock within a transaction
	// Automatically starts transaction, commits on success, rolls back on error
	// The function receives a context that contains the transaction
	WithLock(ctx context.Context, lockKey int64, fn LockFunc) error

	// Close closes the database connection
	Close() error

	// === ENHANCED OPERATIONS (configurable) ===
	
	// WithTxnOptions executes a transaction with specific options
	// Available if EnableRetry is true in config
	WithTxnOptions(ctx context.Context, opts *TransactionOptions, fn TransactionFunc) error
	
	// WithReadOnlyTxn executes a read-only transaction
	// Available if EnableRetry is true in config
	WithReadOnlyTxn(ctx context.Context, fn TransactionFunc) error
	
	// WithRetryableTxn executes a transaction with aggressive retry policy
	// Available if EnableRetry is true in config
	WithRetryableTxn(ctx context.Context, fn TransactionFunc) error

	// WithSavepoint executes a function within a savepoint (nested transaction)
	// Available if EnableSavepoints is true in config
	WithSavepoint(ctx context.Context, name string, fn TransactionFunc) error

	// ExecuteBatch executes multiple operations in a batch
	// Available if EnableBatch is true in config
	ExecuteBatch(ctx context.Context, batch *pgx.Batch) error

	// WithConnection executes a function with a dedicated connection
	// Always available
	WithConnection(ctx context.Context, fn func(conn *pgx.Conn) error) error

	// === HEALTH & INTROSPECTION (always available) ===
	
	// Ping checks database connectivity
	Ping(ctx context.Context) error
	
	// Stats returns database connection pool statistics
	Stats() DatabaseStats

	// GetTransactionInfo returns information about the current transaction
	GetTransactionInfo(ctx context.Context) (*TransactionInfo, error)

	// === MONITORING (configurable) ===
	
	// GetMonitoringMetrics returns current monitoring metrics
	// Available if EnableMonitoring is true in config
	GetMonitoringMetrics() TransactionMetrics
	
	// ResetMetrics clears all monitoring metrics
	// Available if EnableMonitoring is true in config
	ResetMetrics()

	// === CONFIGURATION ===
	
	// GetConfig returns the current configuration
	GetConfig() *DatabaseManagerConfig
	
	// IsFeatureEnabled checks if a specific feature is enabled
	IsFeatureEnabled(feature string) bool
}

// Feature constants for IsFeatureEnabled
const (
	FeatureRetry      = "retry"
	FeatureSavepoints = "savepoints"
	FeatureBatch      = "batch"
	FeatureMonitoring = "monitoring"
	FeatureMetrics    = "metrics"
)
