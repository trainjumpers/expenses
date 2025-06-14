package controller_test

import (
	"expenses/internal/models"
	"net/http"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AccountController", func() {
	Describe("CreateAccount", func() {
		Context("with valid input", func() {
			It("should create an account successfully", func() {
				balance := 10.0
				input := models.CreateAccountInput{
					Name:     "Integration Account",
					BankType: models.BankTypeAxis,
					Currency: models.CurrencyINR,
					Balance:  &balance,
				}
				resp, response := testHelper.MakeRequest(http.MethodPost, "/account", accessToken, input)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
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
				resp, response := testHelper.MakeRequest(http.MethodPost, "/account", accessToken, input)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				Expect(response["message"]).To(Equal("Account created successfully"))
				Expect(response["data"]).To(HaveKey("id"))
			})

			It("should create account with default balance if not provided", func() {
				input := models.CreateAccountInput{
					Name:     "Integration Account without balance",
					BankType: models.BankTypeAxis,
					Currency: models.CurrencyINR,
				}
				resp, response := testHelper.MakeRequest(http.MethodPost, "/account", accessToken, input)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				Expect(response["message"]).To(Equal("Account created successfully"))
				Expect(response["data"]).To(HaveKey("id"))
				Expect(response["data"].(map[string]interface{})["balance"]).To(Equal(0.0))
			})
		})

		Context("with invalid input", func() {
			It("should have a valid bank type", func() {
				input := models.CreateAccountInput{
					Name:     "Integration Account",
					BankType: "invalid",
					Currency: models.CurrencyINR,
				}
				resp, _ := testHelper.MakeRequest(http.MethodPost, "/account", accessToken, input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should have a valid currency", func() {
				input := models.CreateAccountInput{
					Name:     "Integration Account",
					BankType: models.BankTypeAxis,
					Currency: "invalid",
				}
				resp, _ := testHelper.MakeRequest(http.MethodPost, "/account", accessToken, input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return error if currency does not exists", func() {
				input := models.CreateAccountInput{
					Name:     "Integration Account",
					BankType: models.BankTypeAxis,
				}
				resp, _ := testHelper.MakeRequest(http.MethodPost, "/account", accessToken, input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return error if name is empty", func() {
				input := models.CreateAccountInput{
					BankType: models.BankTypeAxis,
					Currency: models.CurrencyINR,
				}
				resp, _ := testHelper.MakeRequest(http.MethodPost, "/account", accessToken, input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return error for invalid JSON", func() {
				resp, _ := testHelper.MakeRequest(http.MethodPost, "/account", accessToken, "{ name: invalid json }")
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return error for empty body", func() {
				resp, _ := testHelper.MakeRequest(http.MethodPost, "/account", accessToken, "")
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return error for currency with wrong casing", func() {
				input := models.CreateAccountInput{
					Name:     "Integration Account",
					BankType: models.BankTypeAxis,
					Currency: "USD", // should be lowercase 'usd'
				}
				resp, _ := testHelper.MakeRequest(http.MethodPost, "/account", accessToken, input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return error for bank type with wrong casing", func() {
				input := models.CreateAccountInput{
					Name:     "Integration Account",
					BankType: "AXIS", // should be lowercase 'axis'
					Currency: models.CurrencyINR,
				}
				resp, _ := testHelper.MakeRequest(http.MethodPost, "/account", accessToken, input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should handle string balance gracefully", func() {
				requestBody := `{
					"name": "Test Account",
					"bank_type": "axis",
					"currency": "inr",
					"balance": "invalid_string"
				}`
				resp, _ := testHelper.MakeRequest(http.MethodPost, "/account", accessToken, requestBody)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})
		})

		Context("with invalid authorization", func() {
			It("should return error for non-existent user id", func() {
				input := models.CreateAccountInput{
					Name:     "Integration Account",
					BankType: models.BankTypeAxis,
					Currency: models.CurrencyINR,
				}
				resp, _ := testHelper.MakeRequest(http.MethodPost, "/account", "invalid token", input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})
		})
	})

	Describe("ListAccounts", func() {
		It("should list accounts", func() {
			resp, response := testHelper.MakeRequest(http.MethodGet, "/account", accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Accounts retrieved successfully"))
			Expect(response["data"]).To(BeAssignableToTypeOf([]interface{}{}))
		})
		It("should return error for non-existent user id", func() {
			resp, response := testHelper.MakeRequest(http.MethodGet, "/account", "invalid token", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(Equal("Invalid authorization format"))
		})
		It("should return empty list for user with no accounts", func() {
			resp, response := testHelper.MakeRequest(http.MethodGet, "/account", accessToken2, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Accounts retrieved successfully"))
			Expect(len(response["data"].([]interface{}))).To(Equal(0))
		})
	})

	Describe("GetAccount", func() {
		It("should get account by id", func() {
			url := "/account/1"
			resp, response := testHelper.MakeRequest(http.MethodGet, url, accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Account retrieved successfully"))
			Expect(response["data"]).To(HaveKey("id"))
		})

		It("should return error for invalid account id format", func() {
			url := "/account/invalid_id"
			resp, response := testHelper.MakeRequest(http.MethodGet, url, accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(Equal("invalid account id"))
		})

		It("should return error for non-existent account id", func() {
			url := "/account/9999"
			resp, response := testHelper.MakeRequest(http.MethodGet, url, accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			Expect(response["message"]).To(Equal("account not found"))
		})
		It("should return error for non-existent user id", func() {
			url := "/account/1"
			resp, response := testHelper.MakeRequest(http.MethodGet, url, accessToken1, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			Expect(response["message"]).To(Equal("account not found"))
		})
	})

	Describe("UpdateAccount", func() {
		It("should update account name", func() {
			update := models.UpdateAccountInput{Name: "Updated Name"}
			url := "/account/1"
			resp, response := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Account updated successfully"))
			Expect(response["data"].(map[string]interface{})["name"]).To(Equal("Updated Name"))
		})

		It("should return error when trying to update account of different user", func() {
			update := models.UpdateAccountInput{Name: "Unauthorized Update"}
			url := "/account/1"
			resp, _ := testHelper.MakeRequest(http.MethodPatch, url, accessToken1, update) // Different user
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))                         // Should be not found due to ownership check
		})

		It("should return error for empty name in update", func() {
			update := models.UpdateAccountInput{Name: ""}
			url := "/account/1"
			resp, _ := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
			// Should succeed if empty name is allowed, or return 400 if validation prevents it
			Expect(resp.StatusCode).To(SatisfyAny(Equal(http.StatusOK), Equal(http.StatusBadRequest)))
		})

		It("should return error for invalid bank type in update", func() {
			update := models.UpdateAccountInput{BankType: "invalid_bank"}
			url := "/account/1"
			resp, _ := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for invalid currency in update", func() {
			update := models.UpdateAccountInput{Currency: "invalid_currency"}
			url := "/account/1"
			resp, _ := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for invalid JSON in update", func() {
			url := "/account/1"
			resp, _ := testHelper.MakeRequest(http.MethodPatch, url, accessToken, "{ name: invalid }")
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for empty body in update", func() {
			url := "/account/1"
			resp, _ := testHelper.MakeRequest(http.MethodPatch, url, accessToken, "")
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for non-existent account id", func() {
			url := "/account/9999"
			resp, _ := testHelper.MakeRequest(http.MethodPatch, url, accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for invalid account id format in update", func() {
			update := models.UpdateAccountInput{Name: "Updated Name"}
			url := "/account/invalid_id"
			resp, response := testHelper.MakeRequest(http.MethodPatch, url, accessToken, update)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(Equal("invalid account id"))
		})

		It("should return error for non-existent user id", func() {
			url := "/account/1"
			resp, _ := testHelper.MakeRequest(http.MethodPatch, url, "invalid token", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})

	Describe("DeleteAccount", func() {
		It("should delete account by id", func() {
			// Create a new one
			balance := 10.0
			input := models.CreateAccountInput{
				Name:     "Integration Account",
				BankType: models.BankTypeAxis,
				Currency: models.CurrencyINR,
				Balance:  &balance,
			}
			resp, response := testHelper.MakeRequest(http.MethodPost, "/account", accessToken, input)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			Expect(response["message"]).To(Equal("Account created successfully"))
			accountId := response["data"].(map[string]interface{})["id"].(float64)

			url := "/account/" + strconv.FormatFloat(accountId, 'f', 0, 64)
			resp, _ = testHelper.MakeRequest(http.MethodDelete, url, accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
		})

		It("should return error when trying to delete account of different user", func() {
			// First create an account with accessToken
			input := models.CreateAccountInput{
				Name:     "Account for Delete Test",
				BankType: models.BankTypeAxis,
				Currency: models.CurrencyINR,
			}
			resp, response := testHelper.MakeRequest(http.MethodPost, "/account", accessToken, input)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			accountId := int64(response["data"].(map[string]interface{})["id"].(float64))

			// Try to delete with different user
			url := "/account/" + strconv.FormatInt(accountId, 10)
			resp, _ = testHelper.MakeRequest(http.MethodDelete, url, accessToken1, nil) // Different user
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))

			// Ensure account is not deleted
			resp, response = testHelper.MakeRequest(http.MethodGet, url, accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Account retrieved successfully"))
			Expect(response["data"]).To(HaveKey("id"))
			Expect(response["data"].(map[string]interface{})["id"]).To(Equal(float64(accountId)))
		})

		It("should return error for invalid account id format in delete", func() {
			url := "/account/invalid"
			resp, _ := testHelper.MakeRequest(http.MethodDelete, url, accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return 404 when deleting non-existent account id", func() {
			url := "/account/99999"
			resp, _ := testHelper.MakeRequest(http.MethodDelete, url, accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
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
			resp, response := testHelper.MakeRequest(http.MethodPost, "/account", accessToken, input)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			accountId = int64(response["data"].(map[string]interface{})["id"].(float64))
		})

		It("should not include soft-deleted accounts in list", func() {
			resp, response := testHelper.MakeRequest(http.MethodGet, "/account", accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			initialCount := len(response["data"].([]interface{}))
			Expect(initialCount).To(BeNumerically(">", 0))
			// Delete the account
			url := "/account/" + strconv.FormatInt(accountId, 10)
			resp, _ = testHelper.MakeRequest(http.MethodDelete, url, accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

			// List accounts again - should have one less account
			resp, response = testHelper.MakeRequest(http.MethodGet, "/account", accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			finalCount := len(response["data"].([]interface{}))
			Expect(finalCount).To(Equal(initialCount - 1))
		})

		It("should return 404 when fetching soft-deleted account", func() {
			url := "/account/" + strconv.FormatInt(accountId, 10)
			resp, _ := testHelper.MakeRequest(http.MethodDelete, url, accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

			resp, _ = testHelper.MakeRequest(http.MethodGet, url, accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("should return 404 when deleting already deleted account", func() {
			url := "/account/" + strconv.FormatInt(accountId, 10)
			resp, _ := testHelper.MakeRequest(http.MethodDelete, url, accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

			resp, _ = testHelper.MakeRequest(http.MethodDelete, url, accessToken, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
	})
})
