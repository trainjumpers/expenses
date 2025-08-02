package base

import (
	"time"

	"github.com/jackc/pgx/v5"
)

// TransactionOptions configures transaction behavior
type TransactionOptions struct {
	// IsolationLevel sets the transaction isolation level
	IsolationLevel pgx.TxIsoLevel

	// AccessMode sets whether transaction is read-only
	AccessMode pgx.TxAccessMode

	// DeferrableMode sets whether transaction is deferrable
	DeferrableMode pgx.TxDeferrableMode

	// Timeout sets maximum transaction duration
	Timeout time.Duration

	// ReadOnly is a convenience flag for read-only transactions
	ReadOnly bool

	// RetryPolicy defines retry behavior for failed transactions
	RetryPolicy *RetryPolicy
}

// RetryPolicy defines how to retry failed transactions
type RetryPolicy struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Backoff    BackoffStrategy
}

type BackoffStrategy int

const (
	BackoffLinear BackoffStrategy = iota
	BackoffExponential
	BackoffFixed
)

// DefaultTransactionOptions returns sensible defaults
func DefaultTransactionOptions() *TransactionOptions {
	return &TransactionOptions{
		IsolationLevel: pgx.ReadCommitted,
		AccessMode:     pgx.ReadWrite,
		DeferrableMode: pgx.NotDeferrable,
		Timeout:        30 * time.Second,
		ReadOnly:       false,
		RetryPolicy: &RetryPolicy{
			MaxRetries: 3,
			BaseDelay:  100 * time.Millisecond,
			MaxDelay:   5 * time.Second,
			Backoff:    BackoffExponential,
		},
	}
}

// ReadOnlyTransactionOptions returns options for read-only transactions
func ReadOnlyTransactionOptions() *TransactionOptions {
	opts := DefaultTransactionOptions()
	opts.AccessMode = pgx.ReadOnly
	opts.ReadOnly = true
	opts.Timeout = 10 * time.Second // Shorter timeout for reads
	return opts
}

// LongRunningTransactionOptions returns options for long-running transactions
func LongRunningTransactionOptions() *TransactionOptions {
	opts := DefaultTransactionOptions()
	opts.Timeout = 5 * time.Minute
	opts.RetryPolicy.MaxRetries = 1 // Less aggressive retry for long operations
	return opts
}
