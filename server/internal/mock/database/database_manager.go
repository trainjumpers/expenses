package mock_database

import (
	"context"
	database "expenses/internal/database/manager"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// MockDatabaseManager implements the DatabaseManager interface for testing
type MockDatabaseManager struct {
	ShouldFailWithTxn bool
	ExecuteQueryError error
	FetchAllError     error
}

// NewMockDatabaseManager creates a new mock database manager
func NewMockDatabaseManager() *MockDatabaseManager {
	return &MockDatabaseManager{}
}

// ExecuteQuery mocks query execution
func (m *MockDatabaseManager) ExecuteQuery(ctx context.Context, query string, args ...interface{}) (rowsAffected int64, err error) {
	if m.ExecuteQueryError != nil {
		return 0, m.ExecuteQueryError
	}
	// Return 1 row affected by default for successful operations
	return 1, nil
}

// FetchOne mocks single row fetching by returning a MockRow
func (m *MockDatabaseManager) FetchOne(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return &MockRow{}
}

// FetchAll mocks multiple row fetching
func (m *MockDatabaseManager) FetchAll(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	if m.FetchAllError != nil {
		return nil, m.FetchAllError
	}
	return &MockRows{}, nil
}

// WithTxn mocks transaction execution
func (m *MockDatabaseManager) WithTxn(ctx context.Context, fn database.TransactionFunc) error {
	if m.ShouldFailWithTxn {
		return fmt.Errorf("mock transaction error")
	}
	// Execute the function with a mock transaction
	return fn(&MockTx{})
}

// WithLock mocks lock execution
func (m *MockDatabaseManager) WithLock(ctx context.Context, lockKey int64, fn database.LockFunc) error {
	// For testing, just execute the function with a mock transaction
	return fn(&MockTx{})
}

// Close mocks closing the database connection
func (m *MockDatabaseManager) Close() error {
	return nil
}

// MockRow implements pgx.Row for testing
type MockRow struct {
	ScanError error
}

func (m *MockRow) Scan(dest ...interface{}) error {
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

func (m *MockRows) Scan(dest ...interface{}) error {
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

func (m *MockRows) Values() ([]interface{}, error) {
	return []interface{}{}, nil
}

func (m *MockRows) RawValues() [][]byte {
	return [][]byte{}
}

func (m *MockRows) Conn() *pgx.Conn {
	return nil
}

// MockTx implements pgx.Tx for testing
type MockTx struct{}

func (m *MockTx) Begin(ctx context.Context) (pgx.Tx, error) {
	return &MockTx{}, nil
}

func (m *MockTx) Commit(ctx context.Context) error {
	return nil
}

func (m *MockTx) Rollback(ctx context.Context) error {
	return nil
}

func (m *MockTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}

func (m *MockTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return &MockBatchResults{}
}

func (m *MockTx) LargeObjects() pgx.LargeObjects {
	return pgx.LargeObjects{}
}

func (m *MockTx) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (m *MockTx) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return &MockRows{}, nil
}

func (m *MockTx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return &MockRow{}
}

func (m *MockTx) Conn() *pgx.Conn {
	return nil
}

// Prepare is not implemented as it's not needed for our mock
func (m *MockTx) Prepare(ctx context.Context, name, sql string) (description *pgconn.StatementDescription, err error) {
	return nil, nil
}

// MockBatchResults implements pgx.BatchResults for testing
type MockBatchResults struct{}

func (m *MockBatchResults) Exec() (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (m *MockBatchResults) Query() (pgx.Rows, error) {
	return &MockRows{}, nil
}

func (m *MockBatchResults) QueryRow() pgx.Row {
	return &MockRow{}
}

func (m *MockBatchResults) Close() error {
	return nil
}
