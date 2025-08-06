package service

import (
	"context"
	mock_database "expenses/internal/mock/database"
	mock "expenses/internal/mock/repository"
	"expenses/internal/models"
	"fmt"
	"strings"
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
			response, err := ruleService.ListRules(ctx, user1, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(response.Rules)).To(Equal(3))
			Expect(response.Total).To(Equal(3))
		})

		It("should return empty list for user with no rules", func() {
			response, err := ruleService.ListRules(ctx, 999, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(response.Rules)).To(Equal(0))
			Expect(response.Total).To(Equal(0))
		})

		It("should only return rules for the requested user", func() {
			response, err := ruleService.ListRules(ctx, user2, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(response.Rules)).To(Equal(1))
			Expect(response.Total).To(Equal(1))
			Expect(response.Rules[0].Name).To(Equal("User2Rule"))
		})

		Context("with pagination", func() {
			BeforeEach(func() {
				// Create additional rules for pagination testing
				for i := 4; i <= 15; i++ {
					input := models.CreateRuleRequest{
						Rule: models.CreateBaseRuleRequest{
							Name:          fmt.Sprintf("Test Rule %d", i),
							Description:   ptrToString(fmt.Sprintf("Rule description %d", i)),
							EffectiveFrom: now,
							CreatedBy:     user1,
						},
						Actions:    []models.CreateRuleActionRequest{{ActionType: models.RuleFieldAmount, ActionValue: "100"}},
						Conditions: []models.CreateRuleConditionRequest{{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals}},
					}
					_, err := ruleService.CreateRule(ctx, input)
					Expect(err).NotTo(HaveOccurred())
				}
			})

			It("should return paginated rules with specified page size", func() {
				query := &models.RuleListQuery{
					Page:     1,
					PageSize: 5,
				}
				response, err := ruleService.ListRules(ctx, user1, query)
				Expect(err).NotTo(HaveOccurred())
				Expect(response.Page).To(Equal(1))
				Expect(response.PageSize).To(Equal(5))
				Expect(response.Total).To(BeNumerically(">=", 15))
				Expect(len(response.Rules)).To(Equal(5))
			})

			It("should handle search filtering", func() {
				searchTerm := "Test Rule 1"
				query := &models.RuleListQuery{
					Page:     1,
					PageSize: 10,
					Search:   &searchTerm,
				}
				response, err := ruleService.ListRules(ctx, user1, query)
				Expect(err).NotTo(HaveOccurred())
				Expect(response.Total).To(BeNumerically(">=", 1))
				for _, rule := range response.Rules {
					Expect(rule.Name).To(ContainSubstring("Test Rule 1"))
				}
			})
		})
	})

	Describe("ListRules with Pagination", func() {
		BeforeEach(func() {
			// Create 15 rules for user1 with different names and descriptions
			for i := 0; i < 15; i++ {
				input := models.CreateRuleRequest{
					Rule: models.CreateBaseRuleRequest{
						Name:          fmt.Sprintf("Test Rule %d", i+1),
						Description:   ptrToString(fmt.Sprintf("Rule description %d", i+1)),
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

		It("should return paginated rules with default values", func() {
			query := models.RuleListQuery{
				Page:     1,
				PageSize: 10,
			}
			response, err := ruleService.ListRules(ctx, user1, &query)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Page).To(Equal(1))
			Expect(response.PageSize).To(Equal(10))
			Expect(response.Total).To(BeNumerically(">=", 15))
			Expect(len(response.Rules)).To(Equal(10))
		})

		It("should handle custom page size", func() {
			query := models.RuleListQuery{
				Page:     2,
				PageSize: 5,
			}
			response, err := ruleService.ListRules(ctx, user1, &query)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Page).To(Equal(2))
			Expect(response.PageSize).To(Equal(5))
			Expect(len(response.Rules)).To(Equal(5))
		})

		It("should filter by search term in name", func() {
			searchTerm := "Test Rule 1"
			query := models.RuleListQuery{
				Page:     1,
				PageSize: 10,
				Search:   &searchTerm,
			}
			response, err := ruleService.ListRules(ctx, user1, &query)
			Expect(err).NotTo(HaveOccurred())
			// Should find Test Rule 1, 10, 11, 12, 13, 14, 15
			Expect(response.Total).To(BeNumerically(">=", 6))

			// Verify search results
			for _, rule := range response.Rules {
				Expect(strings.Contains(rule.Name, "Test Rule 1")).To(BeTrue())
			}
		})

		It("should filter by search term in description", func() {
			searchTerm := "description 5"
			query := models.RuleListQuery{
				Page:     1,
				PageSize: 10,
				Search:   &searchTerm,
			}
			response, err := ruleService.ListRules(ctx, user1, &query)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Total).To(BeNumerically(">=", 1))

			// Verify search results
			found := false
			for _, rule := range response.Rules {
				if rule.Description != nil && strings.Contains(*rule.Description, "description 5") {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue())
		})

		It("should return empty results for non-matching search", func() {
			searchTerm := "nonexistent"
			query := models.RuleListQuery{
				Page:     1,
				PageSize: 10,
				Search:   &searchTerm,
			}
			response, err := ruleService.ListRules(ctx, user1, &query)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Total).To(Equal(0))
			Expect(len(response.Rules)).To(Equal(0))
		})

		It("should only return rules for the requested user", func() {
			query := models.RuleListQuery{
				Page:     1,
				PageSize: 10,
			}
			response, err := ruleService.ListRules(ctx, user2, &query)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Total).To(Equal(1))
			Expect(len(response.Rules)).To(Equal(1))
			Expect(response.Rules[0].Name).To(Equal("User2Rule"))
		})

		It("should handle page beyond available data", func() {
			query := models.RuleListQuery{
				Page:     100,
				PageSize: 10,
			}
			response, err := ruleService.ListRules(ctx, user1, &query)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Page).To(Equal(100))
			Expect(len(response.Rules)).To(Equal(0))
		})

		It("should set default values for invalid parameters", func() {
			query := models.RuleListQuery{
				Page:     0,  // Invalid
				PageSize: -1, // Invalid
			}
			response, err := ruleService.ListRules(ctx, user1, &query)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Page).To(Equal(1))      // Default
			Expect(response.PageSize).To(Equal(10)) // Default
		})

		It("should limit page size to maximum", func() {
			query := models.RuleListQuery{
				Page:     1,
				PageSize: 200, // Above maximum
			}
			response, err := ruleService.ListRules(ctx, user1, &query)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.PageSize).To(Equal(100)) // Maximum
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

	Describe("PutRuleActions", func() {
		var created models.DescribeRuleResponse

		BeforeEach(func() {
			input := models.CreateRuleRequest{
				Rule: models.CreateBaseRuleRequest{
					Name:          "Put Actions Rule",
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

		It("should successfully replace all rule actions", func() {
			req := models.PutRuleActionsRequest{
				Actions: []models.CreateRuleActionRequest{
					{ActionType: models.RuleFieldAmount, ActionValue: "200"},
					{ActionType: models.RuleFieldCategory, ActionValue: "1"},
				},
			}
			resp, err := ruleService.PutRuleActions(ctx, created.Rule.Id, req, user1)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(resp.Actions)).To(Equal(2))
			Expect(resp.Actions[0].ActionValue).To(Equal("200"))
			Expect(resp.Actions[1].ActionValue).To(Equal("1"))
		})

		It("should validate rule ownership before processing", func() {
			req := models.PutRuleActionsRequest{
				Actions: []models.CreateRuleActionRequest{
					{ActionType: models.RuleFieldAmount, ActionValue: "200"},
				},
			}
			_, err := ruleService.PutRuleActions(ctx, created.Rule.Id, req, user2)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for non-existent rule", func() {
			req := models.PutRuleActionsRequest{
				Actions: []models.CreateRuleActionRequest{
					{ActionType: models.RuleFieldAmount, ActionValue: "200"},
				},
			}
			_, err := ruleService.PutRuleActions(ctx, 9999, req, user1)
			Expect(err).To(HaveOccurred())
		})

		It("should validate input before processing", func() {
			req := models.PutRuleActionsRequest{
				Actions: []models.CreateRuleActionRequest{},
			}
			_, err := ruleService.PutRuleActions(ctx, created.Rule.Id, req, user1)
			Expect(err).To(HaveOccurred())
		})

		It("should validate action types", func() {
			req := models.PutRuleActionsRequest{
				Actions: []models.CreateRuleActionRequest{
					{ActionType: "invalid_type", ActionValue: "200"},
				},
			}
			_, err := ruleService.PutRuleActions(ctx, created.Rule.Id, req, user1)
			Expect(err).To(HaveOccurred())
		})

		It("should validate action values for amount type", func() {
			req := models.PutRuleActionsRequest{
				Actions: []models.CreateRuleActionRequest{
					{ActionType: models.RuleFieldAmount, ActionValue: "invalid_amount"},
				},
			}
			_, err := ruleService.PutRuleActions(ctx, created.Rule.Id, req, user1)
			Expect(err).To(HaveOccurred())
		})

		It("should validate action values for category type", func() {
			req := models.PutRuleActionsRequest{
				Actions: []models.CreateRuleActionRequest{
					{ActionType: models.RuleFieldCategory, ActionValue: "invalid_category"},
				},
			}
			_, err := ruleService.PutRuleActions(ctx, created.Rule.Id, req, user1)
			Expect(err).To(HaveOccurred())
		})

		It("should handle empty action values", func() {
			req := models.PutRuleActionsRequest{
				Actions: []models.CreateRuleActionRequest{
					{ActionType: models.RuleFieldName, ActionValue: ""},
				},
			}
			_, err := ruleService.PutRuleActions(ctx, created.Rule.Id, req, user1)
			Expect(err).To(HaveOccurred())
		})

		It("should handle repository errors gracefully", func() {
			// This test would require mocking the repository to return an error
			// For now, we'll test with a scenario that might cause repository errors
			req := models.PutRuleActionsRequest{
				Actions: []models.CreateRuleActionRequest{
					{ActionType: models.RuleFieldAmount, ActionValue: "200"},
				},
			}
			// First delete the rule to cause a repository error
			err := ruleService.DeleteRule(ctx, created.Rule.Id, user1)
			Expect(err).NotTo(HaveOccurred())

			_, err = ruleService.PutRuleActions(ctx, created.Rule.Id, req, user1)
			Expect(err).To(HaveOccurred())
		})

		It("should replace single action with multiple actions", func() {
			req := models.PutRuleActionsRequest{
				Actions: []models.CreateRuleActionRequest{
					{ActionType: models.RuleFieldAmount, ActionValue: "300"},
					{ActionType: models.RuleFieldCategory, ActionValue: "2"},
					{ActionType: models.RuleFieldName, ActionValue: "Updated Name"},
				},
			}
			resp, err := ruleService.PutRuleActions(ctx, created.Rule.Id, req, user1)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(resp.Actions)).To(Equal(3))
		})

		It("should replace multiple actions with single action", func() {
			// First add multiple actions
			req1 := models.PutRuleActionsRequest{
				Actions: []models.CreateRuleActionRequest{
					{ActionType: models.RuleFieldAmount, ActionValue: "300"},
					{ActionType: models.RuleFieldCategory, ActionValue: "2"},
				},
			}
			_, err := ruleService.PutRuleActions(ctx, created.Rule.Id, req1, user1)
			Expect(err).NotTo(HaveOccurred())

			// Then replace with single action
			req2 := models.PutRuleActionsRequest{
				Actions: []models.CreateRuleActionRequest{
					{ActionType: models.RuleFieldName, ActionValue: "Single Action"},
				},
			}
			resp, err := ruleService.PutRuleActions(ctx, created.Rule.Id, req2, user1)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(resp.Actions)).To(Equal(1))
			Expect(resp.Actions[0].ActionValue).To(Equal("Single Action"))
		})
	})

	Describe("PutRuleConditions", func() {
		var created models.DescribeRuleResponse

		BeforeEach(func() {
			input := models.CreateRuleRequest{
				Rule: models.CreateBaseRuleRequest{
					Name:          "Put Conditions Rule",
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

		It("should successfully replace all rule conditions", func() {
			req := models.PutRuleConditionsRequest{
				Conditions: []models.CreateRuleConditionRequest{
					{ConditionType: models.RuleFieldAmount, ConditionValue: "200", ConditionOperator: models.OperatorGreater},
					{ConditionType: models.RuleFieldName, ConditionValue: "test", ConditionOperator: models.OperatorContains},
				},
			}
			resp, err := ruleService.PutRuleConditions(ctx, created.Rule.Id, req, user1)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(resp.Conditions)).To(Equal(2))
			Expect(resp.Conditions[0].ConditionValue).To(Equal("200"))
			Expect(resp.Conditions[0].ConditionOperator).To(Equal(models.OperatorGreater))
			Expect(resp.Conditions[1].ConditionValue).To(Equal("test"))
			Expect(resp.Conditions[1].ConditionOperator).To(Equal(models.OperatorContains))
		})

		It("should validate rule ownership before processing", func() {
			req := models.PutRuleConditionsRequest{
				Conditions: []models.CreateRuleConditionRequest{
					{ConditionType: models.RuleFieldAmount, ConditionValue: "200", ConditionOperator: models.OperatorEquals},
				},
			}
			_, err := ruleService.PutRuleConditions(ctx, created.Rule.Id, req, user2)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for non-existent rule", func() {
			req := models.PutRuleConditionsRequest{
				Conditions: []models.CreateRuleConditionRequest{
					{ConditionType: models.RuleFieldAmount, ConditionValue: "200", ConditionOperator: models.OperatorEquals},
				},
			}
			_, err := ruleService.PutRuleConditions(ctx, 9999, req, user1)
			Expect(err).To(HaveOccurred())
		})

		It("should validate input before processing", func() {
			req := models.PutRuleConditionsRequest{
				Conditions: []models.CreateRuleConditionRequest{},
			}
			_, err := ruleService.PutRuleConditions(ctx, created.Rule.Id, req, user1)
			Expect(err).To(HaveOccurred())
		})

		It("should validate condition types", func() {
			req := models.PutRuleConditionsRequest{
				Conditions: []models.CreateRuleConditionRequest{
					{ConditionType: "invalid_type", ConditionValue: "200", ConditionOperator: models.OperatorEquals},
				},
			}
			_, err := ruleService.PutRuleConditions(ctx, created.Rule.Id, req, user1)
			Expect(err).To(HaveOccurred())
		})

		It("should validate condition values for amount type", func() {
			req := models.PutRuleConditionsRequest{
				Conditions: []models.CreateRuleConditionRequest{
					{ConditionType: models.RuleFieldAmount, ConditionValue: "invalid_amount", ConditionOperator: models.OperatorEquals},
				},
			}
			_, err := ruleService.PutRuleConditions(ctx, created.Rule.Id, req, user1)
			Expect(err).To(HaveOccurred())
		})

		It("should validate condition values for category type", func() {
			req := models.PutRuleConditionsRequest{
				Conditions: []models.CreateRuleConditionRequest{
					{ConditionType: models.RuleFieldCategory, ConditionValue: "invalid_category", ConditionOperator: models.OperatorEquals},
				},
			}
			_, err := ruleService.PutRuleConditions(ctx, created.Rule.Id, req, user1)
			Expect(err).To(HaveOccurred())
		})

		It("should validate condition operators", func() {
			req := models.PutRuleConditionsRequest{
				Conditions: []models.CreateRuleConditionRequest{
					{ConditionType: models.RuleFieldCategory, ConditionValue: "1", ConditionOperator: models.OperatorContains},
				},
			}
			_, err := ruleService.PutRuleConditions(ctx, created.Rule.Id, req, user1)
			Expect(err).To(HaveOccurred())
		})

		It("should handle empty condition values", func() {
			req := models.PutRuleConditionsRequest{
				Conditions: []models.CreateRuleConditionRequest{
					{ConditionType: models.RuleFieldName, ConditionValue: "", ConditionOperator: models.OperatorEquals},
				},
			}
			_, err := ruleService.PutRuleConditions(ctx, created.Rule.Id, req, user1)
			Expect(err).To(HaveOccurred())
		})

		It("should handle repository errors gracefully", func() {
			req := models.PutRuleConditionsRequest{
				Conditions: []models.CreateRuleConditionRequest{
					{ConditionType: models.RuleFieldAmount, ConditionValue: "200", ConditionOperator: models.OperatorEquals},
				},
			}
			// First delete the rule to cause a repository error
			err := ruleService.DeleteRule(ctx, created.Rule.Id, user1)
			Expect(err).NotTo(HaveOccurred())

			_, err = ruleService.PutRuleConditions(ctx, created.Rule.Id, req, user1)
			Expect(err).To(HaveOccurred())
		})

		It("should replace single condition with multiple conditions", func() {
			req := models.PutRuleConditionsRequest{
				Conditions: []models.CreateRuleConditionRequest{
					{ConditionType: models.RuleFieldAmount, ConditionValue: "300", ConditionOperator: models.OperatorGreater},
					{ConditionType: models.RuleFieldName, ConditionValue: "test", ConditionOperator: models.OperatorContains},
					{ConditionType: models.RuleFieldCategory, ConditionValue: "2", ConditionOperator: models.OperatorEquals},
				},
			}
			resp, err := ruleService.PutRuleConditions(ctx, created.Rule.Id, req, user1)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(resp.Conditions)).To(Equal(3))
		})

		It("should replace multiple conditions with single condition", func() {
			// First add multiple conditions
			req1 := models.PutRuleConditionsRequest{
				Conditions: []models.CreateRuleConditionRequest{
					{ConditionType: models.RuleFieldAmount, ConditionValue: "300", ConditionOperator: models.OperatorGreater},
					{ConditionType: models.RuleFieldName, ConditionValue: "test", ConditionOperator: models.OperatorContains},
				},
			}
			_, err := ruleService.PutRuleConditions(ctx, created.Rule.Id, req1, user1)
			Expect(err).NotTo(HaveOccurred())

			// Then replace with single condition
			req2 := models.PutRuleConditionsRequest{
				Conditions: []models.CreateRuleConditionRequest{
					{ConditionType: models.RuleFieldCategory, ConditionValue: "1", ConditionOperator: models.OperatorEquals},
				},
			}
			resp, err := ruleService.PutRuleConditions(ctx, created.Rule.Id, req2, user1)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(resp.Conditions)).To(Equal(1))
			Expect(resp.Conditions[0].ConditionValue).To(Equal("1"))
		})

		It("should validate all operators for amount field type", func() {
			validOperators := []models.RuleOperator{
				models.OperatorEquals,
				models.OperatorGreater,
				models.OperatorLower,
			}

			for _, op := range validOperators {
				req := models.PutRuleConditionsRequest{
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: op},
					},
				}
				_, err := ruleService.PutRuleConditions(ctx, created.Rule.Id, req, user1)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("should validate all operators for string field types", func() {
			validOperators := []models.RuleOperator{
				models.OperatorEquals,
				models.OperatorContains,
			}

			for _, op := range validOperators {
				req := models.PutRuleConditionsRequest{
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldName, ConditionValue: "test", ConditionOperator: op},
					},
				}
				_, err := ruleService.PutRuleConditions(ctx, created.Rule.Id, req, user1)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("should handle concurrent PUT conditions requests", func() {
			req1 := models.PutRuleConditionsRequest{
				Conditions: []models.CreateRuleConditionRequest{
					{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals},
				},
			}
			req2 := models.PutRuleConditionsRequest{
				Conditions: []models.CreateRuleConditionRequest{
					{ConditionType: models.RuleFieldAmount, ConditionValue: "200", ConditionOperator: models.OperatorGreater},
				},
			}

			var wg sync.WaitGroup
			wg.Add(2)
			var err1, err2 error
			var resp1, resp2 models.PutRuleConditionsResponse

			go func() {
				defer wg.Done()
				resp1, err1 = ruleService.PutRuleConditions(ctx, created.Rule.Id, req1, user1)
			}()
			go func() {
				defer wg.Done()
				resp2, err2 = ruleService.PutRuleConditions(ctx, created.Rule.Id, req2, user1)
			}()

			wg.Wait()
			Expect(err1).NotTo(HaveOccurred())
			Expect(err2).NotTo(HaveOccurred())
			// Both should succeed due to locking mechanism
			Expect(len(resp1.Conditions)).To(Equal(1))
			Expect(len(resp2.Conditions)).To(Equal(1))
		})

		It("should handle mixed field types in conditions", func() {
			req := models.PutRuleConditionsRequest{
				Conditions: []models.CreateRuleConditionRequest{
					{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorGreater},
					{ConditionType: models.RuleFieldName, ConditionValue: "expense", ConditionOperator: models.OperatorContains},
					{ConditionType: models.RuleFieldDescription, ConditionValue: "business", ConditionOperator: models.OperatorEquals},
					{ConditionType: models.RuleFieldCategory, ConditionValue: "1", ConditionOperator: models.OperatorEquals},
				},
			}
			resp, err := ruleService.PutRuleConditions(ctx, created.Rule.Id, req, user1)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(resp.Conditions)).To(Equal(4))
		})
	})

	Describe("PutRuleActions - Additional Edge Cases", func() {
		var created models.DescribeRuleResponse

		BeforeEach(func() {
			input := models.CreateRuleRequest{
				Rule: models.CreateBaseRuleRequest{
					Name:          "Edge Case Actions Rule",
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

		It("should handle concurrent PUT actions requests", func() {
			req1 := models.PutRuleActionsRequest{
				Actions: []models.CreateRuleActionRequest{
					{ActionType: models.RuleFieldAmount, ActionValue: "100"},
				},
			}
			req2 := models.PutRuleActionsRequest{
				Actions: []models.CreateRuleActionRequest{
					{ActionType: models.RuleFieldAmount, ActionValue: "200"},
				},
			}

			var wg sync.WaitGroup
			wg.Add(2)
			var err1, err2 error
			var resp1, resp2 models.PutRuleActionsResponse

			go func() {
				defer wg.Done()
				resp1, err1 = ruleService.PutRuleActions(ctx, created.Rule.Id, req1, user1)
			}()
			go func() {
				defer wg.Done()
				resp2, err2 = ruleService.PutRuleActions(ctx, created.Rule.Id, req2, user1)
			}()

			wg.Wait()
			Expect(err1).NotTo(HaveOccurred())
			Expect(err2).NotTo(HaveOccurred())
			// Both should succeed due to locking mechanism
			Expect(len(resp1.Actions)).To(Equal(1))
			Expect(len(resp2.Actions)).To(Equal(1))
		})

		It("should handle mixed field types in actions", func() {
			req := models.PutRuleActionsRequest{
				Actions: []models.CreateRuleActionRequest{
					{ActionType: models.RuleFieldAmount, ActionValue: "150.50"},
					{ActionType: models.RuleFieldName, ActionValue: "Updated Name"},
					{ActionType: models.RuleFieldDescription, ActionValue: "Updated Description"},
					{ActionType: models.RuleFieldCategory, ActionValue: "2"},
				},
			}
			resp, err := ruleService.PutRuleActions(ctx, created.Rule.Id, req, user1)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(resp.Actions)).To(Equal(4))
		})

		It("should handle large number of actions", func() {
			var actions []models.CreateRuleActionRequest
			for i := 0; i < 50; i++ {
				actions = append(actions, models.CreateRuleActionRequest{
					ActionType:  models.RuleFieldAmount,
					ActionValue: "100",
				})
			}
			req := models.PutRuleActionsRequest{Actions: actions}
			resp, err := ruleService.PutRuleActions(ctx, created.Rule.Id, req, user1)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(resp.Actions)).To(Equal(50))
		})

		It("should validate decimal amounts", func() {
			req := models.PutRuleActionsRequest{
				Actions: []models.CreateRuleActionRequest{
					{ActionType: models.RuleFieldAmount, ActionValue: "123.45"},
				},
			}
			resp, err := ruleService.PutRuleActions(ctx, created.Rule.Id, req, user1)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Actions[0].ActionValue).To(Equal("123.45"))
		})

		It("should validate negative amounts", func() {
			req := models.PutRuleActionsRequest{
				Actions: []models.CreateRuleActionRequest{
					{ActionType: models.RuleFieldAmount, ActionValue: "-50.00"},
				},
			}
			resp, err := ruleService.PutRuleActions(ctx, created.Rule.Id, req, user1)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Actions[0].ActionValue).To(Equal("-50.00"))
		})
	})

	Describe("PutRuleConditions - Additional Edge Cases", func() {
		var created models.DescribeRuleResponse

		BeforeEach(func() {
			input := models.CreateRuleRequest{
				Rule: models.CreateBaseRuleRequest{
					Name:          "Edge Case Conditions Rule",
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

		It("should handle large number of conditions", func() {
			var conditions []models.CreateRuleConditionRequest
			for i := 0; i < 50; i++ {
				conditions = append(conditions, models.CreateRuleConditionRequest{
					ConditionType:     models.RuleFieldAmount,
					ConditionValue:    "100",
					ConditionOperator: models.OperatorEquals,
				})
			}
			req := models.PutRuleConditionsRequest{Conditions: conditions}
			resp, err := ruleService.PutRuleConditions(ctx, created.Rule.Id, req, user1)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(resp.Conditions)).To(Equal(50))
		})

		It("should validate decimal amounts in conditions", func() {
			req := models.PutRuleConditionsRequest{
				Conditions: []models.CreateRuleConditionRequest{
					{ConditionType: models.RuleFieldAmount, ConditionValue: "123.45", ConditionOperator: models.OperatorGreater},
				},
			}
			resp, err := ruleService.PutRuleConditions(ctx, created.Rule.Id, req, user1)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Conditions[0].ConditionValue).To(Equal("123.45"))
		})

		It("should validate negative amounts in conditions", func() {
			req := models.PutRuleConditionsRequest{
				Conditions: []models.CreateRuleConditionRequest{
					{ConditionType: models.RuleFieldAmount, ConditionValue: "-50.00", ConditionOperator: models.OperatorLower},
				},
			}
			resp, err := ruleService.PutRuleConditions(ctx, created.Rule.Id, req, user1)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Conditions[0].ConditionValue).To(Equal("-50.00"))
		})

		It("should handle special characters in string conditions", func() {
			req := models.PutRuleConditionsRequest{
				Conditions: []models.CreateRuleConditionRequest{
					{ConditionType: models.RuleFieldName, ConditionValue: "café & résumé", ConditionOperator: models.OperatorContains},
				},
			}
			resp, err := ruleService.PutRuleConditions(ctx, created.Rule.Id, req, user1)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Conditions[0].ConditionValue).To(Equal("café & résumé"))
		})

		It("should handle very long string values", func() {
			longValue := strings.Repeat("a", 200)
			req := models.PutRuleConditionsRequest{
				Conditions: []models.CreateRuleConditionRequest{
					{ConditionType: models.RuleFieldDescription, ConditionValue: longValue, ConditionOperator: models.OperatorEquals},
				},
			}
			resp, err := ruleService.PutRuleConditions(ctx, created.Rule.Id, req, user1)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Conditions[0].ConditionValue).To(Equal(longValue))
		})
	})

	// Helper
})

func ptrToString(s string) *string {
	return &s
}
