package controller_test

import (
	"expenses/internal/models"
	"math"
	"net/http"
	"strconv"
	"strings"
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
		It("should create a transaction for user 2", func() {
			accInput := models.CreateAccountInput{
				Name:     "User2 Account",
				BankType: models.BankTypeAxis,
				Currency: models.CurrencyINR,
			}
			resp, response := testUser2.MakeRequest(http.MethodPost, "/account", accInput)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			accountId := int64(response["data"].(map[string]any)["id"].(float64))

			amount := 100.0
			txInput := models.CreateBaseTransactionInput{
				Name:      "User2 Transaction",
				Amount:    &amount,
				Date:      testDate,
				AccountId: accountId,
			}
			resp, response = testUser2.MakeRequest(http.MethodPost, "/transaction", txInput)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			data := response["data"].(map[string]any)
			Expect(data["name"]).To(Equal("User2 Transaction"))
		})

		It("should not allow user 3 to access user 1's transaction", func() {
			url := "/transaction/1"
			resp, _ := testUser3.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("should create a transaction with category mappings", func() {
			input := map[string]any{
				"name":         "Transaction with mappings",
				"description":  "Test with category and account",
				"amount":       200.00,
				"date":         testDate.Format(time.RFC3339),
				"category_ids": []int64{6, 7},
				"account_id":   3,
			}
			resp, response := testUser2.MakeRequest(http.MethodPost, "/transaction", input)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			Expect(response["message"]).To(Equal("Transaction created successfully"))
			data := response["data"].(map[string]any)
			Expect(data["category_ids"]).To(ContainElements(float64(6), float64(7)))
			Expect(data["account_id"]).To(Equal(float64(3)))
		})

		It("should create transaction without description", func() {
			amount := 85.50
			input := models.CreateBaseTransactionInput{
				Name:      "Transaction without description new",
				Amount:    &amount,
				Date:      testDate,
				AccountId: 3,
			}
			resp, response := testUser2.MakeRequest(http.MethodPost, "/transaction", input)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			Expect(response["message"]).To(Equal("Transaction created successfully"))
			Expect(response["data"]).To(HaveKey("id"))
		})

		Context("Input Validation", func() {
			It("should return validation error for empty name", func() {
				amount := 100.00
				input := models.CreateBaseTransactionInput{
					Name:      "", // Invalid: empty name
					Amount:    &amount,
					Date:      testDate,
					AccountId: 3,
				}
				resp, response := testUser2.MakeRequest(http.MethodPost, "/transaction", input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(ContainSubstring("Error:Field validation"))
			})

			It("should return success for zero amount", func() {
				input := models.CreateBaseTransactionInput{
					Name:      "Valid Transaction",
					Amount:    floatPtr(0),
					Date:      testDate,
					AccountId: 3,
				}
				resp, response := testUser2.MakeRequest(http.MethodPost, "/transaction", input)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				Expect(response["message"]).To(Equal("Transaction created successfully"))
				Expect(response["data"]).To(HaveKey("id"))
			})

			It("should return validation error for future date", func() {
				input := models.CreateBaseTransactionInput{
					Name:      "Valid Transaction",
					Amount:    floatPtr(100.00),
					Date:      futureDate, // Invalid: future date
					AccountId: 3,
				}
				resp, response := testUser2.MakeRequest(http.MethodPost, "/transaction", input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(Equal("transaction date cannot be in the future"))
			})

			It("should return validation error for name too long", func() {
				longName := make([]byte, 201) // 201 characters, exceeds 200 limit
				for i := range longName {
					longName[i] = 'a'
				}

				input := models.CreateBaseTransactionInput{
					Name:      string(longName), // Invalid: too long
					Amount:    floatPtr(100.00),
					Date:      testDate,
					AccountId: 3,
				}
				resp, response := testUser2.MakeRequest(http.MethodPost, "/transaction", input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(ContainSubstring("Error:Field validation"))
			})

			It("should return validation error for description too long", func() {
				longDescription := make([]byte, 1001) // 1001 characters, exceeds 1000 limit
				for i := range longDescription {
					longDescription[i] = 'a'
				}

				input := models.CreateBaseTransactionInput{
					Name:        "Valid Transaction",
					Description: string(longDescription), // Invalid: too long
					Amount:      floatPtr(100.00),
					Date:        testDate,
					AccountId:   3,
				}
				resp, response := testUser2.MakeRequest(http.MethodPost, "/transaction", input)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(ContainSubstring("Error:Field validation"))
			})

			It("should sanitize input by trimming whitespace", func() {
				input := models.CreateBaseTransactionInput{
					Name:        "  Transaction with spaces  ", // Should be trimmed
					Description: "  Description with spaces  ", // Should be trimmed
					Amount:      floatPtr(100.00),
					Date:        testDate,
					AccountId:   3,
				}
				resp, response := testUser2.MakeRequest(http.MethodPost, "/transaction", input)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))

				// Check that the returned data has trimmed values
				data := response["data"].(map[string]any)
				Expect(data["name"]).To(Equal("Transaction with spaces"))
				Expect(data["description"]).To(Equal("Description with spaces"))
			})

			It("should return validation error for invalid category id", func() {
				input := map[string]any{
					"name":         "Transaction with mappings",
					"description":  "Test with category and account",
					"amount":       100.00,
					"date":         testDate.Format(time.RFC3339),
					"category_ids": []int64{9999},
					"account_id":   3,
				}
				resp, response := testUser2.MakeRequest(http.MethodPost, "/transaction", input)
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
				Expect(response["message"]).To(ContainSubstring("category not found"))
			})

			It("should return validation error for invalid account id", func() {
				input := map[string]any{
					"name":        "Transaction with mappings",
					"description": "Test with category and account",
					"amount":      100.00,
					"date":        testDate.Format(time.RFC3339),
					"account_id":  9999,
				}
				resp, response := testUser2.MakeRequest(http.MethodPost, "/transaction", input)
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
				Expect(response["message"]).To(ContainSubstring("account not found"))
			})

			It("should return validation error when creating transaction with a different user's category", func() {
				input := map[string]any{
					"name":         "Transaction with mappings",
					"description":  "Test with category and account",
					"amount":       100.00,
					"date":         testDate.Format(time.RFC3339),
					"category_ids": []int64{1, 2},
					"account_id":   3,
				}
				resp, response := testUser2.MakeRequest(http.MethodPost, "/transaction", input)
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
				Expect(response["message"]).To(ContainSubstring("category not found"))
			})
		})

		It("should return error for non-existent user id", func() {
			input := models.CreateBaseTransactionInput{
				Name:      "Transaction with invalid token",
				Amount:    floatPtr(100.00),
				Date:      testDate,
				AccountId: 3,
			}
			resp, _ := testHelperUnauthenticated.MakeRequest(http.MethodPost, "/transaction", input)
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		})

		It("should return error for invalid JSON", func() {
			resp, _ := testUser2.MakeRequest(http.MethodPost, "/transaction", "{ name: invalid json }")
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for empty body", func() {
			resp, _ := testUser2.MakeRequest(http.MethodPost, "/transaction", "")
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should handle string amount gracefully", func() {
			requestBody := `{
				"name": "Test Transaction",
				"amount": "invalid_string",
				"date": "2023-01-01T00:00:00Z",
				"account_id": 1
			}`
			resp, _ := testUser2.MakeRequest(http.MethodPost, "/transaction", requestBody)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for duplicate transaction", func() {
			amount := 125.75
			input := models.CreateBaseTransactionInput{
				Name:      "Duplicate transaction",
				Amount:    &amount,
				Date:      testDate,
				AccountId: 3,
			}
			resp, response := testUser2.MakeRequest(http.MethodPost, "/transaction", input)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			Expect(response["message"]).To(Equal("Transaction created successfully"))
			Expect(response["data"]).To(HaveKey("id"))

			resp, _ = testUser2.MakeRequest(http.MethodPost, "/transaction", input)
			Expect(resp.StatusCode).To(Equal(http.StatusConflict))
		})

		It("should return error for invalid category id on create", func() {
			amount := 200.00
			input := map[string]any{
				"name":         "Transaction with invalid category",
				"description":  "Test with invalid category id",
				"amount":       amount,
				"date":         testDate.Format(time.RFC3339),
				"category_ids": []int64{99999}, // Invalid category Id
				"account_id":   2,
			}
			resp, _ := testUser2.MakeRequest(http.MethodPost, "/transaction", input)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("should return error for invalid account id on create", func() {
			amount := 200.00
			input := map[string]any{
				"name":         "Transaction with invalid account",
				"description":  "Test with invalid account id",
				"amount":       amount,
				"date":         testDate.Format(time.RFC3339),
				"category_ids": []int64{1},
				"account_id":   99999, // Invalid account Id
			}
			resp, _ := testUser2.MakeRequest(http.MethodPost, "/transaction", input)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("should not allow adding category mapping of a different user on create", func() {
			// Create a category as a different user
			catInput := map[string]any{
				"name": "Other User Category for Create",
				"icon": "other-icon-create",
			}
			catResp, catResponse := testUser2.MakeRequest(http.MethodPost, "/category", catInput)
			Expect(catResp.StatusCode).To(Equal(http.StatusCreated))
			otherCategoryId := int64(catResponse["data"].(map[string]any)["id"].(float64))

			amount := 200.00
			input := map[string]any{
				"name":         "Transaction with other user category",
				"description":  "Should fail",
				"amount":       amount,
				"date":         testDate.Format(time.RFC3339),
				"category_ids": []int64{otherCategoryId},
				"account_id":   2,
			}
			resp, _ := testUser2.MakeRequest(http.MethodPost, "/transaction", input)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("should not allow adding account mapping of a different user on create", func() {
			// Create an account as a different user
			accInput := map[string]any{
				"name":      "Other User Account for Create",
				"bank_type": "axis",
				"currency":  "inr",
				"balance":   10.0,
			}
			accResp, accResponse := testUser1.MakeRequest(http.MethodPost, "/account", accInput)
			Expect(accResp.StatusCode).To(Equal(http.StatusCreated))
			otherAccountId := int64(accResponse["data"].(map[string]any)["id"].(float64))

			amount := 200.00
			input := map[string]any{
				"name":        "Transaction with other user account",
				"description": "Should fail",
				"amount":      amount,
				"date":        testDate.Format(time.RFC3339),
				"account_id":  otherAccountId,
			}
			resp, _ := testUser2.MakeRequest(http.MethodPost, "/transaction", input)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		Context("with malformed tokens", func() {
			It("should return unauthorized or bad request for malformed tokens on create", func() {
				input := models.CreateBaseTransactionInput{
					Name:      "Malformed Token Transaction",
					Amount:    floatPtr(100.00),
					Date:      testDate,
					AccountId: 3,
				}
				checkMalformedTokens(testUser2, http.MethodPost, "/transaction", input)
			})
			It("should return unauthorized or bad request for malformed tokens on list", func() {
				checkMalformedTokens(testUser2, http.MethodGet, "/transaction", nil)
			})
			It("should return unauthorized or bad request for malformed tokens on get", func() {
				url := "/transaction/1"
				checkMalformedTokens(testUser2, http.MethodGet, url, nil)
			})
			It("should return unauthorized or bad request for malformed tokens on update", func() {
				update := map[string]any{"name": "Malformed Update"}
				url := "/transaction/1"
				checkMalformedTokens(testUser2, http.MethodPatch, url, update)
			})
			It("should return unauthorized or bad request for malformed tokens on delete", func() {
				url := "/transaction/1"
				checkMalformedTokens(testUser2, http.MethodDelete, url, nil)
			})
		})
	})

	Describe("ListTransactions", func() {
		It("should list transactions with default pagination", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/transaction", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Transactions retrieved successfully"))
			data := response["data"].(map[string]any)
			Expect(data["total"]).To(Equal(float64(11)))
			transactions := data["transactions"].([]any)
			Expect(len(transactions)).To(Equal(11))
		})

		It("should handle pagination correctly", func() {
			// First page
			resp, response := testUser1.MakeRequest(http.MethodGet, "/transaction?page=1&page_size=3", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			data := response["data"].(map[string]any)
			Expect(data["total"]).To(Equal(float64(11)))
			transactions := data["transactions"].([]any)
			Expect(len(transactions)).To(Equal(3))

			// Second page
			resp, response = testUser1.MakeRequest(http.MethodGet, "/transaction?page=2&page_size=3", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			data = response["data"].(map[string]any)
			transactions = data["transactions"].([]any)
			Expect(len(transactions)).To(Equal(3))

			// Last page
			resp, response = testUser1.MakeRequest(http.MethodGet, "/transaction?page=4&page_size=3", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			data = response["data"].(map[string]any)
			transactions = data["transactions"].([]any)
			Expect(len(transactions)).To(Equal(2)) // Updated to 2 since we now have 11 transactions (11 % 3 = 2)
		})

		It("should filter by account Id", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/transaction?account_id=1", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			data := response["data"].(map[string]any)
			transactions := data["transactions"].([]any)
			Expect(len(transactions)).To(Equal(7)) // Updated to 7 since we added one more transaction to account 1
			for _, tx := range transactions {
				txMap := tx.(map[string]any)
				Expect(txMap["account_id"]).To(Equal(float64(1)))
			}
		})

		It("should filter by category Id", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/transaction?category_id=1", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			data := response["data"].(map[string]any)
			transactions := data["transactions"].([]any)
			Expect(len(transactions)).To(Equal(3))
			for _, tx := range transactions {
				txMap := tx.(map[string]any)
				categoryIds := txMap["category_ids"].([]any)
				found := false
				for _, catId := range categoryIds {
					if catId.(float64) == float64(1) {
						found = true
						break
					}
				}
				Expect(found).To(BeTrue())
			}
		})

		It("should filter uncategorized transactions", func() {
			// Filter for uncategorized transactions
			resp, response := testUser1.MakeRequest(http.MethodGet, "/transaction?uncategorized=true", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			data := response["data"].(map[string]any)
			transactions := data["transactions"].([]any)

			// Should have exactly 1 uncategorized transaction from seed data
			Expect(len(transactions)).To(Equal(1))

			// Verify all returned transactions have no categories
			for _, tx := range transactions {
				txMap := tx.(map[string]any)
				categoryIds := txMap["category_ids"].([]any)
				Expect(categoryIds).To(BeEmpty())
			}

			// Verify the uncategorized transaction is "Cash Withdrawal" from seed data
			txMap := transactions[0].(map[string]any)
			Expect(txMap["name"].(string)).To(Equal("Cash Withdrawal"))
		})

		It("should filter by amount range", func() {
			url := "/transaction?min_amount=50&max_amount=150"
			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			data := response["data"].(map[string]any)
			transactions := data["transactions"].([]any)
			for _, tx := range transactions {
				txMap := tx.(map[string]any)
				amount := txMap["amount"].(float64)
				Expect(amount).To(And(
					BeNumerically(">=", 50),
					BeNumerically("<=", 150),
				))
			}
		})

		It("should filter by date range", func() {
			url := "/transaction?date_from=2023-01-03&date_to=2023-01-05"
			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			data := response["data"].(map[string]any)
			transactions := data["transactions"].([]any)
			Expect(len(transactions)).To(Equal(4))
			for _, tx := range transactions {
				txMap := tx.(map[string]any)
				dateStr := txMap["date"].(string)
				date, err := time.Parse(time.RFC3339, dateStr)
				Expect(err).To(BeNil())
				startDate, _ := time.Parse(time.RFC3339, "2023-01-03T00:00:00Z")
				endDate, _ := time.Parse(time.RFC3339, "2023-01-05T23:59:59Z")
				Expect(date.After(startDate.Add(-time.Second)) && date.Before(endDate.Add(time.Second))).To(BeTrue())
			}
		})

		It("should handle search by name", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/transaction?search=Shopping", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			data := response["data"].(map[string]any)
			transactions := data["transactions"].([]any)
			Expect(len(transactions)).To(Equal(2)) // "Groceries Shopping" and "Online Shopping"
			for _, tx := range transactions {
				txMap := tx.(map[string]any)
				name := txMap["name"].(string)
				Expect(name).To(ContainSubstring("Shopping"))
			}
		})

		It("should handle search by description", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/transaction?search=monthly", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			data := response["data"].(map[string]any)
			transactions := data["transactions"].([]any)
			Expect(len(transactions)).To(Equal(2)) // "Monthly groceries" and "Monthly internet"
			for _, tx := range transactions {
				txMap := tx.(map[string]any)
				description := txMap["description"].(string)
				Expect(strings.ToLower(description)).To(ContainSubstring("monthly"))
			}
		})

		It("should handle multiple filters together", func() {
			url := "/transaction?account_id=1&category_id=1&min_amount=50&max_amount=150&date_from=2023-01-01&date_to=2023-01-03"
			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			data := response["data"].(map[string]any)
			transactions := data["transactions"].([]any)
			for _, tx := range transactions {
				txMap := tx.(map[string]any)
				// Check account
				Expect(txMap["account_id"]).To(Equal(float64(1)))
				// Check amount range
				amount := txMap["amount"].(float64)
				Expect(amount).To(And(
					BeNumerically(">=", 50),
					BeNumerically("<=", 150),
				))
				// Check date range
				date := txMap["date"].(string)
				Expect(date >= "2023-01-01" && date <= "2023-01-03").To(BeTrue())
				// Check category
				categoryIds := txMap["category_ids"].([]any)
				found := false
				for _, catId := range categoryIds {
					if catId.(float64) == float64(1) {
						found = true
						break
					}
				}
				Expect(found).To(BeTrue())
			}
		})

		It("should sort by date ascending", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/transaction?sort_by=date&sort_order=asc", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			data := response["data"].(map[string]any)
			transactions := data["transactions"].([]any)
			var lastDate time.Time
			for _, tx := range transactions {
				txMap := tx.(map[string]any)
				currentDateStr := txMap["date"].(string)
				currentDate, err := time.Parse(time.RFC3339, currentDateStr)
				Expect(err).To(BeNil())
				if !lastDate.IsZero() {
					Expect(currentDate.After(lastDate) || currentDate.Equal(lastDate)).To(BeTrue())
				}
				lastDate = currentDate
			}
		})

		It("should sort by amount descending", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/transaction?sort_by=amount&sort_order=desc", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			data := response["data"].(map[string]any)
			transactions := data["transactions"].([]any)
			var lastAmount float64 = math.MaxFloat64
			for _, tx := range transactions {
				txMap := tx.(map[string]any)
				currentAmount := txMap["amount"].(float64)
				Expect(currentAmount).To(BeNumerically("<=", lastAmount))
				lastAmount = currentAmount
			}
		})

		It("should return error for non-existent user id", func() {
			resp, response := testHelperUnauthenticated.MakeRequest(http.MethodGet, "/transaction", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			Expect(response["message"]).To(Equal("please log in to continue"))
		})

		It("should list empty list for user without transaction", func() {
			resp, response := testUser3.MakeRequest(http.MethodGet, "/transaction", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Transactions retrieved successfully"))
			data := response["data"].(map[string]any)
			Expect(data["total"]).To(Equal(float64(0)))
			Expect(data["transactions"]).To(BeEmpty())
		})

		It("should handle invalid filter values gracefully", func() {
			// Invalid account_id
			resp, _ := testUser1.MakeRequest(http.MethodGet, "/transaction?account_id=invalid", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK)) // Should ignore invalid filter

			// Invalid category_id
			resp, _ = testUser1.MakeRequest(http.MethodGet, "/transaction?category_id=invalid", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK)) // Should ignore invalid filter

			// Invalid amount
			resp, _ = testUser1.MakeRequest(http.MethodGet, "/transaction?min_amount=invalid", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK)) // Should ignore invalid filter

			// Invalid date
			resp, _ = testUser1.MakeRequest(http.MethodGet, "/transaction?date_from=invalid", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK)) // Should ignore invalid filter
		})

		It("should handle edge cases in pagination", func() {
			// Page number less than 1
			resp, response := testUser1.MakeRequest(http.MethodGet, "/transaction?page=0", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			data := response["data"].(map[string]any)
			Expect(data["page"]).To(Equal(float64(1))) // Should default to page 1

			// Page size less than 1
			resp, response = testUser1.MakeRequest(http.MethodGet, "/transaction?page_size=0", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			data = response["data"].(map[string]any)
			Expect(data["page_size"]).To(Equal(float64(15))) // Should default to 15

			// Page number beyond total pages
			resp, response = testUser1.MakeRequest(http.MethodGet, "/transaction?page=100", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			data = response["data"].(map[string]any)
			Expect(data["transactions"]).To(BeEmpty())
		})
	})

	Describe("GetTransaction", func() {
		It("should get transaction by id using seed data", func() {
			url := "/transaction/1"
			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Transaction retrieved successfully"))
			Expect(response["data"]).To(HaveKey("id"))
			Expect(response["data"].(map[string]any)["name"]).To(Equal("Integration Transaction"))
		})

		It("should return error for invalid transaction id format", func() {
			url := "/transaction/invalid_id"
			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(Equal("invalid transaction id"))
		})

		It("should return error for non-existent transaction id", func() {
			url := "/transaction/9999"
			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			Expect(response["message"]).To(Equal("transaction not found"))
		})

		It("should return error when trying to access another user's transaction", func() {
			url := "/transaction/12" // Transaction belonging to user 2
			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			Expect(response["message"]).To(Equal("transaction not found"))
		})
	})

	Describe("UpdateTransaction", func() {
		Context("Input Validation", func() {
			It("should return validation error for future date in update", func() {
				update := models.UpdateBaseTransactionInput{Date: futureDate}
				url := "/transaction/3"
				resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(Equal("transaction date cannot be in the future"))
			})
		})

		It("should update transaction name using seed data", func() {
			update := models.UpdateBaseTransactionInput{Name: "Updated Transaction Name"}
			url := "/transaction/3"
			resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Transaction updated successfully"))
			Expect(response["data"].(map[string]any)["name"]).To(Equal("Updated Transaction Name"))
		})

		It("should update transaction amount using seed data", func() {
			amount := 350.99
			update := models.UpdateBaseTransactionInput{Amount: &amount}
			url := "/transaction/3"
			resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Transaction updated successfully"))
			Expect(response["data"].(map[string]any)["amount"]).To(Equal(amount))
		})

		It("should return error when trying to update transaction of different user", func() {
			url := "/transaction/12" // Transaction belonging to user 2
			input := map[string]any{
				"name":        "Updated Transaction",
				"description": "Updated Description",
				"amount":      150.00,
				"date":        testDate.Format(time.RFC3339),
				"account_id":  1,
			}
			resp, response := testUser1.MakeRequest(http.MethodPatch, url, input)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			Expect(response["message"]).To(Equal("transaction not found"))
		})

		It("should return error for invalid JSON in update", func() {
			url := "/transaction/3"
			resp, _ := testUser1.MakeRequest(http.MethodPatch, url, "{ name: invalid }")
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for empty body in update", func() {
			url := "/transaction/3"
			resp, _ := testUser1.MakeRequest(http.MethodPatch, url, "")
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for non-existent transaction id", func() {
			update := models.UpdateBaseTransactionInput{Name: "Updated Name"}
			url := "/transaction/9999"
			resp, _ := testUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("should return error for invalid transaction id format in update", func() {
			update := models.UpdateBaseTransactionInput{Name: "Updated Name"}
			url := "/transaction/invalid_id"
			resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(Equal("invalid transaction id"))
		})
	})

	Describe("UpdateTransaction Mappings", func() {
		It("should update category mapping", func() {
			update := map[string]any{
				"category_ids": []int64{2, 3},
			}
			url := "/transaction/10"
			resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			data := response["data"].(map[string]any)
			Expect(data["category_ids"]).To(ContainElements(float64(2), float64(3)))
		})

		It("should update account mapping", func() {
			update := map[string]any{
				"account_id": 2,
			}
			url := "/transaction/10"
			resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			data := response["data"].(map[string]any)
			Expect(data["account_id"]).To(Equal(float64(2)))
		})

		It("should not fail when category and account mappings are cleared", func() {
			update := map[string]any{
				"category_ids": []int64{},
			}
			url := "/transaction/10"
			resp, response := testUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			data := response["data"].(map[string]any)
			Expect(data["category_ids"]).To(BeEmpty())
		})

		It("should return error for invalid category id", func() {
			update := map[string]any{
				"category_ids": []int64{99999},
			}
			url := "/transaction/10"
			resp, _ := testUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("should return error for invalid account id", func() {
			update := map[string]any{
				"account_id": 99999,
			}
			url := "/transaction/10"
			resp, _ := testUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("should not allow adding category mapping of a different user", func() {
			// Create a category as a different user
			catInput := map[string]any{
				"name": "Other User Category",
				"icon": "other-icon",
			}
			catResp, catResponse := testUser2.MakeRequest(http.MethodPost, "/category", catInput)
			Expect(catResp.StatusCode).To(Equal(http.StatusCreated))
			otherCategoryId := int64(catResponse["data"].(map[string]any)["id"].(float64))
			update := map[string]any{
				"category_ids": []int64{otherCategoryId},
			}
			url := "/transaction/10"
			resp, _ := testUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
	})

	Describe("DeleteTransaction", func() {
		It("should delete transaction by id", func() {
			createTransactionInput := models.CreateBaseTransactionInput{
				Name:        "Test Transaction",
				Description: "Test Description",
				Amount:      floatPtr(100.00),
				Date:        testDate,
				AccountId:   3,
			}

			resp, response := testUser2.MakeRequest(http.MethodPost, "/transaction", createTransactionInput)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			createdId := int64(response["data"].(map[string]any)["id"].(float64))

			url := "/transaction/" + strconv.FormatInt(createdId, 10)
			resp, _ = testUser2.MakeRequest(http.MethodDelete, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

			resp, _ = testUser2.MakeRequest(http.MethodGet, "/transaction/"+strconv.FormatInt(createdId, 10), nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("should return error when trying to delete transaction of different user", func() {
			url := "/transaction/12" // Transaction belonging to user 2
			resp, response := testUser1.MakeRequest(http.MethodDelete, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			Expect(response["message"]).To(Equal("transaction not found"))
		})

		It("should return error for non-existent transaction id", func() {
			url := "/transaction/9999"
			resp, _ := testUser1.MakeRequest(http.MethodDelete, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("should return error for invalid transaction id format in delete", func() {
			url := "/transaction/invalid_id"
			resp, response := testUser1.MakeRequest(http.MethodDelete, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(Equal("invalid transaction id"))
		})
	})

	Describe("End to End cases", func() {
		It("should create and manipulate new transactions without conflicting with seed data", func() {
			input := models.CreateBaseTransactionInput{
				Name:        "Dynamically Created Transaction",
				Description: "Created during test",
				Amount:      floatPtr(999.99),
				Date:        testDate,
				AccountId:   3,
			}
			resp, response := testUser2.MakeRequest(http.MethodPost, "/transaction", input)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			createdId := int64(response["data"].(map[string]any)["id"].(float64))

			update := models.UpdateBaseTransactionInput{Name: "Updated Dynamic Transaction"}
			url := "/transaction/" + strconv.FormatInt(createdId, 10)
			resp, _ = testUser2.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			resp, _ = testUser2.MakeRequest(http.MethodDelete, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
		})

		It("should handle cross-user isolation properly with seed data", func() {
			// Get all transactions for user 1
			resp, response := testUser1.MakeRequest(http.MethodGet, "/transaction", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			data := response["data"].(map[string]any)
			transactions := data["transactions"].([]any)
			user1TransactionCount := len(transactions)

			// Get all transactions for user 2
			resp, response = testUser2.MakeRequest(http.MethodGet, "/transaction", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			data = response["data"].(map[string]any)
			transactions = data["transactions"].([]any)
			user2TransactionCount := len(transactions)

			// Verify counts match seed data
			Expect(user1TransactionCount).To(Equal(11))
			Expect(user2TransactionCount).To(Equal(12)) // 6 from seed data + 6 created during other tests
		})
	})
})
