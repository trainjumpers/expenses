package service

import (
	"errors"
	"expenses/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var ErrMockError = errors.New("mock error")

// Mock Analytics Repository
type MockAnalyticsRepository struct {
	spendingOverview      *models.SpendingOverviewResponse
	categorySpending      *models.CategorySpendingResponse
	spendingTrends        *models.SpendingTrendsResponse
	accountSpending       *models.AccountSpendingResponse
	topTransactions       *models.TopTransactionsResponse
	monthlyComparison     *models.MonthlyComparisonResponse
	recurringTransactions *models.RecurringTransactionsResponse
	shouldError           bool
}

func (m *MockAnalyticsRepository) GetSpendingOverview(ctx *gin.Context, query models.AnalyticsQuery) (*models.SpendingOverviewResponse, error) {
	if m.shouldError {
		return nil, ErrMockError
	}
	return m.spendingOverview, nil
}

func (m *MockAnalyticsRepository) GetCategorySpending(ctx *gin.Context, query models.AnalyticsQuery) (*models.CategorySpendingResponse, error) {
	if m.shouldError {
		return nil, ErrMockError
	}
	return m.categorySpending, nil
}

func (m *MockAnalyticsRepository) GetSpendingTrends(ctx *gin.Context, query models.AnalyticsQuery, granularity string) (*models.SpendingTrendsResponse, error) {
	if m.shouldError {
		return nil, ErrMockError
	}
	return m.spendingTrends, nil
}

func (m *MockAnalyticsRepository) GetAccountSpending(ctx *gin.Context, query models.AnalyticsQuery) (*models.AccountSpendingResponse, error) {
	if m.shouldError {
		return nil, ErrMockError
	}
	return m.accountSpending, nil
}

func (m *MockAnalyticsRepository) GetTopTransactions(ctx *gin.Context, query models.AnalyticsQuery, limit int) (*models.TopTransactionsResponse, error) {
	if m.shouldError {
		return nil, ErrMockError
	}
	return m.topTransactions, nil
}

func (m *MockAnalyticsRepository) GetMonthlyComparison(ctx *gin.Context, query models.AnalyticsQuery) (*models.MonthlyComparisonResponse, error) {
	if m.shouldError {
		return nil, ErrMockError
	}
	return m.monthlyComparison, nil
}

func (m *MockAnalyticsRepository) GetRecurringTransactions(ctx *gin.Context, query models.AnalyticsQuery) (*models.RecurringTransactionsResponse, error) {
	if m.shouldError {
		return nil, ErrMockError
	}
	return m.recurringTransactions, nil
}

var _ = Describe("AnalyticsService", func() {
	var (
		analyticsService AnalyticsServiceInterface
		mockRepo         *MockAnalyticsRepository
		ctx              *gin.Context
		validQuery       models.AnalyticsQuery
	)

	BeforeEach(func() {
		mockRepo = &MockAnalyticsRepository{}
		analyticsService = NewAnalyticsService(mockRepo)
		ctx = &gin.Context{}
		
		validQuery = models.AnalyticsQuery{
			TimeRange: models.TimeRangeMonth,
			CreatedBy: 1,
		}

		// Set up mock data
		mockRepo.spendingOverview = &models.SpendingOverviewResponse{
			TotalExpenses:    10000.0,
			TotalIncome:      15000.0,
			NetAmount:        5000.0,
			TransactionCount: 50,
			AverageExpense:   200.0,
			AverageIncome:    500.0,
			Period:           "month",
		}

		mockRepo.categorySpending = &models.CategorySpendingResponse{
			Categories: []models.CategorySpendingItem{
				{
					CategoryId:   1,
					CategoryName: "Food",
					Amount:       5000.0,
					Percentage:   50.0,
					Count:        25,
				},
			},
			Uncategorized: models.CategorySpendingItem{
				CategoryName: "Uncategorized",
				Amount:       2000.0,
				Percentage:   20.0,
				Count:        10,
			},
			TotalAmount: 10000.0,
			TotalCount:  50,
		}

		now := time.Now()
		mockRepo.spendingTrends = &models.SpendingTrendsResponse{
			DataPoints: []models.TimeSeriesDataPoint{
				{
					Date:     now,
					Amount:   1000.0,
					Count:    5,
					Income:   1500.0,
					Expenses: 1000.0,
				},
			},
			Period:      "month",
			Granularity: "daily",
		}

		bankName := "HDFC Bank"
		mockRepo.accountSpending = &models.AccountSpendingResponse{
			Accounts: []models.AccountSpendingItem{
				{
					AccountId:   1,
					AccountName: "HDFC Savings",
					BankName:    &bankName,
					Amount:      8000.0,
					Percentage:  80.0,
					Count:       40,
				},
			},
			TotalAmount: 10000.0,
			TotalCount:  50,
		}

		mockRepo.topTransactions = &models.TopTransactionsResponse{
			TopExpenses: []models.TopTransactionItem{
				{
					Id:          1,
					Name:        "Grocery Shopping",
					Amount:      2000.0,
					Date:        now,
					AccountName: "HDFC Savings",
					Categories:  []string{"Food"},
				},
			},
			TopIncome: []models.TopTransactionItem{
				{
					Id:          2,
					Name:        "Salary",
					Amount:      50000.0,
					Date:        now,
					AccountName: "HDFC Savings",
					Categories:  []string{"Income"},
				},
			},
			Limit: 10,
		}

		mockRepo.monthlyComparison = &models.MonthlyComparisonResponse{
			Months: []models.MonthlyComparisonItem{
				{
					Month:        now.AddDate(0, -1, 0),
					MonthName:    "January 2025",
					Amount:       8000.0,
					Count:        40,
					Change:       -10.0,
					ChangeAmount: -1000.0,
				},
				{
					Month:        now,
					MonthName:    "February 2025",
					Amount:       9000.0,
					Count:        45,
					Change:       12.5,
					ChangeAmount: 1000.0,
				},
			},
			Period:      "month",
			TotalMonths: 2,
		}

		mockRepo.recurringTransactions = &models.RecurringTransactionsResponse{
			Patterns: []models.RecurringTransactionPattern{
				{
					Pattern:        "netflix subscription",
					Amount:         799.0,
					Frequency:      "monthly",
					Confidence:     0.9,
					TransactionIds: []int64{1, 2, 3},
					Count:          3,
				},
			},
			TotalAmount: 2397.0,
			Count:       3,
		}
	})

	Describe("GetSpendingOverview", func() {
		Context("with valid query", func() {
			It("should return spending overview", func() {
				result, err := analyticsService.GetSpendingOverview(ctx, validQuery)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(result.TotalExpenses).To(Equal(10000.0))
				Expect(result.TotalIncome).To(Equal(15000.0))
				Expect(result.NetAmount).To(Equal(5000.0))
				Expect(result.TransactionCount).To(Equal(50))
			})
		})

		Context("with invalid query", func() {
			It("should return error for missing created_by", func() {
				invalidQuery := models.AnalyticsQuery{
					TimeRange: models.TimeRangeMonth,
					CreatedBy: 0,
				}
				
				result, err := analyticsService.GetSpendingOverview(ctx, invalidQuery)
				
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("created_by is required"))
			})

			It("should return error for custom time range without dates", func() {
				invalidQuery := models.AnalyticsQuery{
					TimeRange: models.TimeRangeCustom,
					CreatedBy: 1,
				}
				
				result, err := analyticsService.GetSpendingOverview(ctx, invalidQuery)
				
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("start_date and end_date are required"))
			})
		})

		Context("when repository returns error", func() {
			It("should propagate the error", func() {
				mockRepo.shouldError = true
				
				result, err := analyticsService.GetSpendingOverview(ctx, validQuery)
				
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})
		})
	})

	Describe("GetSpendingTrends", func() {
		Context("with valid query and granularity", func() {
			It("should return spending trends", func() {
				result, err := analyticsService.GetSpendingTrends(ctx, validQuery, "daily")
				
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(result.DataPoints).To(HaveLen(1))
				Expect(result.Granularity).To(Equal("daily"))
				Expect(result.Period).To(Equal("month"))
			})
		})

		Context("with invalid granularity", func() {
			It("should return error", func() {
				result, err := analyticsService.GetSpendingTrends(ctx, validQuery, "invalid")
				
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("invalid granularity"))
			})
		})
	})

	Describe("GetAnalyticsSummary", func() {
		Context("with valid query", func() {
			It("should return comprehensive analytics summary", func() {
				result, err := analyticsService.GetAnalyticsSummary(ctx, validQuery)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(result.Overview.TotalExpenses).To(Equal(10000.0))
				Expect(result.CategoryBreakdown.Categories).To(HaveLen(1))
				Expect(result.AccountBreakdown.Accounts).To(HaveLen(1))
				Expect(result.TopTransactions.TopExpenses).To(HaveLen(1))
				Expect(result.MonthlyComparison.Months).To(HaveLen(2))
				Expect(result.RecurringPatterns.Patterns).To(HaveLen(1))
				Expect(result.Period).To(Equal("month"))
				Expect(result.GeneratedAt).ToNot(BeZero())
			})
		})
	})
})
