package database

import (
	"context"
	"expenses/internal/config"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

// TransactionFunc is a function type for transaction operations
type TransactionFunc func(tx pgx.Tx) error

// LockFunc is a function type for lock operations with transaction
type LockFunc func(tx pgx.Tx) error

// DatabaseManager defines the interface for database operations
type DatabaseManager interface {
	// ExecuteQuery executes a query that doesn't return rows (INSERT, UPDATE, DELETE)
	// Returns the number of rows affected and any error
	ExecuteQuery(ctx context.Context, query string, args ...interface{}) (rowsAffected int64, err error)

	// FetchOne executes a query and returns a single row
	// Returns error if no rows found or multiple rows returned
	FetchOne(ctx context.Context, query string, args ...interface{}) pgx.Row

	// FetchAll executes a query and returns multiple rows
	FetchAll(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)

	// WithTxn executes a function within a transaction
	// Automatically commits on success, rolls back on error
	WithTxn(ctx context.Context, fn TransactionFunc) error

	// WithLock executes a function with a table lock within a transaction
	// Automatically starts transaction, commits on success, rolls back on error
	WithLock(ctx context.Context, lockKey int64, fn LockFunc) error

	// Close closes the database connection
	Close() error
}

// NewDatabaseManager creates a new database manager based on DB_TYPE environment variable
func NewDatabaseManager(cfg *config.Config) (DatabaseManager, error) {
	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "postgres" // Default to postgres
	}

	switch dbType {
	case "postgres":
		return NewPostgresDatabaseManager(cfg)
	default:
		return nil, fmt.Errorf("unsupported database type: %s. Only 'postgres' is supported", dbType)
	}
}
