package base_test

import (
	"time"

	"expenses/pkg/database/manager/base"

	"github.com/jackc/pgx/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transaction Options", func() {

	Describe("DefaultTransactionOptions", func() {
		It("should return options with sensible defaults for general use", func() {
			opts := base.DefaultTransactionOptions()

			Expect(opts).NotTo(BeNil())
			Expect(opts.IsolationLevel).To(Equal(pgx.ReadCommitted))
			Expect(opts.AccessMode).To(Equal(pgx.ReadWrite))
			Expect(opts.DeferrableMode).To(Equal(pgx.NotDeferrable))
			Expect(opts.Timeout).To(Equal(30 * time.Second))
			Expect(opts.ReadOnly).To(BeFalse())

			Expect(opts.RetryPolicy).NotTo(BeNil())
			Expect(opts.RetryPolicy.MaxRetries).To(Equal(3))
			Expect(opts.RetryPolicy.BaseDelay).To(Equal(100 * time.Millisecond))
			Expect(opts.RetryPolicy.MaxDelay).To(Equal(5 * time.Second))
			Expect(opts.RetryPolicy.Backoff).To(Equal(base.BackoffExponential))
		})
	})

	Describe("ReadOnlyTransactionOptions", func() {
		It("should return options configured for read-only operations", func() {
			opts := base.ReadOnlyTransactionOptions()

			Expect(opts).NotTo(BeNil())
			Expect(opts.IsolationLevel).To(Equal(pgx.ReadCommitted)) // Usually same as default
			Expect(opts.AccessMode).To(Equal(pgx.ReadOnly))
			Expect(opts.ReadOnly).To(BeTrue())
			Expect(opts.Timeout).To(Equal(10 * time.Second)) // Shorter timeout for reads
		})
	})

	Describe("LongRunningTransactionOptions", func() {
		It("should return options configured for long-running tasks", func() {
			opts := base.LongRunningTransactionOptions()

			Expect(opts).NotTo(BeNil())
			Expect(opts.Timeout).To(Equal(5 * time.Minute))
			Expect(opts.RetryPolicy).NotTo(BeNil())
			Expect(opts.RetryPolicy.MaxRetries).To(Equal(1)) // Less aggressive retry
		})
	})
})
