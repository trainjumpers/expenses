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

	ptrToString := func(s string) *string { return &s }
	now := time.Now()

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
		It("should list rules for the user", func() {
			resp, response := testHelper.MakeRequest(http.MethodGet, "/rule", accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rules fetched successfully"))
			Expect(response["data"]).To(BeAssignableToTypeOf([]interface{}{}))
		})
	})

	Describe("GetRuleById", func() {
		It("should get rule by id", func() {
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testHelper.MakeRequest(http.MethodGet, url, accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rule fetched successfully"))
			Expect(response["data"]).To(HaveKey("rule"))
			rule := response["data"].(map[string]interface{})["rule"].(map[string]interface{})
			Expect(int64(rule["id"].(float64))).To(Equal(ruleId))
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
	})

	Describe("UpdateRule", func() {
		It("should update rule name", func() {
			newName := "Updated Rule Name"
			update := models.UpdateRuleRequest{Name: &newName}
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, response := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Rule updated successfully"))
			Expect(response["data"]).To(HaveKey("rule"))
			rule := response["data"].(map[string]interface{})["rule"].(map[string]interface{})
			Expect(rule["name"]).To(Equal("Updated Rule Name"))
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

		It("should return error for empty body", func() {
			url := "/rule/" + strconv.FormatInt(ruleId, 10)
			resp, _ := testHelper.MakeRequest(http.MethodPatch, url, accessToken, "")
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})

	Describe("DeleteRule", func() {
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

	// Additional tests for UpdateRuleAction, UpdateRuleCondition, ExecuteRules can be added here
	// following similar patterns as above, depending on your API and model structure.
})
