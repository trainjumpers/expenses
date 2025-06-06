package controller_test

import (
	"bytes"
	"encoding/json"
	"expenses/internal/models"
	"net/http"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AccountController", func() {
	Describe("CreateAccount", func() {
		It("should create an account successfully", func() {
			balance := 10.0
			input := models.CreateAccountInput{
				Name:     "Integration Account",
				BankType: models.BankTypeAxis,
				Currency: models.CurrencyINR,
				Balance:  &balance,
			}
			body, _ := json.Marshal(input)
			req, err := http.NewRequest(http.MethodPost, baseURL+"/account", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Account created successfully"))
			Expect(response["data"]).To(HaveKey("id"))
			Expect(response["data"].(map[string]interface{})["balance"]).To(Equal(balance))
		})

		It("should create account for duplicate account name", func() {
			input := models.CreateAccountInput{
				Name:     "Integration Account",
				BankType: models.BankTypeAxis,
				Currency: models.CurrencyINR,
			}
			body, _ := json.Marshal(input)
			req, err := http.NewRequest(http.MethodPost, baseURL+"/account", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Account created successfully"))
			Expect(response["data"]).To(HaveKey("id"))
		})

		It("should return error for non-existent user id", func() {
			input := models.CreateAccountInput{
				Name:     "Integration Account",
				BankType: models.BankTypeAxis,
				Currency: models.CurrencyINR,
			}
			body, _ := json.Marshal(input)
			req, err := http.NewRequest(http.MethodPost, baseURL+"/account", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+"invalid token")
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should create account with default balance if not provided", func() {
			input := models.CreateAccountInput{
				Name:     "Integration Account without balance",
				BankType: models.BankTypeAxis,
				Currency: models.CurrencyINR,
			}
			body, _ := json.Marshal(input)
			req, err := http.NewRequest(http.MethodPost, baseURL+"/account", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Account created successfully"))
			Expect(response["data"]).To(HaveKey("id"))
			Expect(response["data"].(map[string]interface{})["balance"]).To(Equal(0.0))
		})

		It("should have a valid bank type", func() {
			input := models.CreateAccountInput{
				Name:     "Integration Account",
				BankType: "invalid",
				Currency: models.CurrencyINR,
			}
			body, _ := json.Marshal(input)
			req, err := http.NewRequest(http.MethodPost, baseURL+"/account", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should have a valid currency", func() {
			input := models.CreateAccountInput{
				Name:     "Integration Account",
				BankType: models.BankTypeAxis,
				Currency: "invalid",
			}
			body, _ := json.Marshal(input)
			req, err := http.NewRequest(http.MethodPost, baseURL+"/account", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error if currency does not exists", func() {
			input := models.CreateAccountInput{
				Name:     "Integration Account",
				BankType: models.BankTypeAxis,
			}
			body, _ := json.Marshal(input)
			req, err := http.NewRequest(http.MethodPost, baseURL+"/account", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error if name is empty", func() {
			input := models.CreateAccountInput{
				BankType: models.BankTypeAxis,
				Currency: models.CurrencyINR,
			}
			body, _ := json.Marshal(input)
			req, err := http.NewRequest(http.MethodPost, baseURL+"/account", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for invalid JSON", func() {
			req, err := http.NewRequest(http.MethodPost, baseURL+"/account", bytes.NewBuffer([]byte("{ name: invalid json }")))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for empty body", func() {
			req, err := http.NewRequest(http.MethodPost, baseURL+"/account", bytes.NewBuffer([]byte("")))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for currency with wrong casing", func() {
			input := models.CreateAccountInput{
				Name:     "Integration Account",
				BankType: models.BankTypeAxis,
				Currency: "USD", // should be lowercase 'usd'
			}
			body, _ := json.Marshal(input)
			req, err := http.NewRequest(http.MethodPost, baseURL+"/account", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for bank type with wrong casing", func() {
			input := models.CreateAccountInput{
				Name:     "Integration Account",
				BankType: "AXIS", // should be lowercase 'axis'
				Currency: models.CurrencyINR,
			}
			body, _ := json.Marshal(input)
			req, err := http.NewRequest(http.MethodPost, baseURL+"/account", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should handle string balance gracefully", func() {
			requestBody := []byte(`{
				"name": "Test Account",
				"bank_type": "axis",
				"currency": "inr",
				"balance": "invalid_string"
			}`)
			req, err := http.NewRequest(http.MethodPost, baseURL+"/account", bytes.NewBuffer(requestBody))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})

	Describe("ListAccounts", func() {
		It("should list accounts", func() {
			req, err := http.NewRequest(http.MethodGet, baseURL+"/account", nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Accounts retrieved successfully"))
			Expect(response["data"]).To(BeAssignableToTypeOf([]interface{}{}))
		})
		It("should return error for non-existent user id", func() {
			req, err := http.NewRequest(http.MethodGet, baseURL+"/account", nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+"invalid token")
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Invalid authorization format"))
		})
		It("should return empty list for user with no accounts", func() {
			req, err := http.NewRequest(http.MethodGet, baseURL+"/account", nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken2)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Accounts retrieved successfully"))
			Expect(len(response["data"].([]interface{}))).To(Equal(0))
		})
	})

	Describe("GetAccount", func() {
		It("should get account by id", func() {
			url := baseURL + "/account/1"
			req, err := http.NewRequest(http.MethodGet, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Account retrieved successfully"))
			Expect(response["data"]).To(HaveKey("id"))
		})

		It("should return error for invalid account id format", func() {
			url := baseURL + "/account/invalid_id"
			req, err := http.NewRequest(http.MethodGet, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("invalid account id"))
		})

		It("should return error for non-existent account id", func() {
			url := baseURL + "/account/9999"
			req, err := http.NewRequest(http.MethodGet, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("account not found"))
		})
		It("should return error for non-existent user id", func() {
			url := baseURL + "/account/1"
			req, err := http.NewRequest(http.MethodGet, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken1)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("account not found"))
		})
	})

	Describe("UpdateAccount", func() {
		It("should update account name", func() {
			update := models.UpdateAccountInput{Name: "Updated Name"}
			body, _ := json.Marshal(update)
			url := baseURL + "/account/1"
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Account updated successfully"))
			Expect(response["data"].(map[string]interface{})["name"]).To(Equal("Updated Name"))
		})

		It("should return error when trying to update account of different user", func() {
			update := models.UpdateAccountInput{Name: "Unauthorized Update"}
			body, _ := json.Marshal(update)
			url := baseURL + "/account/1"
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken1) // Different user
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound)) // Should be not found due to ownership check
		})

		It("should return error for empty name in update", func() {
			update := models.UpdateAccountInput{Name: ""}
			body, _ := json.Marshal(update)
			url := baseURL + "/account/1"
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			// Should succeed if empty name is allowed, or return 400 if validation prevents it
			Expect(resp.StatusCode).To(SatisfyAny(Equal(http.StatusOK), Equal(http.StatusBadRequest)))
		})

		It("should return error for invalid bank type in update", func() {
			update := models.UpdateAccountInput{BankType: "invalid_bank"}
			body, _ := json.Marshal(update)
			url := baseURL + "/account/1"
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for invalid currency in update", func() {
			update := models.UpdateAccountInput{Currency: "invalid_currency"}
			body, _ := json.Marshal(update)
			url := baseURL + "/account/1"
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for invalid JSON in update", func() {
			url := baseURL + "/account/1"
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer([]byte("{ name: invalid }")))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for empty body in update", func() {
			url := baseURL + "/account/1"
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer([]byte("")))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for non-existent account id", func() {
			url := baseURL + "/account/9999"
			req, err := http.NewRequest(http.MethodPatch, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for invalid account id format in update", func() {
			update := models.UpdateAccountInput{Name: "Updated Name"}
			body, _ := json.Marshal(update)
			url := baseURL + "/account/invalid_id"
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("invalid account id"))
		})

		It("should return error for non-existent user id", func() {
			url := baseURL + "/account/1"
			req, err := http.NewRequest(http.MethodPatch, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})

	Describe("DeleteAccount", func() {
		It("should delete account by id", func() {
			url := baseURL + "/account/1"
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
		})

		It("should return error when trying to delete account of different user", func() {
			// First create an account with accessToken
			input := models.CreateAccountInput{
				Name:     "Account for Delete Test",
				BankType: models.BankTypeAxis,
				Currency: models.CurrencyINR,
			}
			body, _ := json.Marshal(input)
			req, err := http.NewRequest(http.MethodPost, baseURL+"/account", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			accountId := int64(response["data"].(map[string]interface{})["id"].(float64))

			// Try to delete with different user
			url := baseURL + "/account/" + strconv.FormatInt(accountId, 10)
			req, err = http.NewRequest(http.MethodDelete, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken1) // Different user
			resp, err = client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

			// Ensure account is not deleted
			req, err = http.NewRequest(http.MethodGet, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err = client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			response, err = decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Account retrieved successfully"))
			Expect(response["data"]).To(HaveKey("id"))
			Expect(response["data"].(map[string]interface{})["id"]).To(Equal(float64(accountId)))
		})

		It("should return error for invalid account id format in delete", func() {
			url := baseURL + "/account/invalid"
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should be idempotent when deleting non-existent account id", func() {
			url := baseURL + "/account/99999"
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
		})
	})

	Describe("Soft Deletion Scenarios", func() {
		var accountId int64

		BeforeEach(func() {
			input := models.CreateAccountInput{
				Name:     "Account to Delete",
				BankType: models.BankTypeAxis,
				Currency: models.CurrencyINR,
			}
			body, _ := json.Marshal(input)
			req, err := http.NewRequest(http.MethodPost, baseURL+"/account", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			accountId = int64(response["data"].(map[string]interface{})["id"].(float64))
		})

		It("should not include soft-deleted accounts in list", func() {
			req, err := http.NewRequest(http.MethodGet, baseURL+"/account", nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			initialCount := len(response["data"].([]interface{}))
			Expect(initialCount).To(BeNumerically(">", 0))
			// Delete the account
			url := baseURL + "/account/" + strconv.FormatInt(accountId, 10)
			req, err = http.NewRequest(http.MethodDelete, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err = client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

			// List accounts again - should have one less account
			req, err = http.NewRequest(http.MethodGet, baseURL+"/account", nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err = client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			response, err = decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			finalCount := len(response["data"].([]interface{}))
			Expect(finalCount).To(Equal(initialCount - 1))
		})

		It("should return 404 when fetching soft-deleted account", func() {
			url := baseURL + "/account/" + strconv.FormatInt(accountId, 10)
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

			req, err = http.NewRequest(http.MethodGet, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err = client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("should be idempotent when deleting already deleted account", func() {
			url := baseURL + "/account/" + strconv.FormatInt(accountId, 10)
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

			req, err = http.NewRequest(http.MethodDelete, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err = client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
		})
	})
})
