package database

import (
	"context"
	"expenses/internal/config"
	"expenses/pkg/logger"
	"fmt"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TransactionFunc is a function type for transaction operations
type TransactionFunc func(tx pgx.Tx) error

// LockFunc is a function type for lock operations with transaction
type LockFunc func(tx pgx.Tx) error

type DatabaseManager struct {
	pool *pgxpool.Pool
}

func NewDatabaseManager(cfg *config.Config) (*DatabaseManager, error) {
	host := os.Getenv("DB_HOST")
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		return nil, fmt.Errorf("invalid database port number: %w", err)
	}
	user := os.Getenv("DB_USER")
	dbname := os.Getenv("DB_NAME")
	pass := os.Getenv("DB_PASSWORD")
	sslmode := os.Getenv("DB_SSL_MODE")
	if sslmode == "" {
		sslmode = "verify-full"
	}

	psqlSetup := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s&search_path=%s",
		user, pass, host, port, dbname, sslmode, cfg.DBSchema)

	logger.Debugf("Connecting to database")
	pool, err := pgxpool.New(context.Background(), psqlSetup)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	logger.Debugf("Database connected successfully")

	return &DatabaseManager{
		pool: pool,
	}, nil
}

// ExecuteQuery executes a query that doesn't return rows (INSERT, UPDATE, DELETE)
// Returns the number of rows affected and any error
func (dm *DatabaseManager) ExecuteQuery(ctx context.Context, query string, args ...any) (rowsAffected int64, err error) {
	logger.Debugf("Executing query: %s", query)

	result, err := dm.pool.Exec(ctx, query, args...)
	if err != nil {
		logger.Errorf("Failed to execute query: %v", err)
		return 0, err
	}

	return result.RowsAffected(), nil
}

// FetchOne executes a query and returns a single row
// Returns error if no rows found or multiple rows returned
func (dm *DatabaseManager) FetchOne(ctx context.Context, query string, args ...any) pgx.Row {
	logger.Debugf("Fetching single row: %s", query)
	return dm.pool.QueryRow(ctx, query, args...)
}

// FetchAll executes a query and returns multiple rows
func (dm *DatabaseManager) FetchAll(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	logger.Debugf("Fetching multiple rows: %s", query)

	rows, err := dm.pool.Query(ctx, query, args...)
	if err != nil {
		logger.Errorf("Failed to fetch rows: %v", err)
		return nil, err
	}

	return rows, nil
}

// WithTxn executes a function within a transaction
// Automatically commits on success, rolls back on error
func (dm *DatabaseManager) WithTxn(ctx context.Context, fn TransactionFunc) error {
	logger.Debugf("Starting transaction")

	tx, err := dm.pool.Begin(ctx)
	if err != nil {
		logger.Errorf("Failed to begin transaction: %v", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			logger.Errorf("Transaction panicked, rolling back: %v", p)
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				logger.Errorf("Failed to rollback after panic: %v", rollbackErr)
			}
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		logger.Debugf("Transaction function returned error, rolling back: %v", err)
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			logger.Errorf("Failed to rollback transaction: %v", rollbackErr)
			return fmt.Errorf("transaction failed and rollback failed: %w", rollbackErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		logger.Errorf("Failed to commit transaction: %v", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.Debugf("Transaction committed successfully")
	return nil
}

// WithLock executes a function with a table lock within a transaction
// Automatically starts transaction, commits on success, rolls back on error
func (dm *DatabaseManager) WithLock(ctx context.Context, lockQuery string, fn LockFunc) error {
	logger.Debugf("Starting transaction with lock: %s", lockQuery)

	return dm.WithTxn(ctx, func(tx pgx.Tx) error {
		// Acquire the lock
		_, err := tx.Exec(ctx, lockQuery)
		if err != nil {
			logger.Errorf("Failed to acquire lock: %v", err)
			return fmt.Errorf("failed to acquire lock: %w", err)
		}

		logger.Debugf("Lock acquired successfully")

		// Execute the function with the locked transaction
		return fn(tx)
	})
}

// Close closes the database connection pool
func (dm *DatabaseManager) Close() error {
	dm.pool.Close()
	logger.Debugf("Database connection closed")
	return nil
}
