package service_test

import (
	"expenses/internal/models"
	"expenses/internal/service"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RuleEngine", func() {
	var (
		engine         *service.RuleEngine
		testRules      []models.DescribeRuleResponse
		testTxns       []models.TransactionResponse
		testCategories []models.CategoryResponse
		now            time.Time
		userId         int64
	)

	BeforeEach(func() {
		now = time.Now()
		userId = 1

		testCategories = []models.CategoryResponse{
			{Id: 1, Name: "Food", CreatedBy: userId},
			{Id: 2, Name: "Transport", CreatedBy: userId},
			{Id: 3, Name: "Entertainment", CreatedBy: userId},
			{Id: 999, Name: "Other User Category", CreatedBy: 999},
		}

		engine = service.NewRuleEngine(testCategories)

		testTxns = []models.TransactionResponse{
			{
				TransactionBaseResponse: models.TransactionBaseResponse{
					Id:          1,
					Name:        "Grocery Store",
					Description: stringPtr("Weekly groceries"),
					Amount:      150.50,
					Date:        now.Add(-24 * time.Hour),
					CreatedBy:   userId,
					AccountId:   1,
				},
				CategoryIds: []int64{},
			},
			{
				TransactionBaseResponse: models.TransactionBaseResponse{
					Id:          2,
					Name:        "Restaurant Bill",
					Description: stringPtr("Dinner with friends"),
					Amount:      75.25,
					Date:        now.Add(-48 * time.Hour),
					CreatedBy:   userId,
					AccountId:   1,
				},
				CategoryIds: []int64{1},
			},
		}
	})

	Describe("ExecuteRules", func() {
		Context("when no rules exist", func() {
			It("should return empty result", func() {
				result := engine.ExecuteRules([]models.DescribeRuleResponse{}, testTxns)

				Expect(result.Changesets).To(BeEmpty())
				Expect(result.Skipped).To(BeEmpty())
			})
		})

		Context("when no transactions exist", func() {
			BeforeEach(func() {
				testRules = []models.DescribeRuleResponse{
					createTestRule(1, "Test Rule", now.Add(-24*time.Hour), userId),
				}
			})

			It("should return empty result", func() {
				result := engine.ExecuteRules(testRules, []models.TransactionResponse{})

				Expect(result.Changesets).To(BeEmpty())
				Expect(result.Skipped).To(BeEmpty())
			})
		})
	})

	Describe("Rule Condition Evaluation", func() {
		Context("Amount field conditions", func() {
			It("should evaluate equals condition correctly", func() {
				rule := createTestRule(1, "Amount Equals Rule", now.Add(-24*time.Hour), userId)
				rule.Conditions = []models.RuleConditionResponse{
					{
						Id:                1,
						RuleId:            1,
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "150.50",
						ConditionOperator: models.OperatorEquals,
					},
				}
				rule.Actions = []models.RuleActionResponse{
					{
						Id:          1,
						RuleId:      1,
						ActionType:  models.RuleFieldName,
						ActionValue: "Matched Amount",
					},
				}

				result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

				Expect(result.Changesets).To(HaveLen(1))
				Expect(result.Changesets[0].TransactionId).To(Equal(int64(1)))
				Expect(*result.Changesets[0].NameUpdate).To(Equal("Matched Amount"))
			})

			It("should evaluate greater than condition correctly", func() {
				rule := createTestRule(1, "Amount Greater Rule", now.Add(-24*time.Hour), userId)
				rule.Conditions = []models.RuleConditionResponse{
					{
						Id:                1,
						RuleId:            1,
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "100.00",
						ConditionOperator: models.OperatorGreater,
					},
				}
				rule.Actions = []models.RuleActionResponse{
					{
						Id:          1,
						RuleId:      1,
						ActionType:  models.RuleFieldName,
						ActionValue: "Large Purchase",
					},
				}

				result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

				Expect(result.Changesets).To(HaveLen(1))
				Expect(result.Changesets[0].TransactionId).To(Equal(int64(1)))
				Expect(*result.Changesets[0].NameUpdate).To(Equal("Large Purchase"))
			})

			It("should evaluate less than condition correctly", func() {
				rule := createTestRule(1, "Amount Lower Rule", now.Add(-72*time.Hour), userId)
				rule.Conditions = []models.RuleConditionResponse{
					{
						Id:                1,
						RuleId:            1,
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "100.00",
						ConditionOperator: models.OperatorLower,
					},
				}
				rule.Actions = []models.RuleActionResponse{
					{
						Id:          1,
						RuleId:      1,
						ActionType:  models.RuleFieldName,
						ActionValue: "Small Purchase",
					},
				}

				result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

				Expect(result.Changesets).To(HaveLen(1))
				Expect(result.Changesets[0].TransactionId).To(Equal(int64(2)))
				Expect(*result.Changesets[0].NameUpdate).To(Equal("Small Purchase"))
			})

			It("should handle invalid amount values gracefully", func() {
				rule := createTestRule(1, "Invalid Amount Rule", now.Add(-24*time.Hour), userId)
				rule.Conditions = []models.RuleConditionResponse{
					{
						Id:                1,
						RuleId:            1,
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "invalid-amount",
						ConditionOperator: models.OperatorEquals,
					},
				}
				rule.Actions = []models.RuleActionResponse{
					{
						Id:          1,
						RuleId:      1,
						ActionType:  models.RuleFieldName,
						ActionValue: "Should Not Match",
					},
				}

				result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

				Expect(result.Changesets).To(BeEmpty())
			})
		})

		Context("Name field conditions", func() {
			It("should evaluate equals condition (case insensitive)", func() {
				rule := createTestRule(1, "Name Equals Rule", now.Add(-24*time.Hour), userId)
				rule.Conditions = []models.RuleConditionResponse{
					{
						Id:                1,
						RuleId:            1,
						ConditionType:     models.RuleFieldName,
						ConditionValue:    "grocery store",
						ConditionOperator: models.OperatorEquals,
					},
				}
				rule.Actions = []models.RuleActionResponse{
					{
						Id:          1,
						RuleId:      1,
						ActionType:  models.RuleFieldCategory,
						ActionValue: "1",
					},
				}

				result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

				Expect(result.Changesets).To(HaveLen(1))
				Expect(result.Changesets[0].TransactionId).To(Equal(int64(1)))
				Expect(result.Changesets[0].CategoryAdds).To(ContainElement(int64(1)))
			})

			It("should evaluate contains condition", func() {
				rule := createTestRule(1, "Name Contains Rule", now.Add(-72*time.Hour), userId)
				rule.Conditions = []models.RuleConditionResponse{
					{
						Id:                1,
						RuleId:            1,
						ConditionType:     models.RuleFieldName,
						ConditionValue:    "restaurant",
						ConditionOperator: models.OperatorContains,
					},
				}
				rule.Actions = []models.RuleActionResponse{
					{
						Id:          1,
						RuleId:      1,
						ActionType:  models.RuleFieldCategory,
						ActionValue: "3",
					},
				}

				result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

				Expect(result.Changesets).To(HaveLen(1))
				Expect(result.Changesets[0].TransactionId).To(Equal(int64(2)))
				Expect(result.Changesets[0].CategoryAdds).To(ContainElement(int64(3)))
			})
		})

		Context("Description field conditions", func() {

			Context("Multiple conditions (AND logic)", func() {
				It("should apply rule when all conditions match", func() {
					rule := createTestRule(1, "Multiple Conditions Rule", now.Add(-24*time.Hour), userId)
					rule.Conditions = []models.RuleConditionResponse{
						{
							Id:                1,
							RuleId:            1,
							ConditionType:     models.RuleFieldAmount,
							ConditionValue:    "100.00",
							ConditionOperator: models.OperatorGreater,
						},
						{
							Id:                2,
							RuleId:            1,
							ConditionType:     models.RuleFieldName,
							ConditionValue:    "grocery",
							ConditionOperator: models.OperatorContains,
						},
					}
					rule.Actions = []models.RuleActionResponse{
						{
							Id:          1,
							RuleId:      1,
							ActionType:  models.RuleFieldName,
							ActionValue: "Large Grocery Purchase",
						},
					}

					result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

					Expect(result.Changesets).To(HaveLen(1))
					Expect(result.Changesets[0].TransactionId).To(Equal(int64(1)))
					Expect(*result.Changesets[0].NameUpdate).To(Equal("Large Grocery Purchase"))
				})

				It("should not apply rule when some conditions don't match", func() {
					rule := createTestRule(1, "Partial Match Rule", now.Add(-24*time.Hour), userId)
					rule.Conditions = []models.RuleConditionResponse{
						{
							Id:                1,
							RuleId:            1,
							ConditionType:     models.RuleFieldAmount,
							ConditionValue:    "200.00",
							ConditionOperator: models.OperatorGreater,
						},
						{
							Id:                2,
							RuleId:            1,
							ConditionType:     models.RuleFieldName,
							ConditionValue:    "grocery",
							ConditionOperator: models.OperatorContains,
						},
					}
					rule.Actions = []models.RuleActionResponse{
						{
							Id:          1,
							RuleId:      1,
							ActionType:  models.RuleFieldName,
							ActionValue: "Should Not Match",
						},
					}

					result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

					Expect(result.Changesets).To(BeEmpty())
				})
			})
		})

		Describe("Rule Action Application", func() {
			Context("Name actions", func() {
				It("should update transaction name", func() {
					rule := createTestRule(1, "Name Update Rule", now.Add(-24*time.Hour), userId)
					rule.Conditions = []models.RuleConditionResponse{
						{
							Id:                1,
							RuleId:            1,
							ConditionType:     models.RuleFieldAmount,
							ConditionValue:    "100.00",
							ConditionOperator: models.OperatorGreater,
						},
					}
					rule.Actions = []models.RuleActionResponse{
						{
							Id:          1,
							RuleId:      1,
							ActionType:  models.RuleFieldName,
							ActionValue: "Updated Name",
						},
					}

					result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

					Expect(result.Changesets).To(HaveLen(1))
					Expect(result.Changesets[0].TransactionId).To(Equal(int64(1)))
					Expect(*result.Changesets[0].NameUpdate).To(Equal("Updated Name"))
					Expect(result.Changesets[0].UpdatedFields).To(ContainElement(models.RuleFieldName))
				})

				It("should respect first-rule-wins policy", func() {
					rule1 := createTestRule(1, "First Rule", now.Add(-24*time.Hour), userId)
					rule1.Conditions = []models.RuleConditionResponse{
						{
							Id:                1,
							RuleId:            1,
							ConditionType:     models.RuleFieldAmount,
							ConditionValue:    "100.00",
							ConditionOperator: models.OperatorGreater,
						},
					}
					rule1.Actions = []models.RuleActionResponse{
						{
							Id:          1,
							RuleId:      1,
							ActionType:  models.RuleFieldName,
							ActionValue: "First Rule Name",
						},
					}

					rule2 := createTestRule(2, "Second Rule", now.Add(-24*time.Hour), userId)
					rule2.Conditions = []models.RuleConditionResponse{
						{
							Id:                2,
							RuleId:            2,
							ConditionType:     models.RuleFieldAmount,
							ConditionValue:    "100.00",
							ConditionOperator: models.OperatorGreater,
						},
					}
					rule2.Actions = []models.RuleActionResponse{
						{
							Id:          2,
							RuleId:      2,
							ActionType:  models.RuleFieldName,
							ActionValue: "Second Rule Name",
						},
					}

					result := engine.ExecuteRules([]models.DescribeRuleResponse{rule1, rule2}, testTxns)

					Expect(result.Changesets).To(HaveLen(1))
					Expect(result.Changesets[0].TransactionId).To(Equal(int64(1)))
					Expect(*result.Changesets[0].NameUpdate).To(Equal("First Rule Name"))
					Expect(result.Changesets[0].AppliedRules).To(Equal([]int64{1}))
				})
			})

			Context("Description actions", func() {
				It("should update transaction description", func() {
					rule := createTestRule(1, "Description Update Rule", now.Add(-72*time.Hour), userId)
					rule.Conditions = []models.RuleConditionResponse{
						{
							Id:                1,
							RuleId:            1,
							ConditionType:     models.RuleFieldAmount,
							ConditionValue:    "100.00",
							ConditionOperator: models.OperatorGreater,
						},
					}
					rule.Actions = []models.RuleActionResponse{
						{
							Id:          1,
							RuleId:      1,
							ActionType:  models.RuleFieldDescription,
							ActionValue: "Updated Description",
						},
					}

					result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

					Expect(result.Changesets).To(HaveLen(1))
					Expect(result.Changesets[0].TransactionId).To(Equal(int64(1)))
					Expect(*result.Changesets[0].DescUpdate).To(Equal("Updated Description"))
					Expect(result.Changesets[0].UpdatedFields).To(ContainElement(models.RuleFieldDescription))
				})

				It("should respect first-rule-wins policy for descriptions", func() {
					rule1 := createTestRule(1, "First Desc Rule", now.Add(-72*time.Hour), userId)
					rule1.Conditions = []models.RuleConditionResponse{
						{
							Id:                1,
							RuleId:            1,
							ConditionType:     models.RuleFieldAmount,
							ConditionValue:    "100.00",
							ConditionOperator: models.OperatorGreater,
						},
					}
					rule1.Actions = []models.RuleActionResponse{
						{
							Id:          1,
							RuleId:      1,
							ActionType:  models.RuleFieldDescription,
							ActionValue: "First Description",
						},
					}

					rule2 := createTestRule(2, "Second Desc Rule", now.Add(-72*time.Hour), userId)
					rule2.Conditions = []models.RuleConditionResponse{
						{
							Id:                2,
							RuleId:            2,
							ConditionType:     models.RuleFieldAmount,
							ConditionValue:    "100.00",
							ConditionOperator: models.OperatorGreater,
						},
					}
					rule2.Actions = []models.RuleActionResponse{
						{
							Id:          2,
							RuleId:      2,
							ActionType:  models.RuleFieldDescription,
							ActionValue: "Second Description",
						},
					}

					result := engine.ExecuteRules([]models.DescribeRuleResponse{rule1, rule2}, testTxns)

					Expect(result.Changesets).To(HaveLen(1))
					Expect(*result.Changesets[0].DescUpdate).To(Equal("First Description"))
					Expect(result.Changesets[0].AppliedRules).To(Equal([]int64{1}))
				})
			})

			Context("Category actions", func() {
				It("should add category to transaction", func() {
					rule := createTestRule(1, "Category Add Rule", now.Add(-24*time.Hour), userId)
					rule.Conditions = []models.RuleConditionResponse{
						{
							Id:                1,
							RuleId:            1,
							ConditionType:     models.RuleFieldAmount,
							ConditionValue:    "100.00",
							ConditionOperator: models.OperatorGreater,
						},
					}
					rule.Actions = []models.RuleActionResponse{
						{
							Id:          1,
							RuleId:      1,
							ActionType:  models.RuleFieldCategory,
							ActionValue: "2",
						},
					}

					result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

					Expect(result.Changesets).To(HaveLen(1))
					Expect(result.Changesets[0].TransactionId).To(Equal(int64(1)))
					Expect(result.Changesets[0].CategoryAdds).To(ContainElement(int64(2)))
					Expect(result.Changesets[0].UpdatedFields).To(ContainElement(models.RuleFieldCategory))
				})

				It("should skip invalid category IDs", func() {
					rule := createTestRule(1, "Invalid Category Rule", now.Add(-24*time.Hour), userId)
					rule.Conditions = []models.RuleConditionResponse{
						{
							Id:                1,
							RuleId:            1,
							ConditionType:     models.RuleFieldAmount,
							ConditionValue:    "100.00",
							ConditionOperator: models.OperatorGreater,
						},
					}
					rule.Actions = []models.RuleActionResponse{
						{
							Id:          1,
							RuleId:      1,
							ActionType:  models.RuleFieldCategory,
							ActionValue: "invalid-id",
						},
					}

					result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

					Expect(result.Changesets).To(BeEmpty())
				})

				It("should skip categories from different users", func() {
					rule := createTestRule(1, "Other User Category Rule", now.Add(-72*time.Hour), userId)
					rule.Conditions = []models.RuleConditionResponse{
						{
							Id:                1,
							RuleId:            1,
							ConditionType:     models.RuleFieldAmount,
							ConditionValue:    "100.00",
							ConditionOperator: models.OperatorGreater,
						},
					}
					rule.Actions = []models.RuleActionResponse{
						{
							Id:          1,
							RuleId:      1,
							ActionType:  models.RuleFieldCategory,
							ActionValue: "999",
						},
					}

					result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

					Expect(result.Changesets).To(BeEmpty())
				})

				It("should preserve existing categories (additive behavior)", func() {
					rule := createTestRule(1, "Category Additive Rule", now.Add(-72*time.Hour), userId)
					rule.Conditions = []models.RuleConditionResponse{
						{
							Id:                1,
							RuleId:            1,
							ConditionType:     models.RuleFieldAmount,
							ConditionValue:    "50.00",
							ConditionOperator: models.OperatorGreater,
						},
					}
					rule.Actions = []models.RuleActionResponse{
						{
							Id:          1,
							RuleId:      1,
							ActionType:  models.RuleFieldCategory,
							ActionValue: "2",
						},
					}

					result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

					Expect(result.Changesets).To(HaveLen(2))

					// Transaction 1 should get category 2 added
					txn1Changeset := findChangesetByTxnId(result.Changesets, 1)
					Expect(txn1Changeset).ToNot(BeNil())
					Expect(txn1Changeset.CategoryAdds).To(ContainElement(int64(2)))

					// Transaction 2 should get category 2 added (preserving existing category 1)
					txn2Changeset := findChangesetByTxnId(result.Changesets, 2)
					Expect(txn2Changeset).ToNot(BeNil())
					Expect(txn2Changeset.CategoryAdds).To(ContainElement(int64(2)))
				})

				It("should not add duplicate categories", func() {
					rule := createTestRule(1, "Duplicate Category Rule", now.Add(-72*time.Hour), userId)
					rule.Conditions = []models.RuleConditionResponse{
						{
							Id:                1,
							RuleId:            1,
							ConditionType:     models.RuleFieldAmount,
							ConditionValue:    "50.00",
							ConditionOperator: models.OperatorGreater,
						},
					}
					rule.Actions = []models.RuleActionResponse{
						{
							Id:          1,
							RuleId:      1,
							ActionType:  models.RuleFieldCategory,
							ActionValue: "1",
						},
					}

					result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

					// Transaction 2 already has category 1, so no changeset should be generated for it
					Expect(result.Changesets).To(HaveLen(1))
					Expect(result.Changesets[0].TransactionId).To(Equal(int64(1)))
					Expect(result.Changesets[0].CategoryAdds).To(ContainElement(int64(1)))
				})

				It("should handle multiple category actions in same rule", func() {
					rule := createTestRule(1, "Multi Category Action Rule", now.Add(-72*time.Hour), userId)
					rule.Conditions = []models.RuleConditionResponse{
						{
							Id:                1,
							RuleId:            1,
							ConditionType:     models.RuleFieldAmount,
							ConditionValue:    "100.00",
							ConditionOperator: models.OperatorGreater,
						},
					}
					rule.Actions = []models.RuleActionResponse{
						{
							Id:          1,
							RuleId:      1,
							ActionType:  models.RuleFieldCategory,
							ActionValue: "2",
						},
						{
							Id:          2,
							RuleId:      1,
							ActionType:  models.RuleFieldCategory,
							ActionValue: "3",
						},
					}

					result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

					Expect(result.Changesets).To(HaveLen(1))
					changeset := result.Changesets[0]
					Expect(changeset.TransactionId).To(Equal(int64(1)))
					Expect(changeset.CategoryAdds).To(ContainElement(int64(2)))
					Expect(changeset.CategoryAdds).To(ContainElement(int64(3)))
					Expect(changeset.UpdatedFields).To(ContainElement(models.RuleFieldCategory))
				})
			})
		})

		Describe("Rule Effective Date Validation", func() {
			It("should apply rule when effective date is before transaction date", func() {
				rule := createTestRule(1, "Past Effective Rule", now.Add(-48*time.Hour), userId)
				rule.Conditions = []models.RuleConditionResponse{
					{
						Id:                1,
						RuleId:            1,
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "100.00",
						ConditionOperator: models.OperatorGreater,
					},
				}
				rule.Actions = []models.RuleActionResponse{
					{
						Id:          1,
						RuleId:      1,
						ActionType:  models.RuleFieldName,
						ActionValue: "Past Rule Applied",
					},
				}

				result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

				Expect(result.Changesets).To(HaveLen(1))
				Expect(*result.Changesets[0].NameUpdate).To(Equal("Past Rule Applied"))
			})

			It("should not apply rule when effective date is after transaction date", func() {
				rule := createTestRule(1, "Future Effective Rule", now.Add(24*time.Hour), userId)
				rule.Conditions = []models.RuleConditionResponse{
					{
						Id:                1,
						RuleId:            1,
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "100.00",
						ConditionOperator: models.OperatorGreater,
					},
				}
				rule.Actions = []models.RuleActionResponse{
					{
						Id:          1,
						RuleId:      1,
						ActionType:  models.RuleFieldName,
						ActionValue: "Future Rule Applied",
					},
				}

				result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

				Expect(result.Changesets).To(BeEmpty())
			})
		})

		Describe("Edge Cases", func() {
			Context("Empty rule conditions", func() {
				It("should not apply rule when no conditions are defined", func() {
					rule := createTestRule(1, "No Conditions Rule", now.Add(-24*time.Hour), userId)
					rule.Actions = []models.RuleActionResponse{
						{
							Id:          1,
							RuleId:      1,
							ActionType:  models.RuleFieldName,
							ActionValue: "Should Not Apply",
						},
					}

					result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

					Expect(result.Changesets).To(BeEmpty())
				})
			})

			Context("Multiple rules on same transaction", func() {
				It("should apply all applicable rules correctly", func() {
					rule1 := createTestRule(1, "Name Rule", now.Add(-24*time.Hour), userId)
					rule1.Conditions = []models.RuleConditionResponse{
						{
							Id:                1,
							RuleId:            1,
							ConditionType:     models.RuleFieldAmount,
							ConditionValue:    "100.00",
							ConditionOperator: models.OperatorGreater,
						},
					}
					rule1.Actions = []models.RuleActionResponse{
						{
							Id:          1,
							RuleId:      1,
							ActionType:  models.RuleFieldName,
							ActionValue: "Large Purchase",
						},
					}

					rule2 := createTestRule(2, "Category Rule", now.Add(-24*time.Hour), userId)
					rule2.Conditions = []models.RuleConditionResponse{
						{
							Id:                2,
							RuleId:            2,
							ConditionType:     models.RuleFieldAmount,
							ConditionValue:    "100.00",
							ConditionOperator: models.OperatorGreater,
						},
					}
					rule2.Actions = []models.RuleActionResponse{
						{
							Id:          2,
							RuleId:      2,
							ActionType:  models.RuleFieldCategory,
							ActionValue: "2",
						},
					}

					result := engine.ExecuteRules([]models.DescribeRuleResponse{rule1, rule2}, testTxns)

					Expect(result.Changesets).To(HaveLen(1))
					changeset := result.Changesets[0]
					Expect(changeset.TransactionId).To(Equal(int64(1)))
					Expect(*changeset.NameUpdate).To(Equal("Large Purchase"))
					Expect(changeset.CategoryAdds).To(ContainElement(int64(2)))
					Expect(changeset.AppliedRules).To(Equal([]int64{1, 2}))
					Expect(changeset.UpdatedFields).To(ContainElement(models.RuleFieldName))
					Expect(changeset.UpdatedFields).To(ContainElement(models.RuleFieldCategory))
				})
			})
		})

		It("should evaluate equals condition", func() {
			rule := createTestRule(1, "Description Equals Rule", now.Add(-72*time.Hour), userId)
			rule.Conditions = []models.RuleConditionResponse{
				{
					Id:                1,
					RuleId:            1,
					ConditionType:     models.RuleFieldDescription,
					ConditionValue:    "Weekly groceries",
					ConditionOperator: models.OperatorEquals,
				},
			}
			rule.Actions = []models.RuleActionResponse{
				{
					Id:          1,
					RuleId:      1,
					ActionType:  models.RuleFieldCategory,
					ActionValue: "1",
				},
			}

			result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

			Expect(result.Changesets).To(HaveLen(1))
			Expect(result.Changesets[0].TransactionId).To(Equal(int64(1)))
			Expect(result.Changesets[0].CategoryAdds).To(ContainElement(int64(1)))
		})

		It("should evaluate contains condition", func() {
			rule := createTestRule(1, "Description Contains Rule", now.Add(-72*time.Hour), userId)
			rule.Conditions = []models.RuleConditionResponse{
				{
					Id:                1,
					RuleId:            1,
					ConditionType:     models.RuleFieldDescription,
					ConditionValue:    "friends",
					ConditionOperator: models.OperatorContains,
				},
			}
			rule.Actions = []models.RuleActionResponse{
				{
					Id:          1,
					RuleId:      1,
					ActionType:  models.RuleFieldCategory,
					ActionValue: "3",
				},
			}

			result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

			Expect(result.Changesets).To(HaveLen(1))
			Expect(result.Changesets[0].TransactionId).To(Equal(int64(2)))
			Expect(result.Changesets[0].CategoryAdds).To(ContainElement(int64(3)))
		})

		It("should handle nil descriptions", func() {
			txnWithNilDesc := models.TransactionResponse{
				TransactionBaseResponse: models.TransactionBaseResponse{
					Id:          3,
					Name:        "Test Transaction",
					Description: nil,
					Amount:      100.0,
					Date:        now.Add(-24 * time.Hour),
					CreatedBy:   userId,
					AccountId:   1,
				},
				CategoryIds: []int64{},
			}

			rule := createTestRule(1, "Nil Description Rule", now.Add(-72*time.Hour), userId)
			rule.Conditions = []models.RuleConditionResponse{
				{
					Id:                1,
					RuleId:            1,
					ConditionType:     models.RuleFieldDescription,
					ConditionValue:    "",
					ConditionOperator: models.OperatorEquals,
				},
			}
			rule.Actions = []models.RuleActionResponse{
				{
					Id:          1,
					RuleId:      1,
					ActionType:  models.RuleFieldName,
					ActionValue: "No Description",
				},
			}

			result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, []models.TransactionResponse{txnWithNilDesc})

			Expect(result.Changesets).To(HaveLen(1))
			Expect(result.Changesets[0].TransactionId).To(Equal(int64(3)))
			Expect(*result.Changesets[0].NameUpdate).To(Equal("No Description"))
		})
	})

	Context("Category field conditions", func() {
		It("should evaluate equals condition with single category", func() {
			rule := createTestRule(1, "Category Equals Rule", now.Add(-72*time.Hour), userId)
			rule.Conditions = []models.RuleConditionResponse{
				{
					Id:                1,
					RuleId:            1,
					ConditionType:     models.RuleFieldCategory,
					ConditionValue:    "1",
					ConditionOperator: models.OperatorEquals,
				},
			}
			rule.Actions = []models.RuleActionResponse{
				{
					Id:          1,
					RuleId:      1,
					ActionType:  models.RuleFieldName,
					ActionValue: "Has Food Category",
				},
			}

			result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

			Expect(result.Changesets).To(HaveLen(1))
			Expect(result.Changesets[0].TransactionId).To(Equal(int64(2)))
			Expect(*result.Changesets[0].NameUpdate).To(Equal("Has Food Category"))
		})

		It("should evaluate equals condition with multiple categories", func() {
			txnWithMultipleCategories := models.TransactionResponse{
				TransactionBaseResponse: models.TransactionBaseResponse{
					Id:          4,
					Name:        "Multi Category Transaction",
					Description: stringPtr("Has multiple categories"),
					Amount:      200.0,
					Date:        now.Add(-24 * time.Hour),
					CreatedBy:   userId,
					AccountId:   1,
				},
				CategoryIds: []int64{1, 2, 3},
			}

			rule := createTestRule(1, "Multi Category Rule", now.Add(-72*time.Hour), userId)
			rule.Conditions = []models.RuleConditionResponse{
				{
					Id:                1,
					RuleId:            1,
					ConditionType:     models.RuleFieldCategory,
					ConditionValue:    "2",
					ConditionOperator: models.OperatorEquals,
				},
			}
			rule.Actions = []models.RuleActionResponse{
				{
					Id:          1,
					RuleId:      1,
					ActionType:  models.RuleFieldName,
					ActionValue: "Has Transport Category",
				},
			}

			result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, []models.TransactionResponse{txnWithMultipleCategories})

			Expect(result.Changesets).To(HaveLen(1))
			Expect(result.Changesets[0].TransactionId).To(Equal(int64(4)))
			Expect(*result.Changesets[0].NameUpdate).To(Equal("Has Transport Category"))
		})

		It("should handle invalid category IDs in conditions", func() {
			rule := createTestRule(1, "Invalid Category Condition Rule", now.Add(-72*time.Hour), userId)
			rule.Conditions = []models.RuleConditionResponse{
				{
					Id:                1,
					RuleId:            1,
					ConditionType:     models.RuleFieldCategory,
					ConditionValue:    "invalid-id",
					ConditionOperator: models.OperatorEquals,
				},
			}
			rule.Actions = []models.RuleActionResponse{
				{
					Id:          1,
					RuleId:      1,
					ActionType:  models.RuleFieldName,
					ActionValue: "Should Not Match",
				},
			}

			result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

			Expect(result.Changesets).To(BeEmpty())
		})
	})
	Describe("Advanced Edge Cases", func() {
		Context("Rules with no actions", func() {
			It("should not create changeset when no actions are defined", func() {
				rule := createTestRule(1, "No Actions Rule", now.Add(-72*time.Hour), userId)
				rule.Conditions = []models.RuleConditionResponse{
					{
						Id:                1,
						RuleId:            1,
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "100.00",
						ConditionOperator: models.OperatorGreater,
					},
				}
				// No actions defined

				result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

				Expect(result.Changesets).To(BeEmpty())
			})
		})

		Context("Transactions with minimal data", func() {
			It("should handle transactions with minimal required fields", func() {
				minimalTxn := models.TransactionResponse{
					TransactionBaseResponse: models.TransactionBaseResponse{
						Id:          5,
						Name:        "",
						Description: nil,
						Amount:      0.0,
						Date:        now.Add(-24 * time.Hour),
						CreatedBy:   userId,
						AccountId:   1,
					},
					CategoryIds: []int64{},
				}

				rule := createTestRule(1, "Minimal Data Rule", now.Add(-72*time.Hour), userId)
				rule.Conditions = []models.RuleConditionResponse{
					{
						Id:                1,
						RuleId:            1,
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "0.0",
						ConditionOperator: models.OperatorEquals,
					},
				}
				rule.Actions = []models.RuleActionResponse{
					{
						Id:          1,
						RuleId:      1,
						ActionType:  models.RuleFieldName,
						ActionValue: "Zero Amount Transaction",
					},
				}

				result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, []models.TransactionResponse{minimalTxn})

				Expect(result.Changesets).To(HaveLen(1))
				Expect(result.Changesets[0].TransactionId).To(Equal(int64(5)))
				Expect(*result.Changesets[0].NameUpdate).To(Equal("Zero Amount Transaction"))
			})
		})

		Context("Multiple rules with different field combinations", func() {
			It("should handle complex rule interactions", func() {
				rule1 := createTestRule(1, "Amount Rule", now.Add(-72*time.Hour), userId)
				rule1.Conditions = []models.RuleConditionResponse{
					{
						Id:                1,
						RuleId:            1,
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "100.00",
						ConditionOperator: models.OperatorGreater,
					},
				}
				rule1.Actions = []models.RuleActionResponse{
					{
						Id:          1,
						RuleId:      1,
						ActionType:  models.RuleFieldName,
						ActionValue: "Large Amount",
					},
				}

				rule2 := createTestRule(2, "Name Rule", now.Add(-72*time.Hour), userId)
				rule2.Conditions = []models.RuleConditionResponse{
					{
						Id:                2,
						RuleId:            2,
						ConditionType:     models.RuleFieldName,
						ConditionValue:    "grocery",
						ConditionOperator: models.OperatorContains,
					},
				}
				rule2.Actions = []models.RuleActionResponse{
					{
						Id:          2,
						RuleId:      2,
						ActionType:  models.RuleFieldCategory,
						ActionValue: "1",
					},
				}

				rule3 := createTestRule(3, "Description Rule", now.Add(-72*time.Hour), userId)
				rule3.Conditions = []models.RuleConditionResponse{
					{
						Id:                3,
						RuleId:            3,
						ConditionType:     models.RuleFieldDescription,
						ConditionValue:    "weekly",
						ConditionOperator: models.OperatorContains,
					},
				}
				rule3.Actions = []models.RuleActionResponse{
					{
						Id:          3,
						RuleId:      3,
						ActionType:  models.RuleFieldDescription,
						ActionValue: "Recurring Weekly Expense",
					},
				}

				result := engine.ExecuteRules([]models.DescribeRuleResponse{rule1, rule2, rule3}, testTxns)

				// Transaction 1 should match all three rules
				Expect(result.Changesets).To(HaveLen(1))
				changeset := result.Changesets[0]
				Expect(changeset.TransactionId).To(Equal(int64(1)))
				Expect(*changeset.NameUpdate).To(Equal("Large Amount")) // First rule wins for name
				Expect(*changeset.DescUpdate).To(Equal("Recurring Weekly Expense"))
				Expect(changeset.CategoryAdds).To(ContainElement(int64(1)))
				Expect(changeset.AppliedRules).To(Equal([]int64{1, 2, 3}))
			})
		})

		Context("Field deduplication", func() {
			It("should properly deduplicate updated fields", func() {
				rule := createTestRule(1, "Multi Field Update Rule", now.Add(-72*time.Hour), userId)
				rule.Conditions = []models.RuleConditionResponse{
					{
						Id:                1,
						RuleId:            1,
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "100.00",
						ConditionOperator: models.OperatorGreater,
					},
				}
				rule.Actions = []models.RuleActionResponse{
					{
						Id:          1,
						RuleId:      1,
						ActionType:  models.RuleFieldCategory,
						ActionValue: "1",
					},
					{
						Id:          2,
						RuleId:      1,
						ActionType:  models.RuleFieldCategory,
						ActionValue: "2",
					},
					{
						Id:          3,
						RuleId:      1,
						ActionType:  models.RuleFieldName,
						ActionValue: "Updated Name",
					},
				}

				result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

				Expect(result.Changesets).To(HaveLen(1))
				changeset := result.Changesets[0]

				// Should have both category and name fields, but category should appear only once
				categoryFieldCount := 0
				nameFieldCount := 0
				for _, field := range changeset.UpdatedFields {
					if field == models.RuleFieldCategory {
						categoryFieldCount++
					}
					if field == models.RuleFieldName {
						nameFieldCount++
					}
				}
				Expect(categoryFieldCount).To(Equal(1))
				Expect(nameFieldCount).To(Equal(1))
			})
		})
	})

	Describe("Unsupported Operators", func() {
		Context("Amount field with unsupported operators", func() {
			It("should return false for unsupported operators", func() {
				rule := createTestRule(1, "Unsupported Operator Rule", now.Add(-72*time.Hour), userId)
				rule.Conditions = []models.RuleConditionResponse{
					{
						Id:                1,
						RuleId:            1,
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "100.00",
						ConditionOperator: models.OperatorContains, // Unsupported for amount
					},
				}
				rule.Actions = []models.RuleActionResponse{
					{
						Id:          1,
						RuleId:      1,
						ActionType:  models.RuleFieldName,
						ActionValue: "Should Not Match",
					},
				}

				result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

				Expect(result.Changesets).To(BeEmpty())
			})
		})

		Context("String fields with unsupported operators", func() {
			It("should return false for unsupported operators on name field", func() {
				rule := createTestRule(1, "Unsupported String Operator Rule", now.Add(-72*time.Hour), userId)
				rule.Conditions = []models.RuleConditionResponse{
					{
						Id:                1,
						RuleId:            1,
						ConditionType:     models.RuleFieldName,
						ConditionValue:    "grocery",
						ConditionOperator: models.OperatorGreater, // Unsupported for string
					},
				}
				rule.Actions = []models.RuleActionResponse{
					{
						Id:          1,
						RuleId:      1,
						ActionType:  models.RuleFieldName,
						ActionValue: "Should Not Match",
					},
				}

				result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

				Expect(result.Changesets).To(BeEmpty())
			})
		})

		Context("Category field with unsupported operators", func() {
			It("should return false for unsupported operators on category field", func() {
				rule := createTestRule(1, "Unsupported Category Operator Rule", now.Add(-72*time.Hour), userId)
				rule.Conditions = []models.RuleConditionResponse{
					{
						Id:                1,
						RuleId:            1,
						ConditionType:     models.RuleFieldCategory,
						ConditionValue:    "1",
						ConditionOperator: models.OperatorContains, // Unsupported for category
					},
				}
				rule.Actions = []models.RuleActionResponse{
					{
						Id:          1,
						RuleId:      1,
						ActionType:  models.RuleFieldName,
						ActionValue: "Should Not Match",
					},
				}

				result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

				Expect(result.Changesets).To(BeEmpty())
			})
		})
	})

	Describe("Complex Integration Scenarios", func() {
		Context("Rule with all field types in conditions", func() {
			It("should handle comprehensive rule conditions", func() {
				rule := createTestRule(1, "Comprehensive Rule", now.Add(-72*time.Hour), userId)
				rule.Conditions = []models.RuleConditionResponse{
					{
						Id:                1,
						RuleId:            1,
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "70.00",
						ConditionOperator: models.OperatorGreater,
					},
					{
						Id:                2,
						RuleId:            1,
						ConditionType:     models.RuleFieldName,
						ConditionValue:    "restaurant",
						ConditionOperator: models.OperatorContains,
					},
					{
						Id:                3,
						RuleId:            1,
						ConditionType:     models.RuleFieldDescription,
						ConditionValue:    "dinner",
						ConditionOperator: models.OperatorContains,
					},
					{
						Id:                4,
						RuleId:            1,
						ConditionType:     models.RuleFieldCategory,
						ConditionValue:    "1",
						ConditionOperator: models.OperatorEquals,
					},
				}
				rule.Actions = []models.RuleActionResponse{
					{
						Id:          1,
						RuleId:      1,
						ActionType:  models.RuleFieldName,
						ActionValue: "Dinner Out",
					},
				}

				result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

				// Transaction 2 should match all conditions
				Expect(result.Changesets).To(HaveLen(1))
				Expect(result.Changesets[0].TransactionId).To(Equal(int64(2)))
				Expect(*result.Changesets[0].NameUpdate).To(Equal("Dinner Out"))
			})
		})

		Context("Rule with all action types", func() {
			It("should handle rule with name, description, and category actions", func() {
				rule := createTestRule(1, "All Actions Rule", now.Add(-72*time.Hour), userId)
				rule.Conditions = []models.RuleConditionResponse{
					{
						Id:                1,
						RuleId:            1,
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "100.00",
						ConditionOperator: models.OperatorGreater,
					},
				}
				rule.Actions = []models.RuleActionResponse{
					{
						Id:          1,
						RuleId:      1,
						ActionType:  models.RuleFieldName,
						ActionValue: "Large Purchase",
					},
					{
						Id:          2,
						RuleId:      1,
						ActionType:  models.RuleFieldDescription,
						ActionValue: "Automatically categorized large purchase",
					},
					{
						Id:          3,
						RuleId:      1,
						ActionType:  models.RuleFieldCategory,
						ActionValue: "2",
					},
				}

				result := engine.ExecuteRules([]models.DescribeRuleResponse{rule}, testTxns)

				Expect(result.Changesets).To(HaveLen(1))
				changeset := result.Changesets[0]
				Expect(changeset.TransactionId).To(Equal(int64(1)))
				Expect(*changeset.NameUpdate).To(Equal("Large Purchase"))
				Expect(*changeset.DescUpdate).To(Equal("Automatically categorized large purchase"))
				Expect(changeset.CategoryAdds).To(ContainElement(int64(2)))
				Expect(changeset.UpdatedFields).To(ContainElement(models.RuleFieldName))
				Expect(changeset.UpdatedFields).To(ContainElement(models.RuleFieldDescription))
				Expect(changeset.UpdatedFields).To(ContainElement(models.RuleFieldCategory))
			})
		})

		Context("Multiple rules with overlapping conditions", func() {
			It("should handle rules with overlapping but different conditions", func() {
				rule1 := createTestRule(1, "Amount and Name Rule", now.Add(-72*time.Hour), userId)
				rule1.Conditions = []models.RuleConditionResponse{
					{
						Id:                1,
						RuleId:            1,
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "100.00",
						ConditionOperator: models.OperatorGreater,
					},
					{
						Id:                2,
						RuleId:            1,
						ConditionType:     models.RuleFieldName,
						ConditionValue:    "grocery",
						ConditionOperator: models.OperatorContains,
					},
				}
				rule1.Actions = []models.RuleActionResponse{
					{
						Id:          1,
						RuleId:      1,
						ActionType:  models.RuleFieldCategory,
						ActionValue: "1",
					},
				}

				rule2 := createTestRule(2, "Amount Only Rule", now.Add(-72*time.Hour), userId)
				rule2.Conditions = []models.RuleConditionResponse{
					{
						Id:                3,
						RuleId:            2,
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "100.00",
						ConditionOperator: models.OperatorGreater,
					},
				}
				rule2.Actions = []models.RuleActionResponse{
					{
						Id:          2,
						RuleId:      2,
						ActionType:  models.RuleFieldCategory,
						ActionValue: "2",
					},
				}

				result := engine.ExecuteRules([]models.DescribeRuleResponse{rule1, rule2}, testTxns)

				// Transaction 1 should match both rules
				Expect(result.Changesets).To(HaveLen(1))
				changeset := result.Changesets[0]
				Expect(changeset.TransactionId).To(Equal(int64(1)))
				Expect(changeset.CategoryAdds).To(ContainElement(int64(1)))
				Expect(changeset.CategoryAdds).To(ContainElement(int64(2)))
				Expect(changeset.AppliedRules).To(Equal([]int64{1, 2}))
			})
		})
	})
})

func stringPtr(s string) *string {
	return &s
}

func createTestRule(id int64, name string, effectiveFrom time.Time, userId int64) models.DescribeRuleResponse {
	return models.DescribeRuleResponse{
		Rule: models.RuleResponse{
			Id:            id,
			Name:          name,
			EffectiveFrom: effectiveFrom,
			CreatedBy:     userId,
		},
		Actions:    []models.RuleActionResponse{},
		Conditions: []models.RuleConditionResponse{},
	}
}

func findChangesetByTxnId(changesets []service.RuleChangeset, txnId int64) *service.RuleChangeset {
	for _, changeset := range changesets {
		if changeset.TransactionId == txnId {
			return &changeset
		}
	}
	return nil
}
