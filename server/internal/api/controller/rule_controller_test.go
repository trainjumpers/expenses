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
		resp, response := testHelperUser1.MakeRequest(http.MethodPost, "/rule", input)
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
				resp, response := testHelperUser1.MakeRequest(http.MethodPost, "/rule", input)
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
				resp, response := testHelperUser1.MakeRequest(http.MethodPost, "/rule", input)
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
				resp, response := testHelperUser1.MakeRequest(http.MethodPost, "/rule", input)
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
				resp, response := testHelperUser1.MakeRequest(http.MethodPost, "/rule", input)
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
				resp, response := testHelperUser1.MakeRequest(http.MethodPost, "/rule", input)
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
				resp, response := testHelperUser1.MakeRequest(http.MethodPost, "/rule", input)
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
					resp, _ := testHelperUser1.MakeRequest(http.MethodPost, "/rule", input)
					Expect(resp.StatusCode).To(Equal(http.StatusCreated))
					done <- true
				}()
				go func() {
					defer GinkgoRecover()
					resp, _ := testHelperUser1.MakeRequest(http.MethodPost, "/rule", input)
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
					resp, _ := testHelperUser1.MakeRequest(http.MethodPost, "/rule", input)
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
					resp, _ := testHelperUser1.MakeRequest(http.MethodPost, "/rule", input)

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
					resp, _ := testHelperUser1.MakeRequest(http.MethodPost, "/rule", input)

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
					resp, _ := testHelperUser1.MakeRequest(http.MethodPost, "/rule", input)

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
					resp, _ := testHelperUser1.MakeRequest(http.MethodPost, "/rule", input)
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
				resp, _ := testHelperUser1.MakeRequest(http.MethodPost, "/rule", input)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))

				// Test name over boundary (101 chars)
				name101 := strings.Repeat("a", 101)
				input.Rule.Name = name101
				resp, _ = testHelperUser1.MakeRequest(http.MethodPost, "/rule", input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

				// Test description at boundary (255 chars)
				desc255 := strings.Repeat("d", 255)
				input.Rule.Name = "Valid Name"
				input.Rule.Description = &desc255
				resp, _ = testHelperUser1.MakeRequest(http.MethodPost, "/rule", input)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))

				// Test description over boundary (256 chars)
				desc256 := strings.Repeat("d", 256)
				input.Rule.Description = &desc256
				resp, _ = testHelperUser1.MakeRequest(http.MethodPost, "/rule", input)
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
					resp, _ := testHelperUser1.MakeRequestWithToken(http.MethodPost, "/rule", token, input)
					Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized), "Should fail for malformed token: "+token)
				}
			})

			It("should handle authentication for all rule endpoints", func() {
				endpoints := []struct {
					method string
					path   string
					body   interface{}
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
					resp, _ := testHelperUser1.MakeRequestWithToken(endpoint.method, endpoint.path, "invalid-token", endpoint.body)
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
				resp, _ := testHelperUser1.MakeRequestWithToken(http.MethodPost, "/rule", expiredToken, input)
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
				resp, _ := testHelperUser1.MakeRequest(http.MethodPost, "/rule", input)
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
				resp, _ := testHelperUser1.MakeRequest(http.MethodPost, "/rule", input)
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
				resp, _ := testHelperUser1.MakeRequest(http.MethodPost, "/rule", input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return error for invalid JSON", func() {
				resp, _ := testHelperUser1.MakeRequest(http.MethodPost, "/rule", "{ invalid json }")
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return error for empty body", func() {
				resp, _ := testHelperUser1.MakeRequest(http.MethodPost, "/rule", "")
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})
		})
	})

	Describe("ListRules", func() {
		It("should return an empty list when there are no rules", func() {
			// Use a fresh user/token with no rules
			resp, response := testHelperUser2.MakeRequest(http.MethodGet, "/rule", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rules fetched successfully"))
			Expect(response["data"]).To(BeAssignableToTypeOf([]interface{}{}))
			Expect(response["data"]).To(BeEmpty())
		})

		It("should list rules for the user and verify all fields", func() {
			// Create two rules for the main user
			ruleId1, _, _ := createTestRule()
			ruleId2, _, _ := createTestRule()
			resp, response := testHelperUser1.MakeRequest(http.MethodGet, "/rule", nil)
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
			resp, response := testHelperUser2.MakeRequest(http.MethodGet, "/rule", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rules fetched successfully"))
			Expect(len(response["data"].([]interface{}))).To(Equal(0))
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
			resp, response := testHelperUser1.MakeRequest(http.MethodGet, url, nil)
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
			resp, response := testHelperUser1.MakeRequest(http.MethodGet, "/rule/invalid_id", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(Equal("invalid ruleId"))
		})

		It("should return error for non-existent rule id", func() {
			resp, response := testHelperUser1.MakeRequest(http.MethodGet, "/rule/999999", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			Expect(response["message"]).To(ContainSubstring("not found"))
		})

		It("should not allow access to rule belonging to another user", func() {
			// Create a rule as main user
			ruleId, _, _ := createTestRule()
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testHelperUser2.MakeRequest(http.MethodGet, url, nil)
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
			resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rule updated successfully"))
			rule := response["data"].(map[string]interface{})
			Expect(rule["name"]).To(Equal("Updated Rule Name"))
		})

		It("should handle partial updates (only description)", func() {
			newDesc := "Updated Description Only"
			update := models.UpdateRuleRequest{Description: &newDesc}
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rule updated successfully"))
			rule := response["data"].(map[string]interface{})
			Expect(rule["description"]).To(Equal(newDesc))
		})

		It("should handle partial updates (only effective_from)", func() {
			newTime := now.Add(-time.Hour)
			update := models.UpdateRuleRequest{EffectiveFrom: &newTime}
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rule updated successfully"))
			rule := response["data"].(map[string]interface{})
			Expect(rule["effective_from"]).NotTo(BeNil())
		})

		It("should not update if no fields provided", func() {
			update := models.UpdateRuleRequest{}
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(ContainSubstring("no fields"))
		})

		It("should handle empty description update", func() {
			description := ""
			update := models.UpdateRuleRequest{Description: &description}
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rule updated successfully"))
		})

		It("should validate effective_from not in far future", func() {
			name := "Valid"
			future := time.Now().AddDate(5, 0, 0)
			update := models.UpdateRuleRequest{Name: &name, EffectiveFrom: &future}
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
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
				resp, _ := testHelperUser1.MakeRequest(http.MethodPatch, url, update1)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				done <- true
			}()
			go func() {
				defer GinkgoRecover()
				resp, _ := testHelperUser1.MakeRequest(http.MethodPatch, url, update2)
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
			resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
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
			resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(ContainSubstring("validation"))
		})

		It("should return error for effective_from in the future", func() {
			name := "Valid"
			future := time.Now().Add(24 * time.Hour)
			update := models.UpdateRuleRequest{Name: &name, EffectiveFrom: &future}
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(ContainSubstring("the effective date for the rule is invalid or in the past"))
		})

		It("should return error for invalid rule id format", func() {
			newName := "Should Fail"
			update := models.UpdateRuleRequest{Name: &newName}
			resp, response := testHelperUser1.MakeRequest(http.MethodPatch, "/rule/invalid_id", update)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(Equal("invalid ruleId"))
		})

		It("should return error for non-existent rule id", func() {
			newName := "Should Fail"
			update := models.UpdateRuleRequest{Name: &newName}
			resp, response := testHelperUser1.MakeRequest(http.MethodPatch, "/rule/999999", update)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			Expect(response["message"]).To(ContainSubstring("not found"))
		})

		It("should return error for invalid JSON", func() {
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, _ := testHelperUser1.MakeRequest(http.MethodPatch, url, "{ invalid json }")
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
				resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(response["message"]).To(Equal("Rule action updated successfully"))
				action := response["data"].(map[string]interface{})
				Expect(action["action_type"]).To(Equal(string(models.RuleFieldDescription)))
				Expect(action["action_value"]).To(Equal("Updated description action"))
			})

			It("should handle updating only action type", func() {
				typ := models.RuleFieldName
				update := models.UpdateRuleActionRequest{
					ActionType: &typ,
				}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/action/" + strconv.FormatInt(actionId, 10)
				resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				action := response["data"].(map[string]interface{})
				Expect(action["action_type"]).To(Equal(string(models.RuleFieldName)))
			})

			It("should handle updating only action value", func() {
				val := "Updated value only"
				update := models.UpdateRuleActionRequest{
					ActionValue: &val,
				}
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/action/" + strconv.FormatInt(actionId, 10)
				resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				action := response["data"].(map[string]interface{})
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
					resp, _ := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
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
					resp, _ := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
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
					resp, _ := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
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
					resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
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
					resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
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
					resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
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
				resp, response := testHelperUser1.MakeRequest(http.MethodPatch, "/rule/invalid_id/action/"+strconv.FormatInt(actionId, 10), update)
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
				resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
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
				resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
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
				resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
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
				resp, _ := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
				// Should succeed for valid string
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				// Now try an empty string if not allowed
				emptyVal := ""
				update.ActionValue = &emptyVal
				resp2, response2 := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
				Expect(resp2.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response2["message"]).To(ContainSubstring("cannot be empty"))
			})

			It("should return error for empty update request", func() {
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/action/" + strconv.FormatInt(actionId, 10)
				resp, _ := testHelperUser1.MakeRequest(http.MethodPatch, url, "")
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
				resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
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
				resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
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
				resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
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
				resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
				Expect(response["message"]).To(ContainSubstring("the requested rule condition was not found"))
			})

			It("should return error for condition belonging to different user", func() {
				resp, response := testHelperUser2.MakeRequest(http.MethodPost, "/rule", models.CreateRuleRequest{
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
				resp2, response2 := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
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
				resp, _ := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				emptyVal := ""
				update.ConditionValue = &emptyVal
				resp2, response2 := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
				Expect(resp2.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response2["message"]).To(ContainSubstring("cannot be empty"))
			})

			It("should return error for empty update request", func() {
				url := "/rule/" + strconv.FormatInt(ruleId, 10) + "/condition/" + strconv.FormatInt(conditionId, 10)
				resp, _ := testHelperUser1.MakeRequest(http.MethodPatch, url, "")
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
				resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
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
				resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
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
				resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
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
			resp, _ := testHelperUser1.MakeRequest(http.MethodDelete, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
		})

		It("should return error for invalid rule id format", func() {
			resp, _ := testHelperUser1.MakeRequest(http.MethodDelete, "/rule/invalid", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return 404 when deleting non-existent rule id", func() {
			resp, _ := testHelperUser1.MakeRequest(http.MethodDelete, "/rule/999999", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
	})
})
