package mock_database

import (
	"context"
	"fmt"
	"time"

	"expenses/pkg/database/manager/base"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// MockDatabaseManager implements the unified DatabaseManager interface for testing
type MockDatabaseManager struct {
	// Configuration
	config *base.DatabaseManagerConfig

	// Test control flags
	ShouldFailWithTxn bool
	ExecuteQueryError error
	FetchAllError     error

	// Mock data
	mockMetrics base.TransactionMetrics
}

// NewMockDatabaseManager creates a new mock database manager with default config
func NewMockDatabaseManager() *MockDatabaseManager {
	return NewMockDatabaseManagerWithConfig(base.DefaultConfig())
}

// NewMockDatabaseManagerWithConfig creates a new mock database manager with custom config
func NewMockDatabaseManagerWithConfig(config *base.DatabaseManagerConfig) *MockDatabaseManager {
	return &MockDatabaseManager{
		config: config,
		mockMetrics: base.TransactionMetrics{
			ErrorsByType: make(map[string]int64),
		},
	}
}

// === CORE OPERATIONS (always available) ===

// ExecuteQuery mocks query execution
func (m *MockDatabaseManager) ExecuteQuery(ctx context.Context, query string, args ...any) (rowsAffected int64, err error) {
	if m.ExecuteQueryError != nil {
		return 0, m.ExecuteQueryError
	}
	// Return 1 row affected by default for successful operations
	return 1, nil
}

// FetchOne mocks single row fetching by returning a MockRow
func (m *MockDatabaseManager) FetchOne(ctx context.Context, query string, args ...any) pgx.Row {
	return &MockRow{}
}

// FetchAll mocks multiple row fetching
func (m *MockDatabaseManager) FetchAll(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	if m.FetchAllError != nil {
		return nil, m.FetchAllError
	}
	return &MockRows{}, nil
}

// WithTxn mocks transaction execution
func (m *MockDatabaseManager) WithTxn(ctx context.Context, fn base.TransactionFunc) error {
	if m.ShouldFailWithTxn {
		return fmt.Errorf("mock transaction error")
	}

	// Update metrics if monitoring is enabled
	if m.config.EnableMonitoring {
		m.mockMetrics.TotalTransactions++
		m.mockMetrics.CommittedTransactions++
	}

	// Execute the function with a context
	return fn(ctx)
}

// WithLock mocks lock execution
func (m *MockDatabaseManager) WithLock(ctx context.Context, lockKey int64, fn base.LockFunc) error {
	// For testing, just execute the function with a context
	return fn(ctx)
}

// Close mocks closing the database connection
func (m *MockDatabaseManager) Close() error {
	return nil
}

// === ENHANCED OPERATIONS (configurable) ===

// WithTxnOptions mocks transaction with options
func (m *MockDatabaseManager) WithTxnOptions(ctx context.Context, opts *base.TransactionOptions, fn base.TransactionFunc) error {
	if !m.config.EnableRetry {
		return fmt.Errorf("retry feature is disabled")
	}
	return m.WithTxn(ctx, fn)
}

// WithReadOnlyTxn mocks read-only transaction
func (m *MockDatabaseManager) WithReadOnlyTxn(ctx context.Context, fn base.TransactionFunc) error {
	if !m.config.EnableRetry {
		return fmt.Errorf("retry feature is disabled")
	}
	return m.WithTxn(ctx, fn)
}

// WithRetryableTxn mocks retryable transaction
func (m *MockDatabaseManager) WithRetryableTxn(ctx context.Context, fn base.TransactionFunc) error {
	if !m.config.EnableRetry {
		return fmt.Errorf("retry feature is disabled")
	}
	return m.WithTxn(ctx, fn)
}

// WithSavepoint mocks savepoint execution
func (m *MockDatabaseManager) WithSavepoint(ctx context.Context, name string, fn base.TransactionFunc) error {
	if !m.config.EnableSavepoints {
		return fmt.Errorf("savepoints feature is disabled")
	}
	return fn(ctx)
}

// ExecuteBatch mocks batch execution
func (m *MockDatabaseManager) ExecuteBatch(ctx context.Context, batch *pgx.Batch) error {
	if !m.config.EnableBatch {
		return fmt.Errorf("batch feature is disabled")
	}
	return nil
}

// WithConnection mocks connection execution
func (m *MockDatabaseManager) WithConnection(ctx context.Context, fn func(conn *pgx.Conn) error) error {
	return fn(nil) // Pass nil connection for mock
}

// === HEALTH & INTROSPECTION (always available) ===

// Ping mocks database ping
func (m *MockDatabaseManager) Ping(ctx context.Context) error {
	return nil
}

// Stats mocks database statistics
func (m *MockDatabaseManager) Stats() base.DatabaseStats {
	return base.DatabaseStats{
		TotalConnections:    10,
		IdleConnections:     5,
		AcquiredConnections: 3,
		MaxConnections:      25,
		AcquireCount:        100,
		AcquireDuration:     50 * time.Millisecond,
	}
}

// GetTransactionInfo mocks transaction info
func (m *MockDatabaseManager) GetTransactionInfo(ctx context.Context) (*base.TransactionInfo, error) {
	return &base.TransactionInfo{
		ID:             "mock-tx-123",
		StartTime:      time.Now().Add(-100 * time.Millisecond),
		Duration:       100 * time.Millisecond,
		IsReadOnly:     false,
		IsNested:       false,
		IsolationLevel: "READ_COMMITTED",
	}, nil
}

// === MONITORING (configurable) ===

// GetMonitoringMetrics mocks monitoring metrics
func (m *MockDatabaseManager) GetMonitoringMetrics() base.TransactionMetrics {
	if !m.config.EnableMonitoring {
		return base.TransactionMetrics{} // Return empty metrics
	}
	return m.mockMetrics
}

// ResetMetrics mocks metrics reset
func (m *MockDatabaseManager) ResetMetrics() {
	if m.config.EnableMonitoring {
		m.mockMetrics = base.TransactionMetrics{
			ErrorsByType: make(map[string]int64),
		}
	}
}

// === CONFIGURATION ===

// GetConfig returns the mock configuration
func (m *MockDatabaseManager) GetConfig() *base.DatabaseManagerConfig {
	return m.config
}

// IsFeatureEnabled checks if a specific feature is enabled
func (m *MockDatabaseManager) IsFeatureEnabled(feature string) bool {
	switch feature {
	case base.FeatureRetry:
		return m.config.EnableRetry
	case base.FeatureSavepoints:
		return m.config.EnableSavepoints
	case base.FeatureBatch:
		return m.config.EnableBatch
	case base.FeatureMonitoring:
		return m.config.EnableMonitoring
	case base.FeatureMetrics:
		return m.config.EnableMetrics
	default:
		return false
	}
}

// === TEST HELPER METHODS ===

// SetShouldFailWithTxn sets whether transactions should fail
func (m *MockDatabaseManager) SetShouldFailWithTxn(shouldFail bool) {
	m.ShouldFailWithTxn = shouldFail
}

// SetExecuteQueryError sets the error to return from ExecuteQuery
func (m *MockDatabaseManager) SetExecuteQueryError(err error) {
	m.ExecuteQueryError = err
}

// SetFetchAllError sets the error to return from FetchAll
func (m *MockDatabaseManager) SetFetchAllError(err error) {
	m.FetchAllError = err
}

// UpdateMockMetrics updates the mock metrics for testing
func (m *MockDatabaseManager) UpdateMockMetrics(metrics base.TransactionMetrics) {
	m.mockMetrics = metrics
}

// === MOCK IMPLEMENTATIONS ===

// MockRow implements pgx.Row for testing
type MockRow struct {
	ScanError error
}

func (m *MockRow) Scan(dest ...any) error {
	if m.ScanError != nil {
		return m.ScanError
	}
	// For testing, just return nil to indicate successful scan
	return nil
}

// MockRows implements pgx.Rows for testing
type MockRows struct {
	closed bool
}

func (m *MockRows) Close() {
	m.closed = true
}

func (m *MockRows) Next() bool {
	return false // No rows to iterate for basic mock
}

func (m *MockRows) Scan(dest ...any) error {
	return nil
}

func (m *MockRows) Err() error {
	return nil
}

func (m *MockRows) CommandTag() pgconn.CommandTag {
	return pgconn.CommandTag{}
}

func (m *MockRows) FieldDescriptions() []pgconn.FieldDescription {
	return []pgconn.FieldDescription{}
}

func (m *MockRows) Values() ([]any, error) {
	return []any{}, nil
}

func (m *MockRows) RawValues() [][]byte {
	return [][]byte{}
}

func (m *MockRows) Conn() *pgx.Conn {
	return nil
}
