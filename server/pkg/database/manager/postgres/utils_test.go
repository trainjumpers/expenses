package postgres

import (
	"context"
	"encoding/hex"
	"errors"
	"time"

	"expenses/pkg/database/manager/base"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// mockTx implements the pgx.Tx interface for testing
type mockTx struct {
	pgx.Tx
	commitCalled   bool
	rollbackCalled bool
	commitErr      error
	rollbackErr    error
}

func (m *mockTx) Commit(ctx context.Context) error {
	m.commitCalled = true
	return m.commitErr
}

func (m *mockTx) Rollback(ctx context.Context) error {
	m.rollbackCalled = true
	return m.rollbackErr
}

func (m *mockTx) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

var _ = Describe("Postgres Utilities", func() {

	Describe("generateTransactionID", func() {
		It("should generate a unique, valid transaction ID", func() {
			// Generate two IDs
			id1 := generateTransactionID()
			id2 := generateTransactionID()

			// Check that they are different
			Expect(id1).NotTo(Equal(id2))

			// Check the format of the first ID
			Expect(len(id1)).To(Equal(16), "Transaction ID should be 16 characters long")
			_, err := hex.DecodeString(id1)
			Expect(err).NotTo(HaveOccurred(), "Transaction ID should be a valid hex string")

			// Check the format of the second ID
			Expect(len(id2)).To(Equal(16))
			_, err = hex.DecodeString(id2)
			Expect(err).NotTo(HaveOccurred())
		})

		var _ = Describe("Internal Helper Functions", func() {
			var (
				dm *PostgresDatabaseManager
			)

			BeforeEach(func() {
				// A mock pool is not needed as these helpers don't use it directly
				dm = &PostgresDatabaseManager{}
			})

			Describe("calculateBackoffDelay", func() {
				It("should calculate linear backoff delay", func() {
					policy := &base.RetryPolicy{
						BaseDelay: 100 * time.Millisecond,
						MaxDelay:  1 * time.Second,
						Backoff:   base.BackoffLinear,
					}
					Expect(dm.calculateBackoffDelay(1, policy)).To(Equal(100 * time.Millisecond))
					Expect(dm.calculateBackoffDelay(2, policy)).To(Equal(200 * time.Millisecond))
				})

				It("should calculate exponential backoff delay", func() {
					policy := &base.RetryPolicy{
						BaseDelay: 100 * time.Millisecond,
						MaxDelay:  1 * time.Second,
						Backoff:   base.BackoffExponential,
					}
					Expect(dm.calculateBackoffDelay(1, policy)).To(Equal(100 * time.Millisecond)) // 100 * 2^0
					Expect(dm.calculateBackoffDelay(2, policy)).To(Equal(200 * time.Millisecond)) // 100 * 2^1
					Expect(dm.calculateBackoffDelay(3, policy)).To(Equal(400 * time.Millisecond)) // 100 * 2^2
				})

				It("should calculate fixed backoff delay", func() {
					policy := &base.RetryPolicy{
						BaseDelay: 100 * time.Millisecond,
						MaxDelay:  1 * time.Second,
						Backoff:   base.BackoffFixed,
					}
					Expect(dm.calculateBackoffDelay(1, policy)).To(Equal(100 * time.Millisecond))
					Expect(dm.calculateBackoffDelay(2, policy)).To(Equal(100 * time.Millisecond))
				})

				It("should cap the delay at MaxDelay", func() {
					policy := &base.RetryPolicy{
						BaseDelay: 500 * time.Millisecond,
						MaxDelay:  800 * time.Millisecond,
						Backoff:   base.BackoffExponential,
					}
					// 500ms * 2^(2-1) = 1000ms, which is > 800ms, so it should be capped
					Expect(dm.calculateBackoffDelay(2, policy)).To(Equal(800 * time.Millisecond))
				})
			})

			Describe("executeWithCleanup", func() {
				var (
					mockTransaction *mockTx
					ctx             context.Context
				)

				BeforeEach(func() {
					mockTransaction = &mockTx{}
					ctx = context.Background()
				})

				It("should commit on success", func() {
					err := dm.executeWithCleanup(ctx, mockTransaction, ctx, func(txCtx context.Context) error {
						return nil
					})
					Expect(err).NotTo(HaveOccurred())
					Expect(mockTransaction.commitCalled).To(BeTrue())
					Expect(mockTransaction.rollbackCalled).To(BeFalse())
				})

				It("should rollback when the function returns an error", func() {
					expectedErr := errors.New("something went wrong")
					err := dm.executeWithCleanup(ctx, mockTransaction, ctx, func(txCtx context.Context) error {
						return expectedErr
					})
					Expect(err).To(MatchError(expectedErr))
					Expect(mockTransaction.commitCalled).To(BeFalse())
					Expect(mockTransaction.rollbackCalled).To(BeTrue())
				})

				It("should rollback and re-panic when the function panics", func() {
					Expect(func() {
						dm.executeWithCleanup(ctx, mockTransaction, ctx, func(txCtx context.Context) error {
							panic("something terrible happened")
						})
					}).To(PanicWith("something terrible happened"))
					Expect(mockTransaction.commitCalled).To(BeFalse())
					Expect(mockTransaction.rollbackCalled).To(BeTrue())
				})

				It("should return an error if commit fails", func() {
					mockTransaction.commitErr = errors.New("commit failed")
					err := dm.executeWithCleanup(ctx, mockTransaction, ctx, func(txCtx context.Context) error {
						return nil
					})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to commit transaction"))
					Expect(mockTransaction.commitCalled).To(BeTrue())
					Expect(mockTransaction.rollbackCalled).To(BeFalse())
				})
			})

			Describe("isRetryableError", func() {
				It("should identify retryable errors by substring", func() {
					Expect(dm.isRetryableError(errors.New("prefix deadlock detected suffix"))).To(BeTrue())
					Expect(dm.isRetryableError(errors.New("serialization failure"))).To(BeTrue())
					Expect(dm.isRetryableError(errors.New("connection refused"))).To(BeTrue())
					Expect(dm.isRetryableError(errors.New("connection reset"))).To(BeTrue())
					Expect(dm.isRetryableError(errors.New("this is a timeout error"))).To(BeTrue())
				})

				It("should not identify non-retryable errors", func() {
					Expect(dm.isRetryableError(errors.New("invalid syntax for type integer"))).To(BeFalse())
					Expect(dm.isRetryableError(errors.New("null value in column violates not-null constraint"))).To(BeFalse())
				})
			})
		})
	})

	Describe("getCurrentTime", func() {
		It("should return the current time", func() {
			now := time.Now()
			currentTime := getCurrentTime()

			// Check if the returned time is within a reasonable delta of 'now'
			Expect(currentTime).To(BeTemporally("~", now, time.Second), "getCurrentTime should be very close to time.Now()")
		})
	})

})
