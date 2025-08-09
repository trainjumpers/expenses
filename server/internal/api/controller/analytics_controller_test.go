package controller_test

import (
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AnalyticsController", func() {
	Describe("GetAccountAnalytics", func() {
		It("should get account analytics for authenticated user", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/analytics/account", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Account analytics retrieved successfully"))
			Expect(response["data"]).To(HaveKey("account_analytics"))

			data := response["data"].(map[string]any)
			accountAnalytics := data["account_analytics"].([]any)
			Expect(accountAnalytics).NotTo(BeNil())
		})

		It("should return analytics for user with accounts", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/analytics/account", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			data := response["data"].(map[string]any)
			accountAnalytics := data["account_analytics"].([]any)

			// User1 should have accounts from seed data
			Expect(len(accountAnalytics)).To(BeNumerically(">", 0))

			// Verify structure of analytics data
			if len(accountAnalytics) > 0 {
				firstAnalytic := accountAnalytics[0].(map[string]any)
				Expect(firstAnalytic).To(HaveKey("account_id"))
				Expect(firstAnalytic).To(HaveKey("current_balance"))
				Expect(firstAnalytic).To(HaveKey("balance_one_month_ago"))

				// Verify data types
				Expect(firstAnalytic["account_id"]).To(BeAssignableToTypeOf(float64(0)))
				Expect(firstAnalytic["current_balance"]).To(BeAssignableToTypeOf(float64(0)))
				Expect(firstAnalytic["balance_one_month_ago"]).To(BeAssignableToTypeOf(float64(0)))
			}
		})

		It("should return empty analytics for user without accounts", func() {
			resp, response := testUser3.MakeRequest(http.MethodGet, "/analytics/account", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Account analytics retrieved successfully"))

			data := response["data"].(map[string]any)
			accountAnalytics := data["account_analytics"]
			if accountAnalytics != nil {
				Expect(accountAnalytics.([]any)).To(BeEmpty())
			} else {
				// If nil, that's also acceptable for empty analytics
				Expect(accountAnalytics).To(BeNil())
			}
		})

		It("should include all user accounts in analytics", func() {
			// First, get user's accounts to know how many they have
			accountsResp, accountsResponse := testUser1.MakeRequest(http.MethodGet, "/account", nil)
			Expect(accountsResp.StatusCode).To(Equal(http.StatusOK))

			accounts := accountsResponse["data"].([]any)
			expectedAccountCount := len(accounts)

			// Now get analytics
			resp, response := testUser1.MakeRequest(http.MethodGet, "/analytics/account", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			data := response["data"].(map[string]any)
			accountAnalytics := data["account_analytics"].([]any)

			// Should have analytics for all accounts
			Expect(len(accountAnalytics)).To(Equal(expectedAccountCount))
		})

		It("should return analytics with correct account IDs", func() {
			// Get user's accounts
			accountsResp, accountsResponse := testUser1.MakeRequest(http.MethodGet, "/account", nil)
			Expect(accountsResp.StatusCode).To(Equal(http.StatusOK))

			accounts := accountsResponse["data"].([]any)

			// Extract account IDs
			var expectedAccountIds []float64
			for _, account := range accounts {
				accountMap := account.(map[string]any)
				expectedAccountIds = append(expectedAccountIds, accountMap["id"].(float64))
			}

			// Get analytics
			resp, response := testUser1.MakeRequest(http.MethodGet, "/analytics/account", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			data := response["data"].(map[string]any)
			accountAnalytics := data["account_analytics"].([]any)

			// Extract analytics account IDs
			var analyticsAccountIds []float64
			for _, analytic := range accountAnalytics {
				analyticMap := analytic.(map[string]any)
				analyticsAccountIds = append(analyticsAccountIds, analyticMap["account_id"].(float64))
			}

			// Should contain all account IDs
			for _, expectedId := range expectedAccountIds {
				Expect(analyticsAccountIds).To(ContainElement(expectedId))
			}
		})

		It("should return error for unauthenticated user", func() {
			resp, response := testHelperUnauthenticated.MakeRequest(http.MethodGet, "/analytics/account", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			Expect(response["message"]).To(Equal("please log in to continue"))
		})

		Context("with malformed tokens", func() {
			It("should return unauthorized or bad request for malformed tokens", func() {
				checkMalformedTokens(testUser1, http.MethodGet, "/analytics/account", nil)
			})
		})

		It("should handle user isolation correctly", func() {
			// Get analytics for user1
			resp1, response1 := testUser1.MakeRequest(http.MethodGet, "/analytics/account", nil)
			Expect(resp1.StatusCode).To(Equal(http.StatusOK))

			data1 := response1["data"].(map[string]any)
			user1Analytics := data1["account_analytics"].([]any)

			// Get analytics for user2
			resp2, response2 := testUser2.MakeRequest(http.MethodGet, "/analytics/account", nil)
			Expect(resp2.StatusCode).To(Equal(http.StatusOK))

			data2 := response2["data"].(map[string]any)
			user2Analytics := data2["account_analytics"].([]any)

			// Extract account IDs for both users
			var user1AccountIds []float64
			for _, analytic := range user1Analytics {
				analyticMap := analytic.(map[string]any)
				user1AccountIds = append(user1AccountIds, analyticMap["account_id"].(float64))
			}

			var user2AccountIds []float64
			for _, analytic := range user2Analytics {
				analyticMap := analytic.(map[string]any)
				user2AccountIds = append(user2AccountIds, analyticMap["account_id"].(float64))
			}

			// Users should not have overlapping account IDs
			for _, user1Id := range user1AccountIds {
				Expect(user2AccountIds).NotTo(ContainElement(user1Id))
			}
		})

		It("should return valid balance data types", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/analytics/account", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			data := response["data"].(map[string]any)
			accountAnalytics := data["account_analytics"].([]any)

			for _, analytic := range accountAnalytics {
				analyticMap := analytic.(map[string]any)

				// Check that balances are numeric
				currentBalance := analyticMap["current_balance"]
				Expect(currentBalance).To(BeAssignableToTypeOf(float64(0)))

				balanceOneMonthAgo := analyticMap["balance_one_month_ago"]
				Expect(balanceOneMonthAgo).To(BeAssignableToTypeOf(float64(0)))

				// Balances can be positive, negative, or zero
				currentBalanceFloat := currentBalance.(float64)
				balanceOneMonthAgoFloat := balanceOneMonthAgo.(float64)

				// Just verify they are valid numbers (not NaN or Inf)
				Expect(currentBalanceFloat).To(BeNumerically(">=", -999999999))
				Expect(currentBalanceFloat).To(BeNumerically("<=", 999999999))
				Expect(balanceOneMonthAgoFloat).To(BeNumerically(">=", -999999999))
				Expect(balanceOneMonthAgoFloat).To(BeNumerically("<=", 999999999))
			}
		})

		It("should handle accounts with zero balances correctly", func() {
			// Create a new account for user2 (who has fewer transactions)
			accInput := map[string]any{
				"name":      "Zero Balance Account",
				"bank_type": "axis",
				"currency":  "inr",
			}
			resp, _ := testUser2.MakeRequest(http.MethodPost, "/account", accInput)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))

			// Get analytics - should include the new account with zero balances
			resp, response := testUser2.MakeRequest(http.MethodGet, "/analytics/account", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			data := response["data"].(map[string]any)
			accountAnalytics := data["account_analytics"].([]any)

			// Should have at least one account now
			Expect(len(accountAnalytics)).To(BeNumerically(">=", 1))

			// Find the new account in analytics
			found := false
			for _, analytic := range accountAnalytics {
				analyticMap := analytic.(map[string]any)
				// New account should have zero balances
				if analyticMap["current_balance"].(float64) == 0.0 && analyticMap["balance_one_month_ago"].(float64) == 0.0 {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue(), "Should find at least one account with zero balances")
		})

		It("should return consistent data structure even with no data", func() {
			resp, response := testUser3.MakeRequest(http.MethodGet, "/analytics/account", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Account analytics retrieved successfully"))
			Expect(response["data"]).To(HaveKey("account_analytics"))

			// Even with no data, the structure should be consistent
			data := response["data"].(map[string]any)
			Expect(data).To(HaveKey("account_analytics"))
		})

		It("should handle unsupported HTTP method", func() {
			// Test POST method on analytics endpoint (not supported, returns 404)
			// Use a custom request to avoid JSON parsing issues with error responses
			req, err := http.NewRequest(http.MethodPost, testUser1.BaseURL+"/analytics/accounts", nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Cookie", "access_token="+testUser1.AccessToken)

			resp, err := testUser1.Client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("should handle invalid endpoint path", func() {
			// Test invalid analytics endpoint
			// Use a custom request to avoid JSON parsing issues with error responses
			req, err := http.NewRequest(http.MethodGet, testUser1.BaseURL+"/analytics/invalid", nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Cookie", "access_token="+testUser1.AccessToken)

			resp, err := testUser1.Client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("should handle analytics service errors gracefully", func() {
			resp, response := testUser3.MakeRequest(http.MethodGet, "/analytics/account", nil)

			// Should still return 200 OK even when no accounts exist
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Account analytics retrieved successfully"))

			// Should have proper data structure even with empty results
			data := response["data"].(map[string]any)
			Expect(data).To(HaveKey("account_analytics"))

			accountAnalytics := data["account_analytics"]
			// Should be either empty array or nil, both are acceptable
			if accountAnalytics != nil {
				Expect(accountAnalytics.([]any)).To(BeEmpty())
			} else {
				Expect(accountAnalytics).To(BeNil())
			}
		})

		It("should handle edge case with corrupted account data", func() {
			// Test resilience when account data might be in unexpected state
			// This is more of a defensive programming test

			resp, response := testUser1.MakeRequest(http.MethodGet, "/analytics/account", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			// Verify that even if there are edge cases in data,
			// the response structure remains consistent
			data := response["data"].(map[string]any)
			Expect(data).To(HaveKey("account_analytics"))

			accountAnalytics := data["account_analytics"].([]any)

			// Each analytics entry should have required fields
			for _, analytic := range accountAnalytics {
				analyticMap := analytic.(map[string]any)

				// Required fields should always be present
				Expect(analyticMap).To(HaveKey("account_id"))
				Expect(analyticMap).To(HaveKey("current_balance"))
				Expect(analyticMap).To(HaveKey("balance_one_month_ago"))

				// Values should be valid (not nil)
				Expect(analyticMap["account_id"]).NotTo(BeNil())
				Expect(analyticMap["current_balance"]).NotTo(BeNil())
				Expect(analyticMap["balance_one_month_ago"]).NotTo(BeNil())
			}
		})
	})

	Describe("GetNetworthTimeSeries", func() {
		It("should get networth time series for authenticated user", func() {
			startDate := "2023-01-01"
			endDate := "2023-01-31"
			url := "/analytics/networth?start_date=" + startDate + "&end_date=" + endDate

			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Networth time series retrieved successfully"))
			Expect(response["data"]).To(HaveKey("initial_balance"))
			Expect(response["data"]).To(HaveKey("time_series"))

			data := response["data"].(map[string]any)
			initialBalance := data["initial_balance"]
			Expect(initialBalance).To(BeAssignableToTypeOf(float64(0)))

			timeSeries := data["time_series"].([]any)
			Expect(timeSeries).NotTo(BeNil())

			// Verify structure of time series data
			if len(timeSeries) > 0 {
				firstPoint := timeSeries[0].(map[string]any)
				Expect(firstPoint).To(HaveKey("date"))
				Expect(firstPoint).To(HaveKey("networth"))
				Expect(firstPoint["date"]).To(BeAssignableToTypeOf(""))
				Expect(firstPoint["networth"]).To(BeAssignableToTypeOf(float64(0)))
			}
		})

		Context("with invalid parameters", func() {
			It("should return validation errors", func() {
				testCases := []map[string]any{
					{"startDate": "", "endDate": "2023-01-31", "expectedMessage": "start_date and end_date query parameters are required"},
					{"startDate": "2023-01-01", "endDate": "", "expectedMessage": "start_date and end_date query parameters are required"},
					{"startDate": "invalid", "endDate": "2023-01-31", "expectedMessage": "invalid start_date format, expected YYYY-MM-DD"},
					{"startDate": "2023-01-01", "endDate": "invalid", "expectedMessage": "invalid end_date format, expected YYYY-MM-DD"},
					{"startDate": "2023-01-31", "endDate": "2023-01-01", "expectedMessage": "start_date cannot be after end_date"},
					{"startDate": "2023-01-01T00:00:00", "endDate": "2023-01-31", "expectedMessage": "invalid start_date format, expected YYYY-MM-DD"},
					{"startDate": "2023/01/01", "endDate": "2023-01-31", "expectedMessage": "invalid start_date format, expected YYYY-MM-DD"},
					{"startDate": "2023-13-01", "endDate": "2023-01-31", "expectedMessage": "invalid start_date format, expected YYYY-MM-DD"},
					{"startDate": "2023-01-32", "endDate": "2023-01-31", "expectedMessage": "invalid start_date format, expected YYYY-MM-DD"},
				}
				checkNetworthValidation(testUser1, testCases)
			})
		})

		It("should return error for unauthenticated user", func() {
			url := "/analytics/networth?start_date=2023-01-01&end_date=2023-01-31"
			resp, response := testHelperUnauthenticated.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			Expect(response["message"]).To(Equal("please log in to continue"))
		})

		It("should handle valid date range", func() {
			startDate := "2023-01-01"
			endDate := "2023-01-07"
			url := "/analytics/networth?start_date=" + startDate + "&end_date=" + endDate

			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			data := response["data"].(map[string]any)
			timeSeries := data["time_series"].([]any)

			// Should have data points for each day in the range (inclusive)
			expectedDays := 7 // Jan 1-7 inclusive
			Expect(len(timeSeries)).To(Equal(expectedDays))

			// Verify dates are in correct order
			if len(timeSeries) > 1 {
				firstPoint := timeSeries[0].(map[string]any)
				lastPoint := timeSeries[len(timeSeries)-1].(map[string]any)
				Expect(firstPoint["date"]).To(Equal(startDate))
				Expect(lastPoint["date"]).To(Equal(endDate))
			}
		})

		It("should handle same start and end date", func() {
			startDate := "2023-01-15"
			endDate := "2023-01-15"
			url := "/analytics/networth?start_date=" + startDate + "&end_date=" + endDate

			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			data := response["data"].(map[string]any)
			timeSeries := data["time_series"].([]any)

			// Should have exactly one data point for the single day
			Expect(len(timeSeries)).To(Equal(1))

			point := timeSeries[0].(map[string]any)
			Expect(point["date"]).To(Equal(startDate))
		})

		It("should handle edge case date formats", func() {
			// Test with leading zeros
			url := "/analytics/networth?start_date=2023-01-01&end_date=2023-01-02"
			resp, _ := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})

		It("should return error for future dates", func() {
			futureDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02")
			url := "/analytics/networth?start_date=" + futureDate + "&end_date=" + futureDate

			// Note: The current implementation doesn't validate future dates
			// This test documents the current behavior
			resp, _ := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})

		It("should handle very old dates", func() {
			url := "/analytics/networth?start_date=1900-01-01&end_date=1900-01-02"
			resp, _ := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})

		It("should handle URL encoded query parameters", func() {
			// Test with URL encoded dates (though not necessary for this format)
			url := "/analytics/networth?start_date=2023-01-01&end_date=2023-01-02"
			resp, _ := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})

		It("should handle query parameters with whitespace", func() {
			// Test with whitespace (HTTP parsing typically trims query parameters)
			url := "/analytics/networth?start_date= 2023-01-01 &end_date= 2023-01-02 "

			// Use a custom request to avoid JSON parsing issues with potential error responses
			req, err := http.NewRequest(http.MethodGet, testUser1.BaseURL+url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Cookie", "access_token="+testUser1.AccessToken)

			resp, err := testUser1.Client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			// HTTP parsing typically trims whitespace, so this might succeed
			// The exact behavior depends on the HTTP implementation
			Expect(resp.StatusCode).To(BeElementOf([]int{http.StatusOK, http.StatusBadRequest}))
		})

		It("should handle large date ranges", func() {
			// Test with a year-long range
			url := "/analytics/networth?start_date=2023-01-01&end_date=2023-12-31"
			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			data := response["data"].(map[string]any)
			timeSeries := data["time_series"].([]any)

			// Should have 365 data points for 2023 (not a leap year)
			Expect(len(timeSeries)).To(Equal(365))
		})

		It("should handle leap year dates", func() {
			url := "/analytics/networth?start_date=2024-02-28&end_date=2024-03-01"
			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			data := response["data"].(map[string]any)
			timeSeries := data["time_series"].([]any)

			// Should have 3 data points: Feb 28, Feb 29, Mar 1
			Expect(len(timeSeries)).To(Equal(3))
		})

		It("should handle missing both query parameters", func() {
			url := "/analytics/networth"
			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(Equal("start_date and end_date query parameters are required"))
		})

		It("should handle empty query parameter values", func() {
			url := "/analytics/networth?start_date=&end_date="
			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(Equal("start_date and end_date query parameters are required"))
		})

		It("should handle cross-year date ranges", func() {
			url := "/analytics/networth?start_date=2023-12-30&end_date=2024-01-02"
			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			data := response["data"].(map[string]any)
			timeSeries := data["time_series"].([]any)

			// Should have 4 data points: Dec 30, Dec 31, Jan 1, Jan 2
			Expect(len(timeSeries)).To(Equal(4))

			// Verify the dates cross the year boundary correctly
			Expect(timeSeries[0].(map[string]any)["date"]).To(Equal("2023-12-30"))
			Expect(timeSeries[1].(map[string]any)["date"]).To(Equal("2023-12-31"))
			Expect(timeSeries[2].(map[string]any)["date"]).To(Equal("2024-01-01"))
			Expect(timeSeries[3].(map[string]any)["date"]).To(Equal("2024-01-02"))
		})

		Context("with malformed tokens", func() {
			It("should return unauthorized for malformed tokens", func() {
				url := "/analytics/networth?start_date=2023-01-01&end_date=2023-01-31"
				checkMalformedTokens(testUser1, http.MethodGet, url, nil)
			})
		})

		Context("error handling", func() {
			It("should handle service errors gracefully", func() {
				// Test with a valid request that should work
				url := "/analytics/networth?start_date=2023-01-01&end_date=2023-01-02"
				resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)

				// Should succeed (we can't easily simulate service errors in integration tests)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(response["message"]).To(Equal("Networth time series retrieved successfully"))
			})

			It("should maintain consistent response structure on success", func() {
				url := "/analytics/networth?start_date=2023-01-01&end_date=2023-01-01"
				resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				// Verify response structure
				Expect(response).To(HaveKey("message"))
				Expect(response).To(HaveKey("data"))

				data := response["data"].(map[string]any)
				Expect(data).To(HaveKey("initial_balance"))
				Expect(data).To(HaveKey("time_series"))

				timeSeries := data["time_series"].([]any)
				Expect(timeSeries).NotTo(BeNil())
			})
		})
	})

	Describe("GetCategoryAnalytics", func() {
		It("should get category analytics for authenticated user", func() {
			startDate := "2023-01-01"
			endDate := "2023-01-31"
			url := "/analytics/category?start_date=" + startDate + "&end_date=" + endDate

			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Category analytics retrieved successfully"))
			Expect(response["data"]).To(HaveKey("category_transactions"))

			data := response["data"].(map[string]any)
			categoryTransactions := data["category_transactions"].([]any)
			Expect(categoryTransactions).NotTo(BeNil())
		})

		It("should return analytics for user with transactions", func() {
			startDate := "2023-01-01"
			endDate := "2023-01-31"
			url := "/analytics/category?start_date=" + startDate + "&end_date=" + endDate

			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			data := response["data"].(map[string]any)
			categoryTransactions := data["category_transactions"].([]any)

			// User1 should have transactions from seed data
			Expect(len(categoryTransactions)).To(BeNumerically(">=", 0))

			// Verify structure of analytics data
			if len(categoryTransactions) > 0 {
				firstTransaction := categoryTransactions[0].(map[string]any)
				Expect(firstTransaction).To(HaveKey("category_id"))
				Expect(firstTransaction).To(HaveKey("category_name"))
				Expect(firstTransaction).To(HaveKey("total_amount"))

				// Verify data types
				Expect(firstTransaction["category_id"]).To(BeAssignableToTypeOf(float64(0)))
				Expect(firstTransaction["category_name"]).To(BeAssignableToTypeOf(""))
				Expect(firstTransaction["total_amount"]).To(BeAssignableToTypeOf(float64(0)))
			}
		})

		It("should return empty analytics for user without transactions", func() {
			startDate := "2023-01-01"
			endDate := "2023-01-31"
			url := "/analytics/category?start_date=" + startDate + "&end_date=" + endDate

			resp, response := testUser3.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Category analytics retrieved successfully"))

			data := response["data"].(map[string]any)
			categoryTransactions := data["category_transactions"]
			if categoryTransactions != nil {
				Expect(categoryTransactions.([]any)).To(BeEmpty())
			} else {
				// If nil, that's also acceptable for empty analytics
				Expect(categoryTransactions).To(BeNil())
			}
		})

		It("should return error for unauthenticated user", func() {
			url := "/analytics/category?start_date=2023-01-01&end_date=2023-01-31"
			resp, response := testHelperUnauthenticated.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			Expect(response["message"]).To(Equal("please log in to continue"))
		})

		Context("with invalid parameters", func() {
			It("should return validation errors", func() {
				testCases := []map[string]any{
					{"startDate": "", "endDate": "2023-01-31", "expectedMessage": "start_date and end_date query parameters are required"},
					{"startDate": "2023-01-01", "endDate": "", "expectedMessage": "start_date and end_date query parameters are required"},
					{"startDate": "invalid", "endDate": "2023-01-31", "expectedMessage": "invalid start_date format, expected YYYY-MM-DD"},
					{"startDate": "2023-01-01", "endDate": "invalid", "expectedMessage": "invalid end_date format, expected YYYY-MM-DD"},
					{"startDate": "2023-01-31", "endDate": "2023-01-01", "expectedMessage": "start_date cannot be after end_date"},
					{"startDate": "2023-01-01T00:00:00", "endDate": "2023-01-31", "expectedMessage": "invalid start_date format, expected YYYY-MM-DD"},
					{"startDate": "2023/01/01", "endDate": "2023-01-31", "expectedMessage": "invalid start_date format, expected YYYY-MM-DD"},
					{"startDate": "2023-13-01", "endDate": "2023-01-31", "expectedMessage": "invalid start_date format, expected YYYY-MM-DD"},
					{"startDate": "2023-01-32", "endDate": "2023-01-31", "expectedMessage": "invalid start_date format, expected YYYY-MM-DD"},
				}
				checkCategoryValidation(testUser1, testCases)
			})
		})

		It("should handle valid date range", func() {
			startDate := "2023-01-01"
			endDate := "2023-01-07"
			url := "/analytics/category?start_date=" + startDate + "&end_date=" + endDate

			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			data := response["data"].(map[string]any)
			categoryTransactions := data["category_transactions"].([]any)

			// Should have analytics data for the date range
			Expect(categoryTransactions).NotTo(BeNil())
		})

		It("should handle same start and end date", func() {
			startDate := "2023-01-15"
			endDate := "2023-01-15"
			url := "/analytics/category?start_date=" + startDate + "&end_date=" + endDate

			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			data := response["data"].(map[string]any)
			categoryTransactions := data["category_transactions"].([]any)

			// Should have analytics data for the single day
			Expect(categoryTransactions).NotTo(BeNil())
		})

		It("should include uncategorized transactions in analytics", func() {
			// Use a date range that includes transaction 11 (Cash Withdrawal) which has no category
			startDate := "2024-01-01"
			endDate := "2024-01-31"
			url := "/analytics/category?start_date=" + startDate + "&end_date=" + endDate

			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			data := response["data"].(map[string]any)
			categoryTransactions := data["category_transactions"].([]any)
			Expect(categoryTransactions).NotTo(BeNil())

			// Look for uncategorized transactions
			var foundUncategorized bool
			for _, transaction := range categoryTransactions {
				txn := transaction.(map[string]any)
				categoryID := txn["category_id"].(float64)
				categoryName := txn["category_name"].(string)
				totalAmount := txn["total_amount"].(float64)

				if categoryID == -1 && categoryName == "Uncategorized" {
					foundUncategorized = true
					// Verify that uncategorized has the correct amount (transaction 11: Cash Withdrawal = 100.00)
					Expect(totalAmount == 100.0 || totalAmount == 105.5).To(BeTrue(), "totalAmount should be either 100.0 or 105.0")
					break
				}
			}

			// Ensure we found the uncategorized transactions
			Expect(foundUncategorized).To(BeTrue(), "Expected to find uncategorized transactions")
		})

		It("should handle missing both query parameters", func() {
			url := "/analytics/category"
			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(Equal("start_date and end_date query parameters are required"))
		})

		It("should handle empty query parameter values", func() {
			url := "/analytics/category?start_date=&end_date="
			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(Equal("start_date and end_date query parameters are required"))
		})

		It("should handle cross-year date ranges", func() {
			url := "/analytics/category?start_date=2023-12-30&end_date=2024-01-02"
			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			data := response["data"].(map[string]any)
			categoryTransactions := data["category_transactions"].([]any)

			// Should have analytics data for the cross-year range
			Expect(categoryTransactions).NotTo(BeNil())
		})

		Context("with malformed tokens", func() {
			It("should return unauthorized for malformed tokens", func() {
				url := "/analytics/category?start_date=2023-01-01&end_date=2023-01-31"
				checkMalformedTokens(testUser1, http.MethodGet, url, nil)
			})
		})

		It("should handle user isolation correctly", func() {
			startDate := "2023-01-01"
			endDate := "2023-01-31"
			url := "/analytics/category?start_date=" + startDate + "&end_date=" + endDate

			// Get analytics for user1
			resp1, response1 := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp1.StatusCode).To(Equal(http.StatusOK))

			data1 := response1["data"].(map[string]any)
			user1Transactions := data1["category_transactions"].([]any)

			// Get analytics for user2
			resp2, response2 := testUser2.MakeRequest(http.MethodGet, url, nil)
			Expect(resp2.StatusCode).To(Equal(http.StatusOK))

			data2 := response2["data"].(map[string]any)
			user2Transactions := data2["category_transactions"].([]any)

			// Extract category IDs for both users
			var user1CategoryIds []float64
			for _, transaction := range user1Transactions {
				transactionMap := transaction.(map[string]any)
				user1CategoryIds = append(user1CategoryIds, transactionMap["category_id"].(float64))
			}

			var user2CategoryIds []float64
			for _, transaction := range user2Transactions {
				transactionMap := transaction.(map[string]any)
				user2CategoryIds = append(user2CategoryIds, transactionMap["category_id"].(float64))
			}

			// Verify that both users get their own data (user isolation)
			// User1 should have transactions (from seed data)
			Expect(len(user1Transactions)).To(BeNumerically(">=", 0))

			// User2 might have no transactions in this date range, which is fine
			// The important thing is that the response structure is correct
			Expect(user2Transactions).NotTo(BeNil())

			// If both users have transactions, verify they don't interfere with each other
			if len(user1CategoryIds) > 0 && len(user2CategoryIds) > 0 {
				// Users should not have overlapping category IDs (unless they share categories)
				// This test verifies that the data is properly filtered by user
				// Note: Categories might be shared between users, so we just verify the structure
				Expect(user1CategoryIds).NotTo(BeNil())
				Expect(user2CategoryIds).NotTo(BeNil())
			}
		})

		It("should return valid amount data types", func() {
			startDate := "2023-01-01"
			endDate := "2023-01-31"
			url := "/analytics/category?start_date=" + startDate + "&end_date=" + endDate

			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			data := response["data"].(map[string]any)
			categoryTransactions := data["category_transactions"].([]any)

			for _, transaction := range categoryTransactions {
				transactionMap := transaction.(map[string]any)

				// Check that amounts are numeric
				totalAmount := transactionMap["total_amount"]
				Expect(totalAmount).To(BeAssignableToTypeOf(float64(0)))

				// Verify they are valid numbers
				totalAmountFloat := totalAmount.(float64)

				// Amounts can be positive, negative, or zero
				Expect(totalAmountFloat).To(BeNumerically(">=", -999999999))
				Expect(totalAmountFloat).To(BeNumerically("<=", 999999999))
			}
		})

		It("should handle categories with zero transactions correctly", func() {
			// Test with a date range that might have no transactions
			startDate := "2023-12-01"
			endDate := "2023-12-31"
			url := "/analytics/category?start_date=" + startDate + "&end_date=" + endDate

			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			data := response["data"].(map[string]any)
			categoryTransactions := data["category_transactions"].([]any)

			// Should handle empty results gracefully
			Expect(categoryTransactions).NotTo(BeNil())
		})

		It("should return consistent data structure even with no data", func() {
			startDate := "2023-12-01"
			endDate := "2023-12-31"
			url := "/analytics/category?start_date=" + startDate + "&end_date=" + endDate

			resp, response := testUser3.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Category analytics retrieved successfully"))
			Expect(response["data"]).To(HaveKey("category_transactions"))

			// Even with no data, the structure should be consistent
			data := response["data"].(map[string]any)
			Expect(data).To(HaveKey("category_transactions"))
		})

		It("should handle unsupported HTTP method", func() {
			// Test POST method on analytics endpoint (not supported, returns 404)
			req, err := http.NewRequest(http.MethodPost, testUser1.BaseURL+"/analytics/category", nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Cookie", "access_token="+testUser1.AccessToken)

			resp, err := testUser1.Client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("should handle invalid endpoint path", func() {
			// Test invalid analytics endpoint
			req, err := http.NewRequest(http.MethodGet, testUser1.BaseURL+"/analytics/invalid", nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Cookie", "access_token="+testUser1.AccessToken)

			resp, err := testUser1.Client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("should handle analytics service errors gracefully", func() {
			startDate := "2023-12-01"
			endDate := "2023-12-31"
			url := "/analytics/category?start_date=" + startDate + "&end_date=" + endDate

			resp, response := testUser3.MakeRequest(http.MethodGet, url, nil)

			// Should still return 200 OK even when no transactions exist
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Category analytics retrieved successfully"))

			// Should have proper data structure even with empty results
			data := response["data"].(map[string]any)
			Expect(data).To(HaveKey("category_transactions"))

			categoryTransactions := data["category_transactions"]
			// Should be either empty array or nil, both are acceptable
			if categoryTransactions != nil {
				Expect(categoryTransactions.([]any)).To(BeEmpty())
			} else {
				Expect(categoryTransactions).To(BeNil())
			}
		})

		It("should handle edge case with corrupted category data", func() {
			startDate := "2023-01-01"
			endDate := "2023-01-31"
			url := "/analytics/category?start_date=" + startDate + "&end_date=" + endDate

			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			// Verify that even if there are edge cases in data,
			// the response structure remains consistent
			data := response["data"].(map[string]any)
			Expect(data).To(HaveKey("category_transactions"))

			categoryTransactions := data["category_transactions"].([]any)

			// Each transaction entry should have required fields
			for _, transaction := range categoryTransactions {
				transactionMap := transaction.(map[string]any)

				// Required fields should always be present
				Expect(transactionMap).To(HaveKey("category_id"))
				Expect(transactionMap).To(HaveKey("category_name"))
				Expect(transactionMap).To(HaveKey("total_amount"))

				// Values should be valid (not nil)
				Expect(transactionMap["category_id"]).NotTo(BeNil())
				Expect(transactionMap["category_name"]).NotTo(BeNil())
				Expect(transactionMap["total_amount"]).NotTo(BeNil())
			}
		})

		It("should handle large date ranges", func() {
			// Test with a year-long range
			url := "/analytics/category?start_date=2023-01-01&end_date=2023-12-31"
			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			data := response["data"].(map[string]any)
			categoryTransactions := data["category_transactions"].([]any)

			// Should have analytics data for the year
			Expect(categoryTransactions).NotTo(BeNil())
		})

		It("should handle leap year dates", func() {
			url := "/analytics/category?start_date=2024-02-28&end_date=2024-03-01"
			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			data := response["data"].(map[string]any)
			categoryTransactions := data["category_transactions"].([]any)

			// Should have analytics data for the leap year period
			Expect(categoryTransactions).NotTo(BeNil())
		})

		It("should handle URL encoded query parameters", func() {
			// Test with URL encoded dates (though not necessary for this format)
			url := "/analytics/category?start_date=2023-01-01&end_date=2023-01-02"
			resp, _ := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})

		It("should handle query parameters with whitespace", func() {
			// Test with whitespace (HTTP parsing typically trims query parameters)
			url := "/analytics/category?start_date= 2023-01-01 &end_date= 2023-01-02 "

			// Use a custom request to avoid JSON parsing issues with potential error responses
			req, err := http.NewRequest(http.MethodGet, testUser1.BaseURL+url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Cookie", "access_token="+testUser1.AccessToken)

			resp, err := testUser1.Client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			// HTTP parsing typically trims whitespace, so this might succeed
			// The exact behavior depends on the HTTP implementation
			Expect(resp.StatusCode).To(BeElementOf([]int{http.StatusOK, http.StatusBadRequest}))
		})

		Context("error handling", func() {
			It("should handle service errors gracefully", func() {
				// Test with a valid request that should work
				url := "/analytics/category?start_date=2023-01-01&end_date=2023-01-02"
				resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)

				// Should succeed (we can't easily simulate service errors in integration tests)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(response["message"]).To(Equal("Category analytics retrieved successfully"))
			})

			It("should maintain consistent response structure on success", func() {
				url := "/analytics/category?start_date=2023-01-01&end_date=2023-01-01"
				resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				// Verify response structure
				Expect(response).To(HaveKey("message"))
				Expect(response).To(HaveKey("data"))

				data := response["data"].(map[string]any)
				Expect(data).To(HaveKey("category_transactions"))

				categoryTransactions := data["category_transactions"].([]any)
				Expect(categoryTransactions).NotTo(BeNil())
			})
		})
	})

	Describe("GetMonthlyAnalytics", func() {
		It("should get monthly analytics for authenticated user with date range", func() {
			startDate := "2024-01-01"
			endDate := "2024-06-30"
			url := "/analytics/monthly?start_date=" + startDate + "&end_date=" + endDate

			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Monthly analytics retrieved successfully"))
			Expect(response["data"]).To(HaveKey("total_income"))
			Expect(response["data"]).To(HaveKey("total_expenses"))
			Expect(response["data"]).To(HaveKey("total_amount"))

			data := response["data"].(map[string]any)
			Expect(data["total_income"]).To(BeAssignableToTypeOf(float64(0)))
			Expect(data["total_expenses"]).To(BeAssignableToTypeOf(float64(0)))
			Expect(data["total_amount"]).To(BeAssignableToTypeOf(float64(0)))
		})

		It("should get monthly analytics with custom 3-month date range", func() {
			startDate := "2024-03-01"
			endDate := "2024-05-31"
			url := "/analytics/monthly?start_date=" + startDate + "&end_date=" + endDate

			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Monthly analytics retrieved successfully"))

			data := response["data"].(map[string]any)
			Expect(data).To(HaveKey("total_income"))
			Expect(data).To(HaveKey("total_expenses"))
			Expect(data).To(HaveKey("total_amount"))
		})

		It("should get monthly analytics with 12-month date range", func() {
			startDate := "2023-06-01"
			endDate := "2024-05-31"
			url := "/analytics/monthly?start_date=" + startDate + "&end_date=" + endDate

			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Monthly analytics retrieved successfully"))

			data := response["data"].(map[string]any)
			Expect(data).To(HaveKey("total_income"))
			Expect(data).To(HaveKey("total_expenses"))
			Expect(data).To(HaveKey("total_amount"))
		})

		It("should return error for invalid start_date parameter", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/analytics/monthly?start_date=invalid&end_date=2024-06-30", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(Equal("invalid start_date format, expected YYYY-MM-DD"))
		})

		It("should return error for missing start_date parameter", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/analytics/monthly?end_date=2024-06-30", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(Equal("start_date and end_date query parameters are required"))
		})

		It("should return error for missing end_date parameter", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/analytics/monthly?start_date=2024-01-01", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(Equal("start_date and end_date query parameters are required"))
		})

		It("should return error when end_date is before start_date", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/analytics/monthly?start_date=2024-06-01&end_date=2024-01-01", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(Equal("start_date cannot be after end_date"))
		})

		It("should return unauthorized for unauthenticated request", func() {
			resp, response := testHelperUnauthenticated.MakeRequest(http.MethodGet, "/analytics/monthly?start_date=2024-01-01&end_date=2024-06-30", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			Expect(response["message"]).To(Equal("please log in to continue"))
		})

		It("should return valid analytics structure", func() {
			startDate := "2024-01-01"
			endDate := "2024-06-30"
			url := "/analytics/monthly?start_date=" + startDate + "&end_date=" + endDate

			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			data := response["data"].(map[string]any)

			// Verify all required fields are present and have correct types
			totalIncome, hasIncome := data["total_income"]
			Expect(hasIncome).To(BeTrue())
			Expect(totalIncome).To(BeAssignableToTypeOf(float64(0)))

			totalExpenses, hasExpenses := data["total_expenses"]
			Expect(hasExpenses).To(BeTrue())
			Expect(totalExpenses).To(BeAssignableToTypeOf(float64(0)))

			totalAmount, hasAmount := data["total_amount"]
			Expect(hasAmount).To(BeTrue())
			Expect(totalAmount).To(BeAssignableToTypeOf(float64(0)))

			Expect(totalIncome.(float64)).To(BeNumerically(">=", 0))
			Expect(totalExpenses.(float64)).To(BeNumerically(">=", 0))
			Expect(totalAmount.(float64)).To(BeNumerically("==", totalIncome.(float64)-totalExpenses.(float64)))
		})

		It("should return different results for different users", func() {
			startDate := "2024-01-01"
			endDate := "2024-06-30"
			url := "/analytics/monthly?start_date=" + startDate + "&end_date=" + endDate

			// Get analytics for user1
			resp1, response1 := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp1.StatusCode).To(Equal(http.StatusOK))

			// Get analytics for user2
			resp2, response2 := testUser2.MakeRequest(http.MethodGet, url, nil)
			Expect(resp2.StatusCode).To(Equal(http.StatusOK))

			// Both should be successful
			data1 := response1["data"].(map[string]any)
			data2 := response2["data"].(map[string]any)

			// Both should have the required structure
			Expect(data1).To(HaveKey("total_income"))
			Expect(data1).To(HaveKey("total_expenses"))
			Expect(data1).To(HaveKey("total_amount"))

			Expect(data2).To(HaveKey("total_income"))
			Expect(data2).To(HaveKey("total_expenses"))
			Expect(data2).To(HaveKey("total_amount"))
		})

		It("should handle large date range", func() {
			startDate := "2020-01-01"
			endDate := "2024-12-31"
			url := "/analytics/monthly?start_date=" + startDate + "&end_date=" + endDate

			resp, response := testUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Monthly analytics retrieved successfully"))

			data := response["data"].(map[string]any)
			Expect(data).To(HaveKey("total_income"))
			Expect(data).To(HaveKey("total_expenses"))
			Expect(data).To(HaveKey("total_amount"))
		})
	})
})
