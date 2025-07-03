package controller_test

import (
	"expenses/internal/models"
	"net/http"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RuleController", func() {
	var ruleId int64
	var actionId int64
	var conditionId int64

	ptrToString := func(s string) *string { return &s }
	now := time.Now()

	// Helper to create a new rule and return its ID
	createTestRule := func() (int64, int64, int64) {
		input := models.CreateRuleRequest{
			Rule: models.CreateBaseRuleRequest{
				Name:          "Test Rule",
				Description:   ptrToString("A rule for testing"),
				EffectiveFrom: now,
			},
			Actions: []models.CreateRuleActionRequest{
				{
					ActionType:  models.RuleFieldAmount,
					ActionValue: "100",
				},
			},
			Conditions: []models.CreateRuleConditionRequest{
				{
					ConditionType:     models.RuleFieldAmount,
					ConditionValue:    "100",
					ConditionOperator: models.OperatorEquals,
				},
			},
		}
		resp, response := testHelper.MakeRequest(http.MethodPost, "/rule", accessToken, input)
		Expect(resp.StatusCode).To(Equal(http.StatusCreated))
		Expect(response["data"]).To(HaveKey("rule"))
		rule := response["data"].(map[string]interface{})["rule"].(map[string]interface{})
		action := response["data"].(map[string]interface{})["actions"].([]interface{})[0].(map[string]interface{})
		condition := response["data"].(map[string]interface{})["conditions"].([]interface{})[0].(map[string]interface{})
		return int64(rule["id"].(float64)), int64(action["id"].(float64)), int64(condition["id"].(float64))
	}

	Describe("CreateRule", func() {
		Context("with valid input", func() {
			It("should create a rule successfully", func() {
				input := models.CreateRuleRequest{
					Rule: models.CreateBaseRuleRequest{
						Name:          "Test Rule",
						Description:   ptrToString("A rule for testing"),
						EffectiveFrom: now,
					},
					Actions: []models.CreateRuleActionRequest{
						{
							ActionType:  models.RuleFieldAmount,
							ActionValue: "100",
						},
					},
					Conditions: []models.CreateRuleConditionRequest{
						{
							ConditionType:     models.RuleFieldAmount,
							ConditionValue:    "100",
							ConditionOperator: models.OperatorEquals,
						},
					},
				}
				resp, response := testHelper.MakeRequest(http.MethodPost, "/rule", accessToken, input)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				Expect(response["message"]).To(Equal("Rule created successfully"))
				Expect(response["data"]).To(HaveKey("rule"))
				rule := response["data"].(map[string]interface{})["rule"].(map[string]interface{})
				Expect(rule["name"]).To(Equal("Test Rule"))
				ruleId = int64(rule["id"].(float64))
			})

			It("should handle multiple actions of same type", func() {
				input := models.CreateRuleRequest{
					Rule: models.CreateBaseRuleRequest{
						Name:          "Multi Action Rule",
						Description:   ptrToString("Multiple actions"),
						EffectiveFrom: now,
					},
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "100"},
						{ActionType: models.RuleFieldAmount, ActionValue: "200"},
					},
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals},
					},
				}
				resp, response := testHelper.MakeRequest(http.MethodPost, "/rule", accessToken, input)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				actions := response["data"].(map[string]interface{})["actions"].([]interface{})
				Expect(len(actions)).To(Equal(2))
			})

			It("should handle multiple conditions with different operators", func() {
				input := models.CreateRuleRequest{
					Rule: models.CreateBaseRuleRequest{
						Name:          "Multi Condition Rule",
						Description:   ptrToString("Multiple conditions"),
						EffectiveFrom: now,
					},
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "100"},
					},
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals},
						{ConditionType: models.RuleFieldAmount, ConditionValue: "50", ConditionOperator: models.OperatorGreater},
					},
				}
				resp, response := testHelper.MakeRequest(http.MethodPost, "/rule", accessToken, input)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				conditions := response["data"].(map[string]interface{})["conditions"].([]interface{})
				Expect(len(conditions)).To(Equal(2))
			})

			It("should handle very long but valid names (99 chars)", func() {
				longName := ""
				for i := 0; i < 99; i++ {
					longName += "a"
				}
				input := models.CreateRuleRequest{
					Rule: models.CreateBaseRuleRequest{
						Name:          longName,
						Description:   ptrToString("Long name"),
						EffectiveFrom: now,
					},
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "100"},
					},
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals},
					},
				}
				resp, response := testHelper.MakeRequest(http.MethodPost, "/rule", accessToken, input)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				rule := response["data"].(map[string]interface{})["rule"].(map[string]interface{})
				Expect(rule["name"]).To(Equal(longName))
			})

			It("should handle very long but valid descriptions (254 chars)", func() {
				longDesc := ""
				for i := 0; i < 254; i++ {
					longDesc += "d"
				}
				input := models.CreateRuleRequest{
					Rule: models.CreateBaseRuleRequest{
						Name:          "Long Desc Rule",
						Description:   ptrToString(longDesc),
						EffectiveFrom: now,
					},
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "100"},
					},
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals},
					},
				}
				resp, response := testHelper.MakeRequest(http.MethodPost, "/rule", accessToken, input)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				rule := response["data"].(map[string]interface{})["rule"].(map[string]interface{})
				Expect(rule["description"]).To(Equal(longDesc))
			})

			It("should validate action values for category type (must exist)", func() {
				input := models.CreateRuleRequest{
					Rule: models.CreateBaseRuleRequest{
						Name:          "Category Rule",
						Description:   ptrToString("Category action"),
						EffectiveFrom: now,
					},
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldCategory, ActionValue: "nonexistent-category"},
					},
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals},
					},
				}
				resp, response := testHelper.MakeRequest(http.MethodPost, "/rule", accessToken, input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(ContainSubstring("category"))
			})

			It("should handle concurrent rule creation", func() {
				input := models.CreateRuleRequest{
					Rule: models.CreateBaseRuleRequest{
						Name:          "Concurrent Rule",
						Description:   ptrToString("Concurrent"),
						EffectiveFrom: now,
					},
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "100"},
					},
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals},
					},
				}
				done := make(chan bool, 2)
				go func() {
					resp, _ := testHelper.MakeRequest(http.MethodPost, "/rule", accessToken, input)
					Expect(resp.StatusCode).To(Equal(http.StatusCreated))
					done <- true
				}()
				go func() {
					resp, _ := testHelper.MakeRequest(http.MethodPost, "/rule", accessToken, input)
					Expect(resp.StatusCode).To(Equal(http.StatusCreated))
					done <- true
				}()
				Eventually(done, "2s").Should(Receive())
				Eventually(done, "2s").Should(Receive())
			})
		})

		Context("with invalid input", func() {
			It("should return error for missing name", func() {
				input := models.CreateRuleRequest{
					Rule: models.CreateBaseRuleRequest{
						// Name is missing
						Description:   ptrToString("Missing name"),
						EffectiveFrom: now,
						CreatedBy:     1,
					},
					Actions: []models.CreateRuleActionRequest{
						{
							ActionType:  models.RuleFieldAmount,
							ActionValue: "100",
							RuleId:      0,
						},
					},
					Conditions: []models.CreateRuleConditionRequest{
						{
							ConditionType:     models.RuleFieldAmount,
							ConditionValue:    "100",
							ConditionOperator: models.OperatorEquals,
							RuleId:            0,
						},
					},
				}
				resp, _ := testHelper.MakeRequest(http.MethodPost, "/rule", accessToken, input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return error for missing actions", func() {
				input := models.CreateRuleRequest{
					Rule: models.CreateBaseRuleRequest{
						Name:          "No Actions",
						Description:   ptrToString("No actions"),
						EffectiveFrom: now,
						CreatedBy:     1,
					},
					Actions: []models.CreateRuleActionRequest{}, // empty
					Conditions: []models.CreateRuleConditionRequest{
						{
							ConditionType:     models.RuleFieldAmount,
							ConditionValue:    "100",
							ConditionOperator: models.OperatorEquals,
							RuleId:            0,
						},
					},
				}
				resp, _ := testHelper.MakeRequest(http.MethodPost, "/rule", accessToken, input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return error for missing conditions", func() {
				input := models.CreateRuleRequest{
					Rule: models.CreateBaseRuleRequest{
						Name:          "No Conditions",
						Description:   ptrToString("No conditions"),
						EffectiveFrom: now,
						CreatedBy:     1,
					},
					Actions: []models.CreateRuleActionRequest{
						{
							ActionType:  models.RuleFieldAmount,
							ActionValue: "100",
							RuleId:      0,
						},
					},
					Conditions: []models.CreateRuleConditionRequest{}, // empty
				}
				resp, _ := testHelper.MakeRequest(http.MethodPost, "/rule", accessToken, input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return error for invalid JSON", func() {
				resp, _ := testHelper.MakeRequest(http.MethodPost, "/rule", accessToken, "{ invalid json }")
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return error for empty body", func() {
				resp, _ := testHelper.MakeRequest(http.MethodPost, "/rule", accessToken, "")
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})
		})
	})

	Describe("ListRules", func() {
		It("should return an empty list when there are no rules", func() {
			// Use a fresh user/token with no rules
			resp, response := testHelper.MakeRequest(http.MethodGet, "/rule", accessToken1, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rules fetched successfully"))
			Expect(response["data"]).To(BeAssignableToTypeOf([]interface{}{}))
			Expect(response["data"]).To(BeEmpty())
		})

		It("should list rules for the user and verify all fields", func() {
			// Create two rules for the main user
			ruleId1, _, _ := createTestRule()
			ruleId2, _, _ := createTestRule()
			resp, response := testHelper.MakeRequest(http.MethodGet, "/rule", accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rules fetched successfully"))
			data := response["data"].([]interface{})
			var found1, found2 bool
			for _, r := range data {
				rule := r.(map[string]interface{})
				switch id := int64(rule["id"].(float64)); id {
				case ruleId1:
					found1 = true
				case ruleId2:
					found2 = true
				}
				if int64(rule["id"].(float64)) == ruleId1 || int64(rule["id"].(float64)) == ruleId2 {
					Expect(rule).To(HaveKey("name"))
					Expect(rule).To(HaveKey("description"))
					Expect(rule).To(HaveKey("effective_from"))
					Expect(rule).To(HaveKey("created_by"))
				}
			}
			Expect(found1).To(BeTrue())
			Expect(found2).To(BeTrue())
		})

		It("should not list rules created by another user", func() {
			// Create a rule as main user
			createTestRule()
			// List as other user
			resp, response := testHelper.MakeRequest(http.MethodGet, "/rule", accessToken1, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rules fetched successfully"))
			Expect(len(response["data"].([]interface{}))).To(Equal(0))
		})

		It("should return unauthorized for invalid token", func() {
			resp, _ := testHelper.MakeRequest(http.MethodGet, "/rule", "invalidtoken", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		})
	})

	Describe("GetRuleById", func() {
		It("should get rule by id and verify all fields", func() {
			ruleId, actionId, conditionId := createTestRule()
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testHelper.MakeRequest(http.MethodGet, url, accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rule fetched successfully"))
			Expect(response["data"]).To(HaveKey("rule"))
			rule := response["data"].(map[string]interface{})["rule"].(map[string]interface{})
			Expect(int64(rule["id"].(float64))).To(Equal(ruleId))
			Expect(rule).To(HaveKey("name"))
			Expect(rule).To(HaveKey("description"))
			Expect(rule).To(HaveKey("effective_from"))
			Expect(rule).To(HaveKey("created_by"))
			Expect(response["data"]).To(HaveKey("actions"))
			Expect(response["data"]).To(HaveKey("conditions"))
			actions := response["data"].(map[string]any)["actions"].([]interface{})
			conditions := response["data"].(map[string]any)["conditions"].([]interface{})
			Expect(len(actions)).To(BeNumerically(">=", 1))
			Expect(len(conditions)).To(BeNumerically(">=", 1))
			// Check action and condition IDs match
			action := actions[0].(map[string]interface{})
			condition := conditions[0].(map[string]interface{})
			Expect(int64(action["id"].(float64))).To(Equal(actionId))
			Expect(int64(condition["id"].(float64))).To(Equal(conditionId))
		})

		It("should return error for invalid rule id format", func() {
			resp, response := testHelper.MakeRequest(http.MethodGet, "/rule/invalid_id", accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(Equal("invalid ruleId"))
		})

		It("should return error for non-existent rule id", func() {
			resp, response := testHelper.MakeRequest(http.MethodGet, "/rule/999999", accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			Expect(response["message"]).To(ContainSubstring("not found"))
		})

		It("should not allow access to rule belonging to another user", func() {
			// Create a rule as main user
			ruleId, _, _ := createTestRule()
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testHelper.MakeRequest(http.MethodGet, url, accessToken1, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			Expect(response["message"]).To(ContainSubstring("not found"))
		})

		It("should return unauthorized for invalid token", func() {
			ruleId, _, _ := createTestRule()
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, _ := testHelper.MakeRequest(http.MethodGet, url, "invalidtoken", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		})
	})

	Describe("UpdateRule", func() {
		BeforeEach(func() {
			ruleId, actionId, conditionId = createTestRule()
		})

		It("should update rule name", func() {
			newName := "Updated Rule Name"
			update := models.UpdateRuleRequest{Name: &newName}
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rule updated successfully"))
			rule := response["data"].(map[string]interface{})
			Expect(rule["name"]).To(Equal("Updated Rule Name"))
		})

		It("should handle partial updates (only description)", func() {
			newDesc := "Updated Description Only"
			update := models.UpdateRuleRequest{Description: &newDesc}
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rule updated successfully"))
			rule := response["data"].(map[string]interface{})
			Expect(rule["description"]).To(Equal(newDesc))
		})

		It("should handle partial updates (only effective_from)", func() {
			newTime := now.Add(-time.Hour)
			update := models.UpdateRuleRequest{EffectiveFrom: &newTime}
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rule updated successfully"))
			rule := response["data"].(map[string]interface{})
			Expect(rule["effective_from"]).NotTo(BeNil())
		})

		It("should not update if no fields provided", func() {
			update := models.UpdateRuleRequest{}
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(ContainSubstring("no fields"))
		})

		It("should handle empty description update", func() {
			description := ""
			update := models.UpdateRuleRequest{Description: &description}
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rule updated successfully"))
		})

		It("should validate effective_from not in far future", func() {
			name := "Valid"
			future := time.Now().AddDate(5, 0, 0)
			update := models.UpdateRuleRequest{Name: &name, EffectiveFrom: &future}
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(ContainSubstring("effective date"))
		})

		It("should handle concurrent rule updates", func() {
			newName1 := "Concurrent Name 1"
			newName2 := "Concurrent Name 2"
			update1 := models.UpdateRuleRequest{Name: &newName1}
			update2 := models.UpdateRuleRequest{Name: &newName2}
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			done := make(chan bool, 2)
			go func() {
				resp, _ := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update1)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				done <- true
			}()
			go func() {
				resp, _ := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update2)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				done <- true
			}()
			Eventually(done, "2s").Should(Receive())
			Eventually(done, "2s").Should(Receive())
		})

		It("should return error for name longer than 100 chars", func() {
			longName := ""
			for range 101 {
				longName += "a"
			}
			update := models.UpdateRuleRequest{Name: &longName}
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(ContainSubstring("validation"))
		})

		It("should return error for description longer than 255 chars", func() {
			longDesc := ""
			for range 256 {
				longDesc += "a"
			}
			update := models.UpdateRuleRequest{Description: &longDesc}
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(ContainSubstring("validation"))
		})

		It("should return error for effective_from in the future", func() {
			name := "Valid"
			future := time.Now().Add(24 * time.Hour)
			update := models.UpdateRuleRequest{Name: &name, EffectiveFrom: &future}
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(ContainSubstring("the effective date for the rule is invalid or in the past"))
		})

		It("should return error for invalid rule id format", func() {
			newName := "Should Fail"
			update := models.UpdateRuleRequest{Name: &newName}
			resp, response := testHelper.MakeRequest(http.MethodPatch, "/rule/invalid_id", accessToken, update)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(Equal("invalid ruleId"))
		})

		It("should return error for non-existent rule id", func() {
			newName := "Should Fail"
			update := models.UpdateRuleRequest{Name: &newName}
			resp, response := testHelper.MakeRequest(http.MethodPatch, "/rule/999999", accessToken, update)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			Expect(response["message"]).To(ContainSubstring("not found"))
		})

		It("should return error for invalid JSON", func() {
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, _ := testHelper.MakeRequest(http.MethodPatch, url, accessToken, "{ invalid }")
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		Describe("UpdateRuleAction", func() {
			BeforeEach(func() {
				ruleId, actionId, _ = createTestRule()
			})

			It("should return error for non-existent action ID", func() {
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/action/999999"
				typ := models.RuleFieldAmount
				val := "123"
				update := models.UpdateRuleActionRequest{
					ActionType:  &typ,
					ActionValue: &val,
				}
				resp, response := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
				Expect(response["message"]).To(ContainSubstring("not found"))
			})

			It("should return error for action belonging to different rule", func() {
				// Create a second rule, get its actionId
				_, otherActionId, _ := createTestRule()
				// Try to update otherActionId under the first ruleId
				typ := models.RuleFieldAmount
				val := "123"
				update := models.UpdateRuleActionRequest{
					ActionType:  &typ,
					ActionValue: &val,
				}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/action/" + strconv.FormatInt(otherActionId, 10)
				resp, response := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
				Expect(response["message"]).To(ContainSubstring("the requested rule action was not found"))
			})

			It("should return unauthorized for invalid token", func() {
				typ := models.RuleFieldAmount
				val := "123"
				update := models.UpdateRuleActionRequest{
					ActionType:  &typ,
					ActionValue: &val,
				}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/action/" + strconv.FormatInt(actionId, 10)
				resp, _ := testHelper.MakeRequest(http.MethodPatch, url, "invalidtoken", update)
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("should validate action value format for string type", func() {
				typ := models.RuleFieldDescription
				val := "A valid description"
				update := models.UpdateRuleActionRequest{
					ActionType:  &typ,
					ActionValue: &val,
				}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/action/" + strconv.FormatInt(actionId, 10)
				resp, _ := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
				// Should succeed for valid string
				Expect(resp.StatusCode).To(Or(Equal(http.StatusOK), Equal(http.StatusBadRequest)))
				// Now try an empty string if not allowed
				emptyVal := ""
				update.ActionValue = &emptyVal
				resp2, response2 := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
				Expect(resp2.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response2["message"]).To(ContainSubstring("cannot be empty"))
			})

			It("should return error for empty update request", func() {
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/action/" + strconv.FormatInt(actionId, 10)
				resp, _ := testHelper.MakeRequest(http.MethodPatch, url, accessToken, "")
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return error for invalid action type", func() {
				invalidType := models.RuleFieldType("invalid")
				val := "foo"
				update := models.UpdateRuleActionRequest{
					ActionType:  &invalidType,
					ActionValue: &val,
				}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/action/" + strconv.FormatInt(actionId, 10)
				resp, response := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(ContainSubstring("invalid"))
			})

			It("should return error for invalid action value for type", func() {
				typ := models.RuleFieldAmount
				val := "not-a-number"
				update := models.UpdateRuleActionRequest{
					ActionType:  &typ,
					ActionValue: &val,
				}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/action/" + strconv.FormatInt(actionId, 10)
				resp, response := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(ContainSubstring("invalid"))
			})
		})

		Describe("UpdateRuleCondition", func() {
			BeforeEach(func() {
				ruleId, _, conditionId = createTestRule()
			})

			It("should return error for non-existent condition ID", func() {
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/condition/999999"
				typ := models.RuleFieldAmount
				val := "123"
				op := models.OperatorEquals
				update := models.UpdateRuleConditionRequest{
					ConditionType:     &typ,
					ConditionValue:    &val,
					ConditionOperator: &op,
				}
				resp, response := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
				Expect(response["message"]).To(ContainSubstring("not found"))
			})

			It("should return error for condition belonging to different rule", func() {
				_, _, otherConditionId := createTestRule()
				typ := models.RuleFieldAmount
				val := "123"
				op := models.OperatorEquals
				update := models.UpdateRuleConditionRequest{
					ConditionType:     &typ,
					ConditionValue:    &val,
					ConditionOperator: &op,
				}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/condition/" + strconv.FormatInt(otherConditionId, 10)
				resp, response := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
				Expect(response["message"]).To(ContainSubstring("the requested rule condition was not found"))
			})

			It("should return error for condition belonging to different user", func() {
				resp, response := testHelper.MakeRequest(http.MethodPost, "/rule", accessToken1, models.CreateRuleRequest{
					Rule: models.CreateBaseRuleRequest{
						Name:          "Other User Rule",
						Description:   ptrToString("desc"),
						EffectiveFrom: now,
					},
					Actions: []models.CreateRuleActionRequest{
						{
							ActionType:  models.RuleFieldAmount,
							ActionValue: "100",
						},
					},
					Conditions: []models.CreateRuleConditionRequest{
						{
							ConditionType:     models.RuleFieldAmount,
							ConditionValue:    "100",
							ConditionOperator: models.OperatorEquals,
						},
					},
				})
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				otherCondition := response["data"].(map[string]interface{})["conditions"].([]interface{})[0].(map[string]interface{})
				otherRule := response["data"].(map[string]interface{})["rule"].(map[string]interface{})
				otherRuleId := int64(otherRule["id"].(float64))
				otherConditionId := int64(otherCondition["id"].(float64))

				typ := models.RuleFieldAmount
				val := "123"
				op := models.OperatorEquals
				update := models.UpdateRuleConditionRequest{
					ConditionType:     &typ,
					ConditionValue:    &val,
					ConditionOperator: &op,
				}
				url := "/rule/" + strconv.FormatInt(otherRuleId, 10) + "/condition/" + strconv.FormatInt(otherConditionId, 10)
				resp2, response2 := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
				Expect(resp2.StatusCode).To(Equal(http.StatusNotFound))
				Expect(response2["message"]).To(ContainSubstring("not found"))
			})

			It("should return unauthorized for invalid token", func() {
				typ := models.RuleFieldAmount
				val := "123"
				op := models.OperatorEquals
				update := models.UpdateRuleConditionRequest{
					ConditionType:     &typ,
					ConditionValue:    &val,
					ConditionOperator: &op,
				}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/condition/" + strconv.FormatInt(conditionId, 10)
				resp, _ := testHelper.MakeRequest(http.MethodPatch, url, "invalidtoken", update)
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("should validate condition value format for string type", func() {
				typ := models.RuleFieldDescription
				val := "A valid description"
				op := models.OperatorEquals
				update := models.UpdateRuleConditionRequest{
					ConditionType:     &typ,
					ConditionValue:    &val,
					ConditionOperator: &op,
				}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/condition/" + strconv.FormatInt(conditionId, 10)
				resp, _ := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
				Expect(resp.StatusCode).To(Or(Equal(http.StatusOK), Equal(http.StatusBadRequest)))
				emptyVal := ""
				update.ConditionValue = &emptyVal
				resp2, response2 := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
				Expect(resp2.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response2["message"]).To(ContainSubstring("cannot be empty"))
			})

			It("should return error for empty update request", func() {
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/condition/" + strconv.FormatInt(conditionId, 10)
				resp, _ := testHelper.MakeRequest(http.MethodPatch, url, accessToken, "")
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return error for invalid condition type", func() {
				invalidType := models.RuleFieldType("invalid")
				val := "foo"
				op := models.OperatorEquals
				update := models.UpdateRuleConditionRequest{
					ConditionType:     &invalidType,
					ConditionValue:    &val,
					ConditionOperator: &op,
				}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/condition/" + strconv.FormatInt(conditionId, 10)
				resp, response := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(ContainSubstring("invalid"))
			})

			It("should return error for invalid condition value for type", func() {
				typ := models.RuleFieldAmount
				val := "not-a-number"
				op := models.OperatorEquals
				update := models.UpdateRuleConditionRequest{
					ConditionType:     &typ,
					ConditionValue:    &val,
					ConditionOperator: &op,
				}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/condition/" + strconv.FormatInt(conditionId, 10)
				resp, response := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(ContainSubstring("invalid"))
			})

			It("should return error for invalid operator for type", func() {
				typ := models.RuleFieldAmount
				val := "123.45"
				op := models.OperatorContains // not valid for amount
				update := models.UpdateRuleConditionRequest{
					ConditionType:     &typ,
					ConditionValue:    &val,
					ConditionOperator: &op,
				}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/condition/" + strconv.FormatInt(conditionId, 10)
				resp, response := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(ContainSubstring("the operator is not valid for the given condition type"))
			})
		})
	})

	Describe("DeleteRule", func() {
		BeforeEach(func() {
			ruleId, actionId, conditionId = createTestRule()
		})

		It("should delete rule by id", func() {
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, _ := testHelper.MakeRequest(http.MethodDelete, url, accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
		})

		It("should return error for invalid rule id format", func() {
			resp, _ := testHelper.MakeRequest(http.MethodDelete, "/rule/invalid", accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return 404 when deleting non-existent rule id", func() {
			resp, _ := testHelper.MakeRequest(http.MethodDelete, "/rule/999999", accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
	})
})
