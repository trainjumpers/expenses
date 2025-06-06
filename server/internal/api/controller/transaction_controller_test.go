package controller_test

import (
	"bytes"
	"encoding/json"
	"expenses/internal/models"
	"net/http"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Helper function for creating float64 pointers
func floatPtr(f float64) *float64 {
	return &f
}

var _ = Describe("TransactionController", func() {
	var testDate time.Time
	var futureDate time.Time

	BeforeEach(func() {
		testDate, _ = time.Parse("2006-01-02", "2023-01-01")
		futureDate = time.Now().AddDate(0, 0, 1) // Tomorrow
	})

	Describe("CreateTransaction", func() {
		It("should create a transaction successfully", func() {
			amount := 125.75
			input := models.CreateTransactionInput{
				Name:        "New Integration Transaction",
				Description: "New Test Description",
				Amount:      &amount,
				Date:        testDate,
			}
			body, _ := json.Marshal(input)
			req, err := http.NewRequest(http.MethodPost, baseURL+"/transaction", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Transaction created successfully"))
			Expect(response["data"]).To(HaveKey("id"))
		})

		It("should create transaction without description", func() {
			amount := 85.50
			input := models.CreateTransactionInput{
				Name:   "Transaction without description new",
				Amount: &amount,
				Date:   testDate,
			}
			body, _ := json.Marshal(input)
			req, err := http.NewRequest(http.MethodPost, baseURL+"/transaction", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Transaction created successfully"))
			Expect(response["data"]).To(HaveKey("id"))
		})

		Context("Input Validation", func() {
			It("should return validation error for empty name", func() {
				amount := 100.00
				input := models.CreateTransactionInput{
					Name:   "", // Invalid: empty name
					Amount: &amount,
					Date:   testDate,
				}
				body, _ := json.Marshal(input)
				req, err := http.NewRequest(http.MethodPost, baseURL+"/transaction", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+accessToken)
				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				response, err := decodeJSON(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(ContainSubstring("Error:Field validation"))
			})

			It("should return success for zero amount", func() {
				input := models.CreateTransactionInput{
					Name:   "Valid Transaction",
					Amount: floatPtr(0),
					Date:   testDate,
				}
				body, _ := json.Marshal(input)
				req, err := http.NewRequest(http.MethodPost, baseURL+"/transaction", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+accessToken)
				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				response, err := decodeJSON(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(Equal("Transaction created successfully"))
				Expect(response["data"]).To(HaveKey("id"))
			})

			It("should return validation error for future date", func() {
				input := models.CreateTransactionInput{
					Name:   "Valid Transaction",
					Amount: floatPtr(100.00),
					Date:   futureDate, // Invalid: future date
				}
				body, _ := json.Marshal(input)
				req, err := http.NewRequest(http.MethodPost, baseURL+"/transaction", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+accessToken)
				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				response, err := decodeJSON(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(Equal("transaction date cannot be in the future"))
			})

			It("should return validation error for name too long", func() {
				longName := make([]byte, 201) // 201 characters, exceeds 200 limit
				for i := range longName {
					longName[i] = 'a'
				}

				input := models.CreateTransactionInput{
					Name:   string(longName), // Invalid: too long
					Amount: floatPtr(100.00),
					Date:   testDate,
				}
				body, _ := json.Marshal(input)
				req, err := http.NewRequest(http.MethodPost, baseURL+"/transaction", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+accessToken)
				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				response, err := decodeJSON(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(ContainSubstring("Error:Field validation"))
			})

			It("should return validation error for description too long", func() {
				longDescription := make([]byte, 1001) // 1001 characters, exceeds 1000 limit
				for i := range longDescription {
					longDescription[i] = 'a'
				}

				input := models.CreateTransactionInput{
					Name:        "Valid Transaction",
					Description: string(longDescription), // Invalid: too long
					Amount:      floatPtr(100.00),
					Date:        testDate,
				}
				body, _ := json.Marshal(input)
				req, err := http.NewRequest(http.MethodPost, baseURL+"/transaction", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+accessToken)
				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				response, err := decodeJSON(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(ContainSubstring("Error:Field validation"))
			})

			It("should sanitize input by trimming whitespace", func() {
				input := models.CreateTransactionInput{
					Name:        "  Transaction with spaces  ", // Should be trimmed
					Description: "  Description with spaces  ", // Should be trimmed
					Amount:      floatPtr(100.00),
					Date:        testDate,
				}
				body, _ := json.Marshal(input)
				req, err := http.NewRequest(http.MethodPost, baseURL+"/transaction", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+accessToken)
				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				response, err := decodeJSON(resp.Body)
				Expect(err).NotTo(HaveOccurred())

				// Check that the returned data has trimmed values
				data := response["data"].(map[string]interface{})
				Expect(data["name"]).To(Equal("Transaction with spaces"))
			})
		})

		It("should return error for non-existent user id", func() {
			input := models.CreateTransactionInput{
				Name:   "Transaction with invalid token",
				Amount: floatPtr(100.00),
				Date:   testDate,
			}
			body, _ := json.Marshal(input)
			req, err := http.NewRequest(http.MethodPost, baseURL+"/transaction", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+"invalid token")
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for invalid JSON", func() {
			req, err := http.NewRequest(http.MethodPost, baseURL+"/transaction", bytes.NewBuffer([]byte("{ name: invalid json }")))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for empty body", func() {
			req, err := http.NewRequest(http.MethodPost, baseURL+"/transaction", bytes.NewBuffer([]byte("")))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should handle string amount gracefully", func() {
			requestBody := []byte(`{
				"name": "Test Transaction",
				"amount": "invalid_string",
				"date": "2023-01-01T00:00:00Z"
			}`)
			req, err := http.NewRequest(http.MethodPost, baseURL+"/transaction", bytes.NewBuffer(requestBody))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for duplicate transaction", func() {
			amount := 125.75
			input := models.CreateTransactionInput{
				Name:   "Duplicate transaction",
				Amount: &amount,
				Date:   testDate,
			}
			body, _ := json.Marshal(input)
			req, err := http.NewRequest(http.MethodPost, baseURL+"/transaction", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Transaction created successfully"))
			Expect(response["data"]).To(HaveKey("id"))

			req, err = http.NewRequest(http.MethodPost, baseURL+"/transaction", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err = client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusConflict))
		})
	})

	Describe("ListTransactions", func() {
		It("should list transactions", func() {
			req, err := http.NewRequest(http.MethodGet, baseURL+"/transaction", nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Transactions retrieved successfully"))
			Expect(response["data"]).To(BeAssignableToTypeOf([]interface{}{}))

			transactions := response["data"].([]interface{})
			Expect(len(transactions)).To(BeNumerically(">=", 6))
		})

		It("should return error for non-existent user id", func() {
			req, err := http.NewRequest(http.MethodGet, baseURL+"/transaction", nil)
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

		It("should list empty list for user without transaction", func() {
			req, err := http.NewRequest(http.MethodGet, baseURL+"/transaction", nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken2)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Transactions retrieved successfully"))
			Expect(len(response["data"].([]interface{}))).To(Equal(0))
		})
	})

	Describe("GetTransaction", func() {
		It("should get transaction by id using seed data", func() {
			url := baseURL + "/transaction/1"
			req, err := http.NewRequest(http.MethodGet, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Transaction retrieved successfully"))
			Expect(response["data"]).To(HaveKey("id"))
			Expect(response["data"].(map[string]interface{})["name"]).To(Equal("Integration Transaction"))
		})

		It("should return error for invalid transaction id format", func() {
			url := baseURL + "/transaction/invalid_id"
			req, err := http.NewRequest(http.MethodGet, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("invalid transaction id"))
		})

		It("should return error for non-existent transaction id", func() {
			url := baseURL + "/transaction/9999"
			req, err := http.NewRequest(http.MethodGet, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("transaction not found"))
		})

		It("should return error when trying to access another user's transaction", func() {
			url := baseURL + "/transaction/5"
			req, err := http.NewRequest(http.MethodGet, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("transaction not found"))
		})
	})

	Describe("UpdateTransaction", func() {
		Context("Input Validation", func() {

			It("should return validation error for future date in update", func() {
				update := models.UpdateTransactionInput{Date: futureDate}
				body, _ := json.Marshal(update)
				url := baseURL + "/transaction/3"
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
				Expect(response["message"]).To(Equal("transaction date cannot be in the future"))
			})
		})

		It("should update transaction name using seed data", func() {
			update := models.UpdateTransactionInput{Name: "Updated Transaction Name"}
			body, _ := json.Marshal(update)
			url := baseURL + "/transaction/3"
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
			Expect(response["message"]).To(Equal("Transaction updated successfully"))
			Expect(response["data"].(map[string]interface{})["name"]).To(Equal("Updated Transaction Name"))
		})

		It("should update transaction amount using seed data", func() {
			amount := 350.99
			update := models.UpdateTransactionInput{Amount: &amount}
			body, _ := json.Marshal(update)
			url := baseURL + "/transaction/3"
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
			Expect(response["message"]).To(Equal("Transaction updated successfully"))
			Expect(response["data"].(map[string]interface{})["amount"]).To(Equal(amount))
		})

		It("should return error when trying to update transaction of different user", func() {
			update := models.UpdateTransactionInput{Name: "Unauthorized Update"}
			body, _ := json.Marshal(update)
			url := baseURL + "/transaction/5"
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("should return error for invalid JSON in update", func() {
			url := baseURL + "/transaction/3"
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
			url := baseURL + "/transaction/3"
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer([]byte("")))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for non-existent transaction id", func() {
			update := models.UpdateTransactionInput{Name: "Updated Name"}
			body, _ := json.Marshal(update)
			url := baseURL + "/transaction/9999"
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("should return error for invalid transaction id format in update", func() {
			update := models.UpdateTransactionInput{Name: "Updated Name"}
			body, _ := json.Marshal(update)
			url := baseURL + "/transaction/invalid_id"
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
			Expect(response["message"]).To(Equal("invalid transaction id"))
		})
	})

	Describe("DeleteTransaction", func() {
		It("should delete transaction by id using seed data", func() {
			url := baseURL + "/transaction/4"
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
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

		It("should return error when trying to delete transaction of different user", func() {
			url := baseURL + "/transaction/6"
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("should return error for non-existent transaction id", func() {
			url := baseURL + "/transaction/9999"
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("should return error for invalid transaction id format in delete", func() {
			url := baseURL + "/transaction/invalid_id"
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("invalid transaction id"))
		})
	})

	Describe("End to End cases", func() {
		It("should create and manipulate new transactions without conflicting with seed data", func() {
			input := models.CreateTransactionInput{
				Name:        "Dynamically Created Transaction",
				Description: "Created during test",
				Amount:      floatPtr(999.99),
				Date:        testDate,
			}
			body, _ := json.Marshal(input)
			req, err := http.NewRequest(http.MethodPost, baseURL+"/transaction", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())

			createdId := int64(response["data"].(map[string]interface{})["id"].(float64))

			update := models.UpdateTransactionInput{Name: "Updated Dynamic Transaction"}
			body, _ = json.Marshal(update)
			url := baseURL + "/transaction/" + strconv.FormatInt(createdId, 10)
			req, err = http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err = client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			req, err = http.NewRequest(http.MethodDelete, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err = client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
		})

		It("should handle cross-user isolation properly with seed data", func() {
			req, err := http.NewRequest(http.MethodGet, baseURL+"/transaction", nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())

			transactions := response["data"].([]interface{})
			for _, tx := range transactions {
				transaction := tx.(map[string]interface{})
				createdBy := int64(transaction["created_by"].(float64))
				Expect(createdBy).To(Equal(int64(1)))
			}
		})
	})
})
