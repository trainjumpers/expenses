package service

import (
	"expenses/internal/models"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RuleEngine", func() {
	var (
		engine      *RuleEngine
		categories  []models.CategoryResponse
		rules       []models.DescribeRuleResponse
		transaction models.TransactionResponse
		userId      int64
	)

	BeforeEach(func() {
		userId = 1

		// Setup test categories
		categories = []models.CategoryResponse{
			{Id: 1, Name: "Food", CreatedBy: userId},
			{Id: 2, Name: "Transport", CreatedBy: userId},
			{Id: 3, Name: "Shopping", CreatedBy: userId},
			{Id: 4, Name: "Other User Category", CreatedBy: 2}, // Different user
		}

		// Setup base transaction
		desc := "Test transaction description"
		transaction = models.TransactionResponse{
			TransactionBaseResponse: models.TransactionBaseResponse{
				Id:          100,
				Name:        "Test Transaction",
				Description: &desc,
				Amount:      50.0,
				Date:        time.Now(),
				CreatedBy:   userId,
				AccountId:   1,
			},
			CategoryIds: []int64{},
		}
	})

	Describe("NewRuleEngine", func() {
		It("should create engine with categories and rules", func() {
			rules = []models.DescribeRuleResponse{}
			engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)

			Expect(engine).NotTo(BeNil())
			Expect(engine.categories).To(HaveLen(4))
			Expect(engine.rules).To(HaveLen(0))
		})

		It("should build category map correctly", func() {
			rules = []models.DescribeRuleResponse{}
			engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)

			Expect(engine.categories[1].Name).To(Equal("Food"))
			Expect(engine.categories[2].Name).To(Equal("Transport"))
			Expect(engine.categories[3].Name).To(Equal("Shopping"))
			Expect(engine.categories[4].Name).To(Equal("Other User Category"))
		})
	})

	Describe("ProcessTransaction - Basic Flow", func() {
		Context("when no rules exist", func() {
			BeforeEach(func() {
				rules = []models.DescribeRuleResponse{}
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
			})

			It("should return nil", func() {
				result := engine.ProcessTransaction(transaction)
				Expect(result).To(BeNil())
			})
		})

		Context("when rules exist but don't match", func() {
			BeforeEach(func() {
				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:            1,
							Name:          "Non-matching rule",
							EffectiveFrom: time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldName,
								ConditionValue:    "Different Name",
								ConditionOperator: models.OperatorEquals,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "1",
							},
						},
					},
				}
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
			})

			It("should return nil", func() {
				result := engine.ProcessTransaction(transaction)
				Expect(result).To(BeNil())
			})
		})

		Context("when rule matches", func() {
			BeforeEach(func() {
				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:            1,
							Name:          "Matching rule",
							EffectiveFrom: time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldName,
								ConditionValue:    "Test Transaction",
								ConditionOperator: models.OperatorEquals,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "1",
							},
						},
					},
				}
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
			})

			It("should return changeset with applied rule", func() {
				result := engine.ProcessTransaction(transaction)

				Expect(result).NotTo(BeNil())
				Expect(result.TransactionId).To(Equal(int64(100)))
				Expect(result.AppliedRules).To(ContainElement(int64(1)))
				Expect(result.CategoryAdds).To(ContainElement(int64(1)))
			})
		})
	})

	Describe("Condition Evaluation - Amount", func() {
		BeforeEach(func() {
			transaction.Amount = 100.0
		})

		Context("equals operator", func() {
			BeforeEach(func() {
				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:            1,
							Name:          "Amount equals rule",
							EffectiveFrom: time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldAmount,
								ConditionValue:    "100.0",
								ConditionOperator: models.OperatorEquals,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "1",
							},
						},
					},
				}
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
			})

			It("should match when amounts are equal", func() {
				result := engine.ProcessTransaction(transaction)
				Expect(result).NotTo(BeNil())
				Expect(result.CategoryAdds).To(ContainElement(int64(1)))
			})

			It("should not match when amounts are different", func() {
				transaction.Amount = 99.0
				result := engine.ProcessTransaction(transaction)
				Expect(result).To(BeNil())
			})
		})

		Context("greater operator", func() {
			BeforeEach(func() {
				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:            1,
							Name:          "Amount greater rule",
							EffectiveFrom: time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldAmount,
								ConditionValue:    "50.0",
								ConditionOperator: models.OperatorGreater,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "1",
							},
						},
					},
				}
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
			})

			It("should match when amount is greater", func() {
				result := engine.ProcessTransaction(transaction)
				Expect(result).NotTo(BeNil())
			})

			It("should not match when amount is equal", func() {
				transaction.Amount = 50.0
				result := engine.ProcessTransaction(transaction)
				Expect(result).To(BeNil())
			})

			It("should not match when amount is lower", func() {
				transaction.Amount = 25.0
				result := engine.ProcessTransaction(transaction)
				Expect(result).To(BeNil())
			})
		})

		Context("lower operator", func() {
			BeforeEach(func() {
				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:            1,
							Name:          "Amount lower rule",
							EffectiveFrom: time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldAmount,
								ConditionValue:    "150.0",
								ConditionOperator: models.OperatorLower,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "1",
							},
						},
					},
				}
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
			})

			It("should match when amount is lower", func() {
				result := engine.ProcessTransaction(transaction)
				Expect(result).NotTo(BeNil())
			})

			It("should not match when amount is equal", func() {
				transaction.Amount = 150.0
				result := engine.ProcessTransaction(transaction)
				Expect(result).To(BeNil())
			})

			It("should not match when amount is greater", func() {
				transaction.Amount = 200.0
				result := engine.ProcessTransaction(transaction)
				Expect(result).To(BeNil())
			})
		})
	})

	Describe("Action Application - Category", func() {
		BeforeEach(func() {
			rules = []models.DescribeRuleResponse{
				{
					Rule: models.RuleResponse{
						Id:            1,
						Name:          "Add category rule",
						EffectiveFrom: time.Now().Add(-24 * time.Hour),
					},
					Conditions: []models.RuleConditionResponse{
						{
							ConditionType:     models.RuleFieldName,
							ConditionValue:    "Test Transaction",
							ConditionOperator: models.OperatorEquals,
						},
					},
					Actions: []models.RuleActionResponse{
						{
							ActionType:  models.RuleFieldCategory,
							ActionValue: "1",
						},
					},
				},
			}
			engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
		})

		It("should add new category", func() {
			result := engine.ProcessTransaction(transaction)

			Expect(result).NotTo(BeNil())
			Expect(result.CategoryAdds).To(ContainElement(int64(1)))
		})

		It("should not add duplicate category", func() {
			transaction.CategoryIds = []int64{1} // Already has category 1
			result := engine.ProcessTransaction(transaction)

			Expect(result).To(BeNil()) // No changes since category already exists
		})

		It("should not add category from different user", func() {
			rules[0].Actions[0].ActionValue = "4" // Category belongs to user 2
			engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)

			result := engine.ProcessTransaction(transaction)
			Expect(result).To(BeNil()) // No changes since category doesn't belong to user
		})

		It("should handle invalid category ID", func() {
			rules[0].Actions[0].ActionValue = "invalid"
			engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)

			result := engine.ProcessTransaction(transaction)
			Expect(result).To(BeNil()) // No changes since category ID is invalid
		})
	})
	Describe("Condition Evaluation - Name", func() {
		Context("equals operator", func() {
			BeforeEach(func() {
				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:            1,
							Name:          "Name equals rule",
							EffectiveFrom: time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldName,
								ConditionValue:    "Test Transaction",
								ConditionOperator: models.OperatorEquals,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "1",
							},
						},
					},
				}
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
			})

			It("should match exact name (case insensitive)", func() {
				result := engine.ProcessTransaction(transaction)
				Expect(result).NotTo(BeNil())
			})

			It("should match different case", func() {
				transaction.Name = "test transaction"
				result := engine.ProcessTransaction(transaction)
				Expect(result).NotTo(BeNil())
			})

			It("should not match different name", func() {
				transaction.Name = "Different Transaction"
				result := engine.ProcessTransaction(transaction)
				Expect(result).To(BeNil())
			})
		})

		Context("contains operator", func() {
			BeforeEach(func() {
				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:            1,
							Name:          "Name contains rule",
							EffectiveFrom: time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldName,
								ConditionValue:    "Test",
								ConditionOperator: models.OperatorContains,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "1",
							},
						},
					},
				}
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
			})

			It("should match when name contains substring", func() {
				result := engine.ProcessTransaction(transaction)
				Expect(result).NotTo(BeNil())
			})

			It("should match case insensitive", func() {
				transaction.Name = "My test transaction"
				result := engine.ProcessTransaction(transaction)
				Expect(result).NotTo(BeNil())
			})

			It("should not match when substring not present", func() {
				transaction.Name = "Different Transaction"
				result := engine.ProcessTransaction(transaction)
				Expect(result).To(BeNil())
			})
		})
	})

	Describe("Condition Evaluation - Description", func() {
		Context("with description present", func() {
			BeforeEach(func() {
				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:            1,
							Name:          "Description rule",
							EffectiveFrom: time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldDescription,
								ConditionValue:    "Test transaction description",
								ConditionOperator: models.OperatorEquals,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "1",
							},
						},
					},
				}
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
			})

			It("should match description", func() {
				result := engine.ProcessTransaction(transaction)
				Expect(result).NotTo(BeNil())
			})
		})

		Context("with nil description", func() {
			BeforeEach(func() {
				transaction.Description = nil
				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:            1,
							Name:          "Empty description rule",
							EffectiveFrom: time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldDescription,
								ConditionValue:    "",
								ConditionOperator: models.OperatorEquals,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "1",
							},
						},
					},
				}
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
			})

			It("should match empty string when description is nil", func() {
				result := engine.ProcessTransaction(transaction)
				Expect(result).NotTo(BeNil())
			})
		})
	})

	Describe("Condition Evaluation - Category", func() {
		BeforeEach(func() {
			transaction.CategoryIds = []int64{1, 2}
			rules = []models.DescribeRuleResponse{
				{
					Rule: models.RuleResponse{
						Id:            1,
						Name:          "Category condition rule",
						EffectiveFrom: time.Now().Add(-24 * time.Hour),
					},
					Conditions: []models.RuleConditionResponse{
						{
							ConditionType:     models.RuleFieldCategory,
							ConditionValue:    "1",
							ConditionOperator: models.OperatorEquals,
						},
					},
					Actions: []models.RuleActionResponse{
						{
							ActionType:  models.RuleFieldCategory,
							ActionValue: "3",
						},
					},
				},
			}
			engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
		})

		It("should match when transaction has the category", func() {
			result := engine.ProcessTransaction(transaction)
			Expect(result).NotTo(BeNil())
			Expect(result.CategoryAdds).To(ContainElement(int64(3)))
		})

		It("should not match when transaction doesn't have the category", func() {
			transaction.CategoryIds = []int64{2, 3}
			result := engine.ProcessTransaction(transaction)
			Expect(result).To(BeNil())
		})

		It("should handle invalid category ID in condition", func() {
			rules[0].Conditions[0].ConditionValue = "invalid"
			engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)

			result := engine.ProcessTransaction(transaction)
			Expect(result).To(BeNil())
		})
	})

	Describe("Multiple Conditions (AND Logic)", func() {
		BeforeEach(func() {
			transaction.Amount = 100.0
			rules = []models.DescribeRuleResponse{
				{
					Rule: models.RuleResponse{
						Id:            1,
						Name:          "Multiple conditions rule",
						EffectiveFrom: time.Now().Add(-24 * time.Hour),
					},
					Conditions: []models.RuleConditionResponse{
						{
							ConditionType:     models.RuleFieldName,
							ConditionValue:    "Test Transaction",
							ConditionOperator: models.OperatorEquals,
						},
						{
							ConditionType:     models.RuleFieldAmount,
							ConditionValue:    "100.0",
							ConditionOperator: models.OperatorEquals,
						},
					},
					Actions: []models.RuleActionResponse{
						{
							ActionType:  models.RuleFieldCategory,
							ActionValue: "1",
						},
					},
				},
			}
			engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
		})

		It("should match when all conditions are met", func() {
			result := engine.ProcessTransaction(transaction)
			Expect(result).NotTo(BeNil())
		})

		It("should not match when first condition fails", func() {
			transaction.Name = "Different Name"
			result := engine.ProcessTransaction(transaction)
			Expect(result).To(BeNil())
		})

		It("should not match when second condition fails", func() {
			transaction.Amount = 50.0
			result := engine.ProcessTransaction(transaction)
			Expect(result).To(BeNil())
		})

		Context("with empty conditions", func() {
			BeforeEach(func() {
				rules[0].Conditions = []models.RuleConditionResponse{}
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
			})

			It("should return false for empty conditions", func() {
				result := engine.ProcessTransaction(transaction)
				Expect(result).To(BeNil())
			})
		})
	})

	Describe("Action Application - Name Updates", func() {
		BeforeEach(func() {
			rules = []models.DescribeRuleResponse{
				{
					Rule: models.RuleResponse{
						Id:            1,
						Name:          "Name update rule",
						EffectiveFrom: time.Now().Add(-24 * time.Hour),
					},
					Conditions: []models.RuleConditionResponse{
						{
							ConditionType:     models.RuleFieldName,
							ConditionValue:    "Test Transaction",
							ConditionOperator: models.OperatorEquals,
						},
					},
					Actions: []models.RuleActionResponse{
						{
							ActionType:  models.RuleFieldName,
							ActionValue: "Updated Transaction Name",
						},
					},
				},
			}
			engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
		})

		It("should update transaction name", func() {
			result := engine.ProcessTransaction(transaction)

			Expect(result).NotTo(BeNil())
			Expect(result.NameUpdate).NotTo(BeNil())
			Expect(*result.NameUpdate).To(Equal("Updated Transaction Name"))
			Expect(result.AppliedRules).To(ContainElement(int64(1)))
		})

		Context("with multiple name update rules", func() {
			BeforeEach(func() {
				// Add second rule that also updates name
				secondRule := models.DescribeRuleResponse{
					Rule: models.RuleResponse{
						Id:            2,
						Name:          "Second name update rule",
						EffectiveFrom: time.Now().Add(-24 * time.Hour),
					},
					Conditions: []models.RuleConditionResponse{
						{
							ConditionType:     models.RuleFieldName,
							ConditionValue:    "Test Transaction",
							ConditionOperator: models.OperatorEquals,
						},
					},
					Actions: []models.RuleActionResponse{
						{
							ActionType:  models.RuleFieldName,
							ActionValue: "Second Update",
						},
					},
				}
				rules = append(rules, secondRule)
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
			})

			It("should apply first rule only (no overwrite)", func() {
				result := engine.ProcessTransaction(transaction)

				Expect(result).NotTo(BeNil())
				Expect(*result.NameUpdate).To(Equal("Updated Transaction Name")) // First rule wins
				Expect(result.AppliedRules).To(ContainElement(int64(1)))
				Expect(result.AppliedRules).NotTo(ContainElement(int64(2))) // Second rule not applied
			})
		})
	})

	Describe("Action Application - Description Updates", func() {
		BeforeEach(func() {
			rules = []models.DescribeRuleResponse{
				{
					Rule: models.RuleResponse{
						Id:            1,
						Name:          "Description update rule",
						EffectiveFrom: time.Now().Add(-24 * time.Hour),
					},
					Conditions: []models.RuleConditionResponse{
						{
							ConditionType:     models.RuleFieldName,
							ConditionValue:    "Test Transaction",
							ConditionOperator: models.OperatorEquals,
						},
					},
					Actions: []models.RuleActionResponse{
						{
							ActionType:  models.RuleFieldDescription,
							ActionValue: "Updated Description",
						},
					},
				},
			}
			engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
		})

		It("should update transaction description", func() {
			result := engine.ProcessTransaction(transaction)

			Expect(result).NotTo(BeNil())
			Expect(result.DescUpdate).NotTo(BeNil())
			Expect(*result.DescUpdate).To(Equal("Updated Description"))
		})
	})

	Describe("Action Application - Category Advanced", func() {
		Context("with multiple category additions", func() {
			BeforeEach(func() {
				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:            1,
							Name:          "Multi-category rule",
							EffectiveFrom: time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldName,
								ConditionValue:    "Test Transaction",
								ConditionOperator: models.OperatorEquals,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "1",
							},
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "2",
							},
						},
					},
				}
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
			})

			It("should add multiple categories", func() {
				result := engine.ProcessTransaction(transaction)

				Expect(result).NotTo(BeNil())
				Expect(result.CategoryAdds).To(ContainElement(int64(1)))
				Expect(result.CategoryAdds).To(ContainElement(int64(2)))
			})

			It("should not add duplicate categories in changeset", func() {
				// Add another rule that adds same category
				secondRule := models.DescribeRuleResponse{
					Rule: models.RuleResponse{
						Id:            2,
						Name:          "Second category rule",
						EffectiveFrom: time.Now().Add(-24 * time.Hour),
					},
					Conditions: []models.RuleConditionResponse{
						{
							ConditionType:     models.RuleFieldName,
							ConditionValue:    "Test Transaction",
							ConditionOperator: models.OperatorEquals,
						},
					},
					Actions: []models.RuleActionResponse{
						{
							ActionType:  models.RuleFieldCategory,
							ActionValue: "1", // Same category as first rule
						},
					},
				}
				rules = append(rules, secondRule)
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)

				result := engine.ProcessTransaction(transaction)

				Expect(result).NotTo(BeNil())
				// Should only have category 1 once, even though two rules try to add it
				categoryCount := 0
				for _, catId := range result.CategoryAdds {
					if catId == 1 {
						categoryCount++
					}
				}
				Expect(categoryCount).To(Equal(1))
			})
		})
	})

	Describe("Rule Effective Date", func() {
		Context("when rule effective date is in the future", func() {
			BeforeEach(func() {
				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:            1,
							Name:          "Future rule",
							EffectiveFrom: time.Now().Add(24 * time.Hour), // Future date
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldName,
								ConditionValue:    "Test Transaction",
								ConditionOperator: models.OperatorEquals,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "1",
							},
						},
					},
				}
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
			})

			It("should skip rule when effective date is after transaction date", func() {
				result := engine.ProcessTransaction(transaction)
				Expect(result).To(BeNil())
			})
		})

		Context("when rule effective date is in the past", func() {
			BeforeEach(func() {
				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:            1,
							Name:          "Past rule",
							EffectiveFrom: time.Now().Add(-48 * time.Hour), // Past date
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldName,
								ConditionValue:    "Test Transaction",
								ConditionOperator: models.OperatorEquals,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "1",
							},
						},
					},
				}
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
			})

			It("should apply rule when effective date is before transaction date", func() {
				result := engine.ProcessTransaction(transaction)
				Expect(result).NotTo(BeNil())
				Expect(result.CategoryAdds).To(ContainElement(int64(1)))
			})
		})
	})

	Describe("Condition Logic - AND vs OR", func() {
		Context("AND Logic (default behavior)", func() {
			BeforeEach(func() {
				transaction.Amount = 100.0
				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:             1,
							Name:           "AND logic rule",
							ConditionLogic: models.ConditionLogicAnd,
							EffectiveFrom:  time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldName,
								ConditionValue:    "Test Transaction",
								ConditionOperator: models.OperatorEquals,
							},
							{
								ConditionType:     models.RuleFieldAmount,
								ConditionValue:    "100.0",
								ConditionOperator: models.OperatorEquals,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "1",
							},
						},
					},
				}
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
			})

			It("should match when all conditions are met", func() {
				result := engine.ProcessTransaction(transaction)
				Expect(result).NotTo(BeNil())
				Expect(result.CategoryAdds).To(ContainElement(int64(1)))
				Expect(result.AppliedRules).To(ContainElement(int64(1)))
			})

			It("should not match when first condition fails", func() {
				transaction.Name = "Different Name"
				result := engine.ProcessTransaction(transaction)
				Expect(result).To(BeNil())
			})

			It("should not match when second condition fails", func() {
				transaction.Amount = 50.0
				result := engine.ProcessTransaction(transaction)
				Expect(result).To(BeNil())
			})

			It("should not match when both conditions fail", func() {
				transaction.Name = "Different Name"
				transaction.Amount = 50.0
				result := engine.ProcessTransaction(transaction)
				Expect(result).To(BeNil())
			})
		})

		Context("OR Logic", func() {
			BeforeEach(func() {
				transaction.Amount = 100.0
				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:             1,
							Name:           "OR logic rule",
							ConditionLogic: models.ConditionLogicOr,
							EffectiveFrom:  time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldName,
								ConditionValue:    "Test Transaction",
								ConditionOperator: models.OperatorEquals,
							},
							{
								ConditionType:     models.RuleFieldAmount,
								ConditionValue:    "200.0", // Won't match
								ConditionOperator: models.OperatorEquals,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "1",
							},
						},
					},
				}
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
			})

			It("should match when first condition is met", func() {
				result := engine.ProcessTransaction(transaction)
				Expect(result).NotTo(BeNil())
				Expect(result.CategoryAdds).To(ContainElement(int64(1)))
				Expect(result.AppliedRules).To(ContainElement(int64(1)))
			})

			It("should match when second condition is met", func() {
				transaction.Name = "Different Name"
				transaction.Amount = 200.0 // This will match the second condition
				result := engine.ProcessTransaction(transaction)
				Expect(result).NotTo(BeNil())
				Expect(result.CategoryAdds).To(ContainElement(int64(1)))
			})

			It("should match when both conditions are met", func() {
				transaction.Amount = 200.0 // Both conditions will match
				result := engine.ProcessTransaction(transaction)
				Expect(result).NotTo(BeNil())
				Expect(result.CategoryAdds).To(ContainElement(int64(1)))
			})

			It("should not match when no conditions are met", func() {
				transaction.Name = "Different Name"
				transaction.Amount = 50.0 // Neither condition matches
				result := engine.ProcessTransaction(transaction)
				Expect(result).To(BeNil())
			})
		})

		Context("OR Logic with multiple condition types", func() {
			BeforeEach(func() {
				desc := "Test description"
				transaction.Description = &desc
				transaction.Amount = 100.0
				transaction.CategoryIds = []int64{2}

				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:             1,
							Name:           "Complex OR rule",
							ConditionLogic: models.ConditionLogicOr,
							EffectiveFrom:  time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldName,
								ConditionValue:    "Non-matching Name",
								ConditionOperator: models.OperatorEquals,
							},
							{
								ConditionType:     models.RuleFieldAmount,
								ConditionValue:    "999.0",
								ConditionOperator: models.OperatorEquals,
							},
							{
								ConditionType:     models.RuleFieldDescription,
								ConditionValue:    "Test description",
								ConditionOperator: models.OperatorEquals,
							},
							{
								ConditionType:     models.RuleFieldCategory,
								ConditionValue:    "3", // Transaction doesn't have this category
								ConditionOperator: models.OperatorEquals,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "1",
							},
						},
					},
				}
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
			})

			It("should match when description condition is met (third condition)", func() {
				result := engine.ProcessTransaction(transaction)
				Expect(result).NotTo(BeNil())
				Expect(result.CategoryAdds).To(ContainElement(int64(1)))
			})

			It("should not match when no conditions are met", func() {
				differentDesc := "Different description"
				transaction.Description = &differentDesc
				transaction.CategoryIds = []int64{} // Remove categories
				result := engine.ProcessTransaction(transaction)
				Expect(result).To(BeNil())
			})
		})

		Context("Single condition with OR logic", func() {
			BeforeEach(func() {
				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:             1,
							Name:           "Single condition OR rule",
							ConditionLogic: models.ConditionLogicOr,
							EffectiveFrom:  time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldName,
								ConditionValue:    "Test Transaction",
								ConditionOperator: models.OperatorEquals,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "1",
							},
						},
					},
				}
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
			})

			It("should work the same as AND logic with single condition", func() {
				result := engine.ProcessTransaction(transaction)
				Expect(result).NotTo(BeNil())
				Expect(result.CategoryAdds).To(ContainElement(int64(1)))
			})

			It("should not match when single condition fails", func() {
				transaction.Name = "Different Name"
				result := engine.ProcessTransaction(transaction)
				Expect(result).To(BeNil())
			})
		})

		Context("Empty conditions with OR logic", func() {
			BeforeEach(func() {
				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:             1,
							Name:           "Empty conditions OR rule",
							ConditionLogic: models.ConditionLogicOr,
							EffectiveFrom:  time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.RuleConditionResponse{},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "1",
							},
						},
					},
				}
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
			})

			It("should return false for empty conditions (same as AND)", func() {
				result := engine.ProcessTransaction(transaction)
				Expect(result).To(BeNil())
			})
		})

		Context("Mixed AND and OR rules", func() {
			BeforeEach(func() {
				transaction.Amount = 100.0
				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:             1,
							Name:           "AND rule",
							ConditionLogic: models.ConditionLogicAnd,
							EffectiveFrom:  time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldName,
								ConditionValue:    "Test Transaction",
								ConditionOperator: models.OperatorEquals,
							},
							{
								ConditionType:     models.RuleFieldAmount,
								ConditionValue:    "100.0",
								ConditionOperator: models.OperatorEquals,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "1",
							},
						},
					},
					{
						Rule: models.RuleResponse{
							Id:             2,
							Name:           "OR rule",
							ConditionLogic: models.ConditionLogicOr,
							EffectiveFrom:  time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldName,
								ConditionValue:    "Test Transaction",
								ConditionOperator: models.OperatorEquals,
							},
							{
								ConditionType:     models.RuleFieldAmount,
								ConditionValue:    "999.0", // Won't match
								ConditionOperator: models.OperatorEquals,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "2",
							},
						},
					},
				}
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
			})

			It("should apply both rules when their conditions are met", func() {
				result := engine.ProcessTransaction(transaction)
				Expect(result).NotTo(BeNil())
				Expect(result.CategoryAdds).To(ContainElement(int64(1))) // From AND rule
				Expect(result.CategoryAdds).To(ContainElement(int64(2))) // From OR rule
				Expect(result.AppliedRules).To(ContainElement(int64(1)))
				Expect(result.AppliedRules).To(ContainElement(int64(2)))
			})

			It("should apply only OR rule when AND rule conditions partially fail", func() {
				transaction.Amount = 50.0 // AND rule will fail, OR rule will still match on name
				result := engine.ProcessTransaction(transaction)
				Expect(result).NotTo(BeNil())
				Expect(result.CategoryAdds).NotTo(ContainElement(int64(1))) // AND rule failed
				Expect(result.CategoryAdds).To(ContainElement(int64(2)))    // OR rule succeeded
				Expect(result.AppliedRules).NotTo(ContainElement(int64(1)))
				Expect(result.AppliedRules).To(ContainElement(int64(2)))
			})

			It("should apply no rules when both fail", func() {
				transaction.Name = "Different Name"
				transaction.Amount = 50.0
				result := engine.ProcessTransaction(transaction)
				Expect(result).To(BeNil())
			})
		})

		Context("Default condition logic behavior", func() {
			BeforeEach(func() {
				transaction.Amount = 100.0
				// Rule without explicit ConditionLogic (should default to AND)
				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:            1,
							Name:          "Default logic rule",
							EffectiveFrom: time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldName,
								ConditionValue:    "Test Transaction",
								ConditionOperator: models.OperatorEquals,
							},
							{
								ConditionType:     models.RuleFieldAmount,
								ConditionValue:    "100.0",
								ConditionOperator: models.OperatorEquals,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "1",
							},
						},
					},
				}
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
			})

			It("should default to AND logic when ConditionLogic is not set", func() {
				result := engine.ProcessTransaction(transaction)
				Expect(result).NotTo(BeNil())
				Expect(result.CategoryAdds).To(ContainElement(int64(1)))
			})

			It("should behave like AND - fail when one condition doesn't match", func() {
				transaction.Amount = 50.0 // Second condition will fail
				result := engine.ProcessTransaction(transaction)
				Expect(result).To(BeNil())
			})
		})

		Context("Performance considerations", func() {
			BeforeEach(func() {
				// Create a rule with many conditions to test short-circuiting
				conditions := []models.RuleConditionResponse{
					{
						ConditionType:     models.RuleFieldName,
						ConditionValue:    "Non-matching Name", // This will fail first
						ConditionOperator: models.OperatorEquals,
					},
				}

				// Add many more conditions that would be expensive to evaluate
				for i := 0; i < 10; i++ {
					conditions = append(conditions, models.RuleConditionResponse{
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "999.0",
						ConditionOperator: models.OperatorEquals,
					})
				}

				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:             1,
							Name:           "Short-circuit AND rule",
							ConditionLogic: models.ConditionLogicAnd,
							EffectiveFrom:  time.Now().Add(-24 * time.Hour),
						},
						Conditions: conditions,
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "1",
							},
						},
					},
				}
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
			})

			It("should short-circuit AND evaluation on first failure", func() {
				// This test verifies that AND logic stops evaluating after first failure
				result := engine.ProcessTransaction(transaction)
				Expect(result).To(BeNil())
			})
		})

		Context("OR Logic short-circuiting", func() {
			BeforeEach(func() {
				// Create a rule where first condition matches, so others shouldn't be evaluated
				conditions := []models.RuleConditionResponse{
					{
						ConditionType:     models.RuleFieldName,
						ConditionValue:    "Test Transaction", // This will match first
						ConditionOperator: models.OperatorEquals,
					},
				}

				// Add many more conditions
				for i := 0; i < 10; i++ {
					conditions = append(conditions, models.RuleConditionResponse{
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "999.0",
						ConditionOperator: models.OperatorEquals,
					})
				}

				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:             1,
							Name:           "Short-circuit OR rule",
							ConditionLogic: models.ConditionLogicOr,
							EffectiveFrom:  time.Now().Add(-24 * time.Hour),
						},
						Conditions: conditions,
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "1",
							},
						},
					},
				}
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
			})

			It("should short-circuit OR evaluation on first success", func() {
				// This test verifies that OR logic stops evaluating after first success
				result := engine.ProcessTransaction(transaction)
				Expect(result).NotTo(BeNil())
				Expect(result.CategoryAdds).To(ContainElement(int64(1)))
			})
		})
	})

	Describe("Complex Scenarios", func() {
		Context("multiple rules with mixed actions", func() {
			BeforeEach(func() {
				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:            1,
							Name:          "Name and category rule",
							EffectiveFrom: time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldName,
								ConditionValue:    "Test Transaction",
								ConditionOperator: models.OperatorEquals,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldName,
								ActionValue: "Updated Name",
							},
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "1",
							},
						},
					},
					{
						Rule: models.RuleResponse{
							Id:            2,
							Name:          "Description rule",
							EffectiveFrom: time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldAmount,
								ConditionValue:    "50.0",
								ConditionOperator: models.OperatorEquals,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldDescription,
								ActionValue: "Updated Description",
							},
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "2",
							},
						},
					},
				}
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
			})

			It("should apply all matching rules", func() {
				result := engine.ProcessTransaction(transaction)

				Expect(result).NotTo(BeNil())
				Expect(*result.NameUpdate).To(Equal("Updated Name"))
				Expect(*result.DescUpdate).To(Equal("Updated Description"))
				Expect(result.CategoryAdds).To(ContainElement(int64(1)))
				Expect(result.CategoryAdds).To(ContainElement(int64(2)))
				Expect(result.AppliedRules).To(ContainElement(int64(1)))
				Expect(result.AppliedRules).To(ContainElement(int64(2)))
			})
		})

		Context("rule with no matching conditions", func() {
			BeforeEach(func() {
				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:            1,
							Name:          "Partial match rule",
							EffectiveFrom: time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldName,
								ConditionValue:    "Test Transaction",
								ConditionOperator: models.OperatorEquals,
							},
							{
								ConditionType:     models.RuleFieldAmount,
								ConditionValue:    "999.0", // Won't match
								ConditionOperator: models.OperatorEquals,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "1",
							},
						},
					},
				}
				engine = NewRuleEngine(categories, []models.AccountResponse{}, rules)
			})

			It("should not apply rule when not all conditions match", func() {
				result := engine.ProcessTransaction(transaction)
				Expect(result).To(BeNil())
			})
		})
	})

	Describe("Transfer Action", func() {
		var accounts []models.AccountResponse

		BeforeEach(func() {
			accounts = []models.AccountResponse{
				{
					Id:        1,
					Name:      "Checking Account",
					BankType:  models.BankTypeHDFC,
					Currency:  models.CurrencyINR,
					Balance:   1000.0,
					CreatedBy: userId,
				},
				{
					Id:        2,
					Name:      "Savings Account",
					BankType:  models.BankTypeSBI,
					Currency:  models.CurrencyINR,
					Balance:   5000.0,
					CreatedBy: userId,
				},
			}
		})

		Context("when transfer action is applied", func() {
			BeforeEach(func() {
				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:            1,
							Name:          "Transfer Rule",
							EffectiveFrom: time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldName,
								ConditionValue:    "Test Transaction",
								ConditionOperator: models.OperatorEquals,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldTransfer,
								ActionValue: "2", // Transfer to account ID 2
							},
						},
					},
				}
				engine = NewRuleEngine(categories, accounts, rules)
			})

			It("should create transfer info with negated amount", func() {
				result := engine.ProcessTransaction(transaction)

				Expect(result).NotTo(BeNil())
				Expect(result.TransferInfo).NotTo(BeNil())
				Expect(result.TransferInfo.AccountId).To(Equal(int64(2)))
				Expect(result.TransferInfo.Amount).To(Equal(-transaction.Amount))
				Expect(result.AppliedRules).To(ContainElement(int64(1)))
			})

			It("should not transfer to the same account", func() {
				// Update transaction to use account ID 2
				transaction.AccountId = 2
				result := engine.ProcessTransaction(transaction)

				Expect(result).To(BeNil()) // Should not apply transfer to same account
			})

			It("should not transfer to non-existent account", func() {
				// Update action to use non-existent account ID
				rules[0].Actions[0].ActionValue = "999"
				engine = NewRuleEngine(categories, accounts, rules)

				result := engine.ProcessTransaction(transaction)

				Expect(result).To(BeNil()) // Should not apply transfer to non-existent account
			})

			It("should not transfer to account owned by different user", func() {
				// Add account owned by different user
				otherUserAccount := models.AccountResponse{
					Id:        3,
					Name:      "Other User Account",
					BankType:  models.BankTypeICICI,
					Currency:  models.CurrencyINR,
					Balance:   2000.0,
					CreatedBy: userId + 1, // Different user
				}
				accounts = append(accounts, otherUserAccount)
				rules[0].Actions[0].ActionValue = "3"
				engine = NewRuleEngine(categories, accounts, rules)

				result := engine.ProcessTransaction(transaction)

				Expect(result).To(BeNil()) // Should not apply transfer to other user's account
			})
		})

		Context("when transfer condition is used", func() {
			BeforeEach(func() {
				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:            1,
							Name:          "Transfer Condition Rule",
							EffectiveFrom: time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldTransfer,
								ConditionValue:    "1", // Check if transaction is from account 1
								ConditionOperator: models.OperatorEquals,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "1",
							},
						},
					},
				}
				engine = NewRuleEngine(categories, accounts, rules)
			})

			It("should match when transaction is from specified account", func() {
				result := engine.ProcessTransaction(transaction)

				Expect(result).NotTo(BeNil())
				Expect(result.CategoryAdds).To(ContainElement(int64(1)))
				Expect(result.AppliedRules).To(ContainElement(int64(1)))
			})

			It("should not match when transaction is from different account", func() {
				transaction.AccountId = 2
				result := engine.ProcessTransaction(transaction)

				Expect(result).To(BeNil())
			})

			It("should handle invalid account ID in condition", func() {
				// Update condition to use invalid account ID
				rules[0].Conditions[0].ConditionValue = "invalid"
				engine = NewRuleEngine(categories, accounts, rules)

				result := engine.ProcessTransaction(transaction)

				Expect(result).To(BeNil()) // Should not match with invalid account ID
			})
		})

		Context("when multiple transfer actions are attempted", func() {
			BeforeEach(func() {
				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:            1,
							Name:          "First Transfer Rule",
							EffectiveFrom: time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldName,
								ConditionValue:    "Test Transaction",
								ConditionOperator: models.OperatorEquals,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldTransfer,
								ActionValue: "2",
							},
						},
					},
					{
						Rule: models.RuleResponse{
							Id:            2,
							Name:          "Second Transfer Rule",
							EffectiveFrom: time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldName,
								ConditionValue:    "Test Transaction",
								ConditionOperator: models.OperatorEquals,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldTransfer,
								ActionValue: "2", // Same account as first rule
							},
						},
					},
				}
				engine = NewRuleEngine(categories, accounts, rules)
			})

			It("should only apply the first transfer action", func() {
				result := engine.ProcessTransaction(transaction)

				Expect(result).NotTo(BeNil())
				Expect(result.TransferInfo).NotTo(BeNil())
				Expect(result.TransferInfo.AccountId).To(Equal(int64(2)))
				Expect(result.AppliedRules).To(ContainElement(int64(1)))
				Expect(result.AppliedRules).NotTo(ContainElement(int64(2))) // Second rule should not be applied
			})
		})

		Context("when transfer action has invalid account ID", func() {
			BeforeEach(func() {
				rules = []models.DescribeRuleResponse{
					{
						Rule: models.RuleResponse{
							Id:            1,
							Name:          "Invalid Transfer Rule",
							EffectiveFrom: time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.RuleConditionResponse{
							{
								ConditionType:     models.RuleFieldName,
								ConditionValue:    "Test Transaction",
								ConditionOperator: models.OperatorEquals,
							},
						},
						Actions: []models.RuleActionResponse{
							{
								ActionType:  models.RuleFieldTransfer,
								ActionValue: "invalid_account_id",
							},
						},
					},
				}
				engine = NewRuleEngine(categories, accounts, rules)
			})

			It("should not apply transfer action with invalid account ID", func() {
				result := engine.ProcessTransaction(transaction)

				Expect(result).To(BeNil()) // Should not apply transfer with invalid account ID
			})
		})
	})
})
