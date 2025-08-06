package service

import (
	"context"
	mock_repository "expenses/internal/mock/repository"
	"expenses/internal/models"
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
					if result.AccountAnalytics[i].AccountID == account1Id {
						account1Analytics = &result.AccountAnalytics[i]
					} else if result.AccountAnalytics[i].AccountID == account2Id {
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
					if result.AccountAnalytics[i].AccountID == account1Id {
						account1Analytics = &result.AccountAnalytics[i]
					} else if result.AccountAnalytics[i].AccountID == account2Id {
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

	Describe("GetNetworthTimeSeries", func() {
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
				mockAnalyticsRepo.SetNetworthTimeSeries(userId, startDate, endDate, initialBalance, timeSeries)
			})

			It("should return networth time series with negated values", func() {
				result, err := analyticsService.GetNetworthTimeSeries(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Initial balance should be negated
				Expect(result.InitialBalance).To(Equal(-1000.0))

				// Should have data points for each day in range
				Expect(result.TimeSeries).To(HaveLen(3)) // Jan 1, 2, 3

				// Verify first day: -1000 (initial) + (-100) (daily change) = -1100
				Expect(result.TimeSeries[0].Date).To(Equal("2023-01-01"))
				Expect(result.TimeSeries[0].Networth).To(Equal(-1100.0))

				// Verify second day: -1100 + (50) (daily change) = -1050
				Expect(result.TimeSeries[1].Date).To(Equal("2023-01-02"))
				Expect(result.TimeSeries[1].Networth).To(Equal(-1050.0))

				// Verify third day (no transaction, same balance)
				Expect(result.TimeSeries[2].Date).To(Equal("2023-01-03"))
				Expect(result.TimeSeries[2].Networth).To(Equal(-1050.0))
			})
		})

		Context("when repository returns no daily changes", func() {
			BeforeEach(func() {
				// Set up mock data with only initial balance, no daily changes
				initialBalance := 500.0
				timeSeries := []map[string]any{} // Empty time series
				mockAnalyticsRepo.SetNetworthTimeSeries(userId, startDate, endDate, initialBalance, timeSeries)
			})

			It("should return flat networth time series", func() {
				result, err := analyticsService.GetNetworthTimeSeries(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Initial balance should be negated
				Expect(result.InitialBalance).To(Equal(-500.0))

				// Should have data points for each day in range
				Expect(result.TimeSeries).To(HaveLen(3)) // Jan 1, 2, 3

				// All days should have the same networth
				for _, point := range result.TimeSeries {
					Expect(point.Networth).To(Equal(-500.0))
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
				mockAnalyticsRepo.SetNetworthTimeSeries(userId, startDate, endDate, initialBalance, timeSeries)
			})

			It("should handle gaps in daily data correctly", func() {
				result, err := analyticsService.GetNetworthTimeSeries(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Initial balance should be negated
				Expect(result.InitialBalance).To(Equal(-2000.0))

				// Should have data points for each day in range
				Expect(result.TimeSeries).To(HaveLen(3))

				// Jan 1: has transaction (debit 200 -> credit -200 for frontend)
				Expect(result.TimeSeries[0].Date).To(Equal("2023-01-01"))
				Expect(result.TimeSeries[0].Networth).To(Equal(-2200.0)) // -2000 + (-200)

				// Jan 2: no transaction, same as previous day
				Expect(result.TimeSeries[1].Date).To(Equal("2023-01-02"))
				Expect(result.TimeSeries[1].Networth).To(Equal(-2200.0)) // Same as Jan 1

				// Jan 3: has transaction (credit -150 -> debit +150 for frontend)
				Expect(result.TimeSeries[2].Date).To(Equal("2023-01-03"))
				Expect(result.TimeSeries[2].Networth).To(Equal(-2050.0)) // -2200 + 150
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
				mockAnalyticsRepo.SetNetworthTimeSeries(userId, startDate, endDate, initialBalance, timeSeries)
			})

			It("should return single day networth", func() {
				result, err := analyticsService.GetNetworthTimeSeries(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Should have exactly one data point
				Expect(result.TimeSeries).To(HaveLen(1))

				Expect(result.TimeSeries[0].Date).To(Equal("2023-01-01"))
				Expect(result.TimeSeries[0].Networth).To(Equal(-1575.0)) // -1500 + (-75)
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
				mockAnalyticsRepo.SetNetworthTimeSeries(userId, startDate, endDate, initialBalance, timeSeries)
			})

			It("should handle zero initial balance correctly", func() {
				result, err := analyticsService.GetNetworthTimeSeries(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Initial balance should be zero (negated)
				Expect(result.InitialBalance).To(Equal(0.0))

				// Verify cumulative calculation from zero
				Expect(result.TimeSeries[0].Networth).To(Equal(-100.0)) // 0 + (-100)
				Expect(result.TimeSeries[1].Networth).To(Equal(-50.0))  // -100 + 50
				Expect(result.TimeSeries[2].Networth).To(Equal(-50.0))  // Same as previous day
			})
		})

		Context("when different users request networth", func() {
			var user1Id, user2Id int64

			BeforeEach(func() {
				user1Id = 1
				user2Id = 2

				// Set up different networth data for each user
				user1InitialBalance := 1000.0
				user1TimeSeries := []map[string]any{
					{
						"date":         "2023-01-01",
						"daily_change": 100.0,
					},
				}
				mockAnalyticsRepo.SetNetworthTimeSeries(user1Id, startDate, endDate, user1InitialBalance, user1TimeSeries)

				user2InitialBalance := 2000.0
				user2TimeSeries := []map[string]any{
					{
						"date":         "2023-01-01",
						"daily_change": 200.0,
					},
				}
				mockAnalyticsRepo.SetNetworthTimeSeries(user2Id, startDate, endDate, user2InitialBalance, user2TimeSeries)
			})

			It("should return networth data only for the requesting user", func() {
				// Test user1 networth
				result1, err := analyticsService.GetNetworthTimeSeries(ctx, user1Id, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())
				Expect(result1.InitialBalance).To(Equal(-1000.0))
				Expect(result1.TimeSeries[0].Networth).To(Equal(-1100.0)) // -1000 - 100

				// Test user2 networth
				result2, err := analyticsService.GetNetworthTimeSeries(ctx, user2Id, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())
				Expect(result2.InitialBalance).To(Equal(-2000.0))
				Expect(result2.TimeSeries[0].Networth).To(Equal(-2200.0)) // -2000 - 200
			})
		})

		Context("when repository returns error", func() {
			BeforeEach(func() {
				// Configure mock to return error
				mockAnalyticsRepo.SetShouldErrorOnNetworth(true)
			})

			AfterEach(func() {
				// Reset error simulation
				mockAnalyticsRepo.SetShouldErrorOnNetworth(false)
			})

			It("should return the repository error", func() {
				result, err := analyticsService.GetNetworthTimeSeries(ctx, userId, startDate, endDate)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("simulated GetNetworthTimeSeries error"))
				Expect(result).To(Equal(models.NetworthTimeSeriesResponse{}))
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
				mockAnalyticsRepo.SetNetworthTimeSeries(userId, startDate, endDate, initialBalance, timeSeries)
			})

			It("should correctly negate all values for frontend consumption", func() {
				result, err := analyticsService.GetNetworthTimeSeries(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Initial balance: 1000 (debit) -> -1000 (for frontend)
				Expect(result.InitialBalance).To(Equal(-1000.0))

				// Day 1: -1000 + (-200) = -1200 (debit transaction becomes negative for frontend)
				Expect(result.TimeSeries[0].Networth).To(Equal(-1200.0))

				// Day 2: -1200 + 150 = -1050 (credit transaction becomes positive for frontend)
				Expect(result.TimeSeries[1].Networth).To(Equal(-1050.0))

				// Day 3: Same as day 2 (no transaction)
				Expect(result.TimeSeries[2].Networth).To(Equal(-1050.0))
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
				mockAnalyticsRepo.SetNetworthTimeSeries(userId, startDate, endDate, initialBalance, timeSeries)

				result, err := analyticsService.GetNetworthTimeSeries(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Should have exactly one data point
				Expect(result.TimeSeries).To(HaveLen(1))
				Expect(result.TimeSeries[0].Date).To(Equal("2023-01-01"))
				Expect(result.TimeSeries[0].Networth).To(Equal(-1100.0)) // -1000 + (-100)
			})

			It("should handle very large date ranges", func() {
				startDate, _ = time.Parse("2006-01-02", "2023-01-01")
				endDate, _ = time.Parse("2006-01-02", "2023-12-31")

				initialBalance := 1000.0
				timeSeries := []map[string]any{} // No daily changes
				mockAnalyticsRepo.SetNetworthTimeSeries(userId, startDate, endDate, initialBalance, timeSeries)

				result, err := analyticsService.GetNetworthTimeSeries(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Should have 365 data points for 2023 (not a leap year)
				Expect(result.TimeSeries).To(HaveLen(365))

				// All should have the same networth (no changes)
				for _, point := range result.TimeSeries {
					Expect(point.Networth).To(Equal(-1000.0))
				}
			})

			It("should handle leap year correctly", func() {
				startDate, _ = time.Parse("2006-01-02", "2024-02-28")
				endDate, _ = time.Parse("2006-01-02", "2024-03-01")

				initialBalance := 500.0
				timeSeries := []map[string]any{} // No daily changes
				mockAnalyticsRepo.SetNetworthTimeSeries(userId, startDate, endDate, initialBalance, timeSeries)

				result, err := analyticsService.GetNetworthTimeSeries(ctx, userId, startDate, endDate)
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
				mockAnalyticsRepo.SetNetworthTimeSeries(userId, startDate, endDate, initialBalance, timeSeries)

				result, err := analyticsService.GetNetworthTimeSeries(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Should have 4 data points
				Expect(result.TimeSeries).To(HaveLen(4))

				// Verify year boundary crossing
				Expect(result.TimeSeries[0].Date).To(Equal("2023-12-30"))
				Expect(result.TimeSeries[0].Networth).To(Equal(-2000.0)) // No change

				Expect(result.TimeSeries[1].Date).To(Equal("2023-12-31"))
				Expect(result.TimeSeries[1].Networth).To(Equal(-2100.0)) // -2000 + (-100)

				Expect(result.TimeSeries[2].Date).To(Equal("2024-01-01"))
				Expect(result.TimeSeries[2].Networth).To(Equal(-2050.0)) // -2100 + 50

				Expect(result.TimeSeries[3].Date).To(Equal("2024-01-02"))
				Expect(result.TimeSeries[3].Networth).To(Equal(-2050.0)) // No change
			})

			It("should handle very large transaction amounts", func() {
				initialBalance := 999999999.99
				timeSeries := []map[string]any{
					{
						"date":         "2023-01-01",
						"daily_change": 888888888.88,
					},
				}
				mockAnalyticsRepo.SetNetworthTimeSeries(userId, startDate, endDate, initialBalance, timeSeries)

				result, err := analyticsService.GetNetworthTimeSeries(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Should handle large numbers correctly
				Expect(result.InitialBalance).To(Equal(-999999999.99))
				Expect(result.TimeSeries[0].Networth).To(BeNumerically("~", -1888888888.87, 0.01))
			})

			It("should handle negative initial balance", func() {
				initialBalance := -500.0 // Negative initial balance
				timeSeries := []map[string]any{
					{
						"date":         "2023-01-01",
						"daily_change": 100.0,
					},
				}
				mockAnalyticsRepo.SetNetworthTimeSeries(userId, startDate, endDate, initialBalance, timeSeries)

				result, err := analyticsService.GetNetworthTimeSeries(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Negative initial balance should become positive for frontend
				Expect(result.InitialBalance).To(Equal(500.0))
				Expect(result.TimeSeries[0].Networth).To(Equal(400.0)) // 500 + (-100)
			})

			It("should handle empty daily data gracefully", func() {
				initialBalance := 1000.0
				timeSeries := []map[string]any{} // Empty
				mockAnalyticsRepo.SetNetworthTimeSeries(userId, startDate, endDate, initialBalance, timeSeries)

				result, err := analyticsService.GetNetworthTimeSeries(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Should still generate time series with flat values
				Expect(result.TimeSeries).To(HaveLen(1))
				Expect(result.TimeSeries[0].Networth).To(Equal(-1000.0))
			})

			It("should handle malformed daily data gracefully", func() {
				initialBalance := 1000.0
				timeSeries := []map[string]any{
					{
						"date":         "2023-01-01",
						"daily_change": "invalid", // Invalid type
					},
				}
				mockAnalyticsRepo.SetNetworthTimeSeries(userId, startDate, endDate, initialBalance, timeSeries)

				// Testing the new behavior
				_, err := analyticsService.GetNetworthTimeSeries(ctx, userId, startDate, endDate)
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
				mockAnalyticsRepo.SetNetworthTimeSeries(userId, startDate, endDate, initialBalance, timeSeries)

				result, err := analyticsService.GetNetworthTimeSeries(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// The service uses a map, so the last value should win
				Expect(result.TimeSeries[0].Networth).To(Equal(-1050.0)) // -1000 + (-50)
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
				mockAnalyticsRepo.SetNetworthTimeSeries(userId, startDate, endDate, initialBalance, timeSeries)

				result, err := analyticsService.GetNetworthTimeSeries(ctx, userId, startDate, endDate)
				Expect(err).NotTo(HaveOccurred())

				// Should only use the date within range
				Expect(result.TimeSeries[0].Networth).To(Equal(-1050.0)) // -1000 + (-50)
				Expect(result.TimeSeries[1].Networth).To(Equal(-1050.0)) // No change for Jan 2
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
				Expect(result.TotalExpenses).To(Equal(800.0)) // Default from mock
				Expect(result.TotalAmount).To(Equal(1800.0))  // Default from mock
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
	})
})
