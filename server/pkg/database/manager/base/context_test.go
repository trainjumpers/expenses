package base_test

import (
	"context"
	"time"

	"expenses/pkg/database/manager/base"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transaction Context", func() {
	var (
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	Context("when a transaction context is not present", func() {
		It("should not find a transaction context", func() {
			txCtx, ok := base.GetTransactionContext(ctx)
			Expect(ok).To(BeFalse())
			Expect(txCtx).To(BeNil())
		})

		It("should report that it is not in a transaction", func() {
			inTx := base.IsInTransaction(ctx)
			Expect(inTx).To(BeFalse())
		})

		It("should return an empty transaction ID", func() {
			txID := base.GetTransactionID(ctx)
			Expect(txID).To(BeEmpty())
		})
	})

	Context("when a transaction context is present", func() {
		var (
			parentCtx context.Context
			txCtx     *base.TransactionContext
		)

		BeforeEach(func() {
			txCtx = &base.TransactionContext{
				Tx:        nil, // We don't need a real transaction for this test
				StartTime: time.Now(),
				ID:        "test-tx-id-123",
				ReadOnly:  false,
				Nested:    false,
				LockKey:   nil,
			}
			parentCtx = base.WithTransactionContext(context.Background(), txCtx)
		})

		It("should find and return the transaction context", func() {
			retrievedTxCtx, ok := base.GetTransactionContext(parentCtx)
			Expect(ok).To(BeTrue())
			Expect(retrievedTxCtx).To(Equal(txCtx))
			Expect(retrievedTxCtx.ID).To(Equal("test-tx-id-123"))
		})

		It("should report that it is in a transaction", func() {
			inTx := base.IsInTransaction(parentCtx)
			Expect(inTx).To(BeTrue())
		})

		It("should return the correct transaction ID", func() {
			txID := base.GetTransactionID(parentCtx)
			Expect(txID).To(Equal("test-tx-id-123"))
		})
	})
})
