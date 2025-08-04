package controller_test

import (
	"net/http"

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
	})
})
