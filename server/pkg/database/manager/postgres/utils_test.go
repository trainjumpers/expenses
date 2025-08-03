package postgres

import (
	"encoding/hex"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

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
