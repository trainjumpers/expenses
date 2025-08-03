package base

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// TransactionContext holds transaction metadata and state
type TransactionContext struct {
	Tx        pgx.Tx
	StartTime time.Time
	ID        string // Unique transaction ID for logging/tracing
	ReadOnly  bool   // Whether this is a read-only transaction
	Nested    bool   // Whether this is a nested transaction (savepoint)
	LockKey   *int64 // Advisory lock key if applicable
}

// contextKey is a private type for context keys to avoid collisions
type contextKey struct {
	name string
}

var (
	txContextKey = &contextKey{"database_transaction"}
)

// GetTransactionContext retrieves transaction context from context
func GetTransactionContext(ctx context.Context) (*TransactionContext, bool) {
	txCtx, ok := ctx.Value(txContextKey).(*TransactionContext)
	return txCtx, ok
}

// WithTransactionContext adds transaction context to the given context
func WithTransactionContext(ctx context.Context, txCtx *TransactionContext) context.Context {
	return context.WithValue(ctx, txContextKey, txCtx)
}

// IsInTransaction checks if the context has an active transaction
func IsInTransaction(ctx context.Context) bool {
	_, ok := GetTransactionContext(ctx)
	return ok
}

// GetTransactionID returns the transaction ID if in transaction, empty string otherwise
func GetTransactionID(ctx context.Context) string {
	if txCtx, ok := GetTransactionContext(ctx); ok {
		return txCtx.ID
	}
	return ""
}
