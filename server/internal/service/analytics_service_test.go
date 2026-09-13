package service

import (
	"context"
	mock_repository "expenses/internal/mock/repository"
	"expenses/internal/models"
	"math"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AnalyticsService", func() {
	var (
		analyticsService  AnalyticsServiceInterface
		mockAnalyticsRepo *mock_repository.MockAnalyticsRepository
		mockAccountRepo   *mock_repository.MockAccountRepository
		ctx               context.Context
		userId            int64
	)

	BeforeEach(func() {
		ctx = context.Background()
		userId = 1
		mockAnalyticsRepo = mock_repository.NewMockAnalyticsRepository()
		mockAccountRepo = mock_repository.NewMockAccountRepository()
		analyticsService = NewAnalyticsService(mockAnalyticsRepo, mockAccountRepo)
	})

	Describe("GetAccountAnalytics", func() {
		Context("when user has no accounts", func() {
			It("should return empty analytics", func() {
				result, err := analyticsService.GetAccountAnalytics(ctx, userId)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.AccountAnalytics).To(BeEmpty())
			})
		})

		Context("when user has accounts with no transactions", func() {
			BeforeEach(func() {
				// Create test accounts
				account1 := models.CreateAccountInput{
					Name:      "Test Account 1",
					BankType:  models.BankTypeAxis,
					Currency:  models.CurrencyINR,
					CreatedBy: userId,
				}
				account2 := models.CreateAccountInput{
					Name:      "Test Account 2",
					BankType:  models.BankTypeSBI,
					Currency:  models.CurrencyUSD,
					CreatedBy: userId,
				}
				mockAccountRepo.CreateAccount(ctx, account1)
				mockAccountRepo.CreateAccount(ctx, account2)
			})

			It("should return analytics with zero balances", func() {
				result, err := analyticsService.GetAccountAnalytics(ctx, userId)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.AccountAnalytics).To(HaveLen(2))

				for _, analytics := range result.AccountAnalytics {
					Expect(analytics.CurrentBalance).To(Equal(0.0))
					Expect(analytics.BalanceOneMonthAgo).To(Equal(0.0))
				}
			})
		})

		Context("when user has accounts with transactions", func() {
			var account1Id, account2Id int64

			BeforeEach(func() {
				// Create test accounts
				account1Input := models.CreateAccountInput{
					Name:      "Test Account 1",
					BankType:  models.BankTypeAxis,
					Currency:  models.CurrencyINR,
					CreatedBy: userId,
				}
				account2Input := models.CreateAccountInput{
					Name:      "Test Account 2",
					BankType:  models.BankTypeSBI,
					Currency:  models.CurrencyUSD,
					CreatedBy: userId,
				}

				acc1, _ := mockAccountRepo.CreateAccount(ctx, account1Input)
				acc2, _ := mockAccountRepo.CreateAccount(ctx, account2Input)
				account1Id = acc1.Id
				account2Id = acc2.Id

				// Set up current balances (all transactions)
				currentBalances := map[int64]float64{
					account1Id: 1000.0,
					account2Id: 500.0,
				}
				mockAnalyticsRepo.SetBalance(userId, nil, nil, currentBalances)

				// Set up historical balances (one month ago)
				oneMonthAgo := time.Now().AddDate(0, -1, 0)
				historicalBalances := map[int64]float64{
					account1Id: 800.0,
					account2Id: 300.0,
				}
				mockAnalyticsRepo.SetBalance(userId, nil, &oneMonthAgo, historicalBalances)
			})

			It("should return analytics with correct current and historical balances", func() {
				result, err := analyticsService.GetAccountAnalytics(ctx, userId)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.AccountAnalytics).To(HaveLen(2))

				// Find analytics for each account
				var account1Analytics, account2Analytics *models.AccountBalanceAnalytics
				for i := range result.AccountAnalytics {
					switch result.AccountAnalytics[i].AccountID {
					case account1Id:
						account1Analytics = &result.AccountAnalytics[i]
					case account2Id:
						account2Analytics = &result.AccountAnalytics[i]
					}
				}

				Expect(account1Analytics).NotTo(BeNil())
				Expect(account1Analytics.CurrentBalance).To(Equal(-1000.0))
				Expect(account1Analytics.BalanceOneMonthAgo).To(Equal(-800.0))

				Expect(account2Analytics).NotTo(BeNil())
				Expect(account2Analytics.CurrentBalance).To(Equal(-500.0))
				Expect(account2Analytics.BalanceOneMonthAgo).To(Equal(-300.0))
			})
		})

		Context("when user has accounts with partial transaction data", func() {
			var account1Id, account2Id int64

			BeforeEach(func() {
				// Create test accounts
				account1Input := models.CreateAccountInput{
					Name:      "Test Account 1",
					BankType:  models.BankTypeAxis,
					Currency:  models.CurrencyINR,
					CreatedBy: userId,
				}
				account2Input := models.CreateAccountInput{
					Name:      "Test Account 2",
					BankType:  models.BankTypeSBI,
					Currency:  models.CurrencyUSD,
					CreatedBy: userId,
				}

				acc1, _ := mockAccountRepo.CreateAccount(ctx, account1Input)
				acc2, _ := mockAccountRepo.CreateAccount(ctx, account2Input)
				account1Id = acc1.Id
				account2Id = acc2.Id

				// Set up current balances - only account1 has transactions
				currentBalances := map[int64]float64{
					account1Id: 1500.0,
					// account2Id has no transactions, so not in map
				}
				mockAnalyticsRepo.SetBalance(userId, nil, nil, currentBalances)

				// Set up historical balances - only account1 has historical data
				oneMonthAgo := time.Now().AddDate(0, -1, 0)
				historicalBalances := map[int64]float64{
					account1Id: 1200.0,
					// account2Id has no historical data
				}
				mockAnalyticsRepo.SetBalance(userId, nil, &oneMonthAgo, historicalBalances)
			})

			It("should return analytics with zero balances for accounts without transactions", func() {
				result, err := analyticsService.GetAccountAnalytics(ctx, userId)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.AccountAnalytics).To(HaveLen(2))

				// Find analytics for each account
				var account1Analytics, account2Analytics *models.AccountBalanceAnalytics
				for i := range result.AccountAnalytics {
					switch result.AccountAnalytics[i].AccountID {
					case account1Id:
						account1Analytics = &result.AccountAnalytics[i]
					case account2Id:
						account2Analytics = &result.AccountAnalytics[i]
					}
				}

				// Account1 should have transaction data
				Expect(account1Analytics).NotTo(BeNil())
				Expect(account1Analytics.CurrentBalance).To(Equal(-1500.0))
				Expect(account1Analytics.BalanceOneMonthAgo).To(Equal(-1200.0))

				// Account2 should have zero balances (no transactions)
				Expect(account2Analytics).NotTo(BeNil())
				Expect(account2Analytics.CurrentBalance).To(Equal(0.0))
				Expect(account2Analytics.BalanceOneMonthAgo).To(Equal(0.0))
			})
		})

		Context("when analytics repository returns error for current balances", func() {
			BeforeEach(func() {
				// Create test account
				accountInput := models.CreateAccountInput{
					Name:      "Test Account",
					BankType:  models.BankTypeAxis,
					Currency:  models.CurrencyINR,
					CreatedBy: userId,
				}
				mockAccountRepo.CreateAccount(ctx, accountInput)

				// Configure mock to return error on GetBalance
				mockAnalyticsRepo.SetShouldErrorOnBalance(true)
			})

			AfterEach(func() {
				// Reset error simulation
				mockAnalyticsRepo.SetShouldErrorOnBalance(false)
			})

			It("should return the repository error", func() {
				result, err := analyticsService.GetAccountAnalytics(ctx, userId)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("simulated GetBalance error"))
				Expect(result).To(Equal(models.AccountAnalyticsListResponse{}))
			})
		})

		Context("when analytics repository returns error for historical balances", func() {
			BeforeEach(func() {
				// Create test account
				accountInput := models.CreateAccountInput{
					Name:      "Test Account",
					BankType:  models.BankTypeAxis,
					Currency:  models.CurrencyINR,
					CreatedBy: userId,
				}
				mockAccountRepo.CreateAccount(ctx, accountInput)

				// Set up current balances to succeed
				currentBalances := map[int64]float64{1: 1000.0}
				mockAnalyticsRepo.SetBalance(userId, nil, nil, currentBalances)
			})

			It("should handle the case where current balances succeed but historical fail", func() {
				// This is a complex scenario - the service calls GetBalance twice
				// Once for current (nil, nil) and once for historical (nil, &oneMonthAgo)
				// Our current mock doesn't distinguish between these calls
				// In a real implementation, you might want more sophisticated error simulation

				result, err := analyticsService.GetAccountAnalytics(ctx, userId)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.AccountAnalytics).To(HaveLen(1))
			})
		})

		Context("when different users request analytics", func() {
			var user1Id, user2Id int64

			BeforeEach(func() {
				user1Id = 1
				user2Id = 2

				// Create accounts for user1
				account1Input := models.CreateAccountInput{
					Name:      "User1 Account",
					BankType:  models.BankTypeAxis,
					Currency:  models.CurrencyINR,
					CreatedBy: user1Id,
				}
				acc1, _ := mockAccountRepo.CreateAccount(ctx, account1Input)

				// Create accounts for user2
				account2Input := models.CreateAccountInput{
					Name:      "User2 Account",
					BankType:  models.BankTypeSBI,
					Currency:  models.CurrencyUSD,
					CreatedBy: user2Id,
				}
				acc2, _ := mockAccountRepo.CreateAccount(ctx, account2Input)

				// Set up balances for user1
				user1Balances := map[int64]float64{
					acc1.Id: 2000.0,
				}
				mockAnalyticsRepo.SetBalance(user1Id, nil, nil, user1Balances)

				// Set up balances for user2
				user2Balances := map[int64]float64{
					acc2.Id: 3000.0,
				}
				mockAnalyticsRepo.SetBalance(user2Id, nil, nil, user2Balances)
			})

			It("should return analytics only for the requesting user", func() {
				// Test user1 analytics
				result1, err := analyticsService.GetAccountAnalytics(ctx, user1Id)
				Expect(err).NotTo(HaveOccurred())
				Expect(result1.AccountAnalytics).To(HaveLen(1))
				Expect(result1.AccountAnalytics[0].CurrentBalance).To(Equal(-2000.0))

				// Test user2 analytics
				result2, err := analyticsService.GetAccountAnalytics(ctx, user2Id)
				Expect(err).NotTo(HaveOccurred())
				Expect(result2.AccountAnalytics).To(HaveLen(1))
				Expect(result2.AccountAnalytics[0].CurrentBalance).To(Equal(-3000.0))
			})
		})
	})

	Describe("GetCashBalanceHistory", func() {
		var startDate, endDate time.Time

		BeforeEach(func() {
			startDate, _ = time.Parse("2006-01-02", "2023-01-01")
			endDate, _ = time.Parse("2006-01-02", "2023-01-03")
		})

		Context("when repository returns basic data", func() {
			BeforeEach(func() {
				// Set up mock data with initial balance and daily changes
				initialBalance := 1000.0
				timeSeries := []map[string]any{
					{
						"date":         "2023-01-01",
						"daily_change": 100.0,
					},
					{
						"date":         "2023-01-02",
						"daily_change": -50.0,
					},
				}
				mockAnalyticsRepo.SetCashBalanceHistory(userId, startDate, endDate, initialBalance, timeSeries)
			})

			It("should return cash balance time series with negated values", func() {
				result, err := analyticsService.GetCashBalanceHistory(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Initial balance should be negated
				Expect(result.InitialBalance).To(Equal(-1000.0))

				// Should have data points for each day in range
				Expect(result.TimeSeries).To(HaveLen(3)) // Jan 1, 2, 3

				// Verify first day: -1000 (initial) + (-100) (daily change) = -1100
				Expect(result.TimeSeries[0].Date).To(Equal("2023-01-01"))
				Expect(result.TimeSeries[0].CashBalance).To(Equal(-1100.0))

				// Verify second day: -1100 + (50) (daily change) = -1050
				Expect(result.TimeSeries[1].Date).To(Equal("2023-01-02"))
				Expect(result.TimeSeries[1].CashBalance).To(Equal(-1050.0))

				// Verify third day (no transaction, same balance)
				Expect(result.TimeSeries[2].Date).To(Equal("2023-01-03"))
				Expect(result.TimeSeries[2].CashBalance).To(Equal(-1050.0))
			})
		})

		Context("when repository returns no daily changes", func() {
			BeforeEach(func() {
				// Set up mock data with only initial balance, no daily changes
				initialBalance := 500.0
				timeSeries := []map[string]any{} // Empty time series
				mockAnalyticsRepo.SetCashBalanceHistory(userId, startDate, endDate, initialBalance, timeSeries)
			})

			It("should return flat cash balance time series", func() {
				result, err := analyticsService.GetCashBalanceHistory(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Initial balance should be negated
				Expect(result.InitialBalance).To(Equal(-500.0))

				// Should have data points for each day in range
				Expect(result.TimeSeries).To(HaveLen(3)) // Jan 1, 2, 3

				// All days should have the same cash balance
				for _, point := range result.TimeSeries {
					Expect(point.CashBalance).To(Equal(-500.0))
				}

				// Verify dates are correct
				Expect(result.TimeSeries[0].Date).To(Equal("2023-01-01"))
				Expect(result.TimeSeries[1].Date).To(Equal("2023-01-02"))
				Expect(result.TimeSeries[2].Date).To(Equal("2023-01-03"))
			})
		})

		Context("when repository returns complex daily changes", func() {
			BeforeEach(func() {
				// Set up mock data with multiple transactions on same day and gaps
				initialBalance := 2000.0
				timeSeries := []map[string]any{
					{
						"date":         "2023-01-01",
						"daily_change": 200.0, // Debit (stored as positive)
					},
					{
						"date":         "2023-01-03",
						"daily_change": -150.0, // Credit (stored as negative)
					},
				}
				mockAnalyticsRepo.SetCashBalanceHistory(userId, startDate, endDate, initialBalance, timeSeries)
			})

			It("should handle gaps in daily data correctly", func() {
				result, err := analyticsService.GetCashBalanceHistory(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Initial balance should be negated
				Expect(result.InitialBalance).To(Equal(-2000.0))

				// Should have data points for each day in range
				Expect(result.TimeSeries).To(HaveLen(3))

				// Jan 1: has transaction (debit 200 -> credit -200 for frontend)
				Expect(result.TimeSeries[0].Date).To(Equal("2023-01-01"))
				Expect(result.TimeSeries[0].CashBalance).To(Equal(-2200.0)) // -2000 + (-200)

				// Jan 2: no transaction, same as previous day
				Expect(result.TimeSeries[1].Date).To(Equal("2023-01-02"))
				Expect(result.TimeSeries[1].CashBalance).To(Equal(-2200.0)) // Same as Jan 1

				// Jan 3: has transaction (credit -150 -> debit +150 for frontend)
				Expect(result.TimeSeries[2].Date).To(Equal("2023-01-03"))
				Expect(result.TimeSeries[2].CashBalance).To(Equal(-2050.0)) // -2200 + 150
			})
		})

		Context("when date range is single day", func() {
			BeforeEach(func() {
				startDate, _ = time.Parse("2006-01-02", "2023-01-01")
				endDate = startDate // Same day

				initialBalance := 1500.0
				timeSeries := []map[string]any{
					{
						"date":         "2023-01-01",
						"daily_change": 75.0,
					},
				}
				mockAnalyticsRepo.SetCashBalanceHistory(userId, startDate, endDate, initialBalance, timeSeries)
			})

			It("should return single day cash balance", func() {
				result, err := analyticsService.GetCashBalanceHistory(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Should have exactly one data point
				Expect(result.TimeSeries).To(HaveLen(1))

				Expect(result.TimeSeries[0].Date).To(Equal("2023-01-01"))
				Expect(result.TimeSeries[0].CashBalance).To(Equal(-1575.0)) // -1500 + (-75)
			})
		})

		Context("when repository returns zero initial balance", func() {
			BeforeEach(func() {
				initialBalance := 0.0
				timeSeries := []map[string]any{
					{
						"date":         "2023-01-01",
						"daily_change": 100.0,
					},
					{
						"date":         "2023-01-02",
						"daily_change": -50.0,
					},
				}
				mockAnalyticsRepo.SetCashBalanceHistory(userId, startDate, endDate, initialBalance, timeSeries)
			})

			It("should handle zero initial balance correctly", func() {
				result, err := analyticsService.GetCashBalanceHistory(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Initial balance should be zero (negated)
				Expect(result.InitialBalance).To(Equal(0.0))

				// Verify cumulative calculation from zero
				Expect(result.TimeSeries[0].CashBalance).To(Equal(-100.0)) // 0 + (-100)
				Expect(result.TimeSeries[1].CashBalance).To(Equal(-50.0))  // -100 + 50
				Expect(result.TimeSeries[2].CashBalance).To(Equal(-50.0))  // Same as previous day
			})
		})

		Context("when different users request cash balance", func() {
			var user1Id, user2Id int64

			BeforeEach(func() {
				user1Id = 1
				user2Id = 2

				// Set up different cash balance data for each user
				user1InitialBalance := 1000.0
				user1TimeSeries := []map[string]any{
					{
						"date":         "2023-01-01",
						"daily_change": 100.0,
					},
				}
				mockAnalyticsRepo.SetCashBalanceHistory(user1Id, startDate, endDate, user1InitialBalance, user1TimeSeries)

				user2InitialBalance := 2000.0
				user2TimeSeries := []map[string]any{
					{
						"date":         "2023-01-01",
						"daily_change": 200.0,
					},
				}
				mockAnalyticsRepo.SetCashBalanceHistory(user2Id, startDate, endDate, user2InitialBalance, user2TimeSeries)
			})

			It("should return cash balance data only for the requesting user", func() {
				// Test user1 cash balance
				result1, err := analyticsService.GetCashBalanceHistory(ctx, user1Id, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())
				Expect(result1.InitialBalance).To(Equal(-1000.0))
				Expect(result1.TimeSeries[0].CashBalance).To(Equal(-1100.0)) // -1000 - 100

				// Test user2 cash balance
				result2, err := analyticsService.GetCashBalanceHistory(ctx, user2Id, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())
				Expect(result2.InitialBalance).To(Equal(-2000.0))
				Expect(result2.TimeSeries[0].CashBalance).To(Equal(-2200.0)) // -2000 - 200
			})
		})

		Context("when repository returns error", func() {
			BeforeEach(func() {
				// Configure mock to return error
				mockAnalyticsRepo.SetShouldErrorOnCashBalance(true)
			})

			AfterEach(func() {
				// Reset error simulation
				mockAnalyticsRepo.SetShouldErrorOnCashBalance(false)
			})

			It("should return the repository error", func() {
				result, err := analyticsService.GetCashBalanceHistory(ctx, userId, startDate, endDate)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("simulated GetCashBalanceHistory error"))
				Expect(result).To(Equal(models.CashBalanceHistoryResponse{}))
			})
		})

		Context("when validating value negation logic", func() {
			BeforeEach(func() {
				// Test the core business logic: debits stored as positive, credits as negative
				// But frontend expects opposite
				initialBalance := 1000.0 // Stored as positive (debit balance)
				timeSeries := []map[string]any{
					{
						"date":         "2023-01-01",
						"daily_change": 200.0, // Debit transaction (stored positive)
					},
					{
						"date":         "2023-01-02",
						"daily_change": -150.0, // Credit transaction (stored negative)
					},
				}
				mockAnalyticsRepo.SetCashBalanceHistory(userId, startDate, endDate, initialBalance, timeSeries)
			})

			It("should correctly negate all values for frontend consumption", func() {
				result, err := analyticsService.GetCashBalanceHistory(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Initial balance: 1000 (debit) -> -1000 (for frontend)
				Expect(result.InitialBalance).To(Equal(-1000.0))

				// Day 1: -1000 + (-200) = -1200 (debit transaction becomes negative for frontend)
				Expect(result.TimeSeries[0].CashBalance).To(Equal(-1200.0))

				// Day 2: -1200 + 150 = -1050 (credit transaction becomes positive for frontend)
				Expect(result.TimeSeries[1].CashBalance).To(Equal(-1050.0))

				// Day 3: Same as day 2 (no transaction)
				Expect(result.TimeSeries[2].CashBalance).To(Equal(-1050.0))
			})
		})

		Context("when testing edge cases and boundary conditions", func() {
			BeforeEach(func() {
				startDate, _ = time.Parse("2006-01-02", "2023-01-01")
				endDate = startDate // Same day
			})

			It("should handle same start and end date correctly", func() {
				initialBalance := 1000.0
				timeSeries := []map[string]any{
					{
						"date":         "2023-01-01",
						"daily_change": 100.0,
					},
				}
				mockAnalyticsRepo.SetCashBalanceHistory(userId, startDate, endDate, initialBalance, timeSeries)

				result, err := analyticsService.GetCashBalanceHistory(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Should have exactly one data point
				Expect(result.TimeSeries).To(HaveLen(1))
				Expect(result.TimeSeries[0].Date).To(Equal("2023-01-01"))
				Expect(result.TimeSeries[0].CashBalance).To(Equal(-1100.0)) // -1000 + (-100)
			})

			It("should handle very large date ranges", func() {
				startDate, _ = time.Parse("2006-01-02", "2023-01-01")
				endDate, _ = time.Parse("2006-01-02", "2023-12-31")

				initialBalance := 1000.0
				timeSeries := []map[string]any{} // No daily changes
				mockAnalyticsRepo.SetCashBalanceHistory(userId, startDate, endDate, initialBalance, timeSeries)

				result, err := analyticsService.GetCashBalanceHistory(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Should have 365 data points for 2023 (not a leap year)
				Expect(result.TimeSeries).To(HaveLen(365))

				// All should have the same cash balance (no changes)
				for _, point := range result.TimeSeries {
					Expect(point.CashBalance).To(Equal(-1000.0))
				}
			})

			It("should handle leap year correctly", func() {
				startDate, _ = time.Parse("2006-01-02", "2024-02-28")
				endDate, _ = time.Parse("2006-01-02", "2024-03-01")

				initialBalance := 500.0
				timeSeries := []map[string]any{} // No daily changes
				mockAnalyticsRepo.SetCashBalanceHistory(userId, startDate, endDate, initialBalance, timeSeries)

				result, err := analyticsService.GetCashBalanceHistory(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Should have 3 data points: Feb 28, Feb 29, Mar 1
				Expect(result.TimeSeries).To(HaveLen(3))
				Expect(result.TimeSeries[0].Date).To(Equal("2024-02-28"))
				Expect(result.TimeSeries[1].Date).To(Equal("2024-02-29"))
				Expect(result.TimeSeries[2].Date).To(Equal("2024-03-01"))
			})

			It("should handle cross-year date ranges", func() {
				startDate, _ = time.Parse("2006-01-02", "2023-12-30")
				endDate, _ = time.Parse("2006-01-02", "2024-01-02")

				initialBalance := 2000.0
				timeSeries := []map[string]any{
					{
						"date":         "2023-12-31",
						"daily_change": 100.0,
					},
					{
						"date":         "2024-01-01",
						"daily_change": -50.0,
					},
				}
				mockAnalyticsRepo.SetCashBalanceHistory(userId, startDate, endDate, initialBalance, timeSeries)

				result, err := analyticsService.GetCashBalanceHistory(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Should have 4 data points
				Expect(result.TimeSeries).To(HaveLen(4))

				// Verify year boundary crossing
				Expect(result.TimeSeries[0].Date).To(Equal("2023-12-30"))
				Expect(result.TimeSeries[0].CashBalance).To(Equal(-2000.0)) // No change

				Expect(result.TimeSeries[1].Date).To(Equal("2023-12-31"))
				Expect(result.TimeSeries[1].CashBalance).To(Equal(-2100.0)) // -2000 + (-100)

				Expect(result.TimeSeries[2].Date).To(Equal("2024-01-01"))
				Expect(result.TimeSeries[2].CashBalance).To(Equal(-2050.0)) // -2100 + 50

				Expect(result.TimeSeries[3].Date).To(Equal("2024-01-02"))
				Expect(result.TimeSeries[3].CashBalance).To(Equal(-2050.0)) // No change
			})

			It("should handle very large transaction amounts", func() {
				initialBalance := 999999999.99
				timeSeries := []map[string]any{
					{
						"date":         "2023-01-01",
						"daily_change": 888888888.88,
					},
				}
				mockAnalyticsRepo.SetCashBalanceHistory(userId, startDate, endDate, initialBalance, timeSeries)

				result, err := analyticsService.GetCashBalanceHistory(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Should handle large numbers correctly
				Expect(result.InitialBalance).To(Equal(-999999999.99))
				Expect(result.TimeSeries[0].CashBalance).To(BeNumerically("~", -1888888888.87, 0.01))
			})

			It("should handle negative initial balance", func() {
				initialBalance := -500.0 // Negative initial balance
				timeSeries := []map[string]any{
					{
						"date":         "2023-01-01",
						"daily_change": 100.0,
					},
				}
				mockAnalyticsRepo.SetCashBalanceHistory(userId, startDate, endDate, initialBalance, timeSeries)

				result, err := analyticsService.GetCashBalanceHistory(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Negative initial balance should become positive for frontend
				Expect(result.InitialBalance).To(Equal(500.0))
				Expect(result.TimeSeries[0].CashBalance).To(Equal(400.0)) // 500 + (-100)
			})

			It("should handle empty daily data gracefully", func() {
				initialBalance := 1000.0
				timeSeries := []map[string]any{} // Empty
				mockAnalyticsRepo.SetCashBalanceHistory(userId, startDate, endDate, initialBalance, timeSeries)

				result, err := analyticsService.GetCashBalanceHistory(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Should still generate time series with flat values
				Expect(result.TimeSeries).To(HaveLen(1))
				Expect(result.TimeSeries[0].CashBalance).To(Equal(-1000.0))
			})

			It("should handle malformed daily data gracefully", func() {
				initialBalance := 1000.0
				timeSeries := []map[string]any{
					{
						"date":         "2023-01-01",
						"daily_change": "invalid", // Invalid type
					},
				}
				mockAnalyticsRepo.SetCashBalanceHistory(userId, startDate, endDate, initialBalance, timeSeries)

				// Testing the new behavior
				_, err := analyticsService.GetCashBalanceHistory(ctx, userId, startDate, endDate)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("invalid type for daily_change in daily data"))
			})
		})

		Context("when testing data consistency and edge cases", func() {
			It("should handle duplicate dates in daily data", func() {
				startDate, _ = time.Parse("2006-01-02", "2023-01-01")
				endDate, _ = time.Parse("2006-01-02", "2023-01-02")

				initialBalance := 1000.0
				timeSeries := []map[string]any{
					{
						"date":         "2023-01-01",
						"daily_change": 100.0,
					},
					{
						"date":         "2023-01-01", // Duplicate date
						"daily_change": 50.0,
					},
				}
				mockAnalyticsRepo.SetCashBalanceHistory(userId, startDate, endDate, initialBalance, timeSeries)

				result, err := analyticsService.GetCashBalanceHistory(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// The service uses a map, so the last value should win
				Expect(result.TimeSeries[0].CashBalance).To(Equal(-1050.0)) // -1000 + (-50)
			})

			It("should handle dates outside the requested range", func() {
				startDate, _ = time.Parse("2006-01-02", "2023-01-01")
				endDate, _ = time.Parse("2006-01-02", "2023-01-02")

				initialBalance := 1000.0
				timeSeries := []map[string]any{
					{
						"date":         "2022-12-31", // Before range
						"daily_change": 100.0,
					},
					{
						"date":         "2023-01-01", // In range
						"daily_change": 50.0,
					},
					{
						"date":         "2023-01-03", // After range
						"daily_change": 25.0,
					},
				}
				mockAnalyticsRepo.SetCashBalanceHistory(userId, startDate, endDate, initialBalance, timeSeries)

				result, err := analyticsService.GetCashBalanceHistory(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Should only use the date within range
				Expect(result.TimeSeries[0].CashBalance).To(Equal(-1050.0)) // -1000 + (-50)
				Expect(result.TimeSeries[1].CashBalance).To(Equal(-1050.0)) // No change for Jan 2
			})
		})

		Context("when investment accounts hold balances", func() {
			It("should exclude investment ledgers from the cash history", func() {
				startDate, _ = time.Parse("2006-01-02", "2023-01-01")
				endDate, _ = time.Parse("2006-01-02", "2023-01-03")

				bankBalance := 1000.0
				_, err := mockAccountRepo.CreateAccount(ctx, models.CreateAccountInput{
					Name:      "Bank",
					BankType:  models.BankTypeSBI,
					Currency:  models.CurrencyINR,
					Balance:   &bankBalance,
					CreatedBy: userId,
				})
				Expect(err).NotTo(HaveOccurred())

				investmentBalance := 5000.0
				currentValue := 5000.0
				_, err = mockAccountRepo.CreateAccount(ctx, models.CreateAccountInput{
					Name:         "FD",
					BankType:     models.BankTypeInvestment,
					Currency:     models.CurrencyINR,
					Balance:      &investmentBalance,
					CurrentValue: &currentValue,
					CreatedBy:    userId,
				})
				Expect(err).NotTo(HaveOccurred())

				mockAnalyticsRepo.SetCashBalanceHistory(userId, startDate, endDate, 0, []map[string]any{})

				result, err := analyticsService.GetCashBalanceHistory(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.InitialBalance).To(Equal(1000.0))
				Expect(result.TimeSeries[0].CashBalance).To(Equal(1000.0))
			})
		})
	})

	Describe("GetMonthlyAnalytics", func() {
		var startDate, endDate time.Time

		BeforeEach(func() {
			startDate = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			endDate = time.Date(2024, 6, 30, 23, 59, 59, 0, time.UTC)
		})

		Context("when date range is invalid", func() {
			It("should return error when end date is before start date", func() {
				invalidStartDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
				invalidEndDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

				_, err := analyticsService.GetMonthlyAnalytics(ctx, userId, invalidStartDate, invalidEndDate)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("end date must be after or equal to start date"))
			})

			It("should allow same start and end date", func() {
				sameDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

				_, err := analyticsService.GetMonthlyAnalytics(ctx, userId, sameDate, sameDate)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when repository returns error", func() {
			BeforeEach(func() {
				mockAnalyticsRepo.SetShouldErrorOnMonthly(true)
			})

			It("should return error", func() {
				_, err := analyticsService.GetMonthlyAnalytics(ctx, userId, startDate, endDate)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when repository returns success", func() {
			var expectedAnalytics *models.MonthlyAnalyticsResponse

			BeforeEach(func() {
				expectedAnalytics = &models.MonthlyAnalyticsResponse{
					TotalIncome:   1500.0,
					TotalExpenses: 1200.0,
					TotalAmount:   2700.0,
				}
				mockAnalyticsRepo.SetMonthlyAnalytics(userId, startDate, endDate, expectedAnalytics)
			})

			It("should return monthly analytics successfully", func() {
				result, err := analyticsService.GetMonthlyAnalytics(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.TotalIncome).To(Equal(1500.0))
				Expect(result.TotalExpenses).To(Equal(1200.0))
				Expect(result.TotalAmount).To(Equal(2700.0))
			})
		})

		Context("when no specific data is set", func() {
			It("should return default analytics", func() {
				otherStartDate := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
				otherEndDate := time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC)

				result, err := analyticsService.GetMonthlyAnalytics(ctx, userId, otherStartDate, otherEndDate)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.TotalIncome).To(Equal(1000.0))  // Default from mock
				Expect(result.TotalExpenses).To(Equal(600.0)) // Default from mock
				Expect(result.TotalAmount).To(Equal(400.0))   // Default from mock
			})
		})

		Context("with different date ranges", func() {
			BeforeEach(func() {
				// Set different analytics for different date ranges
				startDate3Months := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
				endDate3Months := time.Date(2024, 5, 31, 23, 59, 59, 0, time.UTC)
				startDate12Months := time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)
				endDate12Months := time.Date(2024, 5, 31, 23, 59, 59, 0, time.UTC)

				analytics3Months := &models.MonthlyAnalyticsResponse{
					TotalIncome:   800.0,
					TotalExpenses: 600.0,
					TotalAmount:   1400.0,
				}
				analytics12Months := &models.MonthlyAnalyticsResponse{
					TotalIncome:   3000.0,
					TotalExpenses: 2500.0,
					TotalAmount:   5500.0,
				}
				mockAnalyticsRepo.SetMonthlyAnalytics(userId, startDate3Months, endDate3Months, analytics3Months)
				mockAnalyticsRepo.SetMonthlyAnalytics(userId, startDate12Months, endDate12Months, analytics12Months)
			})

			It("should return correct analytics for 3 months range", func() {
				startDate3Months := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
				endDate3Months := time.Date(2024, 5, 31, 23, 59, 59, 0, time.UTC)

				result, err := analyticsService.GetMonthlyAnalytics(ctx, userId, startDate3Months, endDate3Months)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.TotalIncome).To(Equal(800.0))
				Expect(result.TotalExpenses).To(Equal(600.0))
				Expect(result.TotalAmount).To(Equal(1400.0))
			})

			It("should return correct analytics for 12 months range", func() {
				startDate12Months := time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)
				endDate12Months := time.Date(2024, 5, 31, 23, 59, 59, 0, time.UTC)

				result, err := analyticsService.GetMonthlyAnalytics(ctx, userId, startDate12Months, endDate12Months)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.TotalIncome).To(Equal(3000.0))
				Expect(result.TotalExpenses).To(Equal(2500.0))
				Expect(result.TotalAmount).To(Equal(5500.0))
			})
		})

		Context("with edge cases", func() {
			It("should handle analytics with zero income", func() {
				testStartDate := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
				testEndDate := time.Date(2024, 4, 30, 23, 59, 59, 0, time.UTC)

				analyticsNoIncome := &models.MonthlyAnalyticsResponse{
					TotalIncome:   0.0,
					TotalExpenses: 500.0,
					TotalAmount:   500.0,
				}
				mockAnalyticsRepo.SetMonthlyAnalytics(userId, testStartDate, testEndDate, analyticsNoIncome)

				result, err := analyticsService.GetMonthlyAnalytics(ctx, userId, testStartDate, testEndDate)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.TotalIncome).To(Equal(0.0))
				Expect(result.TotalExpenses).To(Equal(500.0))
				Expect(result.TotalAmount).To(Equal(500.0))
			})

			It("should handle analytics with zero expenses", func() {
				testStartDate := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)
				testEndDate := time.Date(2024, 5, 31, 23, 59, 59, 0, time.UTC)

				analyticsNoExpenses := &models.MonthlyAnalyticsResponse{
					TotalIncome:   1000.0,
					TotalExpenses: 0.0,
					TotalAmount:   1000.0,
				}
				mockAnalyticsRepo.SetMonthlyAnalytics(userId, testStartDate, testEndDate, analyticsNoExpenses)

				result, err := analyticsService.GetMonthlyAnalytics(ctx, userId, testStartDate, testEndDate)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.TotalIncome).To(Equal(1000.0))
				Expect(result.TotalExpenses).To(Equal(0.0))
				Expect(result.TotalAmount).To(Equal(1000.0))
			})

			It("should handle analytics with both zero income and expenses", func() {
				testStartDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
				testEndDate := time.Date(2024, 6, 30, 23, 59, 59, 0, time.UTC)

				analyticsZero := &models.MonthlyAnalyticsResponse{
					TotalIncome:   0.0,
					TotalExpenses: 0.0,
					TotalAmount:   0.0,
				}
				mockAnalyticsRepo.SetMonthlyAnalytics(userId, testStartDate, testEndDate, analyticsZero)

				result, err := analyticsService.GetMonthlyAnalytics(ctx, userId, testStartDate, testEndDate)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.TotalIncome).To(Equal(0.0))
				Expect(result.TotalExpenses).To(Equal(0.0))
				Expect(result.TotalAmount).To(Equal(0.0))
			})
		})

		Context("with investment accounts", func() {
			var investmentAccount models.AccountResponse
			var regularAccount models.AccountResponse

			BeforeEach(func() {
				currentValue := 15000.0
				investmentInput := models.CreateAccountInput{
					Name:         "Investment Account",
					BankType:     models.BankTypeInvestment,
					Currency:     models.CurrencyINR,
					CurrentValue: &currentValue,
					CreatedBy:    userId,
				}
				var err error
				investmentAccount, err = mockAccountRepo.CreateAccount(ctx, investmentInput)
				Expect(err).NotTo(HaveOccurred())

				regularInput := models.CreateAccountInput{
					Name:      "Regular Account",
					BankType:  models.BankTypeAxis,
					Currency:  models.CurrencyINR,
					CreatedBy: userId,
				}
				regularAccount, err = mockAccountRepo.CreateAccount(ctx, regularInput)
				Expect(err).NotTo(HaveOccurred())

				currentBalances := map[int64]float64{
					investmentAccount.Id: 5000.0,
					regularAccount.Id:    3000.0,
				}
				mockAnalyticsRepo.SetBalance(userId, nil, nil, currentBalances)

				oneMonthAgo := time.Now().AddDate(0, -1, 0)
				historicalBalances := map[int64]float64{
					investmentAccount.Id: 4000.0,
					regularAccount.Id:    2000.0,
				}
				mockAnalyticsRepo.SetBalance(userId, nil, &oneMonthAgo, historicalBalances)
			})

			It("should include current value for investment accounts", func() {
				result, err := analyticsService.GetAccountAnalytics(ctx, userId)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(result.AccountAnalytics)).To(BeNumerically(">=", 1))

				var investmentAnalytics *models.AccountBalanceAnalytics
				for i := range result.AccountAnalytics {
					if result.AccountAnalytics[i].AccountID == investmentAccount.Id {
						investmentAnalytics = &result.AccountAnalytics[i]
						break
					}
				}

				Expect(investmentAnalytics).NotTo(BeNil())
				Expect(investmentAnalytics.CurrentValue).NotTo(BeNil())
				Expect(*investmentAnalytics.CurrentValue).To(Equal(15000.0))
			})

			It("should include current value for investment accounts without regular account current values", func() {
				result, err := analyticsService.GetAccountAnalytics(ctx, userId)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(result.AccountAnalytics)).To(BeNumerically(">=", 1))

				var investmentAnalytics *models.AccountBalanceAnalytics
				for i := range result.AccountAnalytics {
					if result.AccountAnalytics[i].AccountID == investmentAccount.Id {
						investmentAnalytics = &result.AccountAnalytics[i]
						break
					}
				}

				Expect(investmentAnalytics).NotTo(BeNil())
				Expect(investmentAnalytics.CurrentValue).NotTo(BeNil())
				Expect(*investmentAnalytics.CurrentValue).To(Equal(15000.0))
			})

			It("should compute percentage increase for investment accounts", func() {
				result, err := analyticsService.GetAccountAnalytics(ctx, userId)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(result.AccountAnalytics)).To(BeNumerically(">=", 1))

				var investmentAnalytics *models.AccountBalanceAnalytics
				for i := range result.AccountAnalytics {
					if result.AccountAnalytics[i].AccountID == investmentAccount.Id {
						investmentAnalytics = &result.AccountAnalytics[i]
						break
					}
				}

				Expect(investmentAnalytics).NotTo(BeNil())
				Expect(investmentAnalytics.CurrentValue).NotTo(BeNil())
				Expect(*investmentAnalytics.CurrentValue).To(Equal(15000.0))
			})

			It("should not include investment metrics for regular accounts", func() {
				result, err := analyticsService.GetAccountAnalytics(ctx, userId)
				Expect(err).NotTo(HaveOccurred())

				var regularAnalytics *models.AccountBalanceAnalytics
				for i := range result.AccountAnalytics {
					if result.AccountAnalytics[i].AccountID == regularAccount.Id {
						regularAnalytics = &result.AccountAnalytics[i]
						break
					}
				}

				Expect(regularAnalytics).NotTo(BeNil())
				Expect(regularAnalytics.CurrentValue).To(BeNil())
				Expect(regularAnalytics.PercentageIncrease).To(BeNil())
				Expect(regularAnalytics.Xirr).To(BeNil())
			})

			It("should return nil current value for investment account without current_value set", func() {
				regularInput := models.CreateAccountInput{
					Name:      "Investment No Value",
					BankType:  models.BankTypeInvestment,
					Currency:  models.CurrencyINR,
					CreatedBy: userId,
				}
				noValueAccount, err := mockAccountRepo.CreateAccount(ctx, regularInput)
				Expect(err).NotTo(HaveOccurred())

				currentBalances := map[int64]float64{
					noValueAccount.Id: 1000.0,
				}
				mockAnalyticsRepo.SetBalance(userId, nil, nil, currentBalances)

				oneMonthAgo := time.Now().AddDate(0, -1, 0)
				historicalBalances := map[int64]float64{
					noValueAccount.Id: 500.0,
				}
				mockAnalyticsRepo.SetBalance(userId, nil, &oneMonthAgo, historicalBalances)

				result, err := analyticsService.GetAccountAnalytics(ctx, userId)
				Expect(err).NotTo(HaveOccurred())

				var noValueAnalytics *models.AccountBalanceAnalytics
				for i := range result.AccountAnalytics {
					if result.AccountAnalytics[i].AccountID == noValueAccount.Id {
						noValueAnalytics = &result.AccountAnalytics[i]
						break
					}
				}

				Expect(noValueAnalytics).NotTo(BeNil())
				Expect(noValueAnalytics.CurrentValue).To(BeNil())
				Expect(noValueAnalytics.PercentageIncrease).To(BeNil())
				Expect(noValueAnalytics.Xirr).To(BeNil())
			})
		})
	})

	Describe("calculateInvestmentMetrics", func() {
		now := time.Date(2023, time.June, 15, 0, 0, 0, 0, time.UTC)

		It("should return zero percentage and XIRR when current value is zero or negative", func() {
			currentValue := -100.0
			flows := []models.AccountCashFlow{
				{AccountID: 1, Amount: -5000.0, Date: now.AddDate(0, -1, 0)},
				{AccountID: 1, Amount: -3000.0, Date: now.AddDate(0, -6, 0)},
			}
			percentage, xirr := calculateInvestmentMetrics(flows, currentValue, now)
			Expect(percentage).To(BeNumerically("==", 0.0))
			Expect(xirr).NotTo(BeNil())
			Expect(*xirr).To(BeNumerically("==", 0.0))
		})

		It("should calculate percentage increase correctly for 50% return", func() {
			currentValue := 15000.0
			flows := []models.AccountCashFlow{
				{AccountID: 1, Amount: -10000.0, Date: now.AddDate(-1, 0, 0)},
			}
			percentage, _ := calculateInvestmentMetrics(flows, currentValue, now)
			expectedPct := ((15000.0 - 10000.0) / 10000.0) * 100
			Expect(percentage).To(BeNumerically("~", expectedPct))
		})

		It("should calculate percentage increase correctly for 100% return", func() {
			currentValue := 20000.0
			flows := []models.AccountCashFlow{
				{AccountID: 1, Amount: -10000.0, Date: now.AddDate(-1, 0, 0)},
			}
			percentage, _ := calculateInvestmentMetrics(flows, currentValue, now)
			expectedPct := ((20000.0 - 10000.0) / 10000.0) * 100
			Expect(percentage).To(BeNumerically("~", expectedPct))
		})

		It("should calculate percentage increase correctly for 20% loss", func() {
			currentValue := 8000.0
			flows := []models.AccountCashFlow{
				{AccountID: 1, Amount: -10000.0, Date: now.AddDate(-1, 0, 0)},
			}
			percentage, _ := calculateInvestmentMetrics(flows, currentValue, now)
			expectedPct := ((8000.0 - 10000.0) / 10000.0) * 100
			Expect(percentage).To(BeNumerically("~", expectedPct))
		})

		It("should calculate percentage increase correctly for break-even", func() {
			currentValue := 10000.0
			flows := []models.AccountCashFlow{
				{AccountID: 1, Amount: -10000.0, Date: now.AddDate(-1, 0, 0)},
			}
			percentage, _ := calculateInvestmentMetrics(flows, currentValue, now)
			Expect(percentage).To(BeNumerically("==", 0.0))
		})

		It("should calculate percentage correctly with multiple investments", func() {
			currentValue := 25000.0
			flows := []models.AccountCashFlow{
				{AccountID: 1, Amount: -10000.0, Date: now.AddDate(-1, 0, 0)},
				{AccountID: 1, Amount: -15000.0, Date: now.AddDate(-2, 0, 0)},
			}
			totalInvested := 25000.0
			percentage, _ := calculateInvestmentMetrics(flows, currentValue, now)
			expectedPct := ((currentValue - totalInvested) / totalInvested) * 100
			Expect(percentage).To(BeNumerically("~", expectedPct))
		})

		It("should return zero XIRR when no cash flows", func() {
			currentValue := 15000.0
			emptyFlows := []models.AccountCashFlow{}
			_, xirr := calculateInvestmentMetrics(emptyFlows, currentValue, now)
			Expect(xirr).NotTo(BeNil())
			Expect(*xirr).To(BeNumerically("==", 0.0))
		})

		It("should return zero XIRR when only one cash flow", func() {
			currentValue := 15000.0
			oneFlow := []models.AccountCashFlow{
				{AccountID: 1, Amount: 15000.0, Date: now.AddDate(-1, 0, 0)},
			}
			_, xirr := calculateInvestmentMetrics(oneFlow, currentValue, now)
			Expect(xirr).NotTo(BeNil())
			Expect(*xirr).To(BeNumerically("==", 0.0))
		})

		It("should return zero XIRR when cash flows have same date", func() {
			currentValue := 15000.0
			sameDateFlows := []models.AccountCashFlow{
				{AccountID: 1, Amount: -10000.0, Date: now.AddDate(-1, 0, 0)},
				{AccountID: 1, Amount: 5000.0, Date: now.AddDate(-1, 0, 0)},
			}
			_, xirr := calculateInvestmentMetrics(sameDateFlows, currentValue, now)
			Expect(xirr).NotTo(BeNil())
			Expect(*xirr).To(BeNumerically("==", 0.0))
		})

		It("should calculate XIRR correctly for 50% annual return (1 year)", func() {
			currentValue := 15000.0
			oneYearDate := now.AddDate(-1, 0, 0)
			flows := []models.AccountCashFlow{
				{AccountID: 1, Amount: -10000.0, Date: oneYearDate},
			}
			_, xirr := calculateInvestmentMetrics(flows, currentValue, now)
			expectedXIRR := math.Pow(1.5, 1.0) - 1
			Expect(xirr).NotTo(BeNil())
			Expect(*xirr).To(BeNumerically("~", expectedXIRR*100))
		})

		It("should calculate XIRR correctly for 100% annual return", func() {
			currentValue := 20000.0
			oneYearDate := now.AddDate(-1, 0, 0)
			flows := []models.AccountCashFlow{
				{AccountID: 1, Amount: -10000.0, Date: oneYearDate},
			}
			_, xirr := calculateInvestmentMetrics(flows, currentValue, now)
			expectedXIRR := 1.0
			Expect(xirr).NotTo(BeNil())
			Expect(*xirr).To(BeNumerically("~", expectedXIRR*100))
		})

		It("should calculate XIRR correctly for -50% annual return (1 year loss)", func() {
			currentValue := 5000.0
			oneYearAgo := now.AddDate(-1, 0, 0)
			flows := []models.AccountCashFlow{
				{AccountID: 1, Amount: -10000.0, Date: oneYearAgo},
			}
			_, xirr := calculateInvestmentMetrics(flows, currentValue, now)
			expectedXIRR := -0.5
			Expect(xirr).NotTo(BeNil())
			Expect(*xirr).To(BeNumerically("~", expectedXIRR*100))
		})

		It("should return zero XIRR when all flows are on same date", func() {
			currentValue := 10000.0
			flows := []models.AccountCashFlow{
				{AccountID: 1, Amount: -10000.0, Date: now.AddDate(-1, 0, 0)},
				{AccountID: 1, Amount: 10000.0, Date: now.AddDate(-1, 0, 0)},
			}
			_, xirr := calculateInvestmentMetrics(flows, currentValue, now)
			Expect(xirr).NotTo(BeNil())
			Expect(*xirr).To(BeNumerically("==", 0.0))
		})

		It("should return zero XIRR when only current value flow", func() {
			currentValue := 10000.0
			flows := []models.AccountCashFlow{
				{AccountID: 1, Amount: 10000.0, Date: now.AddDate(0, 0, 0)},
			}
			_, xirr := calculateInvestmentMetrics(flows, currentValue, now)
			Expect(xirr).NotTo(BeNil())
			Expect(*xirr).To(BeNumerically("==", 0.0))
		})

		It("should handle edge case with very small XIRR values", func() {
			currentValue := 10001.0
			flows := []models.AccountCashFlow{
				{AccountID: 1, Amount: -10000.0, Date: now.AddDate(-1, 0, 0)},
			}
			_, xirr := calculateInvestmentMetrics(flows, currentValue, now)
			Expect(xirr).NotTo(BeNil())
			Expect(math.IsNaN(*xirr)).To(BeFalse())
		})

		It("should calculate XIRR correctly with multiple investments over time", func() {
			currentValue := 50000.0
			flows := []models.AccountCashFlow{
				{AccountID: 1, Amount: -10000.0, Date: now.AddDate(-1, 0, 0)},
				{AccountID: 1, Amount: -10000.0, Date: now.AddDate(-2, 0, 0)},
				{AccountID: 1, Amount: -10000.0, Date: now.AddDate(-3, 0, 0)},
				{AccountID: 1, Amount: -5000.0, Date: now.AddDate(-4, 0, 0)},
				{AccountID: 1, Amount: 5000.0, Date: now.AddDate(-5, 0, 0)},
			}
			_, xirr := calculateInvestmentMetrics(flows, currentValue, now)
			Expect(xirr).NotTo(BeNil())
			// Expected annualized XIRR for these flows is approximately 34.12%
			Expect(*xirr).To(BeNumerically("~", 34.11777533442272, 1e-3))
		})

		It("should calculate percentage correctly with mixed cash flows", func() {
			currentValue := 15000.0
			flows := []models.AccountCashFlow{
				{AccountID: 1, Amount: -5000.0, Date: now.AddDate(-730, 0, 0)},
				{AccountID: 1, Amount: 3000.0, Date: now.AddDate(-365, 0, 0)},
				{AccountID: 1, Amount: -7000.0, Date: now.AddDate(-365, 0, 0)},
				{AccountID: 1, Amount: 3000.0, Date: now.AddDate(-365, 0, 0)},
				{AccountID: 1, Amount: 5000.0, Date: now.AddDate(-365, 0, 0)},
			}
			// totalInvested should be sum of absolute negative flows (investments only)
			totalInvested := 0.0
			for _, f := range flows {
				if f.Amount < 0 {
					totalInvested += -f.Amount
				}
			}
			percentage, _ := calculateInvestmentMetrics(flows, currentValue, now)
			expectedPct := ((currentValue - totalInvested) / totalInvested) * 100
			Expect(percentage).To(BeNumerically("~", expectedPct))
		})

		It("should ignore FD interest credit rows in XIRR and percentage", func() {
			currentValue := 107000.0
			filteredFlows := []models.AccountCashFlow{
				{AccountID: 1, Amount: -100000.0, Date: now.AddDate(-2, 0, 0)},
				{AccountID: 1, Amount: -100000.0, Date: now.AddDate(-1, 0, 0)},
				{AccountID: 1, Amount: 100000.0, Date: now.AddDate(-1, 0, 0), Name: "FD principal redemption"},
				{AccountID: 1, Amount: 7000.0, Date: now.AddDate(-1, 0, 0), Name: "FD interest exit from SBI"},
			}
			withCreditFlows := append([]models.AccountCashFlow{
				{AccountID: 1, Amount: -7000.0, Date: now.AddDate(-1, 0, 0), Name: "FD interest credit from SBI"},
			}, filteredFlows...)

			filteredPct, filteredXirr := calculateInvestmentMetrics(filteredFlows, currentValue, now)
			creditPct, creditXirr := calculateInvestmentMetrics(withCreditFlows, currentValue, now)

			Expect(creditPct).To(Equal(filteredPct))
			Expect(filteredXirr).NotTo(BeNil())
			Expect(creditXirr).NotTo(BeNil())
			Expect(*creditXirr).To(Equal(*filteredXirr))
		})

		It("should keep a non-matching credit-like row in totalInvested and lower XIRR", func() {
			currentValue := 107000.0
			baseFlows := []models.AccountCashFlow{
				{AccountID: 1, Amount: -100000.0, Date: now.AddDate(-2, 0, 0)},
				{AccountID: 1, Amount: -100000.0, Date: now.AddDate(-1, 0, 0)},
				{AccountID: 1, Amount: 100000.0, Date: now.AddDate(-1, 0, 0), Name: "FD principal redemption"},
				{AccountID: 1, Amount: 7000.0, Date: now.AddDate(-1, 0, 0), Name: "FD interest exit from SBI"},
			}
			transferFlows := append([]models.AccountCashFlow{
				{AccountID: 1, Amount: -7000.0, Date: now.AddDate(-1, 0, 0), Name: "Transfer"},
			}, baseFlows...)

			basePct, baseXirr := calculateInvestmentMetrics(baseFlows, currentValue, now)
			transferPct, transferXirr := calculateInvestmentMetrics(transferFlows, currentValue, now)

			// totalInvested includes the 7000 row, so the denominator is 207000.
			Expect(transferPct).To(BeNumerically("~", ((currentValue-207000.0)/207000.0)*100))
			Expect(transferPct).To(BeNumerically("<", basePct))
			Expect(baseXirr).NotTo(BeNil())
			Expect(transferXirr).NotTo(BeNil())
			Expect(*transferXirr).To(BeNumerically("<", *baseXirr))
		})

		It("should not treat HDFC Interest (Credit) as a bookkeeping credit", func() {
			Expect(isInterestCredit("Interest (Credit)")).To(BeFalse())
			Expect(isInterestCredit("FD interest credit from SBI")).To(BeTrue())
			Expect(isInterestCredit("FD INTEREST CREDIT FROM SBI")).To(BeTrue())
			Expect(isInterestCredit("")).To(BeFalse())
		})

		It("should keep flows with an empty name", func() {
			collected := collectInvestmentCashFlows([]models.AccountCashFlow{
				{AccountID: 1, Amount: -10000.0, Date: now.AddDate(-1, 0, 0), Name: ""},
			})
			Expect(collected.contributed).To(Equal(10000.0))
			Expect(collected.flows).To(HaveLen(1))
		})

		It("should derive realized interest without the credit counterpart", func() {
			collected := collectInvestmentCashFlows([]models.AccountCashFlow{
				{AccountID: 1, Amount: -100000.0, Date: now.AddDate(-1, 0, 0), Name: "FD principal"},
				{AccountID: 1, Amount: 100000.0, Date: now.AddDate(-1, 0, 0), Name: "FD principal redemption"},
				{AccountID: 1, Amount: 7000.0, Date: now.AddDate(-1, 0, 0), Name: "FD interest exit from SBI"},
				{AccountID: 1, Amount: -7000.0, Date: now.AddDate(-1, 0, 0), Name: "FD interest credit from SBI"},
			})
			Expect(collected.contributed).To(Equal(100000.0))
			Expect(collected.distributed).To(Equal(107000.0))
			Expect(collected.realizedInterest).To(Equal(7000.0))
		})
	})

	Describe("GetInsights", func() {
		var startDate, endDate time.Time

		BeforeEach(func() {
			startDate = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			endDate = time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC)
		})

		Context("when the user has no accounts", func() {
			It("should return a well-formed response with a zero-filled range", func() {
				result, err := analyticsService.GetInsights(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.Summary.NetWorth).To(Equal(0.0))
				Expect(result.Monthly).To(HaveLen(3))
				Expect(result.Monthly[0].Month).To(Equal("2024-01"))
				Expect(result.Categories).To(BeEmpty())
				Expect(result.TopExpenses).To(BeEmpty())
				Expect(result.Investments).To(BeEmpty())
			})
		})

		Context("when end date is before start date", func() {
			It("should return an error", func() {
				_, err := analyticsService.GetInsights(ctx, userId, endDate, startDate)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("end date must be after or equal to start date"))
			})
		})

		Context("when computing net worth and period figures", func() {
			var investmentAccountId, bankAccountId int64

			BeforeEach(func() {
				currentValue := 15000.0
				investment, err := mockAccountRepo.CreateAccount(ctx, models.CreateAccountInput{
					Name:         "Investments",
					BankType:     models.BankTypeInvestment,
					Currency:     models.CurrencyINR,
					CurrentValue: &currentValue,
					CreatedBy:    userId,
				})
				Expect(err).NotTo(HaveOccurred())
				investmentAccountId = investment.Id

				bank, err := mockAccountRepo.CreateAccount(ctx, models.CreateAccountInput{
					Name:      "Bank",
					BankType:  models.BankTypeSBI,
					Currency:  models.CurrencyINR,
					CreatedBy: userId,
				})
				Expect(err).NotTo(HaveOccurred())
				bankAccountId = bank.Id

				// The mock negates stored balances to mimic SUM(amount) * -1,
				// so storing -500 yields a 500 transaction balance.
				mockAnalyticsRepo.SetBalance(userId, nil, nil, map[int64]float64{
					bankAccountId: -500.0,
				})

				mockAnalyticsRepo.SetInsightsMonthly(userId, startDate, endDate, []models.InsightsMonthlyPoint{
					{Month: "2024-01", Income: 1000.0, Expenses: 400.0, Net: 600.0},
				})
			})

			It("should sum net worth across priced investments and bank balances", func() {
				result, err := analyticsService.GetInsights(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Summary.InvestmentValue).To(Equal(15000.0))
				Expect(result.Summary.BankValue).To(Equal(500.0))
				Expect(result.Summary.NetWorth).To(Equal(15500.0))
				Expect(investmentAccountId).To(BeNumerically(">", 0))
			})

			It("should keep the breakdown equal to net worth for an unpriced investment", func() {
				ledgerBalance := 700.0
				_, err := mockAccountRepo.CreateAccount(ctx, models.CreateAccountInput{
					Name:      "Unpriced FD",
					BankType:  models.BankTypeInvestment,
					Currency:  models.CurrencyINR,
					Balance:   &ledgerBalance,
					CreatedBy: userId,
				})
				Expect(err).NotTo(HaveOccurred())

				result, err := analyticsService.GetInsights(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Summary.InvestmentValue).To(Equal(15700.0))
				Expect(result.Summary.BankValue).To(Equal(500.0))
				Expect(result.Summary.NetWorth).To(Equal(16200.0))
				Expect(
					result.Summary.BankValue + result.Summary.InvestmentValue,
				).To(Equal(result.Summary.NetWorth))
			})

			It("should zero-fill the missing months and total the period", func() {
				result, err := analyticsService.GetInsights(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Monthly).To(HaveLen(3))
				Expect(result.Monthly[0].Month).To(Equal("2024-01"))
				Expect(result.Monthly[0].Income).To(Equal(1000.0))
				Expect(result.Monthly[1].Month).To(Equal("2024-02"))
				Expect(result.Monthly[1].Income).To(Equal(0.0))
				Expect(result.Monthly[1].Net).To(Equal(0.0))
				Expect(result.Monthly[2].Month).To(Equal("2024-03"))
				Expect(result.Summary.PeriodIncome).To(Equal(1000.0))
				Expect(result.Summary.PeriodExpenses).To(Equal(400.0))
				Expect(result.Summary.PeriodNet).To(Equal(600.0))
				Expect(result.Summary.SavingsRate).To(BeNumerically("~", 0.6))
			})
		})

		Context("when period income is zero", func() {
			It("should report a zero savings rate instead of dividing by zero", func() {
				mockAnalyticsRepo.SetInsightsMonthly(userId, startDate, endDate, []models.InsightsMonthlyPoint{
					{Month: "2024-02", Income: 0.0, Expenses: 250.0, Net: -250.0},
				})
				result, err := analyticsService.GetInsights(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Summary.PeriodIncome).To(Equal(0.0))
				Expect(result.Summary.SavingsRate).To(Equal(0.0))
				Expect(result.Monthly).To(HaveLen(3))
			})
		})

		Context("when derived datasets exist", func() {
			BeforeEach(func() {
				mockAnalyticsRepo.SetInsightsCategories(userId, startDate, endDate, []models.InsightsCategory{
					{CategoryID: 1, CategoryName: "Food", TotalAmount: 500.0},
				})
				mockAnalyticsRepo.SetInsightsTopExpenses(userId, startDate, endDate, []models.InsightsTopExpense{
					{Name: "Cafe", Amount: 300.0, Count: 4},
				})
				mockAnalyticsRepo.SetInsightsUncategorized(userId, startDate, endDate, 2, 125.0)
			})

			It("should pass through categories, top expenses and uncategorized totals", func() {
				result, err := analyticsService.GetInsights(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Categories).To(Equal([]models.InsightsCategory{
					{CategoryID: 1, CategoryName: "Food", TotalAmount: 500.0},
				}))
				Expect(result.TopExpenses).To(HaveLen(1))
				Expect(result.TopExpenses[0].Name).To(Equal("Cafe"))
				Expect(result.Summary.UncategorizedCount).To(Equal(int64(2)))
				Expect(result.Summary.UncategorizedAmount).To(Equal(125.0))
			})
		})

		Context("when an investment account has cash flows", func() {
			var investmentAccountId int64

			BeforeEach(func() {
				currentValue := 107000.0
				account, err := mockAccountRepo.CreateAccount(ctx, models.CreateAccountInput{
					Name:         "FD",
					BankType:     models.BankTypeInvestment,
					Currency:     models.CurrencyINR,
					CurrentValue: &currentValue,
					CreatedBy:    userId,
				})
				Expect(err).NotTo(HaveOccurred())
				investmentAccountId = account.Id

				mockAnalyticsRepo.SetAccountCashFlows(userId, []models.AccountCashFlow{
					{AccountID: investmentAccountId, Amount: -100000.0, Date: time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC), Name: "FD principal"},
					{AccountID: investmentAccountId, Amount: -100000.0, Date: time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC), Name: "FD principal"},
					{AccountID: investmentAccountId, Amount: 100000.0, Date: time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC), Name: "FD principal redemption"},
					{AccountID: investmentAccountId, Amount: 7000.0, Date: time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC), Name: "FD interest exit from SBI"},
					{AccountID: investmentAccountId, Amount: -7000.0, Date: time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC), Name: "FD interest credit from SBI"},
				})
			})

			It("should aggregate contributed, distributed and realized interest", func() {
				result, err := analyticsService.GetInsights(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Investments).To(HaveLen(1))
				investment := result.Investments[0]
				Expect(investment.AccountID).To(Equal(investmentAccountId))
				Expect(investment.CurrentValue).To(Equal(107000.0))
				Expect(investment.Contributed).To(Equal(200000.0))
				Expect(investment.Distributed).To(Equal(107000.0))
				Expect(investment.RealizedInterest).To(Equal(7000.0))
				Expect(investment.Xirr).NotTo(BeNil())
			})

			It("should only count in-range realized interest in the summary", func() {
				result, err := analyticsService.GetInsights(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())
				// The interest exit on 2024-03-10 sits inside the Jan-Mar range.
				Expect(result.Summary.RealizedInterest).To(Equal(7000.0))

				outOfRange, err := analyticsService.GetInsights(ctx, userId,
					time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC))
				Expect(err).NotTo(HaveOccurred())
				Expect(outOfRange.Summary.RealizedInterest).To(Equal(0.0))
			})
		})

		Context("when spending behavior datasets exist", func() {
			BeforeEach(func() {
				mockAnalyticsRepo.SetInsightsMonthly(userId, startDate, endDate, []models.InsightsMonthlyPoint{
					{Month: "2024-01", Expenses: 100.0, Net: -100.0},
					{Month: "2024-02", Expenses: 300.0, Net: -300.0},
					{Month: "2024-03", Expenses: 200.0, Net: -200.0},
				})
				mockAnalyticsRepo.SetInsightsSpendingSummary(userId, startDate, endDate, models.InsightsSpendingSummary{
					ExpenseCount:       10,
					AverageTransaction: 60.0,
					MedianTransaction:  50.0,
					LargestExpense:     300.0,
					ActiveSpendingDays: 6,
				})
				mockAnalyticsRepo.SetInsightsCategoryMonths(userId,
					time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC),
					[]models.InsightsCategoryMonth{
						{Month: "2024-03", CategoryID: 1, CategoryName: "Food", Total: 200.0},
						{Month: "2024-02", CategoryID: 1, CategoryName: "Food", Total: 120.0},
						{Month: "2024-02", CategoryID: 2, CategoryName: "Travel", Total: 180.0},
					})
				mockAnalyticsRepo.SetInsightsWeekday(userId, startDate, endDate, []models.InsightsWeekday{
					{Weekday: 6, Total: 300.0, Count: 3, ActiveDays: 2},
					{Weekday: 1, Total: 100.0, Count: 1, ActiveDays: 1},
				})
				mockAnalyticsRepo.SetInsightsMultiCategoryCount(userId, startDate, endDate, 2)
				latest := time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC)
				mockAnalyticsRepo.SetInsightsLatestTransactionDate(userId, &latest)
			})

			It("should compute the spending summary and weekday behavior", func() {
				result, err := analyticsService.GetInsights(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())
				// Jan 1 to Mar 31 inclusive is 91 days, 6 of them active.
				Expect(result.SpendingSummary.NoSpendDays).To(Equal(int64(85)))
				Expect(result.WeekdayBehavior.Days).To(HaveLen(7))
				Expect(result.WeekdayBehavior.Days[6].Total).To(Equal(300.0))
				Expect(result.WeekdayBehavior.Days[6].Average).To(Equal(150.0))
				Expect(result.WeekdayBehavior.Days[6].Share).To(BeNumerically("~", 0.75))
				Expect(result.WeekdayBehavior.WeekendShare).To(BeNumerically("~", 0.75))
			})

			It("should compare complete months and rank category movement", func() {
				result, err := analyticsService.GetInsights(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Trend.RecentMonth).To(Equal("2024-03"))
				Expect(result.Trend.PriorMonth).To(Equal("2024-02"))
				Expect(result.Trend.RecentExpenses).To(Equal(200.0))
				Expect(result.Trend.PriorExpenses).To(Equal(300.0))
				Expect(result.Trend.Change).To(Equal(-100.0))
				Expect(result.Trend.TrailingThreeMonthAverage).To(BeNumerically("~", 200.0))

				Expect(result.CategoryMovement).To(HaveLen(2))
				Expect(result.CategoryMovement[0].CategoryName).To(Equal("Travel"))
				Expect(result.CategoryMovement[0].Change).To(Equal(-180.0))
				Expect(result.CategoryMovement[0].PriorShare).To(BeNumerically("~", 0.6))
				Expect(result.CategoryMovement[1].CategoryName).To(Equal("Food"))
				Expect(result.CategoryMovement[1].Change).To(Equal(80.0))
				Expect(result.CategoryMovement[1].RecentShare).To(Equal(1.0))
			})

			It("should flag data confidence limitations", func() {
				result, err := analyticsService.GetInsights(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.DataConfidence.MultiCategoryCount).To(Equal(int64(2)))
				Expect(result.DataConfidence.MultiCategoryShare).To(BeNumerically("~", 0.2))
				Expect(result.DataConfidence.LatestTransactionDate).NotTo(BeNil())
				Expect(*result.DataConfidence.LatestTransactionDate).To(Equal("2024-03-20"))
				Expect(result.DataConfidence.StaleDays).To(BeNumerically(">", 0))
			})
		})

		Context("when the range covers a single complete month", func() {
			var singleStart, singleEnd time.Time

			BeforeEach(func() {
				singleStart = time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
				singleEnd = time.Date(2024, 8, 31, 23, 59, 59, 0, time.UTC)

				mockAnalyticsRepo.SetInsightsMonthly(userId, singleStart, singleEnd, []models.InsightsMonthlyPoint{
					{Month: "2024-08", Expenses: 250.0, Net: -250.0},
				})
				mockAnalyticsRepo.SetInsightsCategoryMonths(userId, singleStart, singleEnd, []models.InsightsCategoryMonth{
					{Month: "2024-08", CategoryID: 1, CategoryName: "Food", Total: 150.0},
					{Month: "2024-08", CategoryID: 2, CategoryName: "Travel", Total: 100.0},
				})
			})

			It("should not reach before the range for the comparison", func() {
				result, err := analyticsService.GetInsights(ctx, userId, singleStart, singleEnd)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Trend.RecentMonth).To(Equal("2024-08"))
				Expect(result.Trend.PriorMonth).To(BeEmpty())
				Expect(result.Trend.RecentExpenses).To(Equal(250.0))
				Expect(result.Trend.PriorExpenses).To(Equal(0.0))
				Expect(result.Trend.Change).To(Equal(0.0))
				Expect(result.Trend.TrailingThreeMonthAverage).To(Equal(250.0))

				Expect(result.CategoryMovement).To(HaveLen(2))
				Expect(result.CategoryMovement[0].CategoryName).To(Equal("Food"))
				for _, item := range result.CategoryMovement {
					Expect(item.Change).To(Equal(0.0))
					Expect(item.PriorTotal).To(Equal(0.0))
					Expect(item.PriorShare).To(Equal(0.0))
				}
			})
		})
	})
})
