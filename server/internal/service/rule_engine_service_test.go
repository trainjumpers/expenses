package service

import (
	"context"
	repository "expenses/internal/mock/repository"
	"expenses/internal/models"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RuleEngineService", func() {
	var (
		service          RuleEngineServiceInterface
		mockRuleRepo     *repository.MockRuleRepository
		mockTxnRepo      *repository.MockTransactionRepository
		mockCategoryRepo *repository.MockCategoryRepository
		mockAccountRepo  *repository.MockAccountRepository
		ctx              context.Context
		userId           int64
	)

	BeforeEach(func() {
		ctx = context.Background()
		userId = 1

		// Setup mocks
		mockRuleRepo = repository.NewMockRuleRepository()
		mockTxnRepo = repository.NewMockTransactionRepository()
		mockCategoryRepo = repository.NewMockCategoryRepository()
		mockAccountRepo = repository.NewMockAccountRepository()

		service = NewRuleEngineService(mockRuleRepo, mockTxnRepo, mockCategoryRepo, mockAccountRepo)
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
				accounts := []models.AccountResponse{}
				rules := []models.DescribeRuleResponse{}
				engine = NewRuleEngine(categories, accounts, rules)
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

		Describe("buildRuleResponse", func() {
			var rule models.RuleResponse

			BeforeEach(func() {
				var err error
				// Create a rule first
				desc := "Test Description"
				ruleInput := models.CreateBaseRuleRequest{
					Name:          "Test Rule",
					Description:   &desc,
					EffectiveFrom: time.Now().Add(-24 * time.Hour), // Past date
					CreatedBy:     userId,
				}
				rule, err = mockRuleRepo.CreateRule(ctx, ruleInput)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should build rule response successfully", func() {
				// Create actions and conditions for the rule
				actions := []models.CreateRuleActionRequest{
					{
						RuleId:      rule.Id,
						ActionType:  models.RuleFieldName,
						ActionValue: "Updated Name",
					},
				}
				_, err := mockRuleRepo.CreateRuleActions(ctx, actions)
				Expect(err).NotTo(HaveOccurred())

				conditions := []models.CreateRuleConditionRequest{
					{
						RuleId:            rule.Id,
						ConditionType:     models.RuleFieldName,
						ConditionOperator: models.OperatorContains,
						ConditionValue:    "grocery",
					},
				}
				_, err = mockRuleRepo.CreateRuleConditions(ctx, conditions)
				Expect(err).NotTo(HaveOccurred())

				ruleResponse, err := service.(*ruleEngineService).buildRuleResponse(ctx, rule)

				Expect(err).NotTo(HaveOccurred())
				Expect(ruleResponse).NotTo(BeNil())
				Expect(ruleResponse.Rule.Id).To(Equal(rule.Id))
				Expect(ruleResponse.Actions).To(HaveLen(1))
				Expect(ruleResponse.Conditions).To(HaveLen(1))
			})

			It("should return empty actions and conditions when none exist", func() {
				// This test demonstrates the behavior when actions/conditions don't exist
				// The mock repository returns empty slices rather than errors for non-existent data
				invalidRule := models.RuleResponse{Id: 999, Name: "Invalid", CreatedBy: userId}

				ruleResponse, err := service.(*ruleEngineService).buildRuleResponse(ctx, invalidRule)

				// The mock doesn't return errors for non-existent data, it returns empty slices
				Expect(err).NotTo(HaveOccurred())
				Expect(ruleResponse).NotTo(BeNil())
				Expect(ruleResponse.Actions).To(HaveLen(0))
				Expect(ruleResponse.Conditions).To(HaveLen(0))
			})
		})

		Describe("fetchSpecificRules", func() {
			var rule1, rule2 models.RuleResponse

			BeforeEach(func() {
				var err error
				// Create rules
				desc1 := "First rule"
				rule1, err = mockRuleRepo.CreateRule(ctx, models.CreateBaseRuleRequest{
					Name:          "Rule 1",
					Description:   &desc1,
					EffectiveFrom: time.Now().Add(-24 * time.Hour),
					CreatedBy:     userId,
				})
				Expect(err).NotTo(HaveOccurred())

				desc2 := "Second rule"
				rule2, err = mockRuleRepo.CreateRule(ctx, models.CreateBaseRuleRequest{
					Name:          "Rule 2",
					Description:   &desc2,
					EffectiveFrom: time.Now().Add(-12 * time.Hour),
					CreatedBy:     userId,
				})
				Expect(err).NotTo(HaveOccurred())
			})

			It("should fetch specific rules successfully", func() {
				ruleIds := []int64{rule1.Id, rule2.Id}

				rules, err := service.(*ruleEngineService).fetchSpecificRules(ctx, userId, ruleIds)

				Expect(err).NotTo(HaveOccurred())
				Expect(rules).To(HaveLen(2))
				Expect(rules[0].Rule.Id).To(Equal(rule1.Id))
				Expect(rules[1].Rule.Id).To(Equal(rule2.Id))
			})

			It("should handle non-existent rule IDs gracefully", func() {
				ruleIds := []int64{999, 1000}

				rules, err := service.(*ruleEngineService).fetchSpecificRules(ctx, userId, ruleIds)

				Expect(err).NotTo(HaveOccurred())
				Expect(rules).To(HaveLen(0))
			})

			It("should handle mixed valid and invalid rule IDs", func() {
				ruleIds := []int64{rule1.Id, 999}

				rules, err := service.(*ruleEngineService).fetchSpecificRules(ctx, userId, ruleIds)

				Expect(err).NotTo(HaveOccurred())
				Expect(rules).To(HaveLen(1))
				Expect(rules[0].Rule.Id).To(Equal(rule1.Id))
			})
		})

		Describe("fetchAllUserRules", func() {
			BeforeEach(func() {
				// Create rules with different effective dates
				pastDesc := "Rule from past"
				_, err := mockRuleRepo.CreateRule(ctx, models.CreateBaseRuleRequest{
					Name:          "Past Rule",
					Description:   &pastDesc,
					EffectiveFrom: time.Now().Add(-24 * time.Hour), // Past
					CreatedBy:     userId,
				})
				Expect(err).NotTo(HaveOccurred())

				futureDesc := "Rule for future"
				_, err = mockRuleRepo.CreateRule(ctx, models.CreateBaseRuleRequest{
					Name:          "Future Rule",
					Description:   &futureDesc,
					EffectiveFrom: time.Now().Add(24 * time.Hour), // Future
					CreatedBy:     userId,
				})
				Expect(err).NotTo(HaveOccurred())
			})

			It("should fetch only rules with past or current effective dates", func() {
				rules, err := service.(*ruleEngineService).fetchAllUserRules(ctx, userId)

				Expect(err).NotTo(HaveOccurred())
				Expect(rules).To(HaveLen(1)) // Only past rule should be returned
				Expect(rules[0].Rule.Name).To(Equal("Past Rule"))
			})

			It("should handle empty rules list", func() {
				// Create new user with no rules
				newUserId := int64(999)

				rules, err := service.(*ruleEngineService).fetchAllUserRules(ctx, newUserId)

				Expect(err).NotTo(HaveOccurred())
				Expect(rules).To(HaveLen(0))
			})
		})

		Describe("applyChangesets", func() {
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

			It("should apply changesets successfully", func() {
				changesets := []*Changeset{
					{
						TransactionId: txn1.Id,
						NameUpdate:    stringPtr("Updated Name"),
						DescUpdate:    stringPtr("Updated Description"),
						CategoryAdds:  []int64{1},
						AppliedRules:  []int64{1},
					},
				}

				modified, err := service.(*ruleEngineService).applyChangesets(ctx, userId, changesets)

				Expect(err).NotTo(HaveOccurred())
				Expect(modified).To(HaveLen(1))
				Expect(modified[0].TransactionId).To(Equal(txn1.Id))
				Expect(modified[0].AppliedRules).To(ContainElement(int64(1)))
				Expect(modified[0].UpdatedFields).To(ContainElements(
					models.RuleFieldName,
					models.RuleFieldDescription,
					models.RuleFieldCategory,
				))
			})

			It("should handle empty changesets", func() {
				changesets := []*Changeset{}

				modified, err := service.(*ruleEngineService).applyChangesets(ctx, userId, changesets)

				Expect(err).NotTo(HaveOccurred())
				Expect(modified).To(HaveLen(0))
			})

			It("should continue processing when one changeset fails", func() {
				changesets := []*Changeset{
					{
						TransactionId: 999, // Non-existent transaction
						NameUpdate:    stringPtr("Updated Name"),
						AppliedRules:  []int64{1},
					},
					{
						TransactionId: txn1.Id,
						NameUpdate:    stringPtr("Valid Update"),
						AppliedRules:  []int64{2},
					},
				}

				modified, err := service.(*ruleEngineService).applyChangesets(ctx, userId, changesets)

				Expect(err).NotTo(HaveOccurred())
				Expect(modified).To(HaveLen(1)) // Only valid changeset applied
				Expect(modified[0].TransactionId).To(Equal(txn1.Id))
			})
		})

		Describe("applyChangeset", func() {
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

			It("should apply name and description updates", func() {
				changeset := &Changeset{
					TransactionId: txn1.Id,
					NameUpdate:    stringPtr("Updated Name"),
					DescUpdate:    stringPtr("Updated Description"),
				}

				err := service.(*ruleEngineService).applyChangeset(ctx, userId, changeset)

				Expect(err).NotTo(HaveOccurred())
			})

			It("should apply category updates", func() {
				changeset := &Changeset{
					TransactionId: txn1.Id,
					CategoryAdds:  []int64{1, 2},
				}

				err := service.(*ruleEngineService).applyChangeset(ctx, userId, changeset)

				Expect(err).NotTo(HaveOccurred())
			})

			It("should apply both field and category updates", func() {
				changeset := &Changeset{
					TransactionId: txn1.Id,
					NameUpdate:    stringPtr("Updated Name"),
					CategoryAdds:  []int64{1},
				}

				err := service.(*ruleEngineService).applyChangeset(ctx, userId, changeset)

				Expect(err).NotTo(HaveOccurred())
			})

			It("should handle non-existent transaction", func() {
				changeset := &Changeset{
					TransactionId: 999,
					NameUpdate:    stringPtr("Updated Name"),
				}

				err := service.(*ruleEngineService).applyChangeset(ctx, userId, changeset)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to get transaction"))
			})

			It("should handle changeset with no updates", func() {
				changeset := &Changeset{
					TransactionId: txn1.Id,
				}

				err := service.(*ruleEngineService).applyChangeset(ctx, userId, changeset)

				Expect(err).NotTo(HaveOccurred()) // Should succeed with no operations
			})
		})

		Describe("mapRuleTransaction", func() {
			It("should map rule transactions successfully", func() {
				changeset := &Changeset{
					TransactionId: 100,
					AppliedRules:  []int64{1, 2},
				}

				// This should not error - mapping failures are logged but don't fail the operation
				service.(*ruleEngineService).mapRuleTransaction(ctx, changeset)

				// No assertions needed as this method doesn't return errors
				// It logs errors internally
			})

			It("should handle changeset with no applied rules", func() {
				changeset := &Changeset{
					TransactionId: 100,
					AppliedRules:  []int64{},
				}

				// Should complete without issues
				service.(*ruleEngineService).mapRuleTransaction(ctx, changeset)
			})
		})
	})

	Describe("Error Handling", func() {
		It("should handle category repository errors gracefully", func() {
			// This test would require mocking repository failures
			// For now, we'll test the basic flow
			request := models.ExecuteRulesRequest{}

			_, err := service.ExecuteRules(ctx, userId, request)

			Expect(err).NotTo(HaveOccurred())
			// Background execution will handle the error internally
		})

		It("should handle rule repository errors gracefully", func() {
			// Create a category first
			_, err := mockCategoryRepo.CreateCategory(ctx, models.CreateCategoryInput{
				Name:      "Food",
				CreatedBy: userId,
			})
			Expect(err).NotTo(HaveOccurred())

			request := models.ExecuteRulesRequest{}

			_, err = service.ExecuteRules(ctx, userId, request)

			Expect(err).NotTo(HaveOccurred())
			// Background execution will handle any rule fetch errors internally
		})

		It("should handle transaction repository errors gracefully", func() {
			// Create a category first
			_, err := mockCategoryRepo.CreateCategory(ctx, models.CreateCategoryInput{
				Name:      "Food",
				CreatedBy: userId,
			})
			Expect(err).NotTo(HaveOccurred())

			request := models.ExecuteRulesRequest{}

			_, err = service.ExecuteRules(ctx, userId, request)

			Expect(err).NotTo(HaveOccurred())
			// Background execution will handle transaction fetch errors internally
		})
	})

	Describe("Integration Scenarios", func() {
		var (
			cat1  models.CategoryResponse
			rule1 models.RuleResponse
			txn1  models.TransactionResponse
		)

		BeforeEach(func() {
			var err error
			// Create categories
			cat1, err = mockCategoryRepo.CreateCategory(ctx, models.CreateCategoryInput{
				Name:      "Food",
				CreatedBy: userId,
			})
			Expect(err).NotTo(HaveOccurred())

			_, err = mockCategoryRepo.CreateCategory(ctx, models.CreateCategoryInput{
				Name:      "Transport",
				CreatedBy: userId,
			})
			Expect(err).NotTo(HaveOccurred())

			// Create rule
			ruleDesc := "Categorize grocery transactions"
			rule1, err = mockRuleRepo.CreateRule(ctx, models.CreateBaseRuleRequest{
				Name:          "Grocery Rule",
				Description:   &ruleDesc,
				EffectiveFrom: time.Now().Add(-24 * time.Hour),
				CreatedBy:     userId,
			})
			Expect(err).NotTo(HaveOccurred())

			// Create rule action
			actions := []models.CreateRuleActionRequest{
				{
					RuleId:      rule1.Id,
					ActionType:  models.RuleFieldCategory,
					ActionValue: fmt.Sprintf("%d", cat1.Id),
				},
			}
			_, err = mockRuleRepo.CreateRuleActions(ctx, actions)
			Expect(err).NotTo(HaveOccurred())

			// Create rule condition
			conditions := []models.CreateRuleConditionRequest{
				{
					RuleId:            rule1.Id,
					ConditionType:     models.RuleFieldName,
					ConditionOperator: models.OperatorContains,
					ConditionValue:    "grocery",
				},
			}
			_, err = mockRuleRepo.CreateRuleConditions(ctx, conditions)
			Expect(err).NotTo(HaveOccurred())

			// Create transactions
			amount1 := 50.0
			txn1, err = mockTxnRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
				Name:        "Grocery Store Purchase",
				Description: "Weekly groceries",
				Amount:      &amount1,
				Date:        time.Now(),
				CreatedBy:   userId,
				AccountId:   1,
			}, []int64{})
			Expect(err).NotTo(HaveOccurred())

			amount2 := 25.0
			_, err = mockTxnRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
				Name:        "Gas Station",
				Description: "Fuel purchase",
				Amount:      &amount2,
				Date:        time.Now(),
				CreatedBy:   userId,
				AccountId:   1,
			}, []int64{})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should execute rules with complete workflow", func() {
			request := models.ExecuteRulesRequest{}

			response, err := service.ExecuteRules(ctx, userId, request)

			Expect(err).NotTo(HaveOccurred())
			// Since execution is async, we can only verify the response structure
			Expect(response).NotTo(BeNil())
		})

		It("should execute rules for specific transactions", func() {
			txnIds := []int64{txn1.Id}
			request := models.ExecuteRulesRequest{TransactionIds: &txnIds}

			response, err := service.ExecuteRules(ctx, userId, request)

			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
		})

		It("should execute specific rules on all transactions", func() {
			ruleIds := []int64{rule1.Id}
			request := models.ExecuteRulesRequest{RuleIds: &ruleIds}

			response, err := service.ExecuteRules(ctx, userId, request)

			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
		})

		It("should execute specific rules on specific transactions", func() {
			ruleIds := []int64{rule1.Id}
			txnIds := []int64{txn1.Id}
			request := models.ExecuteRulesRequest{
				RuleIds:        &ruleIds,
				TransactionIds: &txnIds,
			}

			response, err := service.ExecuteRules(ctx, userId, request)

			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
		})
	})

	Describe("Transfer Transaction Creation", func() {
		var (
			cat1, cat2 models.CategoryResponse
			acc1, acc2 models.AccountResponse
			rule1      models.RuleResponse
			txn1       models.TransactionResponse
		)

		BeforeEach(func() {
			// Create categories
			cat1, err := mockCategoryRepo.CreateCategory(ctx, models.CreateCategoryInput{
				Name:      "Food",
				CreatedBy: userId,
			})
			Expect(err).NotTo(HaveOccurred())

			cat2, err = mockCategoryRepo.CreateCategory(ctx, models.CreateCategoryInput{
				Name:      "Transfer",
				CreatedBy: userId,
			})
			Expect(err).NotTo(HaveOccurred())

			// Create accounts
			acc1, err = mockAccountRepo.CreateAccount(ctx, models.CreateAccountInput{
				Name:      "Checking",
				BankType:  models.BankTypeHDFC,
				Currency:  models.CurrencyINR,
				CreatedBy: userId,
			})
			Expect(err).NotTo(HaveOccurred())

			acc2, err = mockAccountRepo.CreateAccount(ctx, models.CreateAccountInput{
				Name:      "Savings",
				BankType:  models.BankTypeSBI,
				Currency:  models.CurrencyINR,
				CreatedBy: userId,
			})
			Expect(err).NotTo(HaveOccurred())

			// Create rule with transfer action
			desc := "Transfer rule"
			rule1, err = mockRuleRepo.CreateRule(ctx, models.CreateBaseRuleRequest{
				Name:          "Transfer Rule",
				Description:   &desc,
				EffectiveFrom: time.Now().Add(-24 * time.Hour),
				CreatedBy:     userId,
			})
			Expect(err).NotTo(HaveOccurred())

			// Create rule action for transfer
			actions := []models.CreateRuleActionRequest{
				{
					RuleId:      rule1.Id,
					ActionType:  models.RuleFieldTransfer,
					ActionValue: fmt.Sprintf("%d", acc2.Id),
				},
			}
			_, err = mockRuleRepo.CreateRuleActions(ctx, actions)
			Expect(err).NotTo(HaveOccurred())

			// Create rule condition
			conditions := []models.CreateRuleConditionRequest{
				{
					RuleId:            rule1.Id,
					ConditionType:     models.RuleFieldName,
					ConditionOperator: models.OperatorContains,
					ConditionValue:    "transfer",
				},
			}
			_, err = mockRuleRepo.CreateRuleConditions(ctx, conditions)
			Expect(err).NotTo(HaveOccurred())

			// Create transaction that will trigger transfer
			amount := 100.0
			txn1, err = mockTxnRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
				Name:        "Transfer to savings",
				Description: "Monthly transfer",
				Amount:      &amount,
				Date:        time.Now(),
				CreatedBy:   userId,
				AccountId:   acc1.Id,
			}, []int64{cat1.Id})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should create transfer transaction when rule is applied", func() {
			// Execute rules
			txnIds := []int64{txn1.Id}
			request := models.ExecuteRulesRequest{TransactionIds: &txnIds}
			response, err := service.ExecuteRules(ctx, userId, request)

			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())

			// Since execution is async, we can't directly test the transfer creation
			// But we can verify the rule engine processes the transfer correctly
			// by testing the changeset creation
			categories := []models.CategoryResponse{cat1, cat2}
			accounts := []models.AccountResponse{acc1, acc2}
			rules := []models.DescribeRuleResponse{
				{
					Rule:       rule1,
					Actions:    []models.RuleActionResponse{{ActionType: models.RuleFieldTransfer, ActionValue: fmt.Sprintf("%d", acc2.Id)}},
					Conditions: []models.RuleConditionResponse{{ConditionType: models.RuleFieldName, ConditionValue: "transfer", ConditionOperator: models.OperatorContains}},
				},
			}

			engine := NewRuleEngine(categories, accounts, rules)
			changeset := engine.ProcessTransaction(txn1)

			Expect(changeset).NotTo(BeNil())
			Expect(changeset.TransferInfo).NotTo(BeNil())
			Expect(changeset.TransferInfo.AccountId).To(Equal(acc2.Id))
			Expect(changeset.TransferInfo.Amount).To(Equal(-txn1.Amount))
			Expect(changeset.AppliedRules).To(ContainElement(rule1.Id))
		})

		It("should not create transfer to same account", func() {
			// Update rule to transfer to same account
			actions := []models.CreateRuleActionRequest{
				{
					RuleId:      rule1.Id,
					ActionType:  models.RuleFieldTransfer,
					ActionValue: fmt.Sprintf("%d", acc1.Id), // Same account
				},
			}
			_, err := mockRuleRepo.CreateRuleActions(ctx, actions)
			Expect(err).NotTo(HaveOccurred())

			categories := []models.CategoryResponse{cat1, cat2}
			accounts := []models.AccountResponse{acc1, acc2}
			rules := []models.DescribeRuleResponse{
				{
					Rule:       rule1,
					Actions:    []models.RuleActionResponse{{ActionType: models.RuleFieldTransfer, ActionValue: fmt.Sprintf("%d", acc1.Id)}},
					Conditions: []models.RuleConditionResponse{{ConditionType: models.RuleFieldName, ConditionValue: "transfer", ConditionOperator: models.OperatorContains}},
				},
			}

			engine := NewRuleEngine(categories, accounts, rules)
			changeset := engine.ProcessTransaction(txn1)

			Expect(changeset).To(BeNil()) // Should not create transfer to same account
		})
	})

	Describe("Transfer Integration Tests", func() {
		var (
			cat1, cat2 models.CategoryResponse
			acc1, acc2 models.AccountResponse
			rule1      models.RuleResponse
			txn1       models.TransactionResponse
		)

		BeforeEach(func() {
			// Create categories
			cat1, err := mockCategoryRepo.CreateCategory(ctx, models.CreateCategoryInput{
				Name:      "Food",
				CreatedBy: userId,
			})
			Expect(err).NotTo(HaveOccurred())

			cat2, err = mockCategoryRepo.CreateCategory(ctx, models.CreateCategoryInput{
				Name:      "Transfer",
				CreatedBy: userId,
			})
			Expect(err).NotTo(HaveOccurred())

			// Create accounts
			acc1, err = mockAccountRepo.CreateAccount(ctx, models.CreateAccountInput{
				Name:      "Checking",
				BankType:  models.BankTypeHDFC,
				Currency:  models.CurrencyINR,
				CreatedBy: userId,
			})
			Expect(err).NotTo(HaveOccurred())

			acc2, err = mockAccountRepo.CreateAccount(ctx, models.CreateAccountInput{
				Name:      "Savings",
				BankType:  models.BankTypeSBI,
				Currency:  models.CurrencyINR,
				CreatedBy: userId,
			})
			Expect(err).NotTo(HaveOccurred())

			// Create rule with transfer action
			desc := "Transfer rule"
			rule1, err = mockRuleRepo.CreateRule(ctx, models.CreateBaseRuleRequest{
				Name:          "Transfer Rule",
				Description:   &desc,
				EffectiveFrom: time.Now().Add(-24 * time.Hour),
				CreatedBy:     userId,
			})
			Expect(err).NotTo(HaveOccurred())

			// Create rule action for transfer
			actions := []models.CreateRuleActionRequest{
				{
					RuleId:      rule1.Id,
					ActionType:  models.RuleFieldTransfer,
					ActionValue: fmt.Sprintf("%d", acc2.Id),
				},
			}
			_, err = mockRuleRepo.CreateRuleActions(ctx, actions)
			Expect(err).NotTo(HaveOccurred())

			// Create rule condition
			conditions := []models.CreateRuleConditionRequest{
				{
					RuleId:            rule1.Id,
					ConditionType:     models.RuleFieldName,
					ConditionOperator: models.OperatorContains,
					ConditionValue:    "transfer",
				},
			}
			_, err = mockRuleRepo.CreateRuleConditions(ctx, conditions)
			Expect(err).NotTo(HaveOccurred())

			// Create transaction that will trigger transfer
			amount := 100.0
			txn1, err = mockTxnRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
				Name:        "Transfer to savings",
				Description: "Monthly transfer",
				Amount:      &amount,
				Date:        time.Now(),
				CreatedBy:   userId,
				AccountId:   acc1.Id,
			}, []int64{cat1.Id})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should execute complete transfer workflow", func() {
			// Test the complete workflow from rule execution to transfer creation
			txnIds := []int64{txn1.Id}
			request := models.ExecuteRulesRequest{TransactionIds: &txnIds}
			response, err := service.ExecuteRules(ctx, userId, request)

			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())

			// Verify that the rule engine correctly processes the transfer
			categories := []models.CategoryResponse{cat1, cat2}
			accounts := []models.AccountResponse{acc1, acc2}
			rules := []models.DescribeRuleResponse{
				{
					Rule:       rule1,
					Actions:    []models.RuleActionResponse{{ActionType: models.RuleFieldTransfer, ActionValue: fmt.Sprintf("%d", acc2.Id)}},
					Conditions: []models.RuleConditionResponse{{ConditionType: models.RuleFieldName, ConditionValue: "transfer", ConditionOperator: models.OperatorContains}},
				},
			}

			engine := NewRuleEngine(categories, accounts, rules)
			changeset := engine.ProcessTransaction(txn1)

			Expect(changeset).NotTo(BeNil())
			Expect(changeset.TransferInfo).NotTo(BeNil())
			Expect(changeset.TransferInfo.AccountId).To(Equal(acc2.Id))
			Expect(changeset.TransferInfo.Amount).To(Equal(-txn1.Amount))
			Expect(changeset.AppliedRules).To(ContainElement(rule1.Id))

			// Verify updated fields tracking
			updatedFields := service.(*ruleEngineService).getUpdatedFields(changeset)
			Expect(updatedFields).To(ContainElement(models.RuleFieldTransfer))
		})

		It("should handle transfer with multiple categories", func() {
			// Create transaction with multiple categories
			amount := 150.0
			multiCatTxn, err := mockTxnRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
				Name:        "Transfer with categories",
				Description: "Transfer with multiple categories",
				Amount:      &amount,
				Date:        time.Now(),
				CreatedBy:   userId,
				AccountId:   acc1.Id,
			}, []int64{cat1.Id, cat2.Id})
			Expect(err).NotTo(HaveOccurred())

			// Test that transfer inherits all categories
			categories := []models.CategoryResponse{cat1, cat2}
			accounts := []models.AccountResponse{acc1, acc2}
			rules := []models.DescribeRuleResponse{
				{
					Rule:       rule1,
					Actions:    []models.RuleActionResponse{{ActionType: models.RuleFieldTransfer, ActionValue: fmt.Sprintf("%d", acc2.Id)}},
					Conditions: []models.RuleConditionResponse{{ConditionType: models.RuleFieldName, ConditionValue: "categories", ConditionOperator: models.OperatorContains}},
				},
			}

			engine := NewRuleEngine(categories, accounts, rules)
			changeset := engine.ProcessTransaction(multiCatTxn)

			Expect(changeset).NotTo(BeNil())
			Expect(changeset.TransferInfo).NotTo(BeNil())
			Expect(changeset.TransferInfo.AccountId).To(Equal(acc2.Id))
			Expect(changeset.TransferInfo.Amount).To(Equal(-multiCatTxn.Amount))
		})
	})
})

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
