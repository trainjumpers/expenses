package service

import (
	"context"
	mock_repository "expenses/internal/mock/repository"
	"expenses/internal/models"

	"time"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RuleService CRUD", func() {
	var (
		repo    *mock_repository.MockRuleRepository
		service RuleServiceInterface
		ctx     *gin.Context
	)

	BeforeEach(func() {
		repo = mock_repository.NewMockRuleRepository()
		service = NewRuleService(repo, nil)
		ctx = &gin.Context{}
	})

	Describe("CreateRule", func() {
		It("should create a rule successfully", func() {
			createReq := &models.CreateRuleRequest{BaseRule: models.BaseRule{Name: "Test Rule"}}
			resp, err := service.CreateRule(ctx, createReq)
			Expect(err).To(BeNil())
			Expect(resp.ID).NotTo(BeZero())
			Expect(resp.BaseRule.Name).To(Equal("Test Rule"))
		})
	})

	Describe("GetRuleByID", func() {
		It("should get a rule by ID", func() {
			createReq := &models.CreateRuleRequest{BaseRule: models.BaseRule{Name: "R2"}}
			rule, _ := repo.CreateRule(context.Background(), createReq)
			fetched, err := service.GetRuleByID(ctx, rule.ID)
			Expect(err).To(BeNil())
			Expect(fetched).To(Equal(rule))
		})
		It("should return error if not found", func() {
			fetched, err := service.GetRuleByID(ctx, 9999)
			Expect(err).To(HaveOccurred())
			Expect(fetched).To(BeNil())
		})
	})

	Describe("ListRules", func() {
		It("should return all rules", func() {
			_, _ = repo.CreateRule(context.Background(), &models.CreateRuleRequest{BaseRule: models.BaseRule{Name: "A"}})
			_, _ = repo.CreateRule(context.Background(), &models.CreateRuleRequest{BaseRule: models.BaseRule{Name: "B"}})
			rules, err := service.ListRules(ctx)
			Expect(err).To(BeNil())
			Expect(len(rules)).To(Equal(2))
		})
	})

	Describe("UpdateRule", func() {
		It("should update a rule successfully", func() {
			createReq := &models.CreateRuleRequest{BaseRule: models.BaseRule{Name: "Old"}}
			rule, _ := repo.CreateRule(context.Background(), createReq)
			updateReq := &models.UpdateRuleRequest{ID: rule.ID, BaseRule: models.BaseRule{Name: "New"}}
			err := service.UpdateRule(ctx, rule.ID, updateReq)
			Expect(err).To(BeNil())
			updated, _ := repo.GetRuleByID(context.Background(), rule.ID)
			Expect(updated.BaseRule.Name).To(Equal("New"))
		})
		It("should return error if repo fails", func() {
			err := service.UpdateRule(ctx, 9999, &models.UpdateRuleRequest{ID: 9999, BaseRule: models.BaseRule{Name: "X"}})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("DeleteRule", func() {
		It("should delete a rule successfully", func() {
			createReq := &models.CreateRuleRequest{BaseRule: models.BaseRule{Name: "ToDelete"}}
			rule, _ := repo.CreateRule(context.Background(), createReq)
			err := service.DeleteRule(ctx, rule.ID)
			Expect(err).To(BeNil())
			_, err = repo.GetRuleByID(context.Background(), rule.ID)
			Expect(err).To(HaveOccurred())
		})
		It("should return error if repo fails", func() {
			err := service.DeleteRule(ctx, 9999)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ExecuteRules", func() {
		var (
			userId int64
			txRepo *mock_repository.MockTransactionRepository
		)

		BeforeEach(func() {
			userId = 42
			txRepo = mock_repository.NewMockTransactionRepository()
			service = NewRuleService(repo, txRepo)
		})

		It("should return empty modified/skipped if no rules and no transactions", func() {
			resp, err := service.ExecuteRules(ctx, userId)
			Expect(err).To(BeNil())
			Expect(resp.Modified).To(BeEmpty())
			Expect(resp.Skipped).To(BeEmpty())
		})

		It("should skip all transactions if no rules exist", func() {
			amount := 100.0
			_ = createTestTransaction(txRepo, userId, "T1", amount, nil, nil)
			_ = createTestTransaction(txRepo, userId, "T2", 200.0, nil, nil)
			resp, err := service.ExecuteRules(ctx, userId)
			Expect(err).To(BeNil())
			Expect(resp.Modified).To(BeEmpty())
			Expect(resp.Skipped).To(HaveLen(2))
			for _, s := range resp.Skipped {
				Expect(s.Reason).To(Equal("No matching rule"))
			}
		})

		It("should modify a transaction if a rule matches by amount", func() {
			amount := 123.45
			tx := createTestTransaction(txRepo, userId, "Lunch", amount, nil, nil)
			// Rule: if amount == 123.45, set name to "Updated"
			ruleReq := &models.CreateRuleRequest{
				BaseRule:   models.BaseRule{Name: "AmountRule", EffectiveFrom: tx.Date},
				CreatedBy:  userId,
				Actions:    []models.CreateRuleActionRequest{{BaseRuleAction: models.BaseRuleAction{ActionType: models.RuleFieldName, ActionValue: "Updated"}}},
				Conditions: []models.CreateRuleConditionRequest{{BaseRuleCondition: models.BaseRuleCondition{ConditionType: models.RuleFieldAmount, ConditionValue: "123.45", ConditionOperator: models.OperatorEquals}}},
			}
			_, _ = repo.CreateRule(context.Background(), ruleReq)
			resp, err := service.ExecuteRules(ctx, userId)
			Expect(err).To(BeNil())
			Expect(resp.Modified).To(HaveLen(1))
			Expect(resp.Modified[0].TransactionID).To(Equal(tx.Id))
			Expect(resp.Modified[0].AppliedRules).To(HaveLen(1))
			Expect(resp.Modified[0].UpdatedFields).To(ContainElement(models.RuleFieldName))
			// Confirm transaction was updated
			updated, _ := txRepo.GetTransactionById(ctx, tx.Id, userId)
			Expect(updated.Name).To(Equal("Updated"))
		})

		It("should modify a transaction if a rule matches by name contains", func() {
			amount := 50.0
			tx := createTestTransaction(txRepo, userId, "Dinner Special", amount, nil, nil)
			ruleReq := &models.CreateRuleRequest{
				BaseRule:   models.BaseRule{Name: "NameContains", EffectiveFrom: tx.Date},
				CreatedBy:  userId,
				Actions:    []models.CreateRuleActionRequest{{BaseRuleAction: models.BaseRuleAction{ActionType: models.RuleFieldDescription, ActionValue: "Matched!"}}},
				Conditions: []models.CreateRuleConditionRequest{{BaseRuleCondition: models.BaseRuleCondition{ConditionType: models.RuleFieldName, ConditionValue: "Dinner", ConditionOperator: models.OperatorContains}}},
			}
			_, _ = repo.CreateRule(context.Background(), ruleReq)
			resp, err := service.ExecuteRules(ctx, userId)
			Expect(err).To(BeNil())
			Expect(resp.Modified).To(HaveLen(1))
			updated, _ := txRepo.GetTransactionById(ctx, tx.Id, userId)
			Expect(updated.Description).NotTo(BeNil())
			Expect(*updated.Description).To(Equal("Matched!"))
		})

		It("should apply multiple rules to a transaction", func() {
			amount := 10.0
			tx := createTestTransaction(txRepo, userId, "Coffee", amount, nil, nil)
			// Rule 1: if name == Coffee, set amount to 20
			rule1 := &models.CreateRuleRequest{
				BaseRule:   models.BaseRule{Name: "CoffeeRule", EffectiveFrom: tx.Date},
				CreatedBy:  userId,
				Actions:    []models.CreateRuleActionRequest{{BaseRuleAction: models.BaseRuleAction{ActionType: models.RuleFieldAmount, ActionValue: "20"}}},
				Conditions: []models.CreateRuleConditionRequest{{BaseRuleCondition: models.BaseRuleCondition{ConditionType: models.RuleFieldName, ConditionValue: "Coffee", ConditionOperator: models.OperatorEquals}}},
			}
			// Rule 2: if amount == 10, set description to "Tenner"
			rule2 := &models.CreateRuleRequest{
				BaseRule:   models.BaseRule{Name: "TennerRule", EffectiveFrom: tx.Date},
				CreatedBy:  userId,
				Actions:    []models.CreateRuleActionRequest{{BaseRuleAction: models.BaseRuleAction{ActionType: models.RuleFieldDescription, ActionValue: "Tenner"}}},
				Conditions: []models.CreateRuleConditionRequest{{BaseRuleCondition: models.BaseRuleCondition{ConditionType: models.RuleFieldAmount, ConditionValue: "10", ConditionOperator: models.OperatorEquals}}},
			}
			_, _ = repo.CreateRule(context.Background(), rule1)
			_, _ = repo.CreateRule(context.Background(), rule2)
			resp, err := service.ExecuteRules(ctx, userId)
			Expect(err).To(BeNil())
			Expect(resp.Modified).To(HaveLen(1))
			Expect(resp.Modified[0].AppliedRules).To(HaveLen(2))
			updated, _ := txRepo.GetTransactionById(ctx, tx.Id, userId)
			Expect(updated.Amount).To(Equal(20.0))
			Expect(*updated.Description).To(Equal("Tenner"))
		})

		It("should skip transactions that do not match any rule", func() {
			tx := createTestTransaction(txRepo, userId, "NoMatch", 1.0, nil, nil)
			ruleReq := &models.CreateRuleRequest{
				BaseRule:   models.BaseRule{Name: "Unrelated", EffectiveFrom: tx.Date},
				CreatedBy:  userId,
				Actions:    []models.CreateRuleActionRequest{{BaseRuleAction: models.BaseRuleAction{ActionType: models.RuleFieldName, ActionValue: "ShouldNotApply"}}},
				Conditions: []models.CreateRuleConditionRequest{{BaseRuleCondition: models.BaseRuleCondition{ConditionType: models.RuleFieldAmount, ConditionValue: "999", ConditionOperator: models.OperatorEquals}}},
			}
			_, _ = repo.CreateRule(context.Background(), ruleReq)
			resp, err := service.ExecuteRules(ctx, userId)
			Expect(err).To(BeNil())
			Expect(resp.Modified).To(BeEmpty())
			Expect(resp.Skipped).To(HaveLen(1))
			Expect(resp.Skipped[0].TransactionID).To(Equal(tx.Id))
		})

		It("should apply category update action", func() {
			amount := 77.0
			catId := int64(123)
			tx := createTestTransaction(txRepo, userId, "CatTx", amount, nil, nil)
			ruleReq := &models.CreateRuleRequest{
				BaseRule:   models.BaseRule{Name: "CatRule", EffectiveFrom: tx.Date},
				CreatedBy:  userId,
				Actions:    []models.CreateRuleActionRequest{{BaseRuleAction: models.BaseRuleAction{ActionType: models.RuleFieldCategory, ActionValue: "123"}}},
				Conditions: []models.CreateRuleConditionRequest{{BaseRuleCondition: models.BaseRuleCondition{ConditionType: models.RuleFieldAmount, ConditionValue: "77", ConditionOperator: models.OperatorEquals}}},
			}
			_, _ = repo.CreateRule(context.Background(), ruleReq)
			resp, err := service.ExecuteRules(ctx, userId)
			Expect(err).To(BeNil())
			Expect(resp.Modified).To(HaveLen(1))
			updated, _ := txRepo.GetTransactionById(ctx, tx.Id, userId)
			Expect(updated.CategoryIds).To(ContainElement(catId))
		})

		It("should handle multiple conditions (AND logic)", func() {
			amount := 88.0
			desc := "Special"
			tx := createTestTransaction(txRepo, userId, "MultiCond", amount, &desc, nil)
			ruleReq := &models.CreateRuleRequest{
				BaseRule:  models.BaseRule{Name: "MultiCondRule", EffectiveFrom: tx.Date},
				CreatedBy: userId,
				Actions:   []models.CreateRuleActionRequest{{BaseRuleAction: models.BaseRuleAction{ActionType: models.RuleFieldName, ActionValue: "MC"}}},
				Conditions: []models.CreateRuleConditionRequest{
					{BaseRuleCondition: models.BaseRuleCondition{ConditionType: models.RuleFieldAmount, ConditionValue: "88", ConditionOperator: models.OperatorEquals}},
					{BaseRuleCondition: models.BaseRuleCondition{ConditionType: models.RuleFieldDescription, ConditionValue: "Special", ConditionOperator: models.OperatorEquals}},
				},
			}
			_, _ = repo.CreateRule(context.Background(), ruleReq)
			resp, err := service.ExecuteRules(ctx, userId)
			Expect(err).To(BeNil())
			Expect(resp.Modified).To(HaveLen(1))
			updated, _ := txRepo.GetTransactionById(ctx, tx.Id, userId)
			Expect(updated.Name).To(Equal("MC"))
		})
	})
})

// Helper for creating test transactions
func createTestTransaction(txRepo *mock_repository.MockTransactionRepository, userId int64, name string, amount float64, desc *string, cats []int64) models.TransactionResponse {
	input := models.CreateBaseTransactionInput{
		Name:      name,
		Amount:    &amount,
		Date:      fixedTestDate(),
		CreatedBy: userId,
		AccountId: 1,
	}
	description := ""
	if desc != nil {
		description = *desc
	}
	input.Description = description
	tx, _ := txRepo.CreateTransaction(&gin.Context{}, input, cats)
	return tx
}

func fixedTestDate() time.Time {
	return time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
}
