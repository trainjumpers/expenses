package postgres

import (
	"context"
	"fmt"
	"maps"
	"sync"
	"time"

	"expenses/pkg/database/manager/base"
	"expenses/pkg/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresDatabaseManager implements the unified DatabaseManager interface
type PostgresDatabaseManager struct {
	pool    *pgxpool.Pool
	config  *base.DatabaseManagerConfig
	monitor *TransactionMonitor
}

// NewPostgresDatabaseManager creates a new unified database manager
func NewPostgresDatabaseManager(pool *pgxpool.Pool, config *base.DatabaseManagerConfig) *PostgresDatabaseManager {
	manager := &PostgresDatabaseManager{
		pool:   pool,
		config: config,
	}

	// Initialize monitoring if enabled
	if config.EnableMonitoring {
		manager.monitor = NewTransactionMonitor()
	}

	return manager
}

// === CORE OPERATIONS (always available) ===

// getExecutor returns the appropriate executor (transaction or pool) based on context
func (dm *PostgresDatabaseManager) getExecutor(ctx context.Context) interface {
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
} {
	if txCtx, ok := base.GetTransactionContext(ctx); ok && txCtx.Tx != nil {
		logger.Debugf("Using transaction executor (ID: %s)", txCtx.ID)
		return txCtx.Tx
	}
	logger.Debugf("Using pool executor")
	return dm.pool
}

func (dm *PostgresDatabaseManager) ExecuteQuery(ctx context.Context, query string, args ...any) (rowsAffected int64, err error) {
	logger.Debugf("Executing query: %s", query)

	executor := dm.getExecutor(ctx)
	result, err := executor.Exec(ctx, query, args...)
	if err != nil {
		logger.Errorf("Failed to execute query: %v", err)
		return 0, err
	}

	logger.Debugf("Query executed successfully, rows affected: %d", result.RowsAffected())
	return result.RowsAffected(), nil
}

func (dm *PostgresDatabaseManager) FetchOne(ctx context.Context, query string, args ...any) pgx.Row {
	logger.Debugf("Fetching single row: %s", query)
	executor := dm.getExecutor(ctx)
	return executor.QueryRow(ctx, query, args...)
}

func (dm *PostgresDatabaseManager) FetchAll(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	logger.Debugf("Fetching multiple rows: %s", query)

	executor := dm.getExecutor(ctx)
	rows, err := executor.Query(ctx, query, args...)
	if err != nil {
		logger.Errorf("Failed to fetch rows: %v", err)
		return nil, err
	}
	return rows, nil
}

func (dm *PostgresDatabaseManager) WithTxn(ctx context.Context, fn base.TransactionFunc) error {
	if dm.config.EnableMonitoring && dm.monitor != nil {
		return dm.withMonitoredTransaction(ctx, fn)
	}
	return dm.withBasicTransaction(ctx, fn)
}

func (dm *PostgresDatabaseManager) WithLock(ctx context.Context, lockKey int64, fn base.LockFunc) error {
	logger.Debugf("Starting transaction with lock: %d", lockKey)

	return dm.WithTxn(ctx, func(txCtx context.Context) error {
		// Get the transaction from context for lock acquisition
		transactionCtx, ok := base.GetTransactionContext(txCtx)
		if !ok {
			return fmt.Errorf("transaction not found in context")
		}

		// Store lock key in transaction context
		transactionCtx.LockKey = &lockKey

		// Acquire the lock
		_, err := transactionCtx.Tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", lockKey)
		if err != nil {
			logger.Errorf("Failed to acquire lock: %v", err)
			return fmt.Errorf("failed to acquire lock: %w", err)
		}

		logger.Debugf("Lock acquired successfully (key: %d)", lockKey)

		// Execute the function with the transaction context
		return fn(txCtx)
	})
}

func (dm *PostgresDatabaseManager) Close() error {
	dm.pool.Close()
	logger.Debugf("Database connection closed")
	return nil
}

// === ENHANCED OPERATIONS (configurable) ===

func (dm *PostgresDatabaseManager) WithTxnOptions(ctx context.Context, opts *base.TransactionOptions, fn base.TransactionFunc) error {
	if !dm.config.EnableRetry {
		return fmt.Errorf("retry feature is disabled")
	}

	if opts == nil {
		opts = base.DefaultTransactionOptions()
	}

	// Apply timeout if specified
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// Retry logic
	var lastErr error
	maxRetries := 1
	if opts.RetryPolicy != nil {
		maxRetries = opts.RetryPolicy.MaxRetries + 1
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Apply backoff delay
			delay := dm.calculateBackoffDelay(attempt, opts.RetryPolicy)
			logger.Debugf("Retrying transaction (attempt %d/%d) after %v", attempt+1, maxRetries, delay)
			time.Sleep(delay)
		}

		lastErr = dm.executeTransactionWithOptions(ctx, opts, fn)
		if lastErr == nil {
			return nil // Success
		}

		// Check if error is retryable
		if !dm.isRetryableError(lastErr) {
			break
		}
	}

	return fmt.Errorf("transaction failed after %d attempts: %w", maxRetries, lastErr)
}

func (dm *PostgresDatabaseManager) WithReadOnlyTxn(ctx context.Context, fn base.TransactionFunc) error {
	if !dm.config.EnableRetry {
		return fmt.Errorf("retry feature is disabled")
	}

	opts := base.ReadOnlyTransactionOptions()
	return dm.WithTxnOptions(ctx, opts, fn)
}

func (dm *PostgresDatabaseManager) WithRetryableTxn(ctx context.Context, fn base.TransactionFunc) error {
	if !dm.config.EnableRetry {
		return fmt.Errorf("retry feature is disabled")
	}

	opts := base.DefaultTransactionOptions()
	opts.RetryPolicy.MaxRetries = 5
	opts.RetryPolicy.BaseDelay = 50 * time.Millisecond
	return dm.WithTxnOptions(ctx, opts, fn)
}

func (dm *PostgresDatabaseManager) WithSavepoint(ctx context.Context, name string, fn base.TransactionFunc) error {
	if !dm.config.EnableSavepoints {
		return fmt.Errorf("savepoints feature is disabled")
	}

	// Check if we're already in a transaction
	txCtx, inTx := base.GetTransactionContext(ctx)
	if !inTx {
		return fmt.Errorf("savepoints can only be used within an existing transaction")
	}

	// Create savepoint
	_, err := txCtx.Tx.Exec(ctx, fmt.Sprintf("SAVEPOINT %s", name))
	if err != nil {
		return fmt.Errorf("failed to create savepoint %s: %w", name, err)
	}

	// Create nested transaction context
	nestedTxCtx := &base.TransactionContext{
		Tx:        txCtx.Tx, // Same underlying transaction
		StartTime: getCurrentTime(),
		ID:        generateTransactionID(),
		ReadOnly:  txCtx.ReadOnly,
		Nested:    true,
		LockKey:   txCtx.LockKey,
	}

	nestedCtx := base.WithTransactionContext(ctx, nestedTxCtx)

	// Execute function
	err = fn(nestedCtx)
	if err != nil {
		// Rollback to savepoint
		if rollbackErr := dm.rollbackToSavepoint(ctx, txCtx.Tx, name); rollbackErr != nil {
			logger.Errorf("Failed to rollback to savepoint %s: %v", name, rollbackErr)
		}
		return err
	}

	// Release savepoint
	_, err = txCtx.Tx.Exec(ctx, fmt.Sprintf("RELEASE SAVEPOINT %s", name))
	if err != nil {
		return fmt.Errorf("failed to release savepoint %s: %w", name, err)
	}

	return nil
}

func (dm *PostgresDatabaseManager) ExecuteBatch(ctx context.Context, batch *pgx.Batch) error {
	if !dm.config.EnableBatch {
		return fmt.Errorf("batch feature is disabled")
	}

	// Check if we're in a transaction
	if txCtx, ok := base.GetTransactionContext(ctx); ok {
		results := txCtx.Tx.SendBatch(ctx, batch)
		defer results.Close()

		// Process all results to ensure they complete
		for i := 0; i < batch.Len(); i++ {
			_, err := results.Exec()
			if err != nil {
				return fmt.Errorf("batch operation %d failed: %w", i, err)
			}
		}
		return nil
	}

	// For pool connections
	results := dm.pool.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("batch operation %d failed: %w", i, err)
		}
	}
	return nil
}

func (dm *PostgresDatabaseManager) WithConnection(ctx context.Context, fn func(conn *pgx.Conn) error) error {
	conn, err := dm.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	return fn(conn.Conn())
}

// === HEALTH & INTROSPECTION (always available) ===

func (dm *PostgresDatabaseManager) Ping(ctx context.Context) error {
	return dm.pool.Ping(ctx)
}

func (dm *PostgresDatabaseManager) Stats() base.DatabaseStats {
	stats := dm.pool.Stat()
	return base.DatabaseStats{
		TotalConnections:        stats.TotalConns(),
		IdleConnections:         stats.IdleConns(),
		AcquiredConnections:     stats.AcquiredConns(),
		ConstructingConnections: stats.ConstructingConns(),
		MaxConnections:          stats.MaxConns(),
		AcquireCount:            stats.AcquireCount(),
		AcquireDuration:         stats.AcquireDuration(),
		EmptyAcquireCount:       stats.EmptyAcquireCount(),
		CanceledAcquireCount:    stats.CanceledAcquireCount(),
	}
}

func (dm *PostgresDatabaseManager) GetTransactionInfo(ctx context.Context) (*base.TransactionInfo, error) {
	txCtx, ok := base.GetTransactionContext(ctx)
	if !ok {
		return nil, fmt.Errorf("not in a transaction")
	}

	return &base.TransactionInfo{
		ID:             txCtx.ID,
		StartTime:      txCtx.StartTime,
		Duration:       time.Since(txCtx.StartTime),
		IsReadOnly:     txCtx.ReadOnly,
		IsNested:       txCtx.Nested,
		LockKey:        txCtx.LockKey,
		IsolationLevel: "READ_COMMITTED", // Could be made dynamic
	}, nil
}

// === MONITORING (configurable) ===

func (dm *PostgresDatabaseManager) GetMonitoringMetrics() base.TransactionMetrics {
	if !dm.config.EnableMonitoring || dm.monitor == nil {
		return base.TransactionMetrics{} // Return empty metrics
	}
	return dm.monitor.GetMetrics()
}

func (dm *PostgresDatabaseManager) ResetMetrics() {
	if dm.config.EnableMonitoring && dm.monitor != nil {
		dm.monitor.Reset()
	}
}

// === CONFIGURATION ===

func (dm *PostgresDatabaseManager) GetConfig() *base.DatabaseManagerConfig {
	return dm.config
}

func (dm *PostgresDatabaseManager) IsFeatureEnabled(feature string) bool {
	switch feature {
	case base.FeatureRetry:
		return dm.config.EnableRetry
	case base.FeatureSavepoints:
		return dm.config.EnableSavepoints
	case base.FeatureBatch:
		return dm.config.EnableBatch
	case base.FeatureMonitoring:
		return dm.config.EnableMonitoring
	case base.FeatureMetrics:
		return dm.config.EnableMetrics
	default:
		return false
	}
}

// === HELPER METHODS ===

func (dm *PostgresDatabaseManager) withBasicTransaction(ctx context.Context, fn base.TransactionFunc) error {
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

	// Create transaction context
	txCtx := &base.TransactionContext{
		Tx:        tx,
		StartTime: getCurrentTime(),
		ID:        generateTransactionID(),
		ReadOnly:  false,
		Nested:    false,
	}

	// Create a new context with the transaction
	enhancedCtx := base.WithTransactionContext(ctx, txCtx)

	if err := fn(enhancedCtx); err != nil {
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

	logger.Debugf("Transaction committed successfully (ID: %s)", txCtx.ID)
	return nil
}

func (dm *PostgresDatabaseManager) withMonitoredTransaction(ctx context.Context, fn base.TransactionFunc) error {
	txID := generateTransactionID()
	startTime := getCurrentTime()

	dm.monitor.StartTransaction(txID)

	var committed bool
	err := dm.withBasicTransaction(ctx, func(txCtx context.Context) error {
		// Add transaction ID to context for logging
		enhancedCtx := context.WithValue(txCtx, "transaction_id", txID)

		err := fn(enhancedCtx)
		if err == nil {
			committed = true
		}
		return err
	})

	duration := time.Since(startTime)
	dm.monitor.EndTransaction(txID, duration, committed, err)

	return err
}

// Additional helper methods from enhanced.go
func (dm *PostgresDatabaseManager) executeTransactionWithOptions(ctx context.Context, opts *base.TransactionOptions, fn base.TransactionFunc) error {
	logger.Debugf("Starting transaction with options: isolation=%v, readonly=%v", opts.IsolationLevel, opts.ReadOnly)

	// Begin transaction with options
	txOpts := pgx.TxOptions{
		IsoLevel:       opts.IsolationLevel,
		AccessMode:     opts.AccessMode,
		DeferrableMode: opts.DeferrableMode,
	}

	tx, err := dm.pool.BeginTx(ctx, txOpts)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Create transaction context
	txCtx := &base.TransactionContext{
		Tx:        tx,
		StartTime: getCurrentTime(),
		ID:        generateTransactionID(),
		ReadOnly:  opts.ReadOnly,
		Nested:    false,
	}

	// Add to context
	enhancedCtx := base.WithTransactionContext(ctx, txCtx)

	// Execute with proper cleanup
	return dm.executeWithCleanup(ctx, tx, enhancedCtx, fn)
}

func (dm *PostgresDatabaseManager) executeWithCleanup(ctx context.Context, tx pgx.Tx, txCtx context.Context, fn base.TransactionFunc) error {
	defer func() {
		if p := recover(); p != nil {
			logger.Errorf("Transaction panicked, rolling back: %v", p)
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				logger.Errorf("Failed to rollback after panic: %v", rollbackErr)
			}
			panic(p)
		}
	}()

	if err := fn(txCtx); err != nil {
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

func (dm *PostgresDatabaseManager) calculateBackoffDelay(attempt int, policy *base.RetryPolicy) time.Duration {
	if policy == nil {
		return 100 * time.Millisecond
	}

	var delay time.Duration
	switch policy.Backoff {
	case base.BackoffLinear:
		delay = policy.BaseDelay * time.Duration(attempt)
	case base.BackoffExponential:
		delay = policy.BaseDelay * time.Duration(1<<uint(attempt-1))
	case base.BackoffFixed:
		delay = policy.BaseDelay
	default:
		delay = policy.BaseDelay
	}

	if delay > policy.MaxDelay {
		delay = policy.MaxDelay
	}

	return delay
}

func (dm *PostgresDatabaseManager) isRetryableError(err error) bool {
	// Check for common retryable PostgreSQL errors
	errStr := err.Error()
	retryableErrors := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"deadlock detected",
		"serialization failure",
	}

	for _, retryable := range retryableErrors {
		if contains(errStr, retryable) {
			return true
		}
	}
	return false
}

func (dm *PostgresDatabaseManager) rollbackToSavepoint(ctx context.Context, tx pgx.Tx, name string) error {
	_, err := tx.Exec(ctx, fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", name))
	return err
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TransactionMonitor provides monitoring capabilities
type TransactionMonitor struct {
	metrics *base.TransactionMetrics
	mu      sync.RWMutex
}

// NewTransactionMonitor creates a new transaction monitor
func NewTransactionMonitor() *TransactionMonitor {
	return &TransactionMonitor{
		metrics: &base.TransactionMetrics{
			ErrorsByType: make(map[string]int64),
			MinDuration:  time.Hour, // Initialize to high value
		},
	}
}

// StartTransaction records the start of a transaction
func (tm *TransactionMonitor) StartTransaction(txID string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.metrics.TotalTransactions++
	tm.metrics.ActiveTransactions++

	logger.Debugf("Transaction started: %s (active: %d)", txID, tm.metrics.ActiveTransactions)
}

// EndTransaction records the end of a transaction
func (tm *TransactionMonitor) EndTransaction(txID string, duration time.Duration, committed bool, err error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.metrics.ActiveTransactions--
	tm.metrics.TotalDuration += duration

	// Update duration stats
	if duration > tm.metrics.MaxDuration {
		tm.metrics.MaxDuration = duration
	}
	if duration < tm.metrics.MinDuration {
		tm.metrics.MinDuration = duration
	}

	// Calculate average
	if tm.metrics.TotalTransactions > 0 {
		tm.metrics.AverageDuration = tm.metrics.TotalDuration / time.Duration(tm.metrics.TotalTransactions)
	}

	// Update outcome counters
	if err != nil {
		tm.metrics.FailedTransactions++
		errType := err.Error()
		if len(errType) > 50 {
			errType = errType[:50] // Truncate long errors
		}
		tm.metrics.ErrorsByType[errType]++
	} else if committed {
		tm.metrics.CommittedTransactions++
	} else {
		tm.metrics.RolledBackTransactions++
	}

	logger.Debugf("Transaction ended: %s (duration: %v, committed: %v, error: %v)",
		txID, duration, committed, err)
}

// GetMetrics returns a copy of current metrics
func (tm *TransactionMonitor) GetMetrics() base.TransactionMetrics {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	// Create a copy to avoid race conditions
	errorsCopy := make(map[string]int64)
	maps.Copy(errorsCopy, tm.metrics.ErrorsByType)

	return base.TransactionMetrics{
		TotalTransactions:      tm.metrics.TotalTransactions,
		CommittedTransactions:  tm.metrics.CommittedTransactions,
		RolledBackTransactions: tm.metrics.RolledBackTransactions,
		FailedTransactions:     tm.metrics.FailedTransactions,
		TotalDuration:          tm.metrics.TotalDuration,
		AverageDuration:        tm.metrics.AverageDuration,
		MaxDuration:            tm.metrics.MaxDuration,
		MinDuration:            tm.metrics.MinDuration,
		ActiveTransactions:     tm.metrics.ActiveTransactions,
		ErrorsByType:           errorsCopy,
	}
}

// Reset clears all metrics
func (tm *TransactionMonitor) Reset() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.metrics = &base.TransactionMetrics{
		ErrorsByType: make(map[string]int64),
		MinDuration:  time.Hour,
	}
}
