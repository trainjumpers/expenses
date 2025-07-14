package service_test

import (
	"fmt"
	"time"

	mock_repository "expenses/internal/mock/repository"
	"expenses/internal/models"
	"expenses/internal/service"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RuleEngineService", func() {
	var (
		ruleEngineService service.RuleEngineServiceInterface
		mockRuleRepo      *mock_repository.MockRuleRepository
		mockTxnRepo       *mock_repository.MockTransactionRepository
		mockCategoryRepo  *mock_repository.MockCategoryRepository
		ctx               *gin.Context
		userId            int64
		now               time.Time
	)

	BeforeEach(func() {
		ctx = &gin.Context{}
		mockRuleRepo = mock_repository.NewMockRuleRepository()
		mockTxnRepo = mock_repository.NewMockTransactionRepository()
		mockCategoryRepo = mock_repository.NewMockCategoryRepository()
		ruleEngineService = service.NewRuleEngineService(mockRuleRepo, mockTxnRepo, mockCategoryRepo)
		userId = 1
		now = time.Now()
	})

	Describe("ExecuteRulesForRule", func() {
		Context("when a rule matches multiple transactions", func() {
			var ruleId int64
			const txnCount = 3
			BeforeEach(func() {
				// Create a category for the user
				cat, _ := mockCategoryRepo.CreateCategory(ctx, models.CreateCategoryInput{
					Name:      "Food",
					Icon:      "",
					CreatedBy: userId,
				})
				// Create multiple transactions for the user
				for i := range txnCount {
					amount := 150.0 + float64(i)
					mockTxnRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
						Name:        "Grocery Store",
						Description: "Weekly groceries",
						Amount:      &amount,
						Date:        now.Add(-24 * time.Hour),
						CreatedBy:   userId,
						AccountId:   1,
					}, []int64{})
				}
				// Create a rule that matches all transactions
				rule, _ := mockRuleRepo.CreateRule(ctx, models.CreateBaseRuleRequest{
					Name:          "Grocery Rule",
					EffectiveFrom: now.Add(-48 * time.Hour),
					CreatedBy:     userId,
				})
				ruleId = rule.Id
				mockRuleRepo.CreateRuleActions(ctx, []models.CreateRuleActionRequest{
					{
						ActionType:  models.RuleFieldCategory,
						ActionValue: fmt.Sprintf("%d", cat.Id),
						RuleId:      rule.Id,
					},
				})
				mockRuleRepo.CreateRuleConditions(ctx, []models.CreateRuleConditionRequest{
					{
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "100.00",
						ConditionOperator: models.OperatorGreater,
						RuleId:            rule.Id,
					},
				})
			})

			It("should apply the rule to all transactions and modify them", func() {
				resp, err := ruleEngineService.ExecuteRulesForRule(ctx, ruleId, userId)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.TotalRules).To(Equal(1))
				Expect(resp.ProcessedTxns).To(Equal(txnCount))
				Expect(resp.Modified).To(HaveLen(txnCount))
				for _, mod := range resp.Modified {
					Expect(mod.UpdatedFields).To(ContainElement(models.RuleFieldCategory))
				}
				Expect(resp.Skipped).To(BeEmpty())
			})
		})
	})

	Describe("ExecuteRulesForRule", func() {
		Context("when a rule matches multiple transactions", func() {
			var ruleId int64
			const txnCount = 3
			BeforeEach(func() {
				// Create a category for the user
				cat, _ := mockCategoryRepo.CreateCategory(ctx, models.CreateCategoryInput{
					Name:      "Food",
					Icon:      "",
					CreatedBy: userId,
				})
				// Create multiple transactions for the user
				for i := range txnCount {
					amount := 150.0 + float64(i)
					mockTxnRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
						Name:        "Grocery Store",
						Description: "Weekly groceries",
						Amount:      &amount,
						Date:        now.Add(-24 * time.Hour),
						CreatedBy:   userId,
						AccountId:   1,
					}, []int64{})
				}
				// Create a rule that matches all transactions
				rule, _ := mockRuleRepo.CreateRule(ctx, models.CreateBaseRuleRequest{
					Name:          "Grocery Rule",
					EffectiveFrom: now.Add(-48 * time.Hour),
					CreatedBy:     userId,
				})
				ruleId = rule.Id
				mockRuleRepo.CreateRuleActions(ctx, []models.CreateRuleActionRequest{
					{
						ActionType:  models.RuleFieldCategory,
						ActionValue: fmt.Sprintf("%d", cat.Id),
						RuleId:      rule.Id,
					},
				})
				mockRuleRepo.CreateRuleConditions(ctx, []models.CreateRuleConditionRequest{
					{
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "100.00",
						ConditionOperator: models.OperatorGreater,
						RuleId:            rule.Id,
					},
				})
			})

			It("should apply the rule to all transactions and modify them", func() {
				resp, err := ruleEngineService.ExecuteRulesForRule(ctx, ruleId, userId)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.TotalRules).To(Equal(1))
				Expect(resp.ProcessedTxns).To(Equal(txnCount))
				Expect(resp.Modified).To(HaveLen(txnCount))
				for _, mod := range resp.Modified {
					Expect(mod.UpdatedFields).To(ContainElement(models.RuleFieldCategory))
				}
				Expect(resp.Skipped).To(BeEmpty())
			})
		})
	})

	Describe("ExecuteRules", func() {
		Context("when there are no rules for the user", func() {
			It("should return a zeroed response", func() {
				req := models.ExecuteRulesRequest{
					PageSize: 100,
				}
				resp, err := ruleEngineService.ExecuteRules(ctx, userId, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.TotalRules).To(Equal(0))
				Expect(resp.ProcessedTxns).To(Equal(0))
				Expect(resp.Modified).To(BeEmpty())
				Expect(resp.Skipped).To(BeEmpty())
			})
		})

		Context("when RuleIds are provided in the request", func() {
			var ruleIds []int64
			const txnCount = 3
			BeforeEach(func() {
				// Create a category
				cat, _ := mockCategoryRepo.CreateCategory(ctx, models.CreateCategoryInput{
					Name:      "Groceries",
					Icon:      "",
					CreatedBy: userId,
				})
				// Create 2 rules, but only 1 will be used in RuleIds
				rule1, _ := mockRuleRepo.CreateRule(ctx, models.CreateBaseRuleRequest{
					Name:          "Rule 1",
					EffectiveFrom: now.Add(-time.Hour),
					CreatedBy:     userId,
				})
				rule2, _ := mockRuleRepo.CreateRule(ctx, models.CreateBaseRuleRequest{
					Name:          "Rule 2",
					EffectiveFrom: now.Add(-time.Hour),
					CreatedBy:     userId,
				})
				ruleIds = []int64{rule2.Id}
				mockRuleRepo.CreateRuleActions(ctx, []models.CreateRuleActionRequest{
					{
						ActionType:  models.RuleFieldCategory,
						ActionValue: fmt.Sprintf("%d", cat.Id),
						RuleId:      rule1.Id,
					},
				})
				mockRuleRepo.CreateRuleActions(ctx, []models.CreateRuleActionRequest{
					{
						ActionType:  models.RuleFieldCategory,
						ActionValue: fmt.Sprintf("%d", cat.Id),
						RuleId:      rule2.Id,
					},
				})
				mockRuleRepo.CreateRuleConditions(ctx, []models.CreateRuleConditionRequest{
					{
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "10.00",
						ConditionOperator: models.OperatorGreater,
						RuleId:            rule1.Id,
					},
				})
				mockRuleRepo.CreateRuleConditions(ctx, []models.CreateRuleConditionRequest{
					{
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "10.00",
						ConditionOperator: models.OperatorGreater,
						RuleId:            rule2.Id,
					},
				})
				// Create transactions
				for i := 0; i < txnCount; i++ {
					amt := 20.0 + float64(i)
					mockTxnRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
						Name:        fmt.Sprintf("Txn %d", i),
						Description: "RuleId Transaction",
						Amount:      &amt,
						Date:        now.Add(-time.Duration(i) * time.Hour),
						CreatedBy:   userId,
						AccountId:   1,
					}, []int64{})
				}
			})
			It("should only process the specified rules", func() {
				req := models.ExecuteRulesRequest{
					PageSize: 100,
					RuleIds:  &ruleIds,
				}
				resp, err := ruleEngineService.ExecuteRules(ctx, userId, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.TotalRules).To(Equal(1))
				Expect(resp.ProcessedTxns).To(Equal(txnCount))
				Expect(resp.Modified).To(HaveLen(txnCount))
				for _, mod := range resp.Modified {
					Expect(mod.UpdatedFields).To(ContainElement(models.RuleFieldCategory))
				}
			})
		})

		Context("when TransactionIds are provided in the request", func() {
			var txnIds []int64
			BeforeEach(func() {
				// Create a category and a rule that matches all transactions
				cat, _ := mockCategoryRepo.CreateCategory(ctx, models.CreateCategoryInput{
					Name:      "Groceries",
					Icon:      "",
					CreatedBy: userId,
				})
				rule, _ := mockRuleRepo.CreateRule(ctx, models.CreateBaseRuleRequest{
					Name:          "All Match Rule",
					EffectiveFrom: now.Add(-48 * time.Hour),
					CreatedBy:     userId,
				})
				mockRuleRepo.CreateRuleActions(ctx, []models.CreateRuleActionRequest{
					{
						ActionType:  models.RuleFieldCategory,
						ActionValue: fmt.Sprintf("%d", cat.Id),
						RuleId:      rule.Id,
					},
				})
				mockRuleRepo.CreateRuleConditions(ctx, []models.CreateRuleConditionRequest{
					{
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "10.00",
						ConditionOperator: models.OperatorGreater,
						RuleId:            rule.Id,
					},
				})
				// Create 3 transactions, but only select 2 for processing
				txnIds = make([]int64, 0, 2)
				for i := 0; i < 3; i++ {
					amt := 20.0 + float64(i)
					txn, _ := mockTxnRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
						Name:        fmt.Sprintf("Txn %d", i),
						Description: "Specific Transaction",
						Amount:      &amt,
						Date:        now.Add(-time.Duration(i) * time.Hour),
						CreatedBy:   userId,
						AccountId:   1,
					}, []int64{})
					if i < 2 {
						txnIds = append(txnIds, txn.Id)
					}
				}
			})

			It("should only process the specified transactions", func() {
				req := models.ExecuteRulesRequest{
					PageSize:       100,
					TransactionIds: &txnIds,
				}
				resp, err := ruleEngineService.ExecuteRules(ctx, userId, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.TotalRules).To(Equal(1))
				Expect(resp.ProcessedTxns).To(Equal(2))
				Expect(resp.Modified).To(HaveLen(2))
				for _, mod := range resp.Modified {
					Expect(mod.UpdatedFields).To(ContainElement(models.RuleFieldCategory))
				}
			})
		})

		Context("when there are no transactions for the user", func() {
			BeforeEach(func() {
				rule, _ := mockRuleRepo.CreateRule(ctx, models.CreateBaseRuleRequest{
					Name:          "Test Rule",
					EffectiveFrom: now.Add(-24 * time.Hour),
					CreatedBy:     userId,
				})
				mockRuleRepo.CreateRuleActions(ctx, []models.CreateRuleActionRequest{
					{
						ActionType:  models.RuleFieldName,
						ActionValue: "Updated Name",
						RuleId:      rule.Id,
					},
				})
				mockRuleRepo.CreateRuleConditions(ctx, []models.CreateRuleConditionRequest{
					{
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "100.00",
						ConditionOperator: models.OperatorGreater,
						RuleId:            rule.Id,
					},
				})
			})

			It("should return a zeroed response", func() {
				req := models.ExecuteRulesRequest{
					PageSize: 100,
				}
				resp, err := ruleEngineService.ExecuteRules(ctx, userId, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.TotalRules).To(Equal(1))
				Expect(resp.ProcessedTxns).To(Equal(0))
				Expect(resp.Modified).To(BeEmpty())
				Expect(resp.Skipped).To(BeEmpty())
			})
		})

		Context("when RuleEngine returns skipped transactions", func() {
			var txnId int64
			BeforeEach(func() {
				// Create a transaction for the user
				amount := 50.0
				txn, _ := mockTxnRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
					Name:        "Skipped Transaction",
					Description: "Should be skipped",
					Amount:      &amount,
					Date:        now.Add(-24 * time.Hour),
					CreatedBy:   userId,
					AccountId:   1,
				}, []int64{})
				txnId = txn.Id
				// Create a rule with an effective date after the transaction date (so it will be skipped)
				rule, _ := mockRuleRepo.CreateRule(ctx, models.CreateBaseRuleRequest{
					Name:          "Future Rule",
					EffectiveFrom: now.Add(24 * time.Hour), // future date
					CreatedBy:     userId,
				})
				mockRuleRepo.CreateRuleActions(ctx, []models.CreateRuleActionRequest{
					{
						ActionType:  models.RuleFieldName,
						ActionValue: "Should Not Apply",
						RuleId:      rule.Id,
					},
				})
				mockRuleRepo.CreateRuleConditions(ctx, []models.CreateRuleConditionRequest{
					{
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "50.00",
						ConditionOperator: models.OperatorEquals,
						RuleId:            rule.Id,
					},
				})
			})

			It("should include skipped transactions in the response", func() {
				req := models.ExecuteRulesRequest{
					PageSize: 100,
				}
				resp, err := ruleEngineService.ExecuteRules(ctx, userId, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.TotalRules).To(Equal(1))
				Expect(resp.ProcessedTxns).To(Equal(1))
				Expect(resp.Modified).To(BeEmpty())
				Expect(resp.Skipped).To(HaveLen(1))
				Expect(resp.Skipped[0].TransactionId).To(Equal(txnId))
				// Reason may be empty if RuleEngine doesn't set it, but should exist in the struct
			})
		})

		Context("when rules do not match any transactions", func() {
			BeforeEach(func() {
				amount := 50.0
				mockTxnRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
					Name:        "Unmatched Transaction",
					Description: "No rule matches this",
					Amount:      &amount,
					Date:        now.Add(-24 * time.Hour),
					CreatedBy:   userId,
					AccountId:   1,
				}, []int64{})
				rule, _ := mockRuleRepo.CreateRule(ctx, models.CreateBaseRuleRequest{
					Name:          "No Match Rule",
					EffectiveFrom: now.Add(-48 * time.Hour),
					CreatedBy:     userId,
				})
				mockRuleRepo.CreateRuleActions(ctx, []models.CreateRuleActionRequest{
					{
						ActionType:  models.RuleFieldName,
						ActionValue: "Should Not Apply",
						RuleId:      rule.Id,
					},
				})
				mockRuleRepo.CreateRuleConditions(ctx, []models.CreateRuleConditionRequest{
					{
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "9999.99", // No transaction has this amount
						ConditionOperator: models.OperatorEquals,
						RuleId:            rule.Id,
					},
				})
			})

			It("should not apply any changes and response should show no modifications", func() {
				req := models.ExecuteRulesRequest{
					PageSize: 100,
				}
				resp, err := ruleEngineService.ExecuteRules(ctx, userId, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.TotalRules).To(Equal(1))
				Expect(resp.ProcessedTxns).To(Equal(1))
				Expect(resp.Modified).To(BeEmpty())
				Expect(resp.Skipped).To(BeEmpty())
			})
		})
		Context("when a rule applies all changeset types (name, description, category)", func() {
			BeforeEach(func() {
				// Create a category for the user
				cat, _ := mockCategoryRepo.CreateCategory(ctx, models.CreateCategoryInput{
					Name:      "Food",
					Icon:      "",
					CreatedBy: userId,
				})
				// Create a transaction for the user
				amount := 150.0
				mockTxnRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
					Name:        "Grocery Store",
					Description: "Weekly groceries",
					Amount:      &amount,
					Date:        now.Add(-24 * time.Hour),
					CreatedBy:   userId,
					AccountId:   1,
				}, []int64{})
				// Create a rule that matches the transaction and updates all fields
				rule, _ := mockRuleRepo.CreateRule(ctx, models.CreateBaseRuleRequest{
					Name:          "Grocery Rule",
					EffectiveFrom: now.Add(-time.Hour),
					CreatedBy:     userId,
				})
				mockRuleRepo.CreateRuleActions(ctx, []models.CreateRuleActionRequest{
					{
						ActionType:  models.RuleFieldName,
						ActionValue: "Updated Name",
						RuleId:      rule.Id,
					},
					{
						ActionType:  models.RuleFieldDescription,
						ActionValue: "Updated Description",
						RuleId:      rule.Id,
					},
					{
						ActionType:  models.RuleFieldCategory,
						ActionValue: fmt.Sprintf("%d", cat.Id),
						RuleId:      rule.Id,
					},
				})
				mockRuleRepo.CreateRuleConditions(ctx, []models.CreateRuleConditionRequest{
					{
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "100.00",
						ConditionOperator: models.OperatorGreater,
						RuleId:            rule.Id,
					},
				})
			})

			It("should apply name, description, and category updates", func() {
				req := models.ExecuteRulesRequest{
					PageSize: 100,
				}
				resp, err := ruleEngineService.ExecuteRules(ctx, userId, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.TotalRules).To(Equal(1))
				Expect(resp.ProcessedTxns).To(Equal(1))
				Expect(resp.Modified).To(HaveLen(1))
				mod := resp.Modified[0]
				Expect(mod.UpdatedFields).To(ContainElement(models.RuleFieldName))
				Expect(mod.UpdatedFields).To(ContainElement(models.RuleFieldDescription))
				Expect(mod.UpdatedFields).To(ContainElement(models.RuleFieldCategory))
			})
		})

		Context("when there are more transactions than the page size", func() {
			const pageSize = 5
			const totalTxns = 12

			BeforeEach(func() {
				testNow := time.Now()
				cat, _ := mockCategoryRepo.CreateCategory(ctx, models.CreateCategoryInput{
					Name:      "Groceries",
					Icon:      "",
					CreatedBy: userId,
				})
				rule, _ := mockRuleRepo.CreateRule(ctx, models.CreateBaseRuleRequest{
					Name:          "All Match Rule",
					EffectiveFrom: testNow.Add(-48 * time.Hour),
					CreatedBy:     userId,
				})
				mockRuleRepo.CreateRuleActions(ctx, []models.CreateRuleActionRequest{
					{
						ActionType:  models.RuleFieldCategory,
						ActionValue: fmt.Sprintf("%d", cat.Id),
						RuleId:      rule.Id,
					},
				})
				mockRuleRepo.CreateRuleConditions(ctx, []models.CreateRuleConditionRequest{
					{
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "10.00",
						ConditionOperator: models.OperatorGreater,
						RuleId:            rule.Id,
					},
				})

				// Create more transactions than the page size
				for i := 0; i < totalTxns; i++ {
					amt := 20.0 + float64(i)
					mockTxnRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
						Name:        "Txn " + fmt.Sprintf("%d", int64(i)),
						Description: "Paged Transaction",
						Amount:      &amt,
						Date:        testNow.Add(-time.Duration(i) * time.Hour),
						CreatedBy:   userId,
						AccountId:   1,
					}, []int64{})
				}
			})

			It("should fetch and process all pages of transactions", func() {
				req := models.ExecuteRulesRequest{
					PageSize: pageSize,
				}
				resp, err := ruleEngineService.ExecuteRules(ctx, userId, req)
				fmt.Printf("DEBUG: ExecuteRules response: TotalRules=%d, ProcessedTxns=%d, Modified=%d, Skipped=%d\n", resp.TotalRules, resp.ProcessedTxns, len(resp.Modified), len(resp.Skipped))
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.TotalRules).To(Equal(1))
				Expect(resp.ProcessedTxns).To(Equal(totalTxns))
				Expect(resp.Modified).To(HaveLen(totalTxns))
				for _, mod := range resp.Modified {
					fmt.Printf("DEBUG: Modified transaction: %+v\n", mod)
					Expect(mod.UpdatedFields).To(ContainElement(models.RuleFieldCategory))
				}
			})
		})
	})

	Describe("ExecuteRulesForTransaction", func() {
		Context("when transaction does not exist", func() {
			It("should return an error", func() {
				_, err := ruleEngineService.ExecuteRulesForTransaction(ctx, 999, userId)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to fetch transaction"))
			})
		})
	})

	Describe("ExecuteRulesForRule", func() {
		Context("when rule does not exist", func() {
			It("should return a response with zero processed transactions", func() {
				resp, err := ruleEngineService.ExecuteRulesForRule(ctx, 999, userId)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.TotalRules).To(Equal(0))
				Expect(resp.ProcessedTxns).To(Equal(0))
			})
		})
	})

	Describe("ExecuteRulesForTransaction", func() {
		Context("when a transaction matches a rule", func() {
			var txnId int64
			BeforeEach(func() {
				// Create a category for the user
				cat, _ := mockCategoryRepo.CreateCategory(ctx, models.CreateCategoryInput{
					Name:      "Food",
					Icon:      "",
					CreatedBy: userId,
				})
				// Create a transaction for the user
				amount := 150.0
				txn, _ := mockTxnRepo.CreateTransaction(ctx, models.CreateBaseTransactionInput{
					Name:        "Grocery Store",
					Description: "Weekly groceries",
					Amount:      &amount,
					Date:        now.Add(-24 * time.Hour),
					CreatedBy:   userId,
					AccountId:   1,
				}, []int64{})
				txnId = txn.Id
				// Create a rule that matches the transaction
				rule, _ := mockRuleRepo.CreateRule(ctx, models.CreateBaseRuleRequest{
					Name:          "Grocery Rule",
					EffectiveFrom: now.Add(-time.Hour),
					CreatedBy:     userId,
				})
				mockRuleRepo.CreateRuleActions(ctx, []models.CreateRuleActionRequest{
					{
						ActionType:  models.RuleFieldCategory,
						ActionValue: fmt.Sprintf("%d", cat.Id),
						RuleId:      rule.Id,
					},
				})
				mockRuleRepo.CreateRuleConditions(ctx, []models.CreateRuleConditionRequest{
					{
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "100.00",
						ConditionOperator: models.OperatorGreater,
						RuleId:            rule.Id,
					},
				})
			})

			It("should apply rules to the transaction and modify it", func() {
				resp, err := ruleEngineService.ExecuteRulesForTransaction(ctx, txnId, userId)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.TotalRules).To(Equal(1))
				Expect(resp.ProcessedTxns).To(Equal(1))
				Expect(resp.Modified).To(HaveLen(1))
				Expect(resp.Modified[0].UpdatedFields).To(ContainElement(models.RuleFieldCategory))
				Expect(resp.Skipped).To(BeEmpty())
			})
		})
	})
})
