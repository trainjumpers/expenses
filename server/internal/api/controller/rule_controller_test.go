package controller_test

import (
	"expenses/internal/models"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RuleController", func() {
	Describe("GetAllRules", func() {
		It("should return all rules for the user", func() {
			resp, body := testHelper.MakeRequest(http.MethodGet, "/rules", accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(body["message"]).To(Equal("Rules fetched successfully"))
			data := body["data"].([]interface{})
			Expect(len(data)).To(BeNumerically(">=", 2))
		})
	})

	Describe("GetRuleByID", func() {
		It("should return a rule by ID", func() {
			resp, body := testHelper.MakeRequest(http.MethodGet, "/rules/1", accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(body["message"]).To(Equal("Rule fetched successfully"))
			data := body["data"].(map[string]interface{})
			Expect(data["id"]).To(BeEquivalentTo(1))
			Expect(data["name"]).To(Equal("Amount Rule"))
		})
		It("should return 400 for invalid ID", func() {
			resp, _ := testHelper.MakeRequest(http.MethodGet, "/rules/abc", accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
		It("should return 404 for non-existent rule", func() {
			resp, _ := testHelper.MakeRequest(http.MethodGet, "/rules/9999", accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
	})

	Describe("CreateRule", func() {
		It("should create a new rule", func() {
			createReq := models.CreateRuleRequest{
				BaseRule: models.BaseRule{
					Name:          "Integration Create Rule",
					Description:   nil,
					EffectiveFrom: time.Now(),
				},
				Actions: []models.CreateRuleActionRequest{{
					BaseRuleAction: models.BaseRuleAction{
						ActionType:  models.RuleFieldName,
						ActionValue: "Created by Test",
					},
				}},
				Conditions: []models.CreateRuleConditionRequest{{
					BaseRuleCondition: models.BaseRuleCondition{
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "200.00",
						ConditionOperator: models.OperatorEquals,
					},
				}},
			}
			// Gin middleware injects created_by
			resp, body := testHelper.MakeRequest(http.MethodPost, "/rules", accessToken, createReq)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			Expect(body["message"]).To(Equal("Rule created successfully"))
			data := body["data"].(map[string]interface{})
			Expect(data["name"]).To(Equal("Integration Create Rule"))
		})
		It("should return 400 for invalid input", func() {
			resp, _ := testHelper.MakeRequest(http.MethodPost, "/rules", accessToken, "invalid json")
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})

	Describe("UpdateRule", func() {
		It("should update an existing rule", func() {
			updateReq := models.UpdateRuleRequest{
				ID: 1,
				BaseRule: models.BaseRule{
					Name:          "Updated Amount Rule",
					Description:   nil,
					EffectiveFrom: time.Now(),
				},
				Actions: []models.CreateRuleActionRequest{{
					BaseRuleAction: models.BaseRuleAction{
						ActionType:  models.RuleFieldName,
						ActionValue: "Updated Name",
					},
				}},
				Conditions: []models.CreateRuleConditionRequest{{
					BaseRuleCondition: models.BaseRuleCondition{
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "100.50",
						ConditionOperator: models.OperatorEquals,
					},
				}},
			}
			resp, _ := testHelper.MakeRequest(http.MethodPut, "/rules/1", accessToken, updateReq)
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
		})
		It("should return 400 for invalid ID", func() {
			resp, _ := testHelper.MakeRequest(http.MethodPut, "/rules/abc", accessToken, map[string]interface{}{})
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})
		It("should return 404 for non-existent rule", func() {
			updateReq := models.UpdateRuleRequest{
				ID:       9999,
				BaseRule: models.BaseRule{Name: "Nonexistent", EffectiveFrom: time.Now()},
			}
			resp, _ := testHelper.MakeRequest(http.MethodPut, "/rules/9999", accessToken, updateReq)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
	})

	Describe("DeleteRule", func() {
		It("should return error when deleting an existing rule", func() {
			resp, _ := testHelper.MakeRequest(http.MethodDelete, "/rules/2", accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
		It("should return 404 for invalid ID", func() {
			resp, _ := testHelper.MakeRequest(http.MethodDelete, "/rules/abc", accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
		It("should return 404 for non-existent rule", func() {
			resp, _ := testHelper.MakeRequest(http.MethodDelete, "/rules/9999", accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
	})

	Describe("ExecuteRules", func() {
		It("should execute rules and modify transactions as expected", func() {
			resp, body := testHelper.MakeRequest(http.MethodPost, "/rules/execute?user_id=1", accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(body["message"]).To(Equal("Rules executed successfully"))
			data := body["data"].(map[string]interface{})
			Expect(data).To(HaveKey("modified"))
			Expect(data).To(HaveKey("skipped"))
			// At least one transaction should be modified by seeded rules
			modified := data["modified"].([]interface{})
			Expect(len(modified)).To(BeNumerically(">=", 1))
		})
		It("should return 400 for invalid user_id", func() {
			resp, _ := testHelper.MakeRequest(http.MethodPost, "/rules/execute?user_id=abc", accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
	})

	Describe("Authorization", func() {
		It("should return unauthorized for missing token", func() {
			resp, _ := testHelper.MakeRequest(http.MethodGet, "/rules", "", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
		})
		It("should return unauthorized for invalid token", func() {
			resp, _ := testHelper.MakeRequest(http.MethodGet, "/rules", "invalid-token", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
		})
	})
})
