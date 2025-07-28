package service

import (
	repository "expenses/internal/mock/repository"
	"expenses/internal/models"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gin-gonic/gin"
)

var _ = Describe("RuleEngineService", func() {
	var (
		service          RuleEngineServiceInterface
		mockRuleRepo     *repository.MockRuleRepository
		mockTxnRepo      *repository.MockTransactionRepository
		mockCategoryRepo *repository.MockCategoryRepository
		ctx              *gin.Context
		userId           int64
	)

	BeforeEach(func() {
		ctx = &gin.Context{}
		userId = 1

		// Setup mocks
		mockRuleRepo = repository.NewMockRuleRepository()
		mockTxnRepo = repository.NewMockTransactionRepository()
		mockCategoryRepo = repository.NewMockCategoryRepository()

		service = NewRuleEngineService(mockRuleRepo, mockTxnRepo, mockCategoryRepo)
	})

	Describe("ExecuteRules - Basic Cases", func() {
		It("should return empty response when no rules exist", func() {
			// Create a category first
			_, err := mockCategoryRepo.CreateCategory(ctx, models.CreateCategoryInput{
				Name:      "Food",
				CreatedBy: userId,
			})
			Expect(err).NotTo(HaveOccurred())

			request := models.ExecuteRulesRequest{}

			response, err := service.ExecuteRules(ctx, userId, request)

			Expect(err).NotTo(HaveOccurred())
			Expect(response.TotalRules).To(Equal(0))
			Expect(response.ProcessedTxns).To(Equal(0))
			Expect(response.Modified).To(HaveLen(0))
		})

		It("should handle page size correctly", func() {
			// Create a category
			_, err := mockCategoryRepo.CreateCategory(ctx, models.CreateCategoryInput{
				Name:      "Food",
				CreatedBy: userId,
			})
			Expect(err).NotTo(HaveOccurred())

			request := models.ExecuteRulesRequest{PageSize: 50}

			response, err := service.ExecuteRules(ctx, userId, request)

			Expect(err).NotTo(HaveOccurred())
			Expect(response.TotalRules).To(Equal(0))
			Expect(response.ProcessedTxns).To(Equal(0))
		})

		It("should use default page size for invalid values", func() {
			// Create a category
			_, err := mockCategoryRepo.CreateCategory(ctx, models.CreateCategoryInput{
				Name:      "Food",
				CreatedBy: userId,
			})
			Expect(err).NotTo(HaveOccurred())

			request := models.ExecuteRulesRequest{PageSize: -1} // Invalid

			response, err := service.ExecuteRules(ctx, userId, request)

			Expect(err).NotTo(HaveOccurred())
			Expect(response.TotalRules).To(Equal(0))
		})

		It("should use default page size for oversized values", func() {
			// Create a category
			_, err := mockCategoryRepo.CreateCategory(ctx, models.CreateCategoryInput{
				Name:      "Food",
				CreatedBy: userId,
			})
			Expect(err).NotTo(HaveOccurred())

			request := models.ExecuteRulesRequest{PageSize: 2000} // Too large

			response, err := service.ExecuteRules(ctx, userId, request)

			Expect(err).NotTo(HaveOccurred())
			Expect(response.TotalRules).To(Equal(0))
		})
	})

	Describe("ExecuteRules - Edge Cases", func() {
		It("should handle empty rule IDs gracefully", func() {
			// Create a category
			_, err := mockCategoryRepo.CreateCategory(ctx, models.CreateCategoryInput{
				Name:      "Food",
				CreatedBy: userId,
			})
			Expect(err).NotTo(HaveOccurred())

			emptyRuleIds := []int64{}
			request := models.ExecuteRulesRequest{RuleIds: &emptyRuleIds}

			response, err := service.ExecuteRules(ctx, userId, request)

			Expect(err).NotTo(HaveOccurred())
			Expect(response.TotalRules).To(Equal(0))
			Expect(response.ProcessedTxns).To(Equal(0))
		})

		It("should handle empty transaction IDs gracefully", func() {
			// Create a category
			_, err := mockCategoryRepo.CreateCategory(ctx, models.CreateCategoryInput{
				Name:      "Food",
				CreatedBy: userId,
			})
			Expect(err).NotTo(HaveOccurred())

			emptyTxnIds := []int64{}
			request := models.ExecuteRulesRequest{TransactionIds: &emptyTxnIds}

			response, err := service.ExecuteRules(ctx, userId, request)

			Expect(err).NotTo(HaveOccurred())
			Expect(response.TotalRules).To(Equal(0))
			Expect(response.ProcessedTxns).To(Equal(0))
		})

		It("should handle non-existent rule IDs gracefully", func() {
			// Create a category
			_, err := mockCategoryRepo.CreateCategory(ctx, models.CreateCategoryInput{
				Name:      "Food",
				CreatedBy: userId,
			})
			Expect(err).NotTo(HaveOccurred())

			nonExistentRuleIds := []int64{999, 1000}
			request := models.ExecuteRulesRequest{RuleIds: &nonExistentRuleIds}

			response, err := service.ExecuteRules(ctx, userId, request)

			Expect(err).NotTo(HaveOccurred())
			Expect(response.TotalRules).To(Equal(0)) // No valid rules found
			Expect(response.ProcessedTxns).To(Equal(0))
		})

		It("should handle non-existent transaction IDs gracefully", func() {
			// Create a category
			_, err := mockCategoryRepo.CreateCategory(ctx, models.CreateCategoryInput{
				Name:      "Food",
				CreatedBy: userId,
			})
			Expect(err).NotTo(HaveOccurred())

			nonExistentTxnIds := []int64{999, 1000}
			request := models.ExecuteRulesRequest{TransactionIds: &nonExistentTxnIds}

			response, err := service.ExecuteRules(ctx, userId, request)

			Expect(err).NotTo(HaveOccurred())
			Expect(response.TotalRules).To(Equal(0))
			Expect(response.ProcessedTxns).To(Equal(0)) // No valid transactions found
		})
	})

	Describe("ExecuteRules - With Data", func() {
		var (
			txn1 models.TransactionResponse
		)

		BeforeEach(func() {
			var err error
			// Create category
			_, err = mockCategoryRepo.CreateCategory(ctx, models.CreateCategoryInput{
				Name:      "Food",
				CreatedBy: userId,
			})
			Expect(err).NotTo(HaveOccurred())

			// Create transaction
			amount := 50.0
			txn1, err = mockTxnRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
				Name:        "Grocery Store",
				Description: "Weekly groceries",
				Amount:      &amount,
				Date:        time.Now(),
				CreatedBy:   userId,
				AccountId:   1,
			}, []int64{})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should process specific transactions when transaction IDs provided", func() {
			txnIds := []int64{txn1.Id}
			request := models.ExecuteRulesRequest{TransactionIds: &txnIds}

			response, err := service.ExecuteRules(ctx, userId, request)

			Expect(err).NotTo(HaveOccurred())
			Expect(response.TotalRules).To(Equal(0))    // No rules exist
			Expect(response.ProcessedTxns).To(Equal(0)) // No transactions processed since no rules
			Expect(response.Modified).To(HaveLen(0))    // No modifications since no rules
		})

		It("should handle mixed valid and invalid transaction IDs", func() {
			txnIds := []int64{txn1.Id, 999} // One valid, one invalid
			request := models.ExecuteRulesRequest{TransactionIds: &txnIds}

			response, err := service.ExecuteRules(ctx, userId, request)

			Expect(err).NotTo(HaveOccurred())
			Expect(response.TotalRules).To(Equal(0))
			Expect(response.ProcessedTxns).To(Equal(0)) // No transactions processed since no rules
			Expect(response.Modified).To(HaveLen(0))
		})

		It("should process all transactions when no specific IDs provided", func() {
			request := models.ExecuteRulesRequest{}

			response, err := service.ExecuteRules(ctx, userId, request)

			Expect(err).NotTo(HaveOccurred())
			Expect(response.TotalRules).To(Equal(0))
			Expect(response.ProcessedTxns).To(Equal(0)) // No transactions processed since no rules
			Expect(response.Modified).To(HaveLen(0))
		})

		It("should handle pagination with small page size", func() {
			// Create more transactions
			for i := 2; i <= 3; i++ {
				amount := float64(i * 10)
				_, err := mockTxnRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
					Name:        fmt.Sprintf("Transaction %d", i),
					Description: "Test transaction",
					Amount:      &amount,
					Date:        time.Now(),
					CreatedBy:   userId,
					AccountId:   1,
				}, []int64{})
				Expect(err).NotTo(HaveOccurred())
			}

			request := models.ExecuteRulesRequest{PageSize: 2} // Small page size

			response, err := service.ExecuteRules(ctx, userId, request)

			Expect(err).NotTo(HaveOccurred())
			Expect(response.TotalRules).To(Equal(0))
			Expect(response.ProcessedTxns).To(Equal(0)) // No transactions processed since no rules
			Expect(response.Modified).To(HaveLen(0))
		})
	})

	Describe("Helper Functions", func() {
		Describe("getUpdatedFields", func() {
			It("should return correct updated fields", func() {
				changeset := &Changeset{
					TransactionId: 100,
					NameUpdate:    stringPtr("Updated Name"),
					DescUpdate:    stringPtr("Updated Description"),
					CategoryAdds:  []int64{1, 2},
				}

				fields := service.(*ruleEngineService).getUpdatedFields(changeset)

				Expect(fields).To(ContainElements(
					models.RuleFieldName,
					models.RuleFieldDescription,
					models.RuleFieldCategory,
				))
			})

			It("should return empty for changeset with no updates", func() {
				changeset := &Changeset{
					TransactionId: 100,
				}

				fields := service.(*ruleEngineService).getUpdatedFields(changeset)

				Expect(fields).To(HaveLen(0))
			})

			It("should return only name field when only name is updated", func() {
				changeset := &Changeset{
					TransactionId: 100,
					NameUpdate:    stringPtr("Updated Name"),
				}

				fields := service.(*ruleEngineService).getUpdatedFields(changeset)

				Expect(fields).To(HaveLen(1))
				Expect(fields).To(ContainElement(models.RuleFieldName))
			})

			It("should return only category field when only categories are added", func() {
				changeset := &Changeset{
					TransactionId: 100,
					CategoryAdds:  []int64{1, 2},
				}

				fields := service.(*ruleEngineService).getUpdatedFields(changeset)

				Expect(fields).To(HaveLen(1))
				Expect(fields).To(ContainElement(models.RuleFieldCategory))
			})

			It("should return only description field when only description is updated", func() {
				changeset := &Changeset{
					TransactionId: 100,
					DescUpdate:    stringPtr("Updated Description"),
				}

				fields := service.(*ruleEngineService).getUpdatedFields(changeset)

				Expect(fields).To(HaveLen(1))
				Expect(fields).To(ContainElement(models.RuleFieldDescription))
			})
		})

		Describe("fetchSpecificTransactions", func() {
			var txn1 models.TransactionResponse

			BeforeEach(func() {
				var err error
				amount := 50.0
				txn1, err = mockTxnRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
					Name:        "Test Transaction",
					Description: "Test description",
					Amount:      &amount,
					Date:        time.Now(),
					CreatedBy:   userId,
					AccountId:   1,
				}, []int64{})
				Expect(err).NotTo(HaveOccurred())
			})

			It("should fetch specific transactions successfully", func() {
				txnIds := []int64{txn1.Id}

				transactions, err := service.(*ruleEngineService).fetchSpecificTransactions(ctx, userId, txnIds)

				Expect(err).NotTo(HaveOccurred())
				Expect(transactions).To(HaveLen(1))
				Expect(transactions[0].Id).To(Equal(txn1.Id))
			})

			It("should handle non-existent transaction IDs gracefully", func() {
				txnIds := []int64{999, 1000}

				transactions, err := service.(*ruleEngineService).fetchSpecificTransactions(ctx, userId, txnIds)

				Expect(err).NotTo(HaveOccurred())
				Expect(transactions).To(HaveLen(0)) // No valid transactions found
			})

			It("should handle mixed valid and invalid transaction IDs", func() {
				txnIds := []int64{txn1.Id, 999}

				transactions, err := service.(*ruleEngineService).fetchSpecificTransactions(ctx, userId, txnIds)

				Expect(err).NotTo(HaveOccurred())
				Expect(transactions).To(HaveLen(1)) // Only valid transaction returned
				Expect(transactions[0].Id).To(Equal(txn1.Id))
			})
		})

		Describe("fetchTransactionPage", func() {
			BeforeEach(func() {
				// Create multiple transactions
				for i := 1; i <= 5; i++ {
					amount := float64(i * 10)
					_, err := mockTxnRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
						Name:        fmt.Sprintf("Transaction %d", i),
						Description: "Test transaction",
						Amount:      &amount,
						Date:        time.Now(),
						CreatedBy:   userId,
						AccountId:   1,
					}, []int64{})
					Expect(err).NotTo(HaveOccurred())
				}
			})

			It("should fetch transaction page successfully", func() {
				transactions, err := service.(*ruleEngineService).fetchTransactionPage(ctx, userId, 1, 3)

				Expect(err).NotTo(HaveOccurred())
				Expect(transactions).To(HaveLen(3)) // First page with 3 transactions
			})

			It("should handle page beyond available data", func() {
				transactions, err := service.(*ruleEngineService).fetchTransactionPage(ctx, userId, 10, 3)

				Expect(err).NotTo(HaveOccurred())
				Expect(transactions).To(HaveLen(0)) // No transactions on page 10
			})
		})

		Describe("processTransactions", func() {
			var (
				engine *RuleEngine
				txn1   models.TransactionResponse
			)

			BeforeEach(func() {
				var err error
				// Create category and transaction
				cat1, err := mockCategoryRepo.CreateCategory(ctx, models.CreateCategoryInput{
					Name:      "Food",
					CreatedBy: userId,
				})
				Expect(err).NotTo(HaveOccurred())

				amount := 50.0
				txn1, err = mockTxnRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
					Name:        "Grocery Store",
					Description: "Weekly groceries",
					Amount:      &amount,
					Date:        time.Now(),
					CreatedBy:   userId,
					AccountId:   1,
				}, []int64{})
				Expect(err).NotTo(HaveOccurred())

				// Create a simple rule engine with no rules
				categories := []models.CategoryResponse{cat1}
				rules := []models.DescribeRuleResponse{}
				engine = NewRuleEngine(categories, rules)
			})

			It("should process transactions with no rules", func() {
				transactions := []models.TransactionResponse{txn1}

				changesets := service.(*ruleEngineService).processTransactions(engine, transactions)

				Expect(changesets).To(HaveLen(0)) // No rules, so no changesets
			})

			It("should handle empty transaction list", func() {
				transactions := []models.TransactionResponse{}

				changesets := service.(*ruleEngineService).processTransactions(engine, transactions)

				Expect(changesets).To(HaveLen(0))
			})
		})
	})
})

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
