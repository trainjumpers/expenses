package service

import (
	"context"
	mock_database "expenses/internal/mock/database"
	mock "expenses/internal/mock/repository"
	"expenses/internal/models"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RuleService", func() {
	var (
		ruleService RuleServiceInterface
		mockRepo    *mock.MockRuleRepository
		ctx         context.Context
		now         time.Time
		user1       int64
		user2       int64
		mockTxnRepo *mock.MockTransactionRepository
		mockDB      *mock_database.MockDatabaseManager
	)

	BeforeEach(func() {
		ctx = context.Background()
		mockRepo = mock.NewMockRuleRepository()
		mockTxnRepo = mock.NewMockTransactionRepository()
		mockDB = mock_database.NewMockDatabaseManager()
		ruleService = NewRuleService(mockRepo, mockTxnRepo, mockDB)
		now = time.Now()
		user1 = 1
		user2 = 2
	})

	Describe("CreateRule", func() {
		It("should create a new rule with actions and conditions", func() {
			input := models.CreateRuleRequest{
				Rule: models.CreateBaseRuleRequest{
					Name:          "Test Rule",
					Description:   ptrToString("desc"),
					EffectiveFrom: now,
					CreatedBy:     user1,
				},
				Actions: []models.CreateRuleActionRequest{
					{ActionType: models.RuleFieldAmount, ActionValue: "100"},
				},
				Conditions: []models.CreateRuleConditionRequest{
					{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals},
				},
			}
			resp, err := ruleService.CreateRule(ctx, input)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Rule.Name).To(Equal("Test Rule"))
			Expect(resp.Rule.Description).NotTo(BeNil())
			Expect(resp.Rule.CreatedBy).To(Equal(user1))
			Expect(len(resp.Actions)).To(Equal(1))
			Expect(len(resp.Conditions)).To(Equal(1))
		})

		It("should handle multiple actions and conditions", func() {
			input := models.CreateRuleRequest{
				Rule: models.CreateBaseRuleRequest{
					Name:          "Multi Rule",
					Description:   ptrToString("desc"),
					EffectiveFrom: now,
					CreatedBy:     user1,
				},
				Actions: []models.CreateRuleActionRequest{
					{ActionType: models.RuleFieldAmount, ActionValue: "100"},
					{ActionType: models.RuleFieldCategory, ActionValue: "1"},
				},
				Conditions: []models.CreateRuleConditionRequest{
					{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals},
					{ConditionType: models.RuleFieldCategory, ConditionValue: "1", ConditionOperator: models.OperatorEquals},
				},
			}
			resp, err := ruleService.CreateRule(ctx, input)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(resp.Actions)).To(Equal(2))
			Expect(len(resp.Conditions)).To(Equal(2))
		})

		It("should return error for missing actions", func() {
			input := models.CreateRuleRequest{
				Rule: models.CreateBaseRuleRequest{
					Name:          "No Actions",
					Description:   ptrToString("desc"),
					EffectiveFrom: now,
					CreatedBy:     user1,
				},
				Actions: []models.CreateRuleActionRequest{},
				Conditions: []models.CreateRuleConditionRequest{
					{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals},
				},
			}
			_, err := ruleService.CreateRule(ctx, input)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for missing conditions", func() {
			input := models.CreateRuleRequest{
				Rule: models.CreateBaseRuleRequest{
					Name:          "No Conditions",
					Description:   ptrToString("desc"),
					EffectiveFrom: now,
					CreatedBy:     user1,
				},
				Actions:    []models.CreateRuleActionRequest{{ActionType: models.RuleFieldAmount, ActionValue: "100"}},
				Conditions: []models.CreateRuleConditionRequest{},
			}
			_, err := ruleService.CreateRule(ctx, input)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GetRuleById", func() {
		var created models.DescribeRuleResponse

		BeforeEach(func() {
			input := models.CreateRuleRequest{
				Rule: models.CreateBaseRuleRequest{
					Name:          "Fetch Rule",
					Description:   ptrToString("desc"),
					EffectiveFrom: now,
					CreatedBy:     user1,
				},
				Actions:    []models.CreateRuleActionRequest{{ActionType: models.RuleFieldAmount, ActionValue: "100"}},
				Conditions: []models.CreateRuleConditionRequest{{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals}},
			}
			var err error
			created, err = ruleService.CreateRule(ctx, input)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should get rule by id successfully", func() {
			resp, err := ruleService.GetRuleById(ctx, created.Rule.Id, user1)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Rule.Name).To(Equal("Fetch Rule"))
			Expect(len(resp.Actions)).To(Equal(1))
			Expect(len(resp.Conditions)).To(Equal(1))
		})

		It("should return error for non-existent rule id", func() {
			_, err := ruleService.GetRuleById(ctx, 9999, user1)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when accessing rule of different user", func() {
			_, err := ruleService.GetRuleById(ctx, created.Rule.Id, user2)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ListRules", func() {
		BeforeEach(func() {
			// Create rules for user1
			for i := range 3 {
				input := models.CreateRuleRequest{
					Rule: models.CreateBaseRuleRequest{
						Name:          "Rule" + string(rune('A'+i)),
						Description:   ptrToString("desc"),
						EffectiveFrom: now,
						CreatedBy:     user1,
					},
					Actions:    []models.CreateRuleActionRequest{{ActionType: models.RuleFieldAmount, ActionValue: "100"}},
					Conditions: []models.CreateRuleConditionRequest{{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals}},
				}
				_, err := ruleService.CreateRule(ctx, input)
				Expect(err).NotTo(HaveOccurred())
			}
			// Create rule for user2
			input := models.CreateRuleRequest{
				Rule: models.CreateBaseRuleRequest{
					Name:          "User2Rule",
					Description:   ptrToString("desc"),
					EffectiveFrom: now,
					CreatedBy:     user2,
				},
				Actions:    []models.CreateRuleActionRequest{{ActionType: models.RuleFieldAmount, ActionValue: "100"}},
				Conditions: []models.CreateRuleConditionRequest{{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals}},
			}
			_, err := ruleService.CreateRule(ctx, input)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should list all rules for a specific user", func() {
			rules, err := ruleService.ListRules(ctx, user1)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(rules)).To(Equal(3))
		})

		It("should return empty list for user with no rules", func() {
			rules, err := ruleService.ListRules(ctx, 999)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(rules)).To(Equal(0))
		})

		It("should only return rules for the requested user", func() {
			rules, err := ruleService.ListRules(ctx, user2)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(rules)).To(Equal(1))
			Expect(rules[0].Name).To(Equal("User2Rule"))
		})
	})

	Describe("UpdateRule", func() {
		var created models.DescribeRuleResponse

		BeforeEach(func() {
			input := models.CreateRuleRequest{
				Rule: models.CreateBaseRuleRequest{
					Name:          "Update Rule",
					Description:   ptrToString("desc"),
					EffectiveFrom: now,
					CreatedBy:     user1,
				},
				Actions:    []models.CreateRuleActionRequest{{ActionType: models.RuleFieldAmount, ActionValue: "100"}},
				Conditions: []models.CreateRuleConditionRequest{{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals}},
			}
			var err error
			created, err = ruleService.CreateRule(ctx, input)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should update rule name successfully", func() {
			newName := "Updated Name"
			update := models.UpdateRuleRequest{Name: &newName}
			rule, err := ruleService.UpdateRule(ctx, created.Rule.Id, update, user1)
			Expect(err).NotTo(HaveOccurred())
			Expect(rule.Name).To(Equal(newName))
		})

		It("should update rule description successfully", func() {
			newDesc := "Updated Desc"
			update := models.UpdateRuleRequest{Description: &newDesc}
			rule, err := ruleService.UpdateRule(ctx, created.Rule.Id, update, user1)
			Expect(err).NotTo(HaveOccurred())
			Expect(rule.Description).NotTo(BeNil())
			Expect(*rule.Description).To(Equal(newDesc))
		})

		It("should update rule effective_from successfully", func() {
			newTime := now.Add(-time.Hour)
			update := models.UpdateRuleRequest{EffectiveFrom: &newTime}
			rule, err := ruleService.UpdateRule(ctx, created.Rule.Id, update, user1)
			Expect(err).NotTo(HaveOccurred())
			Expect(rule.EffectiveFrom).To(Equal(newTime))
		})

		It("should return error for non-existent rule id", func() {
			update := models.UpdateRuleRequest{Name: ptrToString("Fail")}
			_, err := ruleService.UpdateRule(ctx, 9999, update, user1)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when updating rule of different user", func() {
			update := models.UpdateRuleRequest{Name: ptrToString("Fail")}
			_, err := ruleService.UpdateRule(ctx, created.Rule.Id, update, user2)
			Expect(err).To(HaveOccurred())
		})

		It("should not update if no fields provided", func() {
			update := models.UpdateRuleRequest{}
			_, err := ruleService.UpdateRule(ctx, created.Rule.Id, update, user1)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle concurrent rule updates", func() {
			newName1 := "Concurrent Name 1"
			newName2 := "Concurrent Name 2"
			update1 := models.UpdateRuleRequest{Name: &newName1}
			update2 := models.UpdateRuleRequest{Name: &newName2}
			var wg sync.WaitGroup
			wg.Add(2)
			var err1, err2 error
			go func() {
				defer wg.Done()
				_, err1 = ruleService.UpdateRule(ctx, created.Rule.Id, update1, user1)
			}()
			go func() {
				defer wg.Done()
				_, err2 = ruleService.UpdateRule(ctx, created.Rule.Id, update2, user1)
			}()
			wg.Wait()
			Expect(err1).NotTo(HaveOccurred())
			Expect(err2).NotTo(HaveOccurred())
		})
	})

	Describe("UpdateRuleAction", func() {
		var created models.DescribeRuleResponse
		var actionId int64

		BeforeEach(func() {
			input := models.CreateRuleRequest{
				Rule: models.CreateBaseRuleRequest{
					Name:          "Action Rule",
					Description:   ptrToString("desc"),
					EffectiveFrom: now,
					CreatedBy:     user1,
				},
				Actions:    []models.CreateRuleActionRequest{{ActionType: models.RuleFieldAmount, ActionValue: "100"}},
				Conditions: []models.CreateRuleConditionRequest{{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals}},
			}
			var err error
			created, err = ruleService.CreateRule(ctx, input)
			Expect(err).NotTo(HaveOccurred())
			actionId = created.Actions[0].Id
		})

		It("should update action value successfully", func() {
			val := "200"
			update := models.UpdateRuleActionRequest{ActionValue: &val}
			action, err := ruleService.UpdateRuleAction(ctx, actionId, created.Rule.Id, update, user1)
			Expect(err).NotTo(HaveOccurred())
			Expect(action.ActionValue).To(Equal(val))
		})

		It("should return error for non-existent action id", func() {
			val := "200"
			update := models.UpdateRuleActionRequest{ActionValue: &val}
			_, err := ruleService.UpdateRuleAction(ctx, 9999, created.Rule.Id, update, user1)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for action belonging to different rule", func() {
			// Create another rule and action
			input := models.CreateRuleRequest{
				Rule: models.CreateBaseRuleRequest{
					Name:          "Other Rule",
					Description:   ptrToString("desc"),
					EffectiveFrom: now,
					CreatedBy:     user1,
				},
				Actions:    []models.CreateRuleActionRequest{{ActionType: models.RuleFieldAmount, ActionValue: "100"}},
				Conditions: []models.CreateRuleConditionRequest{{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals}},
			}
			other, err := ruleService.CreateRule(ctx, input)
			Expect(err).NotTo(HaveOccurred())
			// Try to update actionId under other.Rule.Id
			val := "300"
			update := models.UpdateRuleActionRequest{ActionValue: &val}
			_, err = ruleService.UpdateRuleAction(ctx, actionId, other.Rule.Id, update, user1)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for action belonging to different user", func() {
			val := "200"
			update := models.UpdateRuleActionRequest{ActionValue: &val}
			_, err := ruleService.UpdateRuleAction(ctx, actionId, created.Rule.Id, update, user2)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("UpdateRuleCondition", func() {
		var created models.DescribeRuleResponse
		var condId int64

		BeforeEach(func() {
			input := models.CreateRuleRequest{
				Rule: models.CreateBaseRuleRequest{
					Name:          "Cond Rule",
					Description:   ptrToString("desc"),
					EffectiveFrom: now,
					CreatedBy:     user1,
				},
				Actions:    []models.CreateRuleActionRequest{{ActionType: models.RuleFieldAmount, ActionValue: "100"}},
				Conditions: []models.CreateRuleConditionRequest{{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals}},
			}
			var err error
			created, err = ruleService.CreateRule(ctx, input)
			Expect(err).NotTo(HaveOccurred())
			condId = created.Conditions[0].Id
		})

		It("should update condition value successfully", func() {
			val := "200"
			update := models.UpdateRuleConditionRequest{ConditionValue: &val}
			cond, err := ruleService.UpdateRuleCondition(ctx, condId, created.Rule.Id, update, user1)
			Expect(err).NotTo(HaveOccurred())
			Expect(cond.ConditionValue).To(Equal(val))
		})

		It("should return error for non-existent condition id", func() {
			val := "200"
			update := models.UpdateRuleConditionRequest{ConditionValue: &val}
			_, err := ruleService.UpdateRuleCondition(ctx, 9999, created.Rule.Id, update, user1)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for condition belonging to different rule", func() {
			// Create another rule and condition
			input := models.CreateRuleRequest{
				Rule: models.CreateBaseRuleRequest{
					Name:          "Other Rule",
					Description:   ptrToString("desc"),
					EffectiveFrom: now,
					CreatedBy:     user1,
				},
				Actions:    []models.CreateRuleActionRequest{{ActionType: models.RuleFieldAmount, ActionValue: "100"}},
				Conditions: []models.CreateRuleConditionRequest{{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals}},
			}
			other, err := ruleService.CreateRule(ctx, input)
			Expect(err).NotTo(HaveOccurred())
			// Try to update condId under other.Rule.Id
			val := "300"
			update := models.UpdateRuleConditionRequest{ConditionValue: &val}
			_, err = ruleService.UpdateRuleCondition(ctx, condId, other.Rule.Id, update, user1)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for condition belonging to different user", func() {
			val := "200"
			update := models.UpdateRuleConditionRequest{ConditionValue: &val}
			_, err := ruleService.UpdateRuleCondition(ctx, condId, created.Rule.Id, update, user2)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("DeleteRule", func() {
		var created models.DescribeRuleResponse

		BeforeEach(func() {
			input := models.CreateRuleRequest{
				Rule: models.CreateBaseRuleRequest{
					Name:          "Delete Rule",
					Description:   ptrToString("desc"),
					EffectiveFrom: now,
					CreatedBy:     user1,
				},
				Actions:    []models.CreateRuleActionRequest{{ActionType: models.RuleFieldAmount, ActionValue: "100"}},
				Conditions: []models.CreateRuleConditionRequest{{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals}},
			}
			var err error
			created, err = ruleService.CreateRule(ctx, input)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should delete rule successfully", func() {
			err := ruleService.DeleteRule(ctx, created.Rule.Id, user1)
			Expect(err).NotTo(HaveOccurred())
			_, err = ruleService.GetRuleById(ctx, created.Rule.Id, user1)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when deleting a rule that was already deleted", func() {
			err := ruleService.DeleteRule(ctx, created.Rule.Id, user1)
			Expect(err).NotTo(HaveOccurred())
			err = ruleService.DeleteRule(ctx, created.Rule.Id, user1)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for non-existent rule id", func() {
			err := ruleService.DeleteRule(ctx, 9999, user1)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when deleting rule of different user", func() {
			err := ruleService.DeleteRule(ctx, created.Rule.Id, user2)
			Expect(err).To(HaveOccurred())
		})
	})

	// Helper
})

func ptrToString(s string) *string {
	return &s
}
