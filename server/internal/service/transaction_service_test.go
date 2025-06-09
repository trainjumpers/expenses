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
})
