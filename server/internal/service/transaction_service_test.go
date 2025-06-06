package service

import (
	mock "expenses/internal/mock/repository"
	"expenses/internal/models"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gin-gonic/gin"
)

var _ = Describe("TransactionService", func() {
	var (
		transactionService TransactionServiceInterface
		mockRepo           *mock.MockTransactionRepository
		ctx                *gin.Context
		testDate           time.Time
		futureDate         time.Time
	)

	BeforeEach(func() {
		ctx = &gin.Context{}
		mockRepo = mock.NewMockTransactionRepository()
		transactionService = NewTransactionService(mockRepo)
		testDate, _ = time.Parse("2006-01-02", "2023-01-01")
		futureDate = time.Now().AddDate(0, 0, 1) // Tomorrow
	})

	Describe("CreateTransaction", func() {
		It("should create a new transaction successfully", func() {
			input := models.CreateTransactionInput{
				Name:        "Test Transaction",
				Description: "Test Description",
				Amount:      floatPtr(100.50),
				Date:        testDate,
				CreatedBy:   1,
			}
			tx, err := transactionService.CreateTransaction(ctx, input)
			Expect(err).NotTo(HaveOccurred())
			Expect(tx.Name).To(Equal(input.Name))
			Expect(tx.Description).ToNot(BeNil())
			Expect(*tx.Description).To(Equal(input.Description))
		})

		It("should create a new transaction without description", func() {
			input := models.CreateTransactionInput{
				Name:      "Test Transaction 2",
				Amount:    floatPtr(75.25),
				Date:      testDate,
				CreatedBy: 1,
			}
			tx, err := transactionService.CreateTransaction(ctx, input)
			Expect(err).NotTo(HaveOccurred())
			Expect(tx.Name).To(Equal(input.Name))
		})

		It("should return error for duplicate transaction", func() {
			input := models.CreateTransactionInput{
				Name:        "Duplicate Transaction",
				Description: "Duplicate Description",
				Amount:      floatPtr(100.00),
				Date:        testDate,
				CreatedBy:   1,
			}

			// Create first transaction
			_, err := transactionService.CreateTransaction(ctx, input)
			Expect(err).NotTo(HaveOccurred())

			// Try to create identical transaction
			_, err = transactionService.CreateTransaction(ctx, input)
			Expect(err).To(HaveOccurred())
		})

		It("should allow transactions with different fields", func() {
			input1 := models.CreateTransactionInput{
				Name:        "Transaction 1",
				Description: "Description 1",
				Amount:      floatPtr(100.00),
				Date:        testDate,
				CreatedBy:   1,
			}
			input2 := models.CreateTransactionInput{
				Name:        "Transaction 2",
				Description: "Description 2",
				Amount:      floatPtr(200.00),
				Date:        testDate,
				CreatedBy:   1,
			}

			tx1, err1 := transactionService.CreateTransaction(ctx, input1)
			tx2, err2 := transactionService.CreateTransaction(ctx, input2)

			Expect(err1).NotTo(HaveOccurred())
			Expect(err2).NotTo(HaveOccurred())
			Expect(tx1.Id).NotTo(Equal(tx2.Id))
		})

		It("should fail for future date", func() {
			input := models.CreateTransactionInput{
				Name:        "Test Transaction with incorrect date",
				Description: "Test Description",
				Amount:      floatPtr(100.50),
				Date:        futureDate,
				CreatedBy:   1,
			}
			_, err := transactionService.CreateTransaction(ctx, input)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GetTransactionById", func() {
		var created models.TransactionResponse
		BeforeEach(func() {
			input := models.CreateTransactionInput{
				Name:        "Transaction Get",
				Description: "Get Description",
				Amount:      floatPtr(150.75),
				Date:        testDate,
				CreatedBy:   2,
			}
			var err error
			created, err = transactionService.CreateTransaction(ctx, input)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should get transaction by id", func() {
			tx, err := transactionService.GetTransactionById(ctx, created.Id, 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(tx.Name).To(Equal("Transaction Get"))
			Expect(tx.Amount).To(Equal(150.75))
		})

		It("should return error for non-existent id", func() {
			_, err := transactionService.GetTransactionById(ctx, 9999, 2)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for non-existent user id", func() {
			_, err := transactionService.GetTransactionById(ctx, created.Id, 9999)
			Expect(err).To(HaveOccurred())
		})

		It("should return error while accessing transaction of other user", func() {
			_, err := transactionService.GetTransactionById(ctx, created.Id, 3)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("UpdateTransaction", func() {
		var created models.TransactionResponse
		BeforeEach(func() {
			input := models.CreateTransactionInput{
				Name:        "Transaction Update",
				Description: "Update Description",
				Amount:      floatPtr(200.00),
				Date:        testDate,
				CreatedBy:   3,
			}
			var err error
			created, err = transactionService.CreateTransaction(ctx, input)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should update transaction name", func() {
			update := models.UpdateTransactionInput{Name: "Updated Name"}
			tx, err := transactionService.UpdateTransaction(ctx, created.Id, 3, update)
			Expect(err).NotTo(HaveOccurred())
			Expect(tx.Name).To(Equal("Updated Name"))
		})

		It("should update transaction amount", func() {
			amount := 300.50
			update := models.UpdateTransactionInput{Amount: &amount}
			tx, err := transactionService.UpdateTransaction(ctx, created.Id, 3, update)
			Expect(err).NotTo(HaveOccurred())
			Expect(tx.Amount).To(Equal(300.50))
		})

		It("should update transaction description", func() {
			description := "Updated Description"
			update := models.UpdateTransactionInput{Description: &description}
			tx, err := transactionService.UpdateTransaction(ctx, created.Id, 3, update)
			Expect(err).NotTo(HaveOccurred())
			Expect(tx.Description).ToNot(BeNil())
			Expect(*tx.Description).To(Equal("Updated Description"))
		})

		It("should update transaction date", func() {
			newDate, _ := time.Parse("2006-01-02", "2023-02-01")
			update := models.UpdateTransactionInput{Date: newDate}
			tx, err := transactionService.UpdateTransaction(ctx, created.Id, 3, update)
			Expect(err).NotTo(HaveOccurred())
			Expect(tx.Date.Format("2006-01-02")).To(Equal("2023-02-01"))
		})

		It("should return error for duplicate after update", func() {
			// Create another transaction
			input2 := models.CreateTransactionInput{
				Name:        "Another Transaction",
				Description: "Another Description",
				Amount:      floatPtr(500.00),
				Date:        testDate,
				CreatedBy:   3,
			}
			created2, err := transactionService.CreateTransaction(ctx, input2)
			Expect(err).NotTo(HaveOccurred())

			// Try to update created2 to be identical to created
			update := models.UpdateTransactionInput{
				Name:        "Transaction Update",
				Description: stringPtr("Update Description"),
				Amount:      floatPtr(200.00),
			}
			_, err = transactionService.UpdateTransaction(ctx, created2.Id, 3, update)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for non-existent id", func() {
			update := models.UpdateTransactionInput{Name: "Updated Name"}
			_, err := transactionService.UpdateTransaction(ctx, 9999, 3, update)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for non-existent user id", func() {
			update := models.UpdateTransactionInput{Name: "Updated Name"}
			_, err := transactionService.UpdateTransaction(ctx, created.Id, 9999, update)
			Expect(err).To(HaveOccurred())
		})

		It("should return error while updating transaction of other user", func() {
			update := models.UpdateTransactionInput{Name: "Updated Name"}
			_, err := transactionService.UpdateTransaction(ctx, created.Id, 4, update)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("DeleteTransaction", func() {
		var created models.TransactionResponse
		BeforeEach(func() {
			input := models.CreateTransactionInput{
				Name:        "Transaction Delete",
				Description: "Delete Description",
				Amount:      floatPtr(250.00),
				Date:        testDate,
				CreatedBy:   4,
			}
			var err error
			created, err = transactionService.CreateTransaction(ctx, input)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should delete transaction by id", func() {
			err := transactionService.DeleteTransaction(ctx, created.Id, 4)
			Expect(err).NotTo(HaveOccurred())
			_, err = transactionService.GetTransactionById(ctx, created.Id, 4)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for non-existent id", func() {
			err := transactionService.DeleteTransaction(ctx, 9999, 4)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for non-existent user id", func() {
			err := transactionService.DeleteTransaction(ctx, created.Id, 9999)
			Expect(err).To(HaveOccurred())
		})

		It("should return error while deleting transaction of other user", func() {
			err := transactionService.DeleteTransaction(ctx, created.Id, 3)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ListTransactions", func() {
		BeforeEach(func() {
			for i := 0; i < 3; i++ {
				input := models.CreateTransactionInput{
					Name:        "Transaction List " + string(rune('A'+i)),
					Description: "List Description " + string(rune('A'+i)),
					Amount:      floatPtr(float64(100 + i*50)),
					Date:        testDate.AddDate(0, 0, i), // Different dates to avoid duplicates
					CreatedBy:   5,
				}
				_, err := transactionService.CreateTransaction(ctx, input)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("should list all transactions", func() {
			transactions, err := transactionService.ListTransactions(ctx, 5)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(transactions)).To(BeNumerically(">=", 3))
		})

		It("should return error for non-existent user id", func() {
			_, err := transactionService.ListTransactions(ctx, 9999)
			Expect(err).To(HaveOccurred())
		})

		It("should return error while listing transactions of other user", func() {
			_, err := transactionService.ListTransactions(ctx, 4)
			Expect(err).To(HaveOccurred())
		})
	})
})

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}
