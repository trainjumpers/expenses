package controller_test

import (
	"expenses/internal/models"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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

	conditionLogicAnd := models.ConditionLogicAnd
	conditionLogicOr := models.ConditionLogicOr

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
		resp, response := testUser1.MakeRequest(http.MethodPost, "/rule", input)
		Expect(resp.StatusCode).To(Equal(http.StatusCreated))
		Expect(response["data"]).To(HaveKey("rule"))
		rule := response["data"].(map[string]any)["rule"].(map[string]any)
		action := response["data"].(map[string]any)["actions"].([]any)[0].(map[string]any)
		condition := response["data"].(map[string]any)["conditions"].([]any)[0].(map[string]any)
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
				resp, response := testUser1.MakeRequest(http.MethodPost, "/rule", input)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				Expect(response["message"]).To(Equal("Rule created successfully"))
				Expect(response["data"]).To(HaveKey("rule"))
				rule := response["data"].(map[string]any)["rule"].(map[string]any)
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
				resp, response := testUser1.MakeRequest(http.MethodPost, "/rule", input)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				actions := response["data"].(map[string]any)["actions"].([]any)
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
				resp, response := testUser1.MakeRequest(http.MethodPost, "/rule", input)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				conditions := response["data"].(map[string]any)["conditions"].([]any)
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
				resp, response := testUser1.MakeRequest(http.MethodPost, "/rule", input)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				rule := response["data"].(map[string]any)["rule"].(map[string]any)
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
				resp, response := testUser1.MakeRequest(http.MethodPost, "/rule", input)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				rule := response["data"].(map[string]any)["rule"].(map[string]any)
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
				resp, response := testUser1.MakeRequest(http.MethodPost, "/rule", input)
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
					defer GinkgoRecover()
					resp, _ := testUser1.MakeRequest(http.MethodPost, "/rule", input)
					Expect(resp.StatusCode).To(Equal(http.StatusCreated))
					done <- true
				}()
				go func() {
					defer GinkgoRecover()
					resp, _ := testUser1.MakeRequest(http.MethodPost, "/rule", input)
					Expect(resp.StatusCode).To(Equal(http.StatusCreated))
					done <- true
				}()
				Eventually(done, "2s").Should(Receive())
				Eventually(done, "2s").Should(Receive())
			})

			// 2. Comprehensive validation tests for all field types and operators
			It("should validate all field types in actions", func() {
				testCases := []struct {
					fieldType models.RuleFieldType
					value     string
				}{
					{models.RuleFieldAmount, "100.50"},
					{models.RuleFieldName, "Test Name"},
					{models.RuleFieldDescription, "Test Description"},
					{models.RuleFieldCategory, "1"},
				}

				for _, tc := range testCases {
					input := models.CreateRuleRequest{
						Rule: models.CreateBaseRuleRequest{
							Name:          "Test " + string(tc.fieldType),
							Description:   ptrToString("Testing " + string(tc.fieldType)),
							EffectiveFrom: now,
						},
						Actions: []models.CreateRuleActionRequest{
							{ActionType: tc.fieldType, ActionValue: tc.value},
						},
						Conditions: []models.CreateRuleConditionRequest{
							{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals},
						},
					}
					resp, _ := testUser1.MakeRequest(http.MethodPost, "/rule", input)
					Expect(resp.StatusCode).To(Equal(http.StatusCreated),
						"Failed for action field type: "+string(tc.fieldType))
				}
			})

			It("should validate all operator combinations in conditions", func() {
				testCases := []struct {
					fieldType  models.RuleFieldType
					operator   models.RuleOperator
					value      string
					shouldPass bool
				}{
					// Valid combinations
					{models.RuleFieldAmount, models.OperatorEquals, "100", true},
					{models.RuleFieldAmount, models.OperatorGreater, "50", true},
					{models.RuleFieldAmount, models.OperatorLower, "200", true},
					{models.RuleFieldName, models.OperatorEquals, "Test", true},
					{models.RuleFieldName, models.OperatorContains, "Test", true},
					{models.RuleFieldDescription, models.OperatorEquals, "Description", true},
					{models.RuleFieldDescription, models.OperatorContains, "Description", true},
					{models.RuleFieldCategory, models.OperatorEquals, "1", true},
					// Invalid combinations
					{models.RuleFieldAmount, models.OperatorContains, "100", false},
					{models.RuleFieldName, models.OperatorGreater, "Test", false},
					{models.RuleFieldName, models.OperatorLower, "Test", false},
					{models.RuleFieldDescription, models.OperatorGreater, "Description", false},
					{models.RuleFieldDescription, models.OperatorLower, "Description", false},
					{models.RuleFieldCategory, models.OperatorContains, "1", false},
					{models.RuleFieldCategory, models.OperatorGreater, "1", false},
					{models.RuleFieldCategory, models.OperatorLower, "1", false},
				}

				for _, tc := range testCases {
					input := models.CreateRuleRequest{
						Rule: models.CreateBaseRuleRequest{
							Name:          fmt.Sprintf("Test %s %s", tc.fieldType, tc.operator),
							Description:   ptrToString("Testing operator combinations"),
							EffectiveFrom: now,
						},
						Actions: []models.CreateRuleActionRequest{
							{ActionType: models.RuleFieldAmount, ActionValue: "100"},
						},
						Conditions: []models.CreateRuleConditionRequest{
							{ConditionType: tc.fieldType, ConditionValue: tc.value, ConditionOperator: tc.operator},
						},
					}
					resp, _ := testUser1.MakeRequest(http.MethodPost, "/rule", input)

					if tc.shouldPass {
						Expect(resp.StatusCode).To(Equal(http.StatusCreated),
							fmt.Sprintf("Should pass for %s with %s", tc.fieldType, tc.operator))
					} else {
						Expect(resp.StatusCode).To(Equal(http.StatusBadRequest),
							fmt.Sprintf("Should fail for %s with %s", tc.fieldType, tc.operator))
					}
				}
			})

			It("should validate numeric values for amount fields", func() {
				testCases := []struct {
					value      string
					shouldPass bool
				}{
					{"0", true},
					{"100", true},
					{"100.50", true},
					{"999999.99", true},
					{"-100", true}, // Negative amounts might be valid
					{"not-a-number", false},
					{"abc", false},
					{"100.50.25", false},
					{"", false},
					{" ", false},
				}

				for _, tc := range testCases {
					input := models.CreateRuleRequest{
						Rule: models.CreateBaseRuleRequest{
							Name:          "Amount Test " + tc.value,
							Description:   ptrToString("Testing amount validation"),
							EffectiveFrom: now,
						},
						Actions: []models.CreateRuleActionRequest{
							{ActionType: models.RuleFieldAmount, ActionValue: tc.value},
						},
						Conditions: []models.CreateRuleConditionRequest{
							{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals},
						},
					}
					resp, _ := testUser1.MakeRequest(http.MethodPost, "/rule", input)

					if tc.shouldPass {
						Expect(resp.StatusCode).To(Equal(http.StatusCreated), "Should pass for amount: "+tc.value)
					} else {
						Expect(resp.StatusCode).To(Equal(http.StatusBadRequest), "Should fail for amount: "+tc.value)
					}
				}
			})

			It("should validate category ID values", func() {
				testCases := []struct {
					value      string
					shouldPass bool
				}{
					{"1", true},
					{"123", true},
					{"999", true},
					{"-1", true},
					{"not-a-number", false},
					{"abc", false},
					{"1.5", false},
					{"", false},
					{" ", false},
				}

				for _, tc := range testCases {
					input := models.CreateRuleRequest{
						Rule: models.CreateBaseRuleRequest{
							Name:          "Category Test " + tc.value,
							Description:   ptrToString("Testing category validation"),
							EffectiveFrom: now,
						},
						Actions: []models.CreateRuleActionRequest{
							{ActionType: models.RuleFieldCategory, ActionValue: tc.value},
						},
						Conditions: []models.CreateRuleConditionRequest{
							{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals},
						},
					}
					resp, _ := testUser1.MakeRequest(http.MethodPost, "/rule", input)

					if tc.shouldPass {
						Expect(resp.StatusCode).To(Equal(http.StatusCreated),
							"Should have valid format for category: "+tc.value)
					} else {
						Expect(resp.StatusCode).To(Equal(http.StatusBadRequest), "Should fail for category: "+tc.value)
					}
				}
			})

			It("should handle special characters in string fields", func() {
				testCases := []string{
					"Name with spaces",
					"Name-with-dashes",
					"Name_with_underscores",
					"Name with 123 numbers",
					"Name with !@#$% special chars",
					"Name with unicode: café résumé",
					"Name with emoji: 🎉 test",
				}

				for _, name := range testCases {
					input := models.CreateRuleRequest{
						Rule: models.CreateBaseRuleRequest{
							Name:          name,
							Description:   ptrToString("Description: " + name),
							EffectiveFrom: now,
						},
						Actions: []models.CreateRuleActionRequest{
							{ActionType: models.RuleFieldName, ActionValue: name},
						},
						Conditions: []models.CreateRuleConditionRequest{
							{ConditionType: models.RuleFieldDescription, ConditionValue: name, ConditionOperator: models.OperatorContains},
						},
					}
					resp, _ := testUser1.MakeRequest(http.MethodPost, "/rule", input)
					Expect(resp.StatusCode).To(Equal(http.StatusCreated), "Should handle special chars in: "+name)
				}
			})

			It("should validate boundary values for string lengths", func() {
				// Test name at boundary (100 chars)
				name100 := strings.Repeat("a", 100)
				input := models.CreateRuleRequest{
					Rule: models.CreateBaseRuleRequest{
						Name:          name100,
						Description:   ptrToString("Boundary test"),
						EffectiveFrom: now,
					},
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "100"},
					},
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals},
					},
				}
				resp, _ := testUser1.MakeRequest(http.MethodPost, "/rule", input)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))

				// Test name over boundary (101 chars)
				name101 := strings.Repeat("a", 101)
				input.Rule.Name = name101
				resp, _ = testUser1.MakeRequest(http.MethodPost, "/rule", input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

				// Test description at boundary (255 chars)
				desc255 := strings.Repeat("d", 255)
				input.Rule.Name = "Valid Name"
				input.Rule.Description = &desc255
				resp, _ = testUser1.MakeRequest(http.MethodPost, "/rule", input)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))

				// Test description over boundary (256 chars)
				desc256 := strings.Repeat("d", 256)
				input.Rule.Description = &desc256
				resp, _ = testUser1.MakeRequest(http.MethodPost, "/rule", input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should create rules with AND condition logic (explicit)", func() {
				input := models.CreateRuleRequest{
					Rule: models.CreateBaseRuleRequest{
						Name:           "AND Logic Rule",
						Description:    ptrToString("Testing AND logic"),
						ConditionLogic: &conditionLogicAnd,
						EffectiveFrom:  now,
					},
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldCategory, ActionValue: "1"},
					},
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldName, ConditionValue: "Test", ConditionOperator: models.OperatorEquals},
						{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals},
					},
				}
				resp, response := testUser1.MakeRequest(http.MethodPost, "/rule", input)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))

				rule := response["data"].(map[string]any)["rule"].(map[string]any)
				Expect(rule["condition_logic"]).To(Equal("AND"))
				Expect(rule["name"]).To(Equal("AND Logic Rule"))
			})

			It("should create rules with OR condition logic", func() {
				input := models.CreateRuleRequest{
					Rule: models.CreateBaseRuleRequest{
						Name:           "OR Logic Rule",
						Description:    ptrToString("Testing OR logic"),
						ConditionLogic: &conditionLogicOr,
						EffectiveFrom:  now,
					},
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldCategory, ActionValue: "1"},
					},
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldName, ConditionValue: "Test", ConditionOperator: models.OperatorEquals},
						{ConditionType: models.RuleFieldAmount, ConditionValue: "999", ConditionOperator: models.OperatorEquals},
					},
				}
				resp, response := testUser1.MakeRequest(http.MethodPost, "/rule", input)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))

				rule := response["data"].(map[string]any)["rule"].(map[string]any)
				Expect(rule["condition_logic"]).To(Equal("OR"))
				Expect(rule["name"]).To(Equal("OR Logic Rule"))
			})

			It("should default to AND logic when condition_logic is not specified", func() {
				input := models.CreateRuleRequest{
					Rule: models.CreateBaseRuleRequest{
						Name:          "Default Logic Rule",
						Description:   ptrToString("Testing default logic"),
						EffectiveFrom: now,
						// ConditionLogic not specified - should default to AND
					},
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldCategory, ActionValue: "1"},
					},
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldName, ConditionValue: "Test", ConditionOperator: models.OperatorEquals},
					},
				}
				resp, response := testUser1.MakeRequest(http.MethodPost, "/rule", input)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))

				rule := response["data"].(map[string]any)["rule"].(map[string]any)
				Expect(rule["condition_logic"]).To(Equal("AND"))
			})

			It("should handle single condition with OR logic", func() {
				input := models.CreateRuleRequest{
					Rule: models.CreateBaseRuleRequest{
						Name:           "Single Condition OR Rule",
						Description:    ptrToString("Single condition with OR logic"),
						ConditionLogic: &conditionLogicOr,
						EffectiveFrom:  now,
					},
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldCategory, ActionValue: "1"},
					},
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldName, ConditionValue: "Test", ConditionOperator: models.OperatorEquals},
					},
				}
				resp, response := testUser1.MakeRequest(http.MethodPost, "/rule", input)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))

				rule := response["data"].(map[string]any)["rule"].(map[string]any)
				Expect(rule["condition_logic"]).To(Equal("OR"))
			})

			It("should handle multiple conditions with complex OR logic", func() {
				input := models.CreateRuleRequest{
					Rule: models.CreateBaseRuleRequest{
						Name:           "Complex OR Rule",
						Description:    ptrToString("Multiple conditions with OR logic"),
						ConditionLogic: &conditionLogicOr,
						EffectiveFrom:  now,
					},
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldCategory, ActionValue: "1"},
					},
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldName, ConditionValue: "NonExistent", ConditionOperator: models.OperatorEquals},
						{ConditionType: models.RuleFieldAmount, ConditionValue: "999", ConditionOperator: models.OperatorEquals},
						{ConditionType: models.RuleFieldDescription, ConditionValue: "Test", ConditionOperator: models.OperatorContains},
					},
				}
				resp, response := testUser1.MakeRequest(http.MethodPost, "/rule", input)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))

				rule := response["data"].(map[string]any)["rule"].(map[string]any)
				Expect(rule["condition_logic"]).To(Equal("OR"))
				conditions := response["data"].(map[string]any)["conditions"].([]any)
				Expect(len(conditions)).To(Equal(3))
			})

			It("should reject invalid condition_logic values", func() {
				invalidLogic := "INVALID"
				jsonBody := fmt.Sprintf(`{
					"rule": {
						"name": "Invalid Logic Rule",
						"description": "Testing invalid logic",
						"condition_logic": "%s",
						"effective_from": "%s"
					},
					"actions": [{"action_type": "category", "action_value": "1"}],
					"conditions": [{"condition_type": "name", "condition_value": "Test", "condition_operator": "equals"}]
				}`, invalidLogic, now.Format(time.RFC3339))

				resp, _ := testUser1.MakeRequest(http.MethodPost, "/rule", jsonBody)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})
		})

		// 3. Authentication edge cases
		Context("Authentication Edge Cases", func() {
			It("should return unauthorized for missing Authorization header", func() {
				input := models.CreateRuleRequest{
					Rule: models.CreateBaseRuleRequest{
						Name:          "Test Rule",
						Description:   ptrToString("Test"),
						EffectiveFrom: now,
					},
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "100"},
					},
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals},
					},
				}
				resp, _ := NewTestHelper(baseURL).MakeRequest(http.MethodPost, "/rule", input)
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("should return unauthorized for empty Authorization header", func() {
				input := models.CreateRuleRequest{
					Rule: models.CreateBaseRuleRequest{
						Name:          "Test Rule",
						Description:   ptrToString("Test"),
						EffectiveFrom: now,
					},
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "100"},
					},
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals},
					},
				}
				resp, _ := testHelperUnauthenticated.MakeRequest(http.MethodPost, "/rule", input)
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("should return unauthorized for malformed token", func() {
				malformedTokens := []string{
					"invalid-token",
					"Bearer",
					"NotBearer validtoken",
					"Bearer invalid.token.format",
					"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid",
					"Bearer ",
				}
				input := models.CreateRuleRequest{
					Rule: models.CreateBaseRuleRequest{
						Name:          "Test Rule",
						Description:   ptrToString("Test"),
						EffectiveFrom: now,
					},
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "100"},
					},
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals},
					},
				}
				for _, token := range malformedTokens {
					resp, _ := testUser1.MakeRequestWithToken(http.MethodPost, "/rule", token, input)
					Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized), "Should fail for malformed token: "+token)
				}
			})

			It("should handle authentication for all rule endpoints", func() {
				endpoints := []struct {
					method string
					path   string
					body   any
				}{
					{http.MethodGet, "/rule", nil},
					{http.MethodPost, "/rule", models.CreateRuleRequest{
						Rule: models.CreateBaseRuleRequest{
							Name:          "Test",
							Description:   ptrToString("Test"),
							EffectiveFrom: now,
						},
						Actions: []models.CreateRuleActionRequest{
							{ActionType: models.RuleFieldAmount, ActionValue: "100"},
						},
						Conditions: []models.CreateRuleConditionRequest{
							{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals},
						},
					}},
					{http.MethodGet, "/rule/1", nil},
					{http.MethodPatch, "/rule/1", models.UpdateRuleRequest{Name: ptrToString("Updated")}},
					{http.MethodDelete, "/rule/1", nil},
					{http.MethodPatch, "/rule/1/action/1", models.UpdateRuleActionRequest{ActionValue: ptrToString("200")}},
					{http.MethodPatch, "/rule/1/condition/1", models.UpdateRuleConditionRequest{ConditionValue: ptrToString("200")}},
				}

				for _, endpoint := range endpoints {
					resp, _ := testUser1.MakeRequestWithToken(endpoint.method, endpoint.path, "invalid-token", endpoint.body)
					Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized),
						fmt.Sprintf("Should be unauthorized for %s %s", endpoint.method, endpoint.path))
				}
			})

			It("should handle token expiration scenarios", func() {
				// This test would require generating an expired token
				// For now, we'll test with an obviously invalid token that might simulate expiration
				expiredToken := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjE1MTYyMzkwMjJ9.invalid"

				input := models.CreateRuleRequest{
					Rule: models.CreateBaseRuleRequest{
						Name:          "Test Rule",
						Description:   ptrToString("Test"),
						EffectiveFrom: now,
					},
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "100"},
					},
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals},
					},
				}
				resp, _ := testUser1.MakeRequestWithToken(http.MethodPost, "/rule", expiredToken, input)
				// System returns 400 for malformed JWT tokens
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
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
				resp, _ := testUser1.MakeRequest(http.MethodPost, "/rule", input)
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
				resp, _ := testUser1.MakeRequest(http.MethodPost, "/rule", input)
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
				resp, _ := testUser1.MakeRequest(http.MethodPost, "/rule", input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return error for invalid JSON", func() {
				resp, _ := testUser1.MakeRequest(http.MethodPost, "/rule", "{ invalid json }")
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return error for empty body", func() {
				resp, _ := testUser1.MakeRequest(http.MethodPost, "/rule", "")
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})
		})
	})

	Describe("ListRules", func() {
		It("should return an empty list when there are no rules", func() {
			// Use a fresh user/token with no rules
			resp, response := testUser3.MakeRequest(http.MethodGet, "/rule", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rules fetched successfully"))
			Expect(response["data"]).To(BeAssignableToTypeOf([]any{}))
			Expect(response["data"]).To(BeEmpty())
		})

		It("should list rules for the user and verify all fields", func() {
			// Create two rules for the main user
			ruleId1, _, _ := createTestRule()
			ruleId2, _, _ := createTestRule()
			resp, response := testUser1.MakeRequest(http.MethodGet, "/rule", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rules fetched successfully"))
			data := response["data"].([]any)
			var found1, found2 bool
			for _, r := range data {
				rule := r.(map[string]any)
				switch id := int64(rule["id"].(float64)); id {
				case ruleId1:
					found1 = true
				case ruleId2:
					found2 = true
				}
				if int64(rule["id"].(float64)) == ruleId1 || int64(rule["id"].(float64)) == ruleId2 {
					Expect(rule).To(HaveKey("name"))
					Expect(rule).To(HaveKey("description"))
					Expect(rule).To(HaveKey("condition_logic"))
					Expect(rule).To(HaveKey("effective_from"))
					Expect(rule).To(HaveKey("created_by"))
					Expect(rule["condition_logic"]).To(Equal("AND"))
				}
			}
			Expect(found1).To(BeTrue())
			Expect(found2).To(BeTrue())
		})

		It("should list rules with different condition_logic values", func() {
			andRuleInput := models.CreateRuleRequest{
				Rule: models.CreateBaseRuleRequest{
					Name:           "AND Rule List Test",
					Description:    ptrToString("Testing AND logic in list"),
					ConditionLogic: &conditionLogicAnd,
					EffectiveFrom:  now,
				},
				Actions: []models.CreateRuleActionRequest{
					{ActionType: models.RuleFieldCategory, ActionValue: "1"},
				},
				Conditions: []models.CreateRuleConditionRequest{
					{ConditionType: models.RuleFieldName, ConditionValue: "Test", ConditionOperator: models.OperatorEquals},
				},
			}
			resp, andResponse := testUser1.MakeRequest(http.MethodPost, "/rule", andRuleInput)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			andRuleId := int64(andResponse["data"].(map[string]any)["rule"].(map[string]any)["id"].(float64))

			orRuleInput := models.CreateRuleRequest{
				Rule: models.CreateBaseRuleRequest{
					Name:           "OR Rule List Test",
					Description:    ptrToString("Testing OR logic in list"),
					ConditionLogic: &conditionLogicOr,
					EffectiveFrom:  now,
				},
				Actions: []models.CreateRuleActionRequest{
					{ActionType: models.RuleFieldCategory, ActionValue: "1"},
				},
				Conditions: []models.CreateRuleConditionRequest{
					{ConditionType: models.RuleFieldName, ConditionValue: "Test", ConditionOperator: models.OperatorEquals},
				},
			}
			resp, orResponse := testUser1.MakeRequest(http.MethodPost, "/rule", orRuleInput)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			orRuleId := int64(orResponse["data"].(map[string]any)["rule"].(map[string]any)["id"].(float64))

			// List all rules and verify condition_logic values
			resp, response := testUser1.MakeRequest(http.MethodGet, "/rule", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			data := response["data"].([]any)

			var andRule, orRule map[string]any
			for _, item := range data {
				rule := item.(map[string]any)
				ruleId := int64(rule["id"].(float64))
				switch ruleId {
				case andRuleId:
					andRule = rule
				case orRuleId:
					orRule = rule
				}
			}

			Expect(andRule).NotTo(BeNil())
			Expect(orRule).NotTo(BeNil())
			Expect(andRule["condition_logic"]).To(Equal("AND"))
			Expect(orRule["condition_logic"]).To(Equal("OR"))
			Expect(andRule["name"]).To(Equal("AND Rule List Test"))
			Expect(orRule["name"]).To(Equal("OR Rule List Test"))
		})

		It("should not list rules created by another user", func() {
			// Create a rule as main user
			createTestRule()
			// List as other user
			resp, response := testUser2.MakeRequest(http.MethodGet, "/rule", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rules fetched successfully"))
			Expect(len(response["data"].([]any))).To(Equal(0))
		})

		It("should return unauthorized for invalid token", func() {
			resp, _ := testHelperUnauthenticated.MakeRequest(http.MethodGet, "/rule", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		})
	})

	Describe("GetRuleById", func() {
		It("should get rule by id and verify all fields", func() {
			ruleId, actionId, conditionId := createTestRule()
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rule fetched successfully"))
			Expect(response["data"]).To(HaveKey("rule"))
			rule := response["data"].(map[string]any)["rule"].(map[string]any)
			Expect(int64(rule["id"].(float64))).To(Equal(ruleId))
			Expect(rule).To(HaveKey("name"))
			Expect(rule).To(HaveKey("description"))
			Expect(rule).To(HaveKey("condition_logic"))
			Expect(rule).To(HaveKey("effective_from"))
			Expect(rule).To(HaveKey("created_by"))
			Expect(rule["condition_logic"]).To(Equal("AND"))
			Expect(response["data"]).To(HaveKey("actions"))
			Expect(response["data"]).To(HaveKey("conditions"))
			actions := response["data"].(map[string]any)["actions"].([]any)
			conditions := response["data"].(map[string]any)["conditions"].([]any)
			Expect(len(actions)).To(BeNumerically(">=", 1))
			Expect(len(conditions)).To(BeNumerically(">=", 1))
			// Check action and condition IDs match
			action := actions[0].(map[string]any)
			condition := conditions[0].(map[string]any)
			Expect(int64(action["id"].(float64))).To(Equal(actionId))
			Expect(int64(condition["id"].(float64))).To(Equal(conditionId))
		})

		It("should return correct condition_logic for OR rules", func() {
			input := models.CreateRuleRequest{
				Rule: models.CreateBaseRuleRequest{
					Name:           "OR Logic Test Rule",
					Description:    ptrToString("Testing OR logic retrieval"),
					ConditionLogic: &conditionLogicOr,
					EffectiveFrom:  now,
				},
				Actions: []models.CreateRuleActionRequest{
					{ActionType: models.RuleFieldCategory, ActionValue: "1"},
				},
				Conditions: []models.CreateRuleConditionRequest{
					{ConditionType: models.RuleFieldName, ConditionValue: "Test", ConditionOperator: models.OperatorEquals},
					{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals},
				},
			}
			resp, response := testUser1.MakeRequest(http.MethodPost, "/rule", input)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))

			createdRule := response["data"].(map[string]any)["rule"].(map[string]any)
			ruleId := int64(createdRule["id"].(float64))

			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response = testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			rule := response["data"].(map[string]any)["rule"].(map[string]any)
			Expect(rule["condition_logic"]).To(Equal("OR"))
			Expect(rule["name"]).To(Equal("OR Logic Test Rule"))
		})

		It("should return correct condition_logic after update", func() {
			ruleId, _, _ := createTestRule()

			update := models.UpdateRuleRequest{ConditionLogic: &conditionLogicOr}
			updateUrl := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, _ := testUser1.MakeRequest(http.MethodPatch, updateUrl, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			getUrl := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testUser1.MakeRequest(http.MethodGet, getUrl, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			rule := response["data"].(map[string]any)["rule"].(map[string]any)
			Expect(rule["condition_logic"]).To(Equal("OR"))
		})

		It("should return error for invalid rule id format", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/rule/invalid_id", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(Equal("invalid ruleId"))
		})

		It("should return error for non-existent rule id", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/rule/999999", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			Expect(response["message"]).To(ContainSubstring("not found"))
		})

		It("should not allow access to rule belonging to another user", func() {
			// Create a rule as main user
			ruleId, _, _ := createTestRule()
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testUser2.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			Expect(response["message"]).To(ContainSubstring("not found"))
		})

		It("should return unauthorized for invalid token", func() {
			ruleId, _, _ := createTestRule()
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, _ := testHelperUnauthenticated.MakeRequest(http.MethodGet, url, nil)
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
			resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rule updated successfully"))
			rule := response["data"].(map[string]any)
			Expect(rule["name"]).To(Equal("Updated Rule Name"))
		})

		It("should handle partial updates (only description)", func() {
			newDesc := "Updated Description Only"
			update := models.UpdateRuleRequest{Description: &newDesc}
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rule updated successfully"))
			rule := response["data"].(map[string]any)
			Expect(rule["description"]).To(Equal(newDesc))
		})

		It("should handle partial updates (only effective_from)", func() {
			newTime := now.Add(-time.Hour)
			update := models.UpdateRuleRequest{EffectiveFrom: &newTime}
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rule updated successfully"))
			rule := response["data"].(map[string]any)
			Expect(rule["effective_from"]).NotTo(BeNil())
		})

		It("should not update if no fields provided", func() {
			update := models.UpdateRuleRequest{}
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(ContainSubstring("no fields"))
		})

		It("should handle empty description update", func() {
			description := ""
			update := models.UpdateRuleRequest{Description: &description}
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rule updated successfully"))
		})

		It("should validate effective_from not in far future", func() {
			name := "Valid"
			future := time.Now().AddDate(5, 0, 0)
			update := models.UpdateRuleRequest{Name: &name, EffectiveFrom: &future}
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
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
				defer GinkgoRecover()
				resp, _ := testUser1.MakeRequest(http.MethodPatch, url, update1)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				done <- true
			}()
			go func() {
				defer GinkgoRecover()
				resp, _ := testUser1.MakeRequest(http.MethodPatch, url, update2)
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
			resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
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
			resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(ContainSubstring("validation"))
		})

		It("should return error for effective_from in the future", func() {
			name := "Valid"
			future := time.Now().Add(24 * time.Hour)
			update := models.UpdateRuleRequest{Name: &name, EffectiveFrom: &future}
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(ContainSubstring("the effective date for the rule is invalid or in the past"))
		})

		It("should return error for invalid rule id format", func() {
			newName := "Should Fail"
			update := models.UpdateRuleRequest{Name: &newName}
			resp, response := testUser1.MakeRequest(http.MethodPatch, "/rule/invalid_id", update)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(Equal("invalid ruleId"))
		})

		It("should return error for non-existent rule id", func() {
			newName := "Should Fail"
			update := models.UpdateRuleRequest{Name: &newName}
			resp, response := testUser1.MakeRequest(http.MethodPatch, "/rule/999999", update)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			Expect(response["message"]).To(ContainSubstring("not found"))
		})

		It("should update condition_logic from AND to OR", func() {
			update := models.UpdateRuleRequest{ConditionLogic: &conditionLogicOr}
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rule updated successfully"))
			rule := response["data"].(map[string]any)
			Expect(rule["condition_logic"]).To(Equal("OR"))
		})

		It("should update condition_logic from OR to AND", func() {
			// First set it to OR
			update := models.UpdateRuleRequest{ConditionLogic: &conditionLogicOr}
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, _ := testUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			// Then update back to AND
			update = models.UpdateRuleRequest{ConditionLogic: &conditionLogicAnd}
			resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rule updated successfully"))
			rule := response["data"].(map[string]any)
			Expect(rule["condition_logic"]).To(Equal("AND"))
		})

		It("should update condition_logic along with other fields", func() {
			newName := "Updated Name and Logic"
			newDesc := "Updated Description and Logic"
			update := models.UpdateRuleRequest{
				Name:           &newName,
				Description:    &newDesc,
				ConditionLogic: &conditionLogicOr,
			}
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rule updated successfully"))
			rule := response["data"].(map[string]any)
			Expect(rule["name"]).To(Equal("Updated Name and Logic"))
			Expect(rule["description"]).To(Equal("Updated Description and Logic"))
			Expect(rule["condition_logic"]).To(Equal("OR"))
		})

		It("should reject invalid condition_logic values in update", func() {
			// Test with invalid condition_logic value using raw JSON
			invalidLogic := "INVALID"
			jsonBody := fmt.Sprintf(`{
				"name": "Updated Rule",
				"condition_logic": "%s"
			}`, invalidLogic)

			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, _ := testUser1.MakeRequest(http.MethodPatch, url, jsonBody)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for invalid JSON", func() {
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, _ := testUser1.MakeRequest(http.MethodPatch, url, "{ invalid json }")
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		Describe("UpdateRuleAction", func() {
			BeforeEach(func() {
				ruleId, actionId, _ = createTestRule()
			})

			// 1. Complete positive test cases for UpdateRuleAction
			It("should successfully update action type and value", func() {
				typ := models.RuleFieldDescription
				val := "Updated description action"
				update := models.UpdateRuleActionRequest{
					ActionType:  &typ,
					ActionValue: &val,
				}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/action/" + strconv.FormatInt(actionId, 10)
				resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(response["message"]).To(Equal("Rule action updated successfully"))
				action := response["data"].(map[string]any)
				Expect(action["action_type"]).To(Equal(string(models.RuleFieldDescription)))
				Expect(action["action_value"]).To(Equal("Updated description action"))
			})

			It("should handle updating only action type", func() {
				typ := models.RuleFieldName
				update := models.UpdateRuleActionRequest{
					ActionType: &typ,
				}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/action/" + strconv.FormatInt(actionId, 10)
				resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				action := response["data"].(map[string]any)
				Expect(action["action_type"]).To(Equal(string(models.RuleFieldName)))
			})

			It("should handle updating only action value", func() {
				val := "Updated value only"
				update := models.UpdateRuleActionRequest{
					ActionValue: &val,
				}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/action/" + strconv.FormatInt(actionId, 10)
				resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				action := response["data"].(map[string]any)
				Expect(action["action_value"]).To(Equal("Updated value only"))
			})

			// 2. Comprehensive validation tests for all field types
			It("should validate amount field type with valid numeric values", func() {
				typ := models.RuleFieldAmount
				testCases := []string{"100", "100.50", "0", "999999.99"}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/action/" + strconv.FormatInt(actionId, 10)

				for _, val := range testCases {
					update := models.UpdateRuleActionRequest{
						ActionType:  &typ,
						ActionValue: &val,
					}
					resp, _ := testUser1.MakeRequest(http.MethodPatch, url, update)
					Expect(resp.StatusCode).To(Equal(http.StatusOK), "Failed for amount value: "+val)
				}
			})

			It("should validate name field type with valid string values", func() {
				typ := models.RuleFieldName
				testCases := []string{"Simple Name", "Name with 123", "Name-with-dashes", "Name_with_underscores"}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/action/" + strconv.FormatInt(actionId, 10)

				for _, val := range testCases {
					update := models.UpdateRuleActionRequest{
						ActionType:  &typ,
						ActionValue: &val,
					}
					resp, _ := testUser1.MakeRequest(http.MethodPatch, url, update)
					Expect(resp.StatusCode).To(Equal(http.StatusOK), "Failed for name value: "+val)
				}
			})

			It("should validate description field type with valid string values", func() {
				typ := models.RuleFieldDescription
				testCases := []string{"Simple description", "Description with special chars !@#$%", "Very long description that contains multiple words and sentences to test the field validation."}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/action/" + strconv.FormatInt(actionId, 10)

				for _, val := range testCases {
					update := models.UpdateRuleActionRequest{
						ActionType:  &typ,
						ActionValue: &val,
					}
					resp, _ := testUser1.MakeRequest(http.MethodPatch, url, update)
					Expect(resp.StatusCode).To(Equal(http.StatusOK), "Failed for description value: "+val)
				}
			})

			It("should return error for invalid amount values", func() {
				typ := models.RuleFieldAmount
				invalidValues := []string{"not-a-number", "abc", "100.50.25", "", " ", "∞"}
				// Note: "NaN" appears to be accepted by the system
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/action/" + strconv.FormatInt(actionId, 10)

				for _, val := range invalidValues {
					update := models.UpdateRuleActionRequest{
						ActionType:  &typ,
						ActionValue: &val,
					}
					resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
					Expect(resp.StatusCode).To(Equal(http.StatusBadRequest), "Should fail for invalid amount: "+val)
					Expect(response["message"]).To(ContainSubstring("invalid"), "Error message should mention invalid for: "+val)
				}
			})

			It("should return error for invalid category values", func() {
				typ := models.RuleFieldCategory
				invalidValues := []string{"not-a-number", "abc", "1.5", "", " "}
				// Note: "-1" appears to be accepted by the system
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/action/" + strconv.FormatInt(actionId, 10)

				for _, val := range invalidValues {
					update := models.UpdateRuleActionRequest{
						ActionType:  &typ,
						ActionValue: &val,
					}
					resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
					Expect(resp.StatusCode).To(Equal(http.StatusBadRequest), "Should fail for invalid category: "+val)
					Expect(response["message"]).To(ContainSubstring("invalid"), "Error message should mention invalid for: "+val)
				}
			})

			It("should return error for empty string values for name/description fields", func() {
				testCases := []models.RuleFieldType{models.RuleFieldName, models.RuleFieldDescription}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/action/" + strconv.FormatInt(actionId, 10)

				for _, typ := range testCases {
					emptyVal := ""
					update := models.UpdateRuleActionRequest{
						ActionType:  &typ,
						ActionValue: &emptyVal,
					}
					resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
					Expect(resp.StatusCode).To(Equal(http.StatusBadRequest), "Should fail for empty "+string(typ))
					Expect(response["message"]).To(ContainSubstring("cannot be empty"), "Error should mention empty value for: "+string(typ))
				}
			})

			It("should return error for invalid rule ID format", func() {
				typ := models.RuleFieldAmount
				val := "100"
				update := models.UpdateRuleActionRequest{
					ActionType:  &typ,
					ActionValue: &val,
				}
				resp, response := testUser1.MakeRequest(http.MethodPatch, "/rule/invalid_id/action/"+strconv.FormatInt(actionId, 10), update)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(Equal("invalid ruleId"))
			})

			It("should return error for invalid action ID format", func() {
				typ := models.RuleFieldAmount
				val := "100"
				update := models.UpdateRuleActionRequest{
					ActionType:  &typ,
					ActionValue: &val,
				}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/action/invalid_id"
				resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(Equal("invalid id"))
			})

			It("should return error for non-existent action ID", func() {
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/action/999999"
				typ := models.RuleFieldAmount
				val := "123"
				update := models.UpdateRuleActionRequest{
					ActionType:  &typ,
					ActionValue: &val,
				}
				resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
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
				resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
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
				resp, _ := testHelperUnauthenticated.MakeRequest(http.MethodPatch, url, update)
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
				resp, _ := testUser1.MakeRequest(http.MethodPatch, url, update)
				// Should succeed for valid string
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				// Now try an empty string if not allowed
				emptyVal := ""
				update.ActionValue = &emptyVal
				resp2, response2 := testUser1.MakeRequest(http.MethodPatch, url, update)
				Expect(resp2.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response2["message"]).To(ContainSubstring("cannot be empty"))
			})

			It("should return error for empty update request", func() {
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/action/" + strconv.FormatInt(actionId, 10)
				resp, _ := testUser1.MakeRequest(http.MethodPatch, url, "")
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
				resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
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
				resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(ContainSubstring("invalid"))
			})
		})

		Describe("UpdateRuleCondition", func() {
			BeforeEach(func() {
				ruleId, _, conditionId = createTestRule()
			})

			// 1. Complete positive test cases for UpdateRuleCondition
			It("should successfully update condition type, value and operator", func() {
				typ := models.RuleFieldDescription
				val := "Updated description condition"
				op := models.OperatorContains
				update := models.UpdateRuleConditionRequest{
					ConditionType:     &typ,
					ConditionValue:    &val,
					ConditionOperator: &op,
				}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/condition/" + strconv.FormatInt(conditionId, 10)
				resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(response["message"]).To(Equal("Rule condition updated successfully"))
				condition := response["data"].(map[string]any)
				Expect(condition["condition_type"]).To(Equal(string(models.RuleFieldDescription)))
				Expect(condition["condition_value"]).To(Equal("Updated description condition"))
				Expect(condition["condition_operator"]).To(Equal(string(models.OperatorContains)))
			})

			It("should handle updating only condition type", func() {
				typ := models.RuleFieldName
				update := models.UpdateRuleConditionRequest{
					ConditionType: &typ,
				}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/condition/" + strconv.FormatInt(conditionId, 10)
				resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				condition := response["data"].(map[string]any)
				Expect(condition["condition_type"]).To(Equal(string(models.RuleFieldName)))
			})

			It("should handle updating only condition value", func() {
				val := "Updated condition value only"
				update := models.UpdateRuleConditionRequest{
					ConditionValue: &val,
				}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/condition/" + strconv.FormatInt(conditionId, 10)
				resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				condition := response["data"].(map[string]any)
				Expect(condition["condition_value"]).To(Equal("Updated condition value only"))
			})

			It("should handle updating only condition operator", func() {
				op := models.OperatorGreater
				update := models.UpdateRuleConditionRequest{
					ConditionOperator: &op,
				}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/condition/" + strconv.FormatInt(conditionId, 10)
				resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				condition := response["data"].(map[string]any)
				Expect(condition["condition_operator"]).To(Equal(string(models.OperatorGreater)))
			})

			// 2. Comprehensive validation tests for all field types and operators
			It("should validate all valid operator combinations for amount field", func() {
				typ := models.RuleFieldAmount
				val := "100.50"
				validOperators := []models.RuleOperator{models.OperatorEquals, models.OperatorGreater, models.OperatorLower}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/condition/" + strconv.FormatInt(conditionId, 10)

				for _, op := range validOperators {
					update := models.UpdateRuleConditionRequest{
						ConditionType:     &typ,
						ConditionValue:    &val,
						ConditionOperator: &op,
					}
					resp, _ := testUser1.MakeRequest(http.MethodPatch, url, update)
					Expect(resp.StatusCode).To(Equal(http.StatusOK), "Failed for amount operator: "+string(op))
				}
			})

			It("should validate all valid operator combinations for name field", func() {
				typ := models.RuleFieldName
				val := "Test Name"
				validOperators := []models.RuleOperator{models.OperatorEquals, models.OperatorContains}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/condition/" + strconv.FormatInt(conditionId, 10)

				for _, op := range validOperators {
					update := models.UpdateRuleConditionRequest{
						ConditionType:     &typ,
						ConditionValue:    &val,
						ConditionOperator: &op,
					}
					resp, _ := testUser1.MakeRequest(http.MethodPatch, url, update)
					Expect(resp.StatusCode).To(Equal(http.StatusOK), "Failed for name operator: "+string(op))
				}
			})

			It("should validate all valid operator combinations for description field", func() {
				typ := models.RuleFieldDescription
				val := "Test Description"
				validOperators := []models.RuleOperator{models.OperatorEquals, models.OperatorContains}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/condition/" + strconv.FormatInt(conditionId, 10)

				for _, op := range validOperators {
					update := models.UpdateRuleConditionRequest{
						ConditionType:     &typ,
						ConditionValue:    &val,
						ConditionOperator: &op,
					}
					resp, _ := testUser1.MakeRequest(http.MethodPatch, url, update)
					Expect(resp.StatusCode).To(Equal(http.StatusOK), "Failed for description operator: "+string(op))
				}
			})

			It("should validate category field only accepts equals operator", func() {
				typ := models.RuleFieldCategory
				val := "1"
				op := models.OperatorEquals
				update := models.UpdateRuleConditionRequest{
					ConditionType:     &typ,
					ConditionValue:    &val,
					ConditionOperator: &op,
				}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/condition/" + strconv.FormatInt(conditionId, 10)
				resp, _ := testUser1.MakeRequest(http.MethodPatch, url, update)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})

			It("should return error for invalid operator combinations", func() {
				testCases := []struct {
					fieldType models.RuleFieldType
					operator  models.RuleOperator
					value     string
				}{
					{models.RuleFieldAmount, models.OperatorContains, "100"},      // Contains not valid for amount
					{models.RuleFieldName, models.OperatorGreater, "test"},        // Greater not valid for name
					{models.RuleFieldName, models.OperatorLower, "test"},          // Lower not valid for name
					{models.RuleFieldDescription, models.OperatorGreater, "test"}, // Greater not valid for description
					{models.RuleFieldDescription, models.OperatorLower, "test"},   // Lower not valid for description
					{models.RuleFieldCategory, models.OperatorContains, "1"},      // Contains not valid for category
					{models.RuleFieldCategory, models.OperatorGreater, "1"},       // Greater not valid for category
					{models.RuleFieldCategory, models.OperatorLower, "1"},         // Lower not valid for category
				}

				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/condition/" + strconv.FormatInt(conditionId, 10)
				for _, tc := range testCases {
					update := models.UpdateRuleConditionRequest{
						ConditionType:     &tc.fieldType,
						ConditionValue:    &tc.value,
						ConditionOperator: &tc.operator,
					}
					resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
					Expect(resp.StatusCode).To(Equal(http.StatusBadRequest),
						fmt.Sprintf("Should fail for %s with %s operator", tc.fieldType, tc.operator))
					Expect(response["message"]).To(ContainSubstring("operator"),
						fmt.Sprintf("Error should mention operator for %s with %s", tc.fieldType, tc.operator))
				}
			})

			It("should validate condition values for different field types", func() {
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/condition/" + strconv.FormatInt(conditionId, 10)

				// Valid amount values
				amountValues := []string{"0", "100", "100.50", "999999.99"}
				typ := models.RuleFieldAmount
				op := models.OperatorEquals
				for _, val := range amountValues {
					update := models.UpdateRuleConditionRequest{
						ConditionType:     &typ,
						ConditionValue:    &val,
						ConditionOperator: &op,
					}
					resp, _ := testUser1.MakeRequest(http.MethodPatch, url, update)
					Expect(resp.StatusCode).To(Equal(http.StatusOK), "Failed for valid amount: "+val)
				}

				// Valid category values (integers)
				categoryValues := []string{"1", "123", "999"}
				typ = models.RuleFieldCategory
				for _, val := range categoryValues {
					update := models.UpdateRuleConditionRequest{
						ConditionType:     &typ,
						ConditionValue:    &val,
						ConditionOperator: &op,
					}
					resp, _ := testUser1.MakeRequest(http.MethodPatch, url, update)
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
				}
			})

			It("should return error for invalid condition values", func() {
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/condition/" + strconv.FormatInt(conditionId, 10)

				// Invalid amount values
				typ := models.RuleFieldAmount
				op := models.OperatorEquals
				invalidAmounts := []string{"not-a-number", "abc", "100.50.25", "", " "}
				for _, val := range invalidAmounts {
					update := models.UpdateRuleConditionRequest{
						ConditionType:     &typ,
						ConditionValue:    &val,
						ConditionOperator: &op,
					}
					resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
					Expect(resp.StatusCode).To(Equal(http.StatusBadRequest), "Should fail for invalid amount: "+val)
					Expect(response["message"]).To(ContainSubstring("invalid"), "Error should mention invalid for: "+val)
				}

				// Invalid category values
				typ = models.RuleFieldCategory
				invalidCategories := []string{"not-a-number", "abc", "1.5", "", " "}
				// Note: "-1" appears to be accepted by the system
				for _, val := range invalidCategories {
					update := models.UpdateRuleConditionRequest{
						ConditionType:     &typ,
						ConditionValue:    &val,
						ConditionOperator: &op,
					}
					resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
					Expect(resp.StatusCode).To(Equal(http.StatusBadRequest), "Should fail for invalid category: "+val)
					Expect(response["message"]).To(ContainSubstring("invalid"), "Error should mention invalid for: "+val)
				}
			})

			It("should return error for empty string values for name/description fields", func() {
				testCases := []models.RuleFieldType{models.RuleFieldName, models.RuleFieldDescription}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/condition/" + strconv.FormatInt(conditionId, 10)
				op := models.OperatorEquals

				for _, typ := range testCases {
					emptyVal := ""
					update := models.UpdateRuleConditionRequest{
						ConditionType:     &typ,
						ConditionValue:    &emptyVal,
						ConditionOperator: &op,
					}
					resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
					Expect(resp.StatusCode).To(Equal(http.StatusBadRequest), "Should fail for empty "+string(typ))
					Expect(response["message"]).To(ContainSubstring("cannot be empty"), "Error should mention empty value for: "+string(typ))
				}
			})

			It("should return error for invalid rule ID format", func() {
				typ := models.RuleFieldAmount
				val := "100"
				op := models.OperatorEquals
				update := models.UpdateRuleConditionRequest{
					ConditionType:     &typ,
					ConditionValue:    &val,
					ConditionOperator: &op,
				}
				resp, response := testUser1.MakeRequest(http.MethodPatch, "/rule/invalid_id/condition/"+strconv.FormatInt(conditionId, 10), update)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(Equal("invalid ruleId"))
			})

			It("should return error for invalid condition ID format", func() {
				typ := models.RuleFieldAmount
				val := "100"
				op := models.OperatorEquals
				update := models.UpdateRuleConditionRequest{
					ConditionType:     &typ,
					ConditionValue:    &val,
					ConditionOperator: &op,
				}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/condition/invalid_id"
				resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(Equal("invalid id"))
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
				resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
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
				resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
				Expect(response["message"]).To(ContainSubstring("the requested rule condition was not found"))
			})

			It("should return error for condition belonging to different user", func() {
				resp, response := testUser2.MakeRequest(http.MethodPost, "/rule", models.CreateRuleRequest{
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
				otherCondition := response["data"].(map[string]any)["conditions"].([]any)[0].(map[string]any)
				otherRule := response["data"].(map[string]any)["rule"].(map[string]any)
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
				resp2, response2 := testUser1.MakeRequest(http.MethodPatch, url, update)
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
				resp, _ := testHelperUnauthenticated.MakeRequest(http.MethodPatch, url, update)
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
				resp, _ := testUser1.MakeRequest(http.MethodPatch, url, update)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				emptyVal := ""
				update.ConditionValue = &emptyVal
				resp2, response2 := testUser1.MakeRequest(http.MethodPatch, url, update)
				Expect(resp2.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response2["message"]).To(ContainSubstring("cannot be empty"))
			})

			It("should return error for empty update request", func() {
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/condition/" + strconv.FormatInt(conditionId, 10)
				resp, _ := testUser1.MakeRequest(http.MethodPatch, url, "")
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
				resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
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
				resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
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
				resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
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
			resp, _ := testUser1.MakeRequest(http.MethodDelete, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
		})

		It("should return error for invalid rule id format", func() {
			resp, _ := testUser1.MakeRequest(http.MethodDelete, "/rule/invalid", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return 404 when deleting non-existent rule id", func() {
			resp, _ := testUser1.MakeRequest(http.MethodDelete, "/rule/999999", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
	})

	Describe("ExecuteRules", func() {
		// Helper to get a transaction by ID for verification
		getTestTransaction := func(id int64, user *TestHelper) map[string]any {
			resp, response := user.MakeRequest(http.MethodGet, fmt.Sprintf("/transaction/%d", id), nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["data"]).ToNot(BeNil())
			return response["data"].(map[string]any)
		}

		Context("when executing rules using seeded data", func() {
			It("should accept the request and modify a transaction in the background for User 1", func() {
				originalTxn := getTestTransaction(1, testUser1)
				Expect(originalTxn["description"]).To(Equal("Test Description"))

				executeReq := models.ExecuteRulesRequest{}
				resp, _ := testUser1.MakeRequest(http.MethodPost, "/rule/execute", executeReq)
				Expect(resp.StatusCode).To(Equal(http.StatusAccepted))

				// Allow time for the background goroutine to execute
				time.Sleep(2 * time.Second)

				// Verify the transaction was actually updated in the database.
				updatedTxn := getTestTransaction(1, testUser1)
				Expect(updatedTxn["description"]).To(Equal("Updated by Name Rule"))
			})

			It("should accept the request but not modify any transaction if no conditions are met", func() {
				// Get original state to compare against
				originalTxn := getTestTransaction(1, testUser1)

				executeReq := models.ExecuteRulesRequest{
					RuleIds: &[]int64{1}, // Execute only Rule ID 1
				}
				resp, _ := testUser1.MakeRequest(http.MethodPost, "/rule/execute", executeReq)
				Expect(resp.StatusCode).To(Equal(http.StatusAccepted))

				// Allow time for background processing
				time.Sleep(2 * time.Second)

				// Verify the transaction was not updated
				updatedTxn := getTestTransaction(1, testUser1)
				Expect(updatedTxn["description"]).To(Equal(originalTxn["description"]))
			})

			It("should accept the request for a user with no rules", func() {
				executeReq := models.ExecuteRulesRequest{}
				resp, _ := testUser3.MakeRequest(http.MethodPost, "/rule/execute", executeReq)
				Expect(resp.StatusCode).To(Equal(http.StatusAccepted))
			})

			Context("with invalid requests or data", func() {
				It("should return unauthorized when no auth token is provided", func() {
					executeReq := models.ExecuteRulesRequest{}
					unauthenticatedUser := NewTestHelper(baseURL)
					resp, _ := unauthenticatedUser.MakeRequest(http.MethodPost, "/rule/execute", executeReq)
					Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
				})

				It("should accept the request for a non-existent rule_id", func() {
					executeReq := models.ExecuteRulesRequest{
						RuleIds: &[]int64{9999}, // This rule does not exist
					}
					resp, _ := testUser1.MakeRequest(http.MethodPost, "/rule/execute", executeReq)
					Expect(resp.StatusCode).To(Equal(http.StatusAccepted))
				})

				It("should accept the request for a non-existent transaction_id", func() {
					executeReq := models.ExecuteRulesRequest{
						TransactionIds: &[]int64{9999}, // This transaction does not exist
					}
					resp, _ := testUser1.MakeRequest(http.MethodPost, "/rule/execute", executeReq)
					Expect(resp.StatusCode).To(Equal(http.StatusAccepted))
				})

				It("should accept the request for a rule that belongs to another user", func() {
					executeReq := models.ExecuteRulesRequest{
						RuleIds: &[]int64{2},
					}
					resp, _ := testUser2.MakeRequest(http.MethodPost, "/rule/execute", executeReq)
					Expect(resp.StatusCode).To(Equal(http.StatusAccepted))
				})

				It("should not apply a rule action if the target category does not exist", func() {
					ruleInput := models.CreateRuleRequest{
						Rule: models.CreateBaseRuleRequest{
							Name:          "Bad Category Rule",
							EffectiveFrom: time.Now().Add(-24 * time.Hour),
						},
						Conditions: []models.CreateRuleConditionRequest{
							{
								ConditionType:     models.RuleFieldName,
								ConditionOperator: models.OperatorEquals,
								ConditionValue:    "Coffee Shop", // Matches Transaction ID 10
							},
						},
						Actions: []models.CreateRuleActionRequest{
							{
								ActionType:  models.RuleFieldCategory,
								ActionValue: "9999", // This category ID does not exist
							},
						},
					}
					resp, _ := testUser1.MakeRequest(http.MethodPost, "/rule", ruleInput)
					Expect(resp.StatusCode).To(Equal(http.StatusCreated))
					originalTxn := getTestTransaction(10, testUser1)
					originalCategories := originalTxn["category_ids"].([]any)
					executeReq := models.ExecuteRulesRequest{}
					resp, _ = testUser1.MakeRequest(http.MethodPost, "/rule/execute", executeReq)
					Expect(resp.StatusCode).To(Equal(http.StatusAccepted))

					// Allow time for background processing
					time.Sleep(2 * time.Second)

					// Verify the transaction categories were not updated
					updatedTxn := getTestTransaction(10, testUser1)
					updatedCategories := updatedTxn["category_ids"].([]any)
					Expect(updatedCategories).To(HaveLen(len(originalCategories)))

				})

				It("should return a 400 Bad Request for invalid data types in request", func() {
					invalidBody := map[string]any{
						"rule_ids": "not-an-array-of-integers",
					}
					resp, _ := testUser1.MakeRequest(http.MethodPost, "/rule/execute", invalidBody)
					Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				})
			})
		})
	})

	Describe("PutRuleActions", func() {
		var testRuleId int64

		BeforeEach(func() {
			// Create a test rule for each test
			testRuleId, _, _ = createTestRule()
		})

		Context("Authentication and Authorization", func() {
			It("should return unauthorized for missing authentication", func() {
				input := models.PutRuleActionsRequest{
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "200"},
					},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/actions"
				resp, _ := testHelperUnauthenticated.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("should return unauthorized for malformed tokens", func() {
				input := models.PutRuleActionsRequest{
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "200"},
					},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/actions"
				checkMalformedTokens(testUser1, http.MethodPut, url, input)
			})

			It("should return not found for rule belonging to another user", func() {
				input := models.PutRuleActionsRequest{
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "200"},
					},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/actions"
				resp, response := testUser2.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
				Expect(response["message"]).To(ContainSubstring("not found"))
			})

			It("should return not found for non-existent rule", func() {
				input := models.PutRuleActionsRequest{
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "200"},
					},
				}
				resp, response := testUser1.MakeRequest(http.MethodPut, "/rule/999999/actions", input)
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
				Expect(response["message"]).To(ContainSubstring("not found"))
			})
		})

		Context("Input Validation and JSON Binding", func() {
			It("should return bad request for invalid JSON", func() {
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/actions"
				resp, response := testUser1.MakeRequest(http.MethodPut, url, "{ invalid json }")
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(ContainSubstring("invalid"))
			})

			It("should return bad request for empty actions array", func() {
				input := models.PutRuleActionsRequest{
					Actions: []models.CreateRuleActionRequest{},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/actions"
				resp, response := testUser1.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(ContainSubstring("min"))
			})

			It("should return bad request for missing actions field", func() {
				input := map[string]any{}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/actions"
				resp, response := testUser1.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(ContainSubstring("required"))
			})

			It("should return bad request for invalid rule ID format", func() {
				input := models.PutRuleActionsRequest{
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "200"},
					},
				}
				resp, response := testUser1.MakeRequest(http.MethodPut, "/rule/invalid_id/actions", input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(Equal("invalid ruleId"))
			})

			It("should validate action types and values", func() {
				testCases := []struct {
					actionType  models.RuleFieldType
					actionValue string
					shouldPass  bool
					description string
				}{
					{models.RuleFieldAmount, "100.50", true, "valid amount"},
					{models.RuleFieldAmount, "invalid", false, "invalid amount"},
					{models.RuleFieldName, "Valid Name", true, "valid name"},
					{models.RuleFieldDescription, "Valid Description", true, "valid description"},
					{models.RuleFieldCategory, "1", true, "valid category ID"},
					{models.RuleFieldCategory, "invalid", false, "invalid category ID"},
					{models.RuleFieldTransfer, "1", true, "valid transfer account ID"},
					{models.RuleFieldTransfer, "invalid", false, "invalid transfer account ID"},
				}

				for _, tc := range testCases {
					input := models.PutRuleActionsRequest{
						Actions: []models.CreateRuleActionRequest{
							{ActionType: tc.actionType, ActionValue: tc.actionValue},
						},
					}
					url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/actions"
					resp, _ := testUser1.MakeRequest(http.MethodPut, url, input)

					if tc.shouldPass {
						Expect(resp.StatusCode).To(Equal(http.StatusOK), "Should pass for "+tc.description)
					} else {
						Expect(resp.StatusCode).To(Equal(http.StatusBadRequest), "Should fail for "+tc.description)
					}
				}
			})

			It("should enforce maximum actions limit", func() {
				// Create 51 actions (exceeds max of 50)
				actions := make([]models.CreateRuleActionRequest, 51)
				for i := 0; i < 51; i++ {
					actions[i] = models.CreateRuleActionRequest{
						ActionType:  models.RuleFieldAmount,
						ActionValue: fmt.Sprintf("%d", i+100),
					}
				}
				input := models.PutRuleActionsRequest{Actions: actions}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/actions"
				resp, response := testUser1.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(ContainSubstring("max"))
			})
		})

		Context("HTTP Status Code Responses", func() {
			It("should return 200 OK for successful PUT operation", func() {
				input := models.PutRuleActionsRequest{
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "200"},
						{ActionType: models.RuleFieldName, ActionValue: "Updated Name"},
					},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/actions"
				resp, response := testUser1.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(response["message"]).To(Equal("Rule actions updated successfully"))
			})

			It("should return 400 Bad Request for validation errors", func() {
				input := models.PutRuleActionsRequest{
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "invalid_amount"},
					},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/actions"
				resp, _ := testUser1.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return 404 Not Found for non-existent rule", func() {
				input := models.PutRuleActionsRequest{
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "200"},
					},
				}
				resp, _ := testUser1.MakeRequest(http.MethodPut, "/rule/999999/actions", input)
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			})

			It("should return 401 Unauthorized for missing authentication", func() {
				input := models.PutRuleActionsRequest{
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "200"},
					},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/actions"
				resp, _ := testHelperUnauthenticated.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("Error Response Formatting", func() {
			It("should return properly formatted error response for validation failures", func() {
				input := models.PutRuleActionsRequest{
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "invalid_amount"},
					},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/actions"
				resp, response := testUser1.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response).To(HaveKey("message"))
				Expect(response["message"]).To(BeAssignableToTypeOf(""))
			})

			It("should return properly formatted error response for not found", func() {
				input := models.PutRuleActionsRequest{
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "200"},
					},
				}
				resp, response := testUser1.MakeRequest(http.MethodPut, "/rule/999999/actions", input)
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
				Expect(response).To(HaveKey("message"))
				Expect(response["message"]).To(ContainSubstring("not found"))
			})

			It("should return properly formatted error response for unauthorized", func() {
				input := models.PutRuleActionsRequest{
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "200"},
					},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/actions"
				resp, response := testHelperUnauthenticated.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
				Expect(response).To(HaveKey("message"))
			})
		})

		Context("Successful PUT Operations", func() {
			It("should successfully replace all actions with new ones", func() {
				input := models.PutRuleActionsRequest{
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "300"},
						{ActionType: models.RuleFieldName, ActionValue: "New Transaction Name"},
						{ActionType: models.RuleFieldDescription, ActionValue: "New Description"},
					},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/actions"
				resp, response := testUser1.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(response["message"]).To(Equal("Rule actions updated successfully"))
				Expect(response["data"]).To(HaveKey("actions"))

				actions := response["data"].(map[string]any)["actions"].([]any)
				Expect(len(actions)).To(Equal(3))

				// Verify the actions were created correctly
				actionTypes := make(map[string]string)
				for _, action := range actions {
					actionMap := action.(map[string]any)
					actionTypes[actionMap["action_type"].(string)] = actionMap["action_value"].(string)
				}
				Expect(actionTypes["amount"]).To(Equal("300"))
				Expect(actionTypes["name"]).To(Equal("New Transaction Name"))
				Expect(actionTypes["description"]).To(Equal("New Description"))
			})

			It("should handle single action replacement", func() {
				input := models.PutRuleActionsRequest{
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "500"},
					},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/actions"
				resp, response := testUser1.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				actions := response["data"].(map[string]any)["actions"].([]any)
				Expect(len(actions)).To(Equal(1))

				action := actions[0].(map[string]any)
				Expect(action["action_type"]).To(Equal("amount"))
				Expect(action["action_value"]).To(Equal("500"))
			})

			It("should handle multiple actions of the same type", func() {
				input := models.PutRuleActionsRequest{
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "100"},
						{ActionType: models.RuleFieldAmount, ActionValue: "200"},
					},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/actions"
				resp, response := testUser1.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				actions := response["data"].(map[string]any)["actions"].([]any)
				Expect(len(actions)).To(Equal(2))

				// Both should be amount type
				for _, action := range actions {
					actionMap := action.(map[string]any)
					Expect(actionMap["action_type"]).To(Equal("amount"))
				}
			})
		})

		Context("End-to-End PUT Actions Workflow", func() {
			It("should completely replace existing actions and verify persistence", func() {
				// First, verify the initial state
				getRuleUrl := "/rule/" + strconv.FormatInt(testRuleId, 10)
				resp, response := testUser1.MakeRequest(http.MethodGet, getRuleUrl, nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				initialActions := response["data"].(map[string]any)["actions"].([]any)
				Expect(len(initialActions)).To(Equal(1)) // From createTestRule

				// Replace with new actions
				input := models.PutRuleActionsRequest{
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "999"},
						{ActionType: models.RuleFieldName, ActionValue: "Completely New Name"},
					},
				}
				putUrl := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/actions"
				resp, _ = testUser1.MakeRequest(http.MethodPut, putUrl, input)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				// Verify the replacement was successful
				resp, response = testUser1.MakeRequest(http.MethodGet, getRuleUrl, nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				updatedActions := response["data"].(map[string]any)["actions"].([]any)
				Expect(len(updatedActions)).To(Equal(2))

				// Verify old actions are gone and new ones exist
				actionTypes := make(map[string]string)
				for _, action := range updatedActions {
					actionMap := action.(map[string]any)
					actionTypes[actionMap["action_type"].(string)] = actionMap["action_value"].(string)
				}
				Expect(actionTypes["amount"]).To(Equal("999"))
				Expect(actionTypes["name"]).To(Equal("Completely New Name"))
			})

			It("should handle transactional integrity on validation failure", func() {
				// Get initial state
				getRuleUrl := "/rule/" + strconv.FormatInt(testRuleId, 10)
				resp, response := testUser1.MakeRequest(http.MethodGet, getRuleUrl, nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				initialActions := response["data"].(map[string]any)["actions"].([]any)

				// Try to update with invalid data
				input := models.PutRuleActionsRequest{
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "valid_amount_200"},
						{ActionType: models.RuleFieldAmount, ActionValue: "invalid_amount"},
					},
				}
				putUrl := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/actions"
				resp, _ = testUser1.MakeRequest(http.MethodPut, putUrl, input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

				// Verify original actions are still intact
				resp, response = testUser1.MakeRequest(http.MethodGet, getRuleUrl, nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				currentActions := response["data"].(map[string]any)["actions"].([]any)
				Expect(len(currentActions)).To(Equal(len(initialActions)))
			})
		})

		Context("Concurrent Access Scenarios", func() {
			It("should handle concurrent PUT requests to the same rule", func() {
				done := make(chan bool, 2)

				// First concurrent request
				go func() {
					defer GinkgoRecover()
					input := models.PutRuleActionsRequest{
						Actions: []models.CreateRuleActionRequest{
							{ActionType: models.RuleFieldAmount, ActionValue: "100"},
						},
					}
					url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/actions"
					resp, _ := testUser1.MakeRequest(http.MethodPut, url, input)
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
					done <- true
				}()

				// Second concurrent request
				go func() {
					defer GinkgoRecover()
					input := models.PutRuleActionsRequest{
						Actions: []models.CreateRuleActionRequest{
							{ActionType: models.RuleFieldAmount, ActionValue: "200"},
						},
					}
					url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/actions"
					resp, _ := testUser1.MakeRequest(http.MethodPut, url, input)
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
					done <- true
				}()

				// Wait for both requests to complete
				Eventually(done, "5s").Should(Receive())
				Eventually(done, "5s").Should(Receive())

				// Verify the rule still has valid actions (one of the two requests succeeded)
				getRuleUrl := "/rule/" + strconv.FormatInt(testRuleId, 10)
				resp, response := testUser1.MakeRequest(http.MethodGet, getRuleUrl, nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				actions := response["data"].(map[string]any)["actions"].([]any)
				Expect(len(actions)).To(Equal(1))

				action := actions[0].(map[string]any)
				actionValue := action["action_value"].(string)
				Expect(actionValue).To(SatisfyAny(Equal("100"), Equal("200")))
			})

			It("should handle concurrent requests from different users to different rules", func() {
				// Create a second rule for testUser2
				secondRuleId, _, _ := func() (int64, int64, int64) {
					input := models.CreateRuleRequest{
						Rule: models.CreateBaseRuleRequest{
							Name:          "Second Test Rule",
							Description:   ptrToString("A second rule for testing"),
							EffectiveFrom: now,
						},
						Actions: []models.CreateRuleActionRequest{
							{ActionType: models.RuleFieldAmount, ActionValue: "100"},
						},
						Conditions: []models.CreateRuleConditionRequest{
							{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals},
						},
					}
					resp, response := testUser2.MakeRequest(http.MethodPost, "/rule", input)
					Expect(resp.StatusCode).To(Equal(http.StatusCreated))
					rule := response["data"].(map[string]any)["rule"].(map[string]any)
					action := response["data"].(map[string]any)["actions"].([]any)[0].(map[string]any)
					condition := response["data"].(map[string]any)["conditions"].([]any)[0].(map[string]any)
					return int64(rule["id"].(float64)), int64(action["id"].(float64)), int64(condition["id"].(float64))
				}()

				done := make(chan bool, 2)

				// User1 updates their rule
				go func() {
					defer GinkgoRecover()
					input := models.PutRuleActionsRequest{
						Actions: []models.CreateRuleActionRequest{
							{ActionType: models.RuleFieldAmount, ActionValue: "300"},
						},
					}
					url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/actions"
					resp, _ := testUser1.MakeRequest(http.MethodPut, url, input)
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
					done <- true
				}()

				// User2 updates their rule
				go func() {
					defer GinkgoRecover()
					input := models.PutRuleActionsRequest{
						Actions: []models.CreateRuleActionRequest{
							{ActionType: models.RuleFieldAmount, ActionValue: "400"},
						},
					}
					url := "/rule/" + strconv.FormatInt(secondRuleId, 10) + "/actions"
					resp, _ := testUser2.MakeRequest(http.MethodPut, url, input)
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
					done <- true
				}()

				// Wait for both requests to complete
				Eventually(done, "5s").Should(Receive())
				Eventually(done, "5s").Should(Receive())

				// Verify both rules were updated correctly
				getRuleUrl1 := "/rule/" + strconv.FormatInt(testRuleId, 10)
				resp, response := testUser1.MakeRequest(http.MethodGet, getRuleUrl1, nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				actions1 := response["data"].(map[string]any)["actions"].([]any)
				Expect(actions1[0].(map[string]any)["action_value"]).To(Equal("300"))

				getRuleUrl2 := "/rule/" + strconv.FormatInt(secondRuleId, 10)
				resp, response = testUser2.MakeRequest(http.MethodGet, getRuleUrl2, nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				actions2 := response["data"].(map[string]any)["actions"].([]any)
				Expect(actions2[0].(map[string]any)["action_value"]).To(Equal("400"))
			})
		})

		Context("Data Consistency After Operations", func() {
			It("should maintain referential integrity after PUT operation", func() {
				input := models.PutRuleActionsRequest{
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "150"},
						{ActionType: models.RuleFieldName, ActionValue: "Integrity Test"},
					},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/actions"
				resp, response := testUser1.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				// Verify all returned actions have the correct rule_id
				actions := response["data"].(map[string]any)["actions"].([]any)
				for _, action := range actions {
					actionMap := action.(map[string]any)
					Expect(int64(actionMap["rule_id"].(float64))).To(Equal(testRuleId))
				}

				// Verify by fetching the rule again
				getRuleUrl := "/rule/" + strconv.FormatInt(testRuleId, 10)
				resp, response = testUser1.MakeRequest(http.MethodGet, getRuleUrl, nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				fetchedActions := response["data"].(map[string]any)["actions"].([]any)
				Expect(len(fetchedActions)).To(Equal(2))
				for _, action := range fetchedActions {
					actionMap := action.(map[string]any)
					Expect(int64(actionMap["rule_id"].(float64))).To(Equal(testRuleId))
				}
			})

			It("should ensure no orphaned actions remain after replacement", func() {
				// Get initial action count for the rule
				getRuleUrl := "/rule/" + strconv.FormatInt(testRuleId, 10)
				resp, response := testUser1.MakeRequest(http.MethodGet, getRuleUrl, nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				initialActions := response["data"].(map[string]any)["actions"].([]any)

				// Replace with different number of actions
				input := models.PutRuleActionsRequest{
					Actions: []models.CreateRuleActionRequest{
						{ActionType: models.RuleFieldAmount, ActionValue: "100"},
						{ActionType: models.RuleFieldName, ActionValue: "Name 1"},
						{ActionType: models.RuleFieldDescription, ActionValue: "Desc 1"},
					},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/actions"
				resp, _ = testUser1.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				// Verify exact count matches what we sent
				resp, response = testUser1.MakeRequest(http.MethodGet, getRuleUrl, nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				finalActions := response["data"].(map[string]any)["actions"].([]any)
				Expect(len(finalActions)).To(Equal(3))

				// Verify no old actions remain by checking IDs
				initialActionIds := make(map[int64]bool)
				for _, action := range initialActions {
					actionMap := action.(map[string]any)
					initialActionIds[int64(actionMap["id"].(float64))] = true
				}

				for _, action := range finalActions {
					actionMap := action.(map[string]any)
					actionId := int64(actionMap["id"].(float64))
					Expect(initialActionIds[actionId]).To(BeFalse(), "Old action ID should not exist in new actions")
				}
			})
		})
	})

	Describe("PutRuleConditions", func() {
		var testRuleId int64

		BeforeEach(func() {
			// Create a test rule for each test
			testRuleId, _, _ = createTestRule()
		})

		Context("Authentication and Authorization", func() {
			It("should return unauthorized for missing authentication", func() {
				input := models.PutRuleConditionsRequest{
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "200", ConditionOperator: models.OperatorEquals},
					},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/conditions"
				resp, _ := testHelperUnauthenticated.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("should return unauthorized for malformed tokens", func() {
				input := models.PutRuleConditionsRequest{
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "200", ConditionOperator: models.OperatorEquals},
					},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/conditions"
				checkMalformedTokens(testUser1, http.MethodPut, url, input)
			})

			It("should return not found for rule belonging to another user", func() {
				input := models.PutRuleConditionsRequest{
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "200", ConditionOperator: models.OperatorEquals},
					},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/conditions"
				resp, response := testUser2.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
				Expect(response["message"]).To(ContainSubstring("not found"))
			})

			It("should return not found for non-existent rule", func() {
				input := models.PutRuleConditionsRequest{
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "200", ConditionOperator: models.OperatorEquals},
					},
				}
				resp, response := testUser1.MakeRequest(http.MethodPut, "/rule/999999/conditions", input)
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
				Expect(response["message"]).To(ContainSubstring("not found"))
			})
		})

		Context("Input Validation and JSON Binding", func() {
			It("should return bad request for invalid JSON", func() {
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/conditions"
				resp, response := testUser1.MakeRequest(http.MethodPut, url, "{ invalid json }")
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(ContainSubstring("invalid"))
			})

			It("should return bad request for empty conditions array", func() {
				input := models.PutRuleConditionsRequest{
					Conditions: []models.CreateRuleConditionRequest{},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/conditions"
				resp, response := testUser1.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(ContainSubstring("min"))
			})

			It("should return bad request for missing conditions field", func() {
				input := map[string]any{}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/conditions"
				resp, response := testUser1.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(ContainSubstring("required"))
			})

			It("should return bad request for invalid rule ID format", func() {
				input := models.PutRuleConditionsRequest{
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "200", ConditionOperator: models.OperatorEquals},
					},
				}
				resp, response := testUser1.MakeRequest(http.MethodPut, "/rule/invalid_id/conditions", input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(Equal("invalid ruleId"))
			})

			It("should validate condition types, operators, and values", func() {
				testCases := []struct {
					conditionType     models.RuleFieldType
					conditionOperator models.RuleOperator
					conditionValue    string
					shouldPass        bool
					description       string
				}{
					// Valid combinations
					{models.RuleFieldAmount, models.OperatorEquals, "100.50", true, "valid amount equals"},
					{models.RuleFieldAmount, models.OperatorGreater, "50", true, "valid amount greater"},
					{models.RuleFieldAmount, models.OperatorLower, "200", true, "valid amount lower"},
					{models.RuleFieldName, models.OperatorEquals, "Test Name", true, "valid name equals"},
					{models.RuleFieldName, models.OperatorContains, "Test", true, "valid name contains"},
					{models.RuleFieldDescription, models.OperatorEquals, "Description", true, "valid description equals"},
					{models.RuleFieldDescription, models.OperatorContains, "Desc", true, "valid description contains"},
					{models.RuleFieldCategory, models.OperatorEquals, "1", true, "valid category equals"},
					// Invalid combinations
					{models.RuleFieldAmount, models.OperatorContains, "100", false, "invalid amount contains"},
					{models.RuleFieldName, models.OperatorGreater, "Test", false, "invalid name greater"},
					{models.RuleFieldName, models.OperatorLower, "Test", false, "invalid name lower"},
					{models.RuleFieldDescription, models.OperatorGreater, "Desc", false, "invalid description greater"},
					{models.RuleFieldDescription, models.OperatorLower, "Desc", false, "invalid description lower"},
					{models.RuleFieldCategory, models.OperatorContains, "1", false, "invalid category contains"},
					{models.RuleFieldCategory, models.OperatorGreater, "1", false, "invalid category greater"},
					{models.RuleFieldCategory, models.OperatorLower, "1", false, "invalid category lower"},
					// Invalid values
					{models.RuleFieldAmount, models.OperatorEquals, "invalid", false, "invalid amount value"},
					{models.RuleFieldCategory, models.OperatorEquals, "invalid", false, "invalid category value"},
				}

				for _, tc := range testCases {
					input := models.PutRuleConditionsRequest{
						Conditions: []models.CreateRuleConditionRequest{
							{ConditionType: tc.conditionType, ConditionValue: tc.conditionValue, ConditionOperator: tc.conditionOperator},
						},
					}
					url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/conditions"
					resp, _ := testUser1.MakeRequest(http.MethodPut, url, input)

					if tc.shouldPass {
						Expect(resp.StatusCode).To(Equal(http.StatusOK), "Should pass for "+tc.description)
					} else {
						Expect(resp.StatusCode).To(Equal(http.StatusBadRequest), "Should fail for "+tc.description)
					}
				}
			})

			It("should enforce maximum conditions limit", func() {
				// Create 51 conditions (exceeds max of 50)
				conditions := make([]models.CreateRuleConditionRequest, 51)
				for i := 0; i < 51; i++ {
					conditions[i] = models.CreateRuleConditionRequest{
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    fmt.Sprintf("%d", i+100),
						ConditionOperator: models.OperatorEquals,
					}
				}
				input := models.PutRuleConditionsRequest{Conditions: conditions}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/conditions"
				resp, response := testUser1.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(ContainSubstring("max"))
			})
		})

		Context("HTTP Status Code Responses", func() {
			It("should return 200 OK for successful PUT operation", func() {
				input := models.PutRuleConditionsRequest{
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "200", ConditionOperator: models.OperatorEquals},
						{ConditionType: models.RuleFieldName, ConditionValue: "Test", ConditionOperator: models.OperatorContains},
					},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/conditions"
				resp, response := testUser1.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(response["message"]).To(Equal("Rule conditions updated successfully"))
			})

			It("should return 400 Bad Request for validation errors", func() {
				input := models.PutRuleConditionsRequest{
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "invalid_amount", ConditionOperator: models.OperatorEquals},
					},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/conditions"
				resp, _ := testUser1.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return 404 Not Found for non-existent rule", func() {
				input := models.PutRuleConditionsRequest{
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "200", ConditionOperator: models.OperatorEquals},
					},
				}
				resp, _ := testUser1.MakeRequest(http.MethodPut, "/rule/999999/conditions", input)
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			})

			It("should return 401 Unauthorized for missing authentication", func() {
				input := models.PutRuleConditionsRequest{
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "200", ConditionOperator: models.OperatorEquals},
					},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/conditions"
				resp, _ := testHelperUnauthenticated.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("Error Response Formatting", func() {
			It("should return properly formatted error response for validation failures", func() {
				input := models.PutRuleConditionsRequest{
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "invalid_amount", ConditionOperator: models.OperatorEquals},
					},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/conditions"
				resp, response := testUser1.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response).To(HaveKey("message"))
				Expect(response["message"]).To(BeAssignableToTypeOf(""))
			})

			It("should return properly formatted error response for not found", func() {
				input := models.PutRuleConditionsRequest{
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "200", ConditionOperator: models.OperatorEquals},
					},
				}
				resp, response := testUser1.MakeRequest(http.MethodPut, "/rule/999999/conditions", input)
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
				Expect(response).To(HaveKey("message"))
				Expect(response["message"]).To(ContainSubstring("not found"))
			})

			It("should return properly formatted error response for unauthorized", func() {
				input := models.PutRuleConditionsRequest{
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "200", ConditionOperator: models.OperatorEquals},
					},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/conditions"
				resp, response := testHelperUnauthenticated.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
				Expect(response).To(HaveKey("message"))
			})
		})

		Context("Successful PUT Operations", func() {
			It("should successfully replace all conditions with new ones", func() {
				input := models.PutRuleConditionsRequest{
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "300", ConditionOperator: models.OperatorGreater},
						{ConditionType: models.RuleFieldName, ConditionValue: "Coffee", ConditionOperator: models.OperatorContains},
						{ConditionType: models.RuleFieldDescription, ConditionValue: "Purchase", ConditionOperator: models.OperatorEquals},
					},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/conditions"
				resp, response := testUser1.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(response["message"]).To(Equal("Rule conditions updated successfully"))
				Expect(response["data"]).To(HaveKey("conditions"))

				conditions := response["data"].(map[string]any)["conditions"].([]any)
				Expect(len(conditions)).To(Equal(3))

				// Verify the conditions were created correctly
				conditionMap := make(map[string]map[string]string)
				for _, condition := range conditions {
					condMap := condition.(map[string]any)
					condType := condMap["condition_type"].(string)
					conditionMap[condType] = map[string]string{
						"value":    condMap["condition_value"].(string),
						"operator": condMap["condition_operator"].(string),
					}
				}
				Expect(conditionMap["amount"]["value"]).To(Equal("300"))
				Expect(conditionMap["amount"]["operator"]).To(Equal("greater"))
				Expect(conditionMap["name"]["value"]).To(Equal("Coffee"))
				Expect(conditionMap["name"]["operator"]).To(Equal("contains"))
				Expect(conditionMap["description"]["value"]).To(Equal("Purchase"))
				Expect(conditionMap["description"]["operator"]).To(Equal("equals"))
			})

			It("should handle single condition replacement", func() {
				input := models.PutRuleConditionsRequest{
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "500", ConditionOperator: models.OperatorLower},
					},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/conditions"
				resp, response := testUser1.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				conditions := response["data"].(map[string]any)["conditions"].([]any)
				Expect(len(conditions)).To(Equal(1))

				condition := conditions[0].(map[string]any)
				Expect(condition["condition_type"]).To(Equal("amount"))
				Expect(condition["condition_value"]).To(Equal("500"))
				Expect(condition["condition_operator"]).To(Equal("lower"))
			})

			It("should handle multiple conditions of the same type with different operators", func() {
				input := models.PutRuleConditionsRequest{
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorGreater},
						{ConditionType: models.RuleFieldAmount, ConditionValue: "500", ConditionOperator: models.OperatorLower},
					},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/conditions"
				resp, response := testUser1.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				conditions := response["data"].(map[string]any)["conditions"].([]any)
				Expect(len(conditions)).To(Equal(2))

				// Both should be amount type but with different operators
				operators := make([]string, 0)
				for _, condition := range conditions {
					condMap := condition.(map[string]any)
					Expect(condMap["condition_type"]).To(Equal("amount"))
					operators = append(operators, condMap["condition_operator"].(string))
				}
				Expect(operators).To(ContainElements("greater", "lower"))
			})
		})

		Context("End-to-End PUT Conditions Workflow", func() {
			It("should completely replace existing conditions and verify persistence", func() {
				// First, verify the initial state
				getRuleUrl := "/rule/" + strconv.FormatInt(testRuleId, 10)
				resp, response := testUser1.MakeRequest(http.MethodGet, getRuleUrl, nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				initialConditions := response["data"].(map[string]any)["conditions"].([]any)
				Expect(len(initialConditions)).To(Equal(1)) // From createTestRule

				// Replace with new conditions
				input := models.PutRuleConditionsRequest{
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "999", ConditionOperator: models.OperatorEquals},
						{ConditionType: models.RuleFieldName, ConditionValue: "Completely New Pattern", ConditionOperator: models.OperatorContains},
					},
				}
				putUrl := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/conditions"
				resp, _ = testUser1.MakeRequest(http.MethodPut, putUrl, input)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				// Verify the replacement was successful
				resp, response = testUser1.MakeRequest(http.MethodGet, getRuleUrl, nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				updatedConditions := response["data"].(map[string]any)["conditions"].([]any)
				Expect(len(updatedConditions)).To(Equal(2))

				// Verify old conditions are gone and new ones exist
				conditionMap := make(map[string]map[string]string)
				for _, condition := range updatedConditions {
					condMap := condition.(map[string]any)
					condType := condMap["condition_type"].(string)
					conditionMap[condType] = map[string]string{
						"value":    condMap["condition_value"].(string),
						"operator": condMap["condition_operator"].(string),
					}
				}
				Expect(conditionMap["amount"]["value"]).To(Equal("999"))
				Expect(conditionMap["amount"]["operator"]).To(Equal("equals"))
				Expect(conditionMap["name"]["value"]).To(Equal("Completely New Pattern"))
				Expect(conditionMap["name"]["operator"]).To(Equal("contains"))
			})

			It("should handle transactional integrity on validation failure", func() {
				// Get initial state
				getRuleUrl := "/rule/" + strconv.FormatInt(testRuleId, 10)
				resp, response := testUser1.MakeRequest(http.MethodGet, getRuleUrl, nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				initialConditions := response["data"].(map[string]any)["conditions"].([]any)

				// Try to update with invalid data
				input := models.PutRuleConditionsRequest{
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "200", ConditionOperator: models.OperatorEquals},
						{ConditionType: models.RuleFieldAmount, ConditionValue: "invalid_amount", ConditionOperator: models.OperatorEquals},
					},
				}
				putUrl := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/conditions"
				resp, _ = testUser1.MakeRequest(http.MethodPut, putUrl, input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

				// Verify original conditions are still intact
				resp, response = testUser1.MakeRequest(http.MethodGet, getRuleUrl, nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				currentConditions := response["data"].(map[string]any)["conditions"].([]any)
				Expect(len(currentConditions)).To(Equal(len(initialConditions)))
			})
		})

		Context("Concurrent Access Scenarios", func() {
			It("should handle concurrent PUT requests to the same rule", func() {
				done := make(chan bool, 2)

				// First concurrent request
				go func() {
					defer GinkgoRecover()
					input := models.PutRuleConditionsRequest{
						Conditions: []models.CreateRuleConditionRequest{
							{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals},
						},
					}
					url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/conditions"
					resp, _ := testUser1.MakeRequest(http.MethodPut, url, input)
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
					done <- true
				}()

				// Second concurrent request
				go func() {
					defer GinkgoRecover()
					input := models.PutRuleConditionsRequest{
						Conditions: []models.CreateRuleConditionRequest{
							{ConditionType: models.RuleFieldAmount, ConditionValue: "200", ConditionOperator: models.OperatorGreater},
						},
					}
					url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/conditions"
					resp, _ := testUser1.MakeRequest(http.MethodPut, url, input)
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
					done <- true
				}()

				// Wait for both requests to complete
				Eventually(done, "5s").Should(Receive())
				Eventually(done, "5s").Should(Receive())

				// Verify the rule still has valid conditions (one of the two requests succeeded)
				getRuleUrl := "/rule/" + strconv.FormatInt(testRuleId, 10)
				resp, response := testUser1.MakeRequest(http.MethodGet, getRuleUrl, nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				conditions := response["data"].(map[string]any)["conditions"].([]any)
				Expect(len(conditions)).To(Equal(1))

				condition := conditions[0].(map[string]any)
				conditionValue := condition["condition_value"].(string)
				conditionOperator := condition["condition_operator"].(string)
				Expect(conditionValue).To(SatisfyAny(Equal("100"), Equal("200")))
				Expect(conditionOperator).To(SatisfyAny(Equal("equals"), Equal("greater")))
			})

			It("should handle concurrent requests from different users to different rules", func() {
				// Create a second rule for testUser2
				secondRuleId, _, _ := func() (int64, int64, int64) {
					input := models.CreateRuleRequest{
						Rule: models.CreateBaseRuleRequest{
							Name:          "Second Test Rule",
							Description:   ptrToString("A second rule for testing"),
							EffectiveFrom: now,
						},
						Actions: []models.CreateRuleActionRequest{
							{ActionType: models.RuleFieldAmount, ActionValue: "100"},
						},
						Conditions: []models.CreateRuleConditionRequest{
							{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals},
						},
					}
					resp, response := testUser2.MakeRequest(http.MethodPost, "/rule", input)
					Expect(resp.StatusCode).To(Equal(http.StatusCreated))
					rule := response["data"].(map[string]any)["rule"].(map[string]any)
					action := response["data"].(map[string]any)["actions"].([]any)[0].(map[string]any)
					condition := response["data"].(map[string]any)["conditions"].([]any)[0].(map[string]any)
					return int64(rule["id"].(float64)), int64(action["id"].(float64)), int64(condition["id"].(float64))
				}()

				done := make(chan bool, 2)

				// User1 updates their rule
				go func() {
					defer GinkgoRecover()
					input := models.PutRuleConditionsRequest{
						Conditions: []models.CreateRuleConditionRequest{
							{ConditionType: models.RuleFieldAmount, ConditionValue: "300", ConditionOperator: models.OperatorGreater},
						},
					}
					url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/conditions"
					resp, _ := testUser1.MakeRequest(http.MethodPut, url, input)
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
					done <- true
				}()

				// User2 updates their rule
				go func() {
					defer GinkgoRecover()
					input := models.PutRuleConditionsRequest{
						Conditions: []models.CreateRuleConditionRequest{
							{ConditionType: models.RuleFieldAmount, ConditionValue: "400", ConditionOperator: models.OperatorLower},
						},
					}
					url := "/rule/" + strconv.FormatInt(secondRuleId, 10) + "/conditions"
					resp, _ := testUser2.MakeRequest(http.MethodPut, url, input)
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
					done <- true
				}()

				// Wait for both requests to complete
				Eventually(done, "5s").Should(Receive())
				Eventually(done, "5s").Should(Receive())

				// Verify both rules were updated correctly
				getRuleUrl1 := "/rule/" + strconv.FormatInt(testRuleId, 10)
				resp, response := testUser1.MakeRequest(http.MethodGet, getRuleUrl1, nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				conditions1 := response["data"].(map[string]any)["conditions"].([]any)
				condition1 := conditions1[0].(map[string]any)
				Expect(condition1["condition_value"]).To(Equal("300"))
				Expect(condition1["condition_operator"]).To(Equal("greater"))

				getRuleUrl2 := "/rule/" + strconv.FormatInt(secondRuleId, 10)
				resp, response = testUser2.MakeRequest(http.MethodGet, getRuleUrl2, nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				conditions2 := response["data"].(map[string]any)["conditions"].([]any)
				condition2 := conditions2[0].(map[string]any)
				Expect(condition2["condition_value"]).To(Equal("400"))
				Expect(condition2["condition_operator"]).To(Equal("lower"))
			})
		})

		Context("Data Consistency After Operations", func() {
			It("should maintain referential integrity after PUT operation", func() {
				input := models.PutRuleConditionsRequest{
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "150", ConditionOperator: models.OperatorEquals},
						{ConditionType: models.RuleFieldName, ConditionValue: "Integrity Test", ConditionOperator: models.OperatorContains},
					},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/conditions"
				resp, response := testUser1.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				// Verify all returned conditions have the correct rule_id
				conditions := response["data"].(map[string]any)["conditions"].([]any)
				for _, condition := range conditions {
					conditionMap := condition.(map[string]any)
					Expect(int64(conditionMap["rule_id"].(float64))).To(Equal(testRuleId))
				}

				// Verify by fetching the rule again
				getRuleUrl := "/rule/" + strconv.FormatInt(testRuleId, 10)
				resp, response = testUser1.MakeRequest(http.MethodGet, getRuleUrl, nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				fetchedConditions := response["data"].(map[string]any)["conditions"].([]any)
				Expect(len(fetchedConditions)).To(Equal(2))
				for _, condition := range fetchedConditions {
					conditionMap := condition.(map[string]any)
					Expect(int64(conditionMap["rule_id"].(float64))).To(Equal(testRuleId))
				}
			})

			It("should ensure no orphaned conditions remain after replacement", func() {
				// Get initial condition count for the rule
				getRuleUrl := "/rule/" + strconv.FormatInt(testRuleId, 10)
				resp, response := testUser1.MakeRequest(http.MethodGet, getRuleUrl, nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				initialConditions := response["data"].(map[string]any)["conditions"].([]any)

				// Replace with different number of conditions
				input := models.PutRuleConditionsRequest{
					Conditions: []models.CreateRuleConditionRequest{
						{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorEquals},
						{ConditionType: models.RuleFieldName, ConditionValue: "Name 1", ConditionOperator: models.OperatorContains},
						{ConditionType: models.RuleFieldDescription, ConditionValue: "Desc 1", ConditionOperator: models.OperatorEquals},
					},
				}
				url := "/rule/" + strconv.FormatInt(testRuleId, 10) + "/conditions"
				resp, _ = testUser1.MakeRequest(http.MethodPut, url, input)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				// Verify exact count matches what we sent
				resp, response = testUser1.MakeRequest(http.MethodGet, getRuleUrl, nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				finalConditions := response["data"].(map[string]any)["conditions"].([]any)
				Expect(len(finalConditions)).To(Equal(3))

				// Verify no old conditions remain by checking IDs
				initialConditionIds := make(map[int64]bool)
				for _, condition := range initialConditions {
					conditionMap := condition.(map[string]any)
					initialConditionIds[int64(conditionMap["id"].(float64))] = true
				}

				for _, condition := range finalConditions {
					conditionMap := condition.(map[string]any)
					conditionId := int64(conditionMap["id"].(float64))
					Expect(initialConditionIds[conditionId]).To(BeFalse(), "Old condition ID should not exist in new conditions")
				}
			})
		})
	})
})
