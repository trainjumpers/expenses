package service

import (
	repository "expenses/internal/mock/repository"
	"expenses/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TransactionService", func() {
	var (
		transactionService TransactionServiceInterface
		mockRepo           *repository.MockTransactionRepository
		categoryMockRepo   *repository.MockCategoryRepository
		accountMockRepo    *repository.MockAccountRepository
		ctx                *gin.Context
		testDate           time.Time
	)

	BeforeEach(func() {
		ctx = &gin.Context{}
		mockRepo = repository.NewMockTransactionRepository()
		categoryMockRepo = repository.NewMockCategoryRepository()
		accountMockRepo = repository.NewMockAccountRepository()
		transactionService = NewTransactionService(mockRepo, categoryMockRepo, accountMockRepo)
		testDate, _ = time.Parse("2006-01-02", "2023-01-01")
	})

	Describe("CreateTransaction", func() {
		var cat1 models.CategoryResponse
		var acc1 models.AccountResponse

		BeforeEach(func() {
			cat1, _ = categoryMockRepo.CreateCategory(ctx, models.CreateCategoryInput{Name: "Food", CreatedBy: 1})
			acc1, _ = accountMockRepo.CreateAccount(ctx, models.CreateAccountInput{Name: "HDFC", BankType: "hdfc", Currency: "inr", CreatedBy: 1})
		})

		It("should create a new transaction successfully", func() {
			amount := 150.0
			input := models.CreateTransactionInput{
				CreateBaseTransactionInput: models.CreateBaseTransactionInput{
					Name:        "Lunch",
					Description: "Bought lunch",
					Amount:      &amount,
					Date:        testDate,
					CreatedBy:   1,
					AccountId:   acc1.Id,
				},
				CategoryIds: []int64{cat1.Id},
			}

			resp, err := transactionService.CreateTransaction(ctx, input)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Name).To(Equal("Lunch"))
			Expect(resp.CategoryIds).To(ContainElement(cat1.Id))
			Expect(resp.AccountId).To(Equal(acc1.Id))
		})

		It("should fail if category does not exist", func() {
			amount := 150.0
			input := models.CreateTransactionInput{
				CreateBaseTransactionInput: models.CreateBaseTransactionInput{
					Name:      "Lunch",
					Amount:    &amount,
					Date:      testDate,
					CreatedBy: 1,
					AccountId: acc1.Id,
				},
				CategoryIds: []int64{999}, // Non-existent
			}
			_, err := transactionService.CreateTransaction(ctx, input)
			Expect(err).To(HaveOccurred())
		})

		It("should fail if account does not exist", func() {
			amount := 150.0
			input := models.CreateTransactionInput{
				CreateBaseTransactionInput: models.CreateBaseTransactionInput{
					Name:      "Lunch",
					Amount:    &amount,
					Date:      testDate,
					CreatedBy: 1,
					AccountId: 999, // Non-existent
				},
				CategoryIds: []int64{cat1.Id},
			}
			_, err := transactionService.CreateTransaction(ctx, input)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("UpdateTransaction", func() {
		var createdTx models.TransactionResponse
		var cat1, cat2 models.CategoryResponse
		var acc1, acc2 models.AccountResponse

		BeforeEach(func() {
			cat1, _ = categoryMockRepo.CreateCategory(ctx, models.CreateCategoryInput{Name: "Groceries", CreatedBy: 1})
			cat2, _ = categoryMockRepo.CreateCategory(ctx, models.CreateCategoryInput{Name: "Entertainment", CreatedBy: 1})
			acc1, _ = accountMockRepo.CreateAccount(ctx, models.CreateAccountInput{Name: "ICICI", BankType: "icici", Currency: "inr", CreatedBy: 1})
			acc2, _ = accountMockRepo.CreateAccount(ctx, models.CreateAccountInput{Name: "Axis", BankType: "axis", Currency: "inr", CreatedBy: 1})
			amount := 500.0
			input := models.CreateTransactionInput{
				CreateBaseTransactionInput: models.CreateBaseTransactionInput{
					Name:      "Initial",
					Amount:    &amount,
					Date:      testDate,
					CreatedBy: 1,
					AccountId: acc1.Id,
				},
				CategoryIds: []int64{cat1.Id},
			}
			createdTx, _ = transactionService.CreateTransaction(ctx, input)
		})

		It("should update transaction details and mappings", func() {
			newName := "Updated Transaction"
			newAmount := 1000.0
			updateInput := models.UpdateTransactionInput{
				UpdateBaseTransactionInput: models.UpdateBaseTransactionInput{
					Name:      newName,
					Amount:    &newAmount,
					AccountId: &acc2.Id,
				},
				CategoryIds: &[]int64{cat2.Id},
			}
			resp, err := transactionService.UpdateTransaction(ctx, createdTx.Id, 1, updateInput)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Name).To(Equal(newName))
			Expect(resp.Amount).To(Equal(newAmount))
			Expect(resp.CategoryIds).To(ContainElement(cat2.Id))
			Expect(resp.AccountId).To(Equal(acc2.Id))
		})

		It("should fail to update with non-existent account", func() {
			accountId := int64(999)
			updateInput := models.UpdateTransactionInput{
				UpdateBaseTransactionInput: models.UpdateBaseTransactionInput{
					AccountId: &accountId,
				},
			}
			_, err := transactionService.UpdateTransaction(ctx, createdTx.Id, 1, updateInput)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GetTransactionById", func() {
		var createdTx models.TransactionResponse
		BeforeEach(func() {
			cat, _ := categoryMockRepo.CreateCategory(ctx, models.CreateCategoryInput{Name: "Test", CreatedBy: 5})
			acc, _ := accountMockRepo.CreateAccount(ctx, models.CreateAccountInput{Name: "Test", BankType: "sbi", Currency: "inr", CreatedBy: 5})
			amount := 123.0
			input := models.CreateTransactionInput{
				CreateBaseTransactionInput: models.CreateBaseTransactionInput{
					Name:      "ForGet",
					Amount:    &amount,
					Date:      testDate,
					CreatedBy: 5,
					AccountId: acc.Id,
				},
				CategoryIds: []int64{cat.Id},
			}
			createdTx, _ = transactionService.CreateTransaction(ctx, input)
		})

		It("should get a transaction by its ID", func() {
			resp, err := transactionService.GetTransactionById(ctx, createdTx.Id, 5)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Id).To(Equal(createdTx.Id))
			Expect(resp.Name).To(Equal("ForGet"))
		})

		It("should fail for a non-existent ID", func() {
			_, err := transactionService.GetTransactionById(ctx, 999, 5)
			Expect(err).To(HaveOccurred())
		})

		It("should fail if user ID does not match", func() {
			_, err := transactionService.GetTransactionById(ctx, createdTx.Id, 999)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("DeleteTransaction", func() {
		var createdTx models.TransactionResponse
		var cat1 models.CategoryResponse
		var acc1 models.AccountResponse

		BeforeEach(func() {
			cat1, _ = categoryMockRepo.CreateCategory(ctx, models.CreateCategoryInput{Name: "DeleteTest", CreatedBy: 3})
			acc1, _ = accountMockRepo.CreateAccount(ctx, models.CreateAccountInput{Name: "DeleteAccount", BankType: "sbi", Currency: "inr", CreatedBy: 3})
			amount := 200.0
			input := models.CreateTransactionInput{
				CreateBaseTransactionInput: models.CreateBaseTransactionInput{
					Name:        "Transaction to Delete",
					Description: "Will be deleted",
					Amount:      &amount,
					Date:        testDate,
					CreatedBy:   3,
					AccountId:   acc1.Id,
				},
				CategoryIds: []int64{cat1.Id},
			}
			createdTx, _ = transactionService.CreateTransaction(ctx, input)
		})

		It("should delete a transaction successfully", func() {
			err := transactionService.DeleteTransaction(ctx, createdTx.Id, 3)
			Expect(err).NotTo(HaveOccurred())

			// Verify transaction is deleted by trying to get it
			_, err = transactionService.GetTransactionById(ctx, createdTx.Id, 3)
			Expect(err).To(HaveOccurred())
		})

		It("should fail to delete a non-existent transaction", func() {
			err := transactionService.DeleteTransaction(ctx, 999, 3)
			Expect(err).To(HaveOccurred())
		})

		It("should fail to delete transaction of different user", func() {
			err := transactionService.DeleteTransaction(ctx, createdTx.Id, 999)
			Expect(err).To(HaveOccurred())
		})

		It("should fail to delete already deleted transaction", func() {
			// Delete once
			err := transactionService.DeleteTransaction(ctx, createdTx.Id, 3)
			Expect(err).NotTo(HaveOccurred())

			// Try to delete again
			err = transactionService.DeleteTransaction(ctx, createdTx.Id, 3)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("UpdateTransaction - Additional Edge Cases", func() {
		var createdTx models.TransactionResponse
		var cat1, cat2 models.CategoryResponse
		var acc1, acc2 models.AccountResponse

		BeforeEach(func() {
			cat1, _ = categoryMockRepo.CreateCategory(ctx, models.CreateCategoryInput{Name: "OriginalCategory", CreatedBy: 4})
			cat2, _ = categoryMockRepo.CreateCategory(ctx, models.CreateCategoryInput{Name: "UpdatedCategory", CreatedBy: 4})
			acc1, _ = accountMockRepo.CreateAccount(ctx, models.CreateAccountInput{Name: "OriginalAccount", BankType: "hdfc", Currency: "inr", CreatedBy: 4})
			acc2, _ = accountMockRepo.CreateAccount(ctx, models.CreateAccountInput{Name: "UpdatedAccount", BankType: "axis", Currency: "inr", CreatedBy: 4})
			amount := 300.0
			input := models.CreateTransactionInput{
				CreateBaseTransactionInput: models.CreateBaseTransactionInput{
					Name:        "Original Transaction",
					Description: "Original description",
					Amount:      &amount,
					Date:        testDate,
					CreatedBy:   4,
					AccountId:   acc1.Id,
				},
				CategoryIds: []int64{cat1.Id},
			}
			createdTx, _ = transactionService.CreateTransaction(ctx, input)
		})

		It("should fail to update transaction with non-existent category", func() {
			nonExistentCategoryIds := []int64{999}
			updateInput := models.UpdateTransactionInput{
				CategoryIds: &nonExistentCategoryIds,
			}
			_, err := transactionService.UpdateTransaction(ctx, createdTx.Id, 4, updateInput)
			Expect(err).To(HaveOccurred())
		})

		It("should fail to update non-existent transaction", func() {
			updateInput := models.UpdateTransactionInput{
				UpdateBaseTransactionInput: models.UpdateBaseTransactionInput{
					Name: "Updated Name",
				},
			}
			_, err := transactionService.UpdateTransaction(ctx, 999, 4, updateInput)
			Expect(err).To(HaveOccurred())
		})

		It("should fail to update transaction of different user", func() {
			updateInput := models.UpdateTransactionInput{
				UpdateBaseTransactionInput: models.UpdateBaseTransactionInput{
					Name: "Updated Name",
				},
			}
			_, err := transactionService.UpdateTransaction(ctx, createdTx.Id, 999, updateInput)
			Expect(err).To(HaveOccurred())
		})

		It("should update only category mappings without base transaction fields", func() {
			newCategoryIds := []int64{cat2.Id}
			updateInput := models.UpdateTransactionInput{
				CategoryIds: &newCategoryIds,
			}
			resp, err := transactionService.UpdateTransaction(ctx, createdTx.Id, 4, updateInput)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.CategoryIds).To(ContainElement(cat2.Id))
			Expect(resp.CategoryIds).NotTo(ContainElement(cat1.Id))
			// Base fields should remain unchanged
			Expect(resp.Name).To(Equal("Original Transaction"))
		})

		It("should clear category mappings when empty array is provided", func() {
			emptyCategoryIds := []int64{}
			updateInput := models.UpdateTransactionInput{
				CategoryIds: &emptyCategoryIds,
			}
			resp, err := transactionService.UpdateTransaction(ctx, createdTx.Id, 4, updateInput)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.CategoryIds).To(BeEmpty())
		})

		It("should update both base fields and category mappings", func() {
			newName := "Completely Updated Transaction"
			newAmount := 500.0
			newCategoryIds := []int64{cat2.Id}
			updateInput := models.UpdateTransactionInput{
				UpdateBaseTransactionInput: models.UpdateBaseTransactionInput{
					Name:      newName,
					Amount:    &newAmount,
					AccountId: &acc2.Id,
				},
				CategoryIds: &newCategoryIds,
			}
			resp, err := transactionService.UpdateTransaction(ctx, createdTx.Id, 4, updateInput)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Name).To(Equal(newName))
			Expect(resp.Amount).To(Equal(newAmount))
			Expect(resp.AccountId).To(Equal(acc2.Id))
			Expect(resp.CategoryIds).To(ContainElement(cat2.Id))
			Expect(resp.CategoryIds).NotTo(ContainElement(cat1.Id))
		})

		It("should handle partial updates with nil description", func() {
			updateInput := models.UpdateTransactionInput{
				UpdateBaseTransactionInput: models.UpdateBaseTransactionInput{
					Name:        "Updated Name Only",
					Description: nil, // Explicitly setting to nil
				},
			}
			resp, err := transactionService.UpdateTransaction(ctx, createdTx.Id, 4, updateInput)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Name).To(Equal("Updated Name Only"))
		})

		It("should update description successfully", func() {
			newDescription := "This is a completely new description"
			updateInput := models.UpdateTransactionInput{
				UpdateBaseTransactionInput: models.UpdateBaseTransactionInput{
					Description: &newDescription,
				},
			}
			resp, err := transactionService.UpdateTransaction(ctx, createdTx.Id, 4, updateInput)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Description).NotTo(BeNil())
			Expect(*resp.Description).To(Equal(newDescription))
			// Other fields should remain unchanged
			Expect(resp.Name).To(Equal("Original Transaction"))
			Expect(resp.Amount).To(Equal(300.0))
		})

		It("should update date successfully", func() {
			newDate, _ := time.Parse("2006-01-02", "2023-02-15")
			updateInput := models.UpdateTransactionInput{
				UpdateBaseTransactionInput: models.UpdateBaseTransactionInput{
					Date: newDate,
				},
			}
			resp, err := transactionService.UpdateTransaction(ctx, createdTx.Id, 4, updateInput)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Date.Format("2006-01-02")).To(Equal("2023-02-15"))
			// Other fields should remain unchanged
			Expect(resp.Name).To(Equal("Original Transaction"))
			Expect(resp.Amount).To(Equal(300.0))
		})

		It("should fail when updating to future date", func() {
			futureDate := time.Now().AddDate(0, 0, 1) // Tomorrow
			updateInput := models.UpdateTransactionInput{
				UpdateBaseTransactionInput: models.UpdateBaseTransactionInput{
					Date: futureDate,
				},
			}
			_, err := transactionService.UpdateTransaction(ctx, createdTx.Id, 4, updateInput)
			Expect(err).To(HaveOccurred())
			// Should be a date validation error
		})

		It("should fail when update creates duplicate transaction", func() {
			// First create another transaction that we'll try to duplicate
			amount := 300.0
			duplicateInput := models.CreateTransactionInput{
				CreateBaseTransactionInput: models.CreateBaseTransactionInput{
					Name:        "Duplicate Target",
					Description: "Duplicate description",
					Amount:      &amount,
					Date:        testDate,
					CreatedBy:   4,
					AccountId:   acc1.Id,
				},
				CategoryIds: []int64{cat1.Id},
			}
			_, err := transactionService.CreateTransaction(ctx, duplicateInput)
			Expect(err).NotTo(HaveOccurred())

			// Now try to update the original transaction to match the duplicate
			duplicateDescription := "Duplicate description"
			updateInput := models.UpdateTransactionInput{
				UpdateBaseTransactionInput: models.UpdateBaseTransactionInput{
					Name:        "Duplicate Target",
					Description: &duplicateDescription,
					// Amount and date remain the same (300.0 and testDate)
				},
			}
			_, err = transactionService.UpdateTransaction(ctx, createdTx.Id, 4, updateInput)
			Expect(err).To(HaveOccurred())
			// Should be a transaction already exists error
		})

		It("should succeed when updating to duplicate values for different user", func() {
			// Create category and account for different user
			cat7, _ := categoryMockRepo.CreateCategory(ctx, models.CreateCategoryInput{Name: "User7Category", CreatedBy: 7})
			acc7, _ := accountMockRepo.CreateAccount(ctx, models.CreateAccountInput{Name: "User7Account", BankType: "sbi", Currency: "inr", CreatedBy: 7})

			// Create a transaction for a different user with same details
			amount := 300.0
			userTwoInput := models.CreateTransactionInput{
				CreateBaseTransactionInput: models.CreateBaseTransactionInput{
					Name:        "Cross User Transaction",
					Description: "User 2 description",
					Amount:      &amount,
					Date:        testDate,
					CreatedBy:   7, // Different user
					AccountId:   acc7.Id,
				},
				CategoryIds: []int64{cat7.Id},
			}
			_, err := transactionService.CreateTransaction(ctx, userTwoInput)
			Expect(err).NotTo(HaveOccurred())

			// Now update our transaction to have same name/description/amount/date as the other user's
			// This should succeed because it's a different user
			userTwoDescription := "User 2 description"
			updateInput := models.UpdateTransactionInput{
				UpdateBaseTransactionInput: models.UpdateBaseTransactionInput{
					Name:        "Cross User Transaction",
					Description: &userTwoDescription,
					// Amount and date remain the same
				},
			}
			resp, err := transactionService.UpdateTransaction(ctx, createdTx.Id, 4, updateInput)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Name).To(Equal("Cross User Transaction"))
			Expect(*resp.Description).To(Equal("User 2 description"))
		})

		It("should succeed when updating one field breaks potential duplicate", func() {
			// Create a transaction that would be duplicate except for amount
			differentAmount := 500.0
			nearDuplicateInput := models.CreateTransactionInput{
				CreateBaseTransactionInput: models.CreateBaseTransactionInput{
					Name:        "Original Transaction",
					Description: "Original description",
					Amount:      &differentAmount, // Different amount
					Date:        testDate,
					CreatedBy:   4,
					AccountId:   acc1.Id,
				},
				CategoryIds: []int64{cat1.Id},
			}
			_, err := transactionService.CreateTransaction(ctx, nearDuplicateInput)
			Expect(err).NotTo(HaveOccurred())

			// Now update our original transaction - this should succeed because the amounts are different
			newDescription := "Updated to be similar but not duplicate"
			updateInput := models.UpdateTransactionInput{
				UpdateBaseTransactionInput: models.UpdateBaseTransactionInput{
					Description: &newDescription,
					// Name, amount, and date make it non-duplicate
				},
			}
			resp, err := transactionService.UpdateTransaction(ctx, createdTx.Id, 4, updateInput)
			Expect(err).NotTo(HaveOccurred())
			Expect(*resp.Description).To(Equal("Updated to be similar but not duplicate"))
		})
	})

	Describe("ListTransactions", func() {
		var cat1 models.CategoryResponse
		var acc1 models.AccountResponse

		BeforeEach(func() {
			cat1, _ = categoryMockRepo.CreateCategory(ctx, models.CreateCategoryInput{Name: "ListTest", CreatedBy: 6})
			acc1, _ = accountMockRepo.CreateAccount(ctx, models.CreateAccountInput{Name: "ListAccount", BankType: "icici", Currency: "inr", CreatedBy: 6})
		})

		It("should list transactions for a user", func() {
			// Create multiple transactions
			amount1 := 100.0
			amount2 := 200.0
			input1 := models.CreateTransactionInput{
				CreateBaseTransactionInput: models.CreateBaseTransactionInput{
					Name:      "Transaction 1",
					Amount:    &amount1,
					Date:      testDate,
					CreatedBy: 6,
					AccountId: acc1.Id,
				},
				CategoryIds: []int64{cat1.Id},
			}
			input2 := models.CreateTransactionInput{
				CreateBaseTransactionInput: models.CreateBaseTransactionInput{
					Name:      "Transaction 2",
					Amount:    &amount2,
					Date:      testDate,
					CreatedBy: 6,
					AccountId: acc1.Id,
				},
				CategoryIds: []int64{cat1.Id},
			}

			_, err := transactionService.CreateTransaction(ctx, input1)
			Expect(err).NotTo(HaveOccurred())
			_, err = transactionService.CreateTransaction(ctx, input2)
			Expect(err).NotTo(HaveOccurred())

			transactions, err := transactionService.ListTransactions(ctx, 6)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(transactions)).To(Equal(2))
		})

		It("should return error when listing transactions for user with no transactions", func() {
			_, err := transactionService.ListTransactions(ctx, 999)
			Expect(err).To(HaveOccurred())
		})
	})
})
