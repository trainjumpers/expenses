package service

import (
	mockDatabase "expenses/internal/mock/database"
	repository "expenses/internal/mock/repository"
	"expenses/internal/models"
	"strings"
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
		mockDB             *mockDatabase.MockDatabaseManager
		ctx                *gin.Context
		testDate           time.Time
		cat1               models.CategoryResponse
		cat2               models.CategoryResponse
		cat3               models.CategoryResponse
		acc1               models.AccountResponse
		acc2               models.AccountResponse
		userId             int64
	)

	BeforeEach(func() {
		ctx = &gin.Context{}
		mockRepo = repository.NewMockTransactionRepository()
		categoryMockRepo = repository.NewMockCategoryRepository()
		accountMockRepo = repository.NewMockAccountRepository()
		mockDB = mockDatabase.NewMockDatabaseManager()
		transactionService = NewTransactionService(mockRepo, categoryMockRepo, accountMockRepo, mockDB)
		testDate, _ = time.Parse("2006-01-02", "2023-01-01")
		userId = 1

		// Create test categories and accounts
		var err error
		cat1, err = categoryMockRepo.CreateCategory(ctx, models.CreateCategoryInput{Name: "Food", CreatedBy: userId})
		Expect(err).NotTo(HaveOccurred())
		Expect(cat1.Id).To(BeNumerically(">", 0))
		cat2, err = categoryMockRepo.CreateCategory(ctx, models.CreateCategoryInput{Name: "Entertainment", CreatedBy: userId})
		Expect(err).NotTo(HaveOccurred())
		Expect(cat2.Id).To(BeNumerically(">", 0))
		cat3, err = categoryMockRepo.CreateCategory(ctx, models.CreateCategoryInput{Name: "Bills", CreatedBy: userId})
		Expect(err).NotTo(HaveOccurred())
		Expect(cat3.Id).To(BeNumerically(">", 0))
		acc1, err = accountMockRepo.CreateAccount(ctx, models.CreateAccountInput{Name: "HDFC", BankType: "hdfc", Currency: "inr", CreatedBy: userId})
		Expect(err).NotTo(HaveOccurred())
		Expect(acc1.Id).To(BeNumerically(">", 0))
		acc2, err = accountMockRepo.CreateAccount(ctx, models.CreateAccountInput{Name: "Chase", BankType: "chase", Currency: "usd", CreatedBy: userId})
		Expect(err).NotTo(HaveOccurred())
		Expect(acc2.Id).To(BeNumerically(">", 0))

		// Create test transactions
		amount1 := 100.0
		mockRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
			Name:        "Groceries",
			Description: "Weekly groceries",
			Amount:      &amount1,
			Date:        testDate,
			CreatedBy:   userId,
			AccountId:   acc1.Id,
		}, []int64{cat1.Id})

		amount2 := 50.0
		mockRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
			Name:        "Movie tickets",
			Description: "Weekend movie",
			Amount:      &amount2,
			Date:        testDate.AddDate(0, 0, 1),
			CreatedBy:   userId,
			AccountId:   acc2.Id,
		}, []int64{cat2.Id})

		amount3 := 200.0
		mockRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
			Name:        "Electricity bill",
			Description: "Monthly bill",
			Amount:      &amount3,
			Date:        testDate.AddDate(0, 0, 2),
			CreatedBy:   userId,
			AccountId:   acc1.Id,
		}, []int64{cat3.Id})

		amount4 := 75.0
		mockRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
			Name:        "Restaurant",
			Description: "Dinner with friends",
			Amount:      &amount4,
			Date:        testDate.AddDate(0, 0, 3),
			CreatedBy:   userId,
			AccountId:   acc2.Id,
		}, []int64{cat1.Id, cat2.Id})

		// Create an uncategorized transaction for testing
		amount5 := 25.0
		mockRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
			Name:        "Cash withdrawal",
			Description: "ATM withdrawal",
			Amount:      &amount5,
			Date:        testDate.AddDate(0, 0, 4),
			CreatedBy:   userId,
			AccountId:   acc1.Id,
		}, []int64{}) // No categories
	})

	Describe("CreateTransaction", func() {
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

		BeforeEach(func() {
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
			resp, err := transactionService.UpdateTransaction(ctx, createdTx.Id, userId, updateInput)
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
			resp, err := transactionService.UpdateTransaction(ctx, createdTx.Id, userId, updateInput)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.CategoryIds).To(ContainElement(cat2.Id))
			Expect(resp.CategoryIds).NotTo(ContainElement(cat1.Id))
		})

		It("should clear category mappings when empty array is provided", func() {
			emptyCategoryIds := []int64{}
			updateInput := models.UpdateTransactionInput{
				CategoryIds: &emptyCategoryIds,
			}
			resp, err := transactionService.UpdateTransaction(ctx, createdTx.Id, userId, updateInput)
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
			resp, err := transactionService.UpdateTransaction(ctx, createdTx.Id, userId, updateInput)
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
			resp, err := transactionService.UpdateTransaction(ctx, createdTx.Id, userId, updateInput)
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
			resp, err := transactionService.UpdateTransaction(ctx, createdTx.Id, userId, updateInput)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Description).NotTo(BeNil())
			Expect(*resp.Description).To(Equal(newDescription))
			// Other fields should remain unchanged
			Expect(resp.Name).To(Equal("Initial"))
			Expect(resp.Amount).To(Equal(500.0))
		})

		It("should update date successfully", func() {
			newDate, _ := time.Parse("2006-01-02", "2023-02-15")
			updateInput := models.UpdateTransactionInput{
				UpdateBaseTransactionInput: models.UpdateBaseTransactionInput{
					Date: newDate,
				},
			}
			resp, err := transactionService.UpdateTransaction(ctx, createdTx.Id, userId, updateInput)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Date.Format("2006-01-02")).To(Equal("2023-02-15"))
			// Other fields should remain unchanged
			Expect(resp.Name).To(Equal("Initial"))
			Expect(resp.Amount).To(Equal(500.0))
		})

		It("should fail when updating to future date", func() {
			futureDate := time.Now().AddDate(0, 0, 1) // Tomorrow
			updateInput := models.UpdateTransactionInput{
				UpdateBaseTransactionInput: models.UpdateBaseTransactionInput{
					Date: futureDate,
				},
			}
			_, err := transactionService.UpdateTransaction(ctx, createdTx.Id, userId, updateInput)
			Expect(err).To(HaveOccurred())
			// Should be a date validation error
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
			resp, err := transactionService.UpdateTransaction(ctx, createdTx.Id, userId, updateInput)
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
					CreatedBy:   userId,
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
			resp, err := transactionService.UpdateTransaction(ctx, createdTx.Id, userId, updateInput)
			Expect(err).NotTo(HaveOccurred())
			Expect(*resp.Description).To(Equal("Updated to be similar but not duplicate"))
		})
	})

	Describe("GetTransactionById", func() {
		var createdTx models.TransactionResponse
		BeforeEach(func() {
			var err error
			cat, err := categoryMockRepo.CreateCategory(ctx, models.CreateCategoryInput{Name: "GetTest Category", CreatedBy: 5})
			Expect(err).NotTo(HaveOccurred())
			Expect(cat.Id).To(BeNumerically(">", 0))
			acc, err := accountMockRepo.CreateAccount(ctx, models.CreateAccountInput{Name: "Test", BankType: "sbi", Currency: "inr", CreatedBy: 5})
			Expect(err).NotTo(HaveOccurred())
			Expect(acc.Id).To(BeNumerically(">", 0))
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
			createdTx, err = transactionService.CreateTransaction(ctx, input)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should get a transaction by its Id", func() {
			resp, err := transactionService.GetTransactionById(ctx, createdTx.Id, 5)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Id).To(Equal(createdTx.Id))
			Expect(resp.Name).To(Equal("ForGet"))
		})

		It("should fail for a non-existent Id", func() {
			_, err := transactionService.GetTransactionById(ctx, 999, 5)
			Expect(err).To(HaveOccurred())
		})

		It("should fail if user Id does not match", func() {
			_, err := transactionService.GetTransactionById(ctx, createdTx.Id, 999)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("DeleteTransaction", func() {
		var createdTx models.TransactionResponse
		var cat1 models.CategoryResponse
		var acc1 models.AccountResponse

		BeforeEach(func() {
			var err error
			cat1, err = categoryMockRepo.CreateCategory(ctx, models.CreateCategoryInput{Name: "DeleteTest Category", CreatedBy: 3})
			Expect(err).NotTo(HaveOccurred())
			Expect(cat1.Id).To(BeNumerically(">", 0))
			acc1, err = accountMockRepo.CreateAccount(ctx, models.CreateAccountInput{Name: "DeleteAccount", BankType: "sbi", Currency: "inr", CreatedBy: 3})
			Expect(err).NotTo(HaveOccurred())
			Expect(acc1.Id).To(BeNumerically(">", 0))
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
			createdTx, err = transactionService.CreateTransaction(ctx, input)
			Expect(err).NotTo(HaveOccurred())
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

	Describe("ListTransactions", func() {
		It("should return all transactions with default pagination", func() {
			result, err := transactionService.ListTransactions(ctx, userId, models.TransactionListQuery{})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Total).To(Equal(5))
			Expect(result.Page).To(Equal(1))
			Expect(result.PageSize).To(Equal(15))
			Expect(result.Transactions).To(HaveLen(5))
		})

		It("should filter by account Id", func() {
			accountId := int64(1)
			result, err := transactionService.ListTransactions(ctx, userId, models.TransactionListQuery{
				AccountId: &accountId,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Total).To(Equal(3))
			for _, tx := range result.Transactions {
				Expect(tx.AccountId).To(Equal(accountId))
			}
		})

		It("should filter by category Id", func() {
			categoryId := int64(1) // Food category
			result, err := transactionService.ListTransactions(ctx, userId, models.TransactionListQuery{
				CategoryId: &categoryId,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Total).To(Equal(2))
			for _, tx := range result.Transactions {
				found := false
				for _, catId := range tx.CategoryIds {
					if catId == categoryId {
						found = true
						break
					}
				}
				Expect(found).To(BeTrue())
			}
		})

		It("should filter uncategorized transactions", func() {
			uncategorized := true
			result, err := transactionService.ListTransactions(ctx, userId, models.TransactionListQuery{
				Uncategorized: &uncategorized,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Total).To(Equal(1)) // Only one transaction without categories
			for _, tx := range result.Transactions {
				Expect(tx.CategoryIds).To(BeEmpty())
			}
		})

		It("should filter by amount range", func() {
			minAmount := 75.0
			maxAmount := 150.0
			result, err := transactionService.ListTransactions(ctx, userId, models.TransactionListQuery{
				MinAmount: &minAmount,
				MaxAmount: &maxAmount,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Total).To(Equal(2))
			for _, tx := range result.Transactions {
				Expect(tx.Amount).To(And(
					BeNumerically(">=", minAmount),
					BeNumerically("<=", maxAmount),
				))
			}
		})

		It("should filter by date range", func() {
			dateFrom := testDate
			dateTo := testDate.AddDate(0, 0, 1)
			result, err := transactionService.ListTransactions(ctx, userId, models.TransactionListQuery{
				DateFrom: &dateFrom,
				DateTo:   &dateTo,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Total).To(Equal(2))
			for _, tx := range result.Transactions {
				Expect(tx.Date).To(And(
					BeTemporally(">=", dateFrom),
					BeTemporally("<=", dateTo),
				))
			}
		})

		It("should search by name and description", func() {
			search := "bill"
			result, err := transactionService.ListTransactions(ctx, userId, models.TransactionListQuery{
				Search: &search,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Total).To(Equal(1))
			for _, tx := range result.Transactions {
				searchFound := strings.Contains(strings.ToLower(tx.Name), search) ||
					(tx.Description != nil && strings.Contains(strings.ToLower(*tx.Description), search))
				Expect(searchFound).To(BeTrue())
			}
		})

		It("should sort by amount ascending", func() {
			result, err := transactionService.ListTransactions(ctx, userId, models.TransactionListQuery{
				SortBy:    "amount",
				SortOrder: "asc",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Total).To(Equal(5))
			for i := 1; i < len(result.Transactions); i++ {
				Expect(result.Transactions[i].Amount).To(BeNumerically(">=", result.Transactions[i-1].Amount))
			}
		})

		It("should sort by date descending", func() {
			result, err := transactionService.ListTransactions(ctx, userId, models.TransactionListQuery{
				SortBy:    "date",
				SortOrder: "desc",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Total).To(Equal(5))
			for i := 1; i < len(result.Transactions); i++ {
				Expect(result.Transactions[i].Date).To(BeTemporally("<=", result.Transactions[i-1].Date))
			}
		})

		It("should handle pagination", func() {
			result, err := transactionService.ListTransactions(ctx, userId, models.TransactionListQuery{
				Page:     2,
				PageSize: 2,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Total).To(Equal(5))
			Expect(result.Page).To(Equal(2))
			Expect(result.PageSize).To(Equal(2))
			Expect(result.Transactions).To(HaveLen(2))
		})

		It("should handle multiple filters together", func() {
			accountId := int64(1)
			minAmount := 100.0
			dateFrom := testDate
			search := "bill"
			result, err := transactionService.ListTransactions(ctx, userId, models.TransactionListQuery{
				AccountId: &accountId,
				MinAmount: &minAmount,
				DateFrom:  &dateFrom,
				Search:    &search,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Total).To(Equal(1))
			for _, tx := range result.Transactions {
				Expect(tx.AccountId).To(Equal(accountId))
				Expect(tx.Amount).To(BeNumerically(">=", minAmount))
				Expect(tx.Date).To(BeTemporally(">=", dateFrom))
				searchFound := strings.Contains(strings.ToLower(tx.Name), search) ||
					(tx.Description != nil && strings.Contains(strings.ToLower(*tx.Description), search))
				Expect(searchFound).To(BeTrue())
			}
		})

		It("should return empty result for non-existent filters", func() {
			accountId := int64(999)
			result, err := transactionService.ListTransactions(ctx, userId, models.TransactionListQuery{
				AccountId: &accountId,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Total).To(Equal(0))
			Expect(result.Transactions).To(BeEmpty())
		})
	})

	Describe("TransactionService validation helpers", func() {
		var (
			transactionService *TransactionService
			categoryMockRepo   *repository.MockCategoryRepository
			accountMockRepo    *repository.MockAccountRepository
			ctx                *gin.Context
		)

		BeforeEach(func() {
			ctx = &gin.Context{}
			categoryMockRepo = repository.NewMockCategoryRepository()
			accountMockRepo = repository.NewMockAccountRepository()
			transactionService = &TransactionService{
				repo:         nil, // not needed for these tests
				categoryRepo: categoryMockRepo,
				accountRepo:  accountMockRepo,
				db:           nil,
			}
		})

		Describe("validateDateNotInFuture", func() {
			It("should return error if date is in the future", func() {
				future := time.Now().Add(48 * time.Hour)
				err := transactionService.validateDateNotInFuture(future)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("transaction date cannot be in the future"))
			})
			It("should not return error if date is today or in the past", func() {
				past := time.Now().Add(-24 * time.Hour)
				err := transactionService.validateDateNotInFuture(past)
				Expect(err).NotTo(HaveOccurred())
				today := time.Now()
				err = transactionService.validateDateNotInFuture(today)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Describe("validateAccountExists", func() {
			It("should return error if account does not exist", func() {
				err := transactionService.validateAccountExists(ctx, 9999, 1)
				Expect(err).To(HaveOccurred())
			})
			It("should not return error if account exists", func() {
				acc, _ := accountMockRepo.CreateAccount(ctx, models.CreateAccountInput{Name: "Test", BankType: "hdfc", Currency: "inr", CreatedBy: 1})
				err := transactionService.validateAccountExists(ctx, acc.Id, 1)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Describe("validateCategoryExists", func() {
			It("should return nil if categoryIds is empty", func() {
				err := transactionService.validateCategoryExists(ctx, []int64{}, 1)
				Expect(err).NotTo(HaveOccurred())
			})
			It("should return error if any category does not exist", func() {
				cat, _ := categoryMockRepo.CreateCategory(ctx, models.CreateCategoryInput{Name: "Test", CreatedBy: 1})
				err := transactionService.validateCategoryExists(ctx, []int64{cat.Id, 9999}, 1)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("category not found"))
			})
			It("should not return error if all categories exist", func() {
				cat1, _ := categoryMockRepo.CreateCategory(ctx, models.CreateCategoryInput{Name: "Test1", CreatedBy: 1})
				cat2, _ := categoryMockRepo.CreateCategory(ctx, models.CreateCategoryInput{Name: "Test2", CreatedBy: 1})
				err := transactionService.validateCategoryExists(ctx, []int64{cat1.Id, cat2.Id}, 1)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
