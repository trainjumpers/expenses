package controller

import (
	"bytes"
	"encoding/json"
	"expenses/internal/models"
	"net/http"

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
	})
})
