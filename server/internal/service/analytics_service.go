package service

import (
	"expenses/internal/models"
	"expenses/internal/repository"
	"fmt"
	"math"
	"time"

	"github.com/gin-gonic/gin"
)

type AnalyticsServiceInterface interface {
	GetSpendingOverview(ctx *gin.Context, query models.AnalyticsQuery) (*models.SpendingOverviewResponse, error)
	GetCategorySpending(ctx *gin.Context, query models.AnalyticsQuery) (*models.CategorySpendingResponse, error)
	GetSpendingTrends(ctx *gin.Context, query models.AnalyticsQuery, granularity string) (*models.SpendingTrendsResponse, error)
	GetAccountSpending(ctx *gin.Context, query models.AnalyticsQuery) (*models.AccountSpendingResponse, error)
	GetTopTransactions(ctx *gin.Context, query models.AnalyticsQuery, limit int) (*models.TopTransactionsResponse, error)
	GetMonthlyComparison(ctx *gin.Context, query models.AnalyticsQuery) (*models.MonthlyComparisonResponse, error)
	GetRecurringTransactions(ctx *gin.Context, query models.AnalyticsQuery) (*models.RecurringTransactionsResponse, error)
	GetAnalyticsSummary(ctx *gin.Context, query models.AnalyticsQuery) (*models.AnalyticsSummaryResponse, error)
	GetAnalyticsInsights(ctx *gin.Context, query models.AnalyticsQuery) (*models.AnalyticsInsightsResponse, error)
}

type AnalyticsService struct {
	analyticsRepo repository.AnalyticsRepositoryInterface
}

func NewAnalyticsService(analyticsRepo repository.AnalyticsRepositoryInterface) AnalyticsServiceInterface {
	return &AnalyticsService{
		analyticsRepo: analyticsRepo,
	}
}

func (s *AnalyticsService) GetSpendingOverview(ctx *gin.Context, query models.AnalyticsQuery) (*models.SpendingOverviewResponse, error) {
	if err := s.validateAnalyticsQuery(query); err != nil {
		return nil, err
	}

	return s.analyticsRepo.GetSpendingOverview(ctx, query)
}

func (s *AnalyticsService) GetCategorySpending(ctx *gin.Context, query models.AnalyticsQuery) (*models.CategorySpendingResponse, error) {
	if err := s.validateAnalyticsQuery(query); err != nil {
		return nil, err
	}

	return s.analyticsRepo.GetCategorySpending(ctx, query)
}

func (s *AnalyticsService) GetSpendingTrends(ctx *gin.Context, query models.AnalyticsQuery, granularity string) (*models.SpendingTrendsResponse, error) {
	if err := s.validateAnalyticsQuery(query); err != nil {
		return nil, err
	}

	// Validate granularity
	validGranularities := map[string]bool{
		"daily":   true,
		"weekly":  true,
		"monthly": true,
	}
	if !validGranularities[granularity] {
		return nil, fmt.Errorf("invalid granularity: %s. Must be one of: daily, weekly, monthly", granularity)
	}

	// Auto-adjust granularity based on time range for better UX
	if granularity == "" {
		granularity = s.getOptimalGranularity(query.TimeRange)
	}

	return s.analyticsRepo.GetSpendingTrends(ctx, query, granularity)
}

func (s *AnalyticsService) GetAccountSpending(ctx *gin.Context, query models.AnalyticsQuery) (*models.AccountSpendingResponse, error) {
	if err := s.validateAnalyticsQuery(query); err != nil {
		return nil, err
	}

	return s.analyticsRepo.GetAccountSpending(ctx, query)
}

func (s *AnalyticsService) GetTopTransactions(ctx *gin.Context, query models.AnalyticsQuery, limit int) (*models.TopTransactionsResponse, error) {
	if err := s.validateAnalyticsQuery(query); err != nil {
		return nil, err
	}

	// Validate and set default limit
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	return s.analyticsRepo.GetTopTransactions(ctx, query, limit)
}

func (s *AnalyticsService) GetMonthlyComparison(ctx *gin.Context, query models.AnalyticsQuery) (*models.MonthlyComparisonResponse, error) {
	if err := s.validateAnalyticsQuery(query); err != nil {
		return nil, err
	}

	return s.analyticsRepo.GetMonthlyComparison(ctx, query)
}

func (s *AnalyticsService) GetRecurringTransactions(ctx *gin.Context, query models.AnalyticsQuery) (*models.RecurringTransactionsResponse, error) {
	if err := s.validateAnalyticsQuery(query); err != nil {
		return nil, err
	}

	return s.analyticsRepo.GetRecurringTransactions(ctx, query)
}

func (s *AnalyticsService) GetAnalyticsSummary(ctx *gin.Context, query models.AnalyticsQuery) (*models.AnalyticsSummaryResponse, error) {
	if err := s.validateAnalyticsQuery(query); err != nil {
		return nil, err
	}

	// Fetch all analytics data concurrently for better performance
	type result struct {
		overview          *models.SpendingOverviewResponse
		categoryBreakdown *models.CategorySpendingResponse
		accountBreakdown  *models.AccountSpendingResponse
		topTransactions   *models.TopTransactionsResponse
		monthlyComparison *models.MonthlyComparisonResponse
		recurringPatterns *models.RecurringTransactionsResponse
		err               error
	}

	resultChan := make(chan result, 1)

	go func() {
		var res result

		// Get spending overview
		res.overview, res.err = s.analyticsRepo.GetSpendingOverview(ctx, query)
		if res.err != nil {
			resultChan <- res
			return
		}

		// Get category breakdown
		res.categoryBreakdown, res.err = s.analyticsRepo.GetCategorySpending(ctx, query)
		if res.err != nil {
			resultChan <- res
			return
		}

		// Get account breakdown
		res.accountBreakdown, res.err = s.analyticsRepo.GetAccountSpending(ctx, query)
		if res.err != nil {
			resultChan <- res
			return
		}

		// Get top transactions
		res.topTransactions, res.err = s.analyticsRepo.GetTopTransactions(ctx, query, 5)
		if res.err != nil {
			resultChan <- res
			return
		}

		// Get monthly comparison
		res.monthlyComparison, res.err = s.analyticsRepo.GetMonthlyComparison(ctx, query)
		if res.err != nil {
			resultChan <- res
			return
		}

		// Get recurring patterns
		res.recurringPatterns, res.err = s.analyticsRepo.GetRecurringTransactions(ctx, query)
		if res.err != nil {
			resultChan <- res
			return
		}

		resultChan <- res
	}()

	res := <-resultChan
	if res.err != nil {
		return nil, fmt.Errorf("failed to get analytics summary: %w", res.err)
	}

	return &models.AnalyticsSummaryResponse{
		Overview:            *res.overview,
		CategoryBreakdown:   *res.categoryBreakdown,
		AccountBreakdown:    *res.accountBreakdown,
		TopTransactions:     *res.topTransactions,
		MonthlyComparison:   *res.monthlyComparison,
		RecurringPatterns:   *res.recurringPatterns,
		Period:              string(query.TimeRange),
		GeneratedAt:         time.Now(),
	}, nil
}

func (s *AnalyticsService) GetAnalyticsInsights(ctx *gin.Context, query models.AnalyticsQuery) (*models.AnalyticsInsightsResponse, error) {
	if err := s.validateAnalyticsQuery(query); err != nil {
		return nil, err
	}

	// Get data needed for insights
	overview, err := s.analyticsRepo.GetSpendingOverview(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get overview for insights: %w", err)
	}

	categoryBreakdown, err := s.analyticsRepo.GetCategorySpending(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get category breakdown for insights: %w", err)
	}

	monthlyComparison, err := s.analyticsRepo.GetMonthlyComparison(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly comparison for insights: %w", err)
	}

	recurringPatterns, err := s.analyticsRepo.GetRecurringTransactions(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get recurring patterns for insights: %w", err)
	}

	// Generate insights based on the data
	insights := s.generateInsights(overview, categoryBreakdown, monthlyComparison, recurringPatterns)

	return &models.AnalyticsInsightsResponse{
		Insights:    insights,
		Count:       len(insights),
		GeneratedAt: time.Now(),
	}, nil
}

// Helper methods

func (s *AnalyticsService) validateAnalyticsQuery(query models.AnalyticsQuery) error {
	if query.CreatedBy <= 0 {
		return fmt.Errorf("created_by is required")
	}

	if query.TimeRange == models.TimeRangeCustom {
		if query.StartDate == nil || query.EndDate == nil {
			return fmt.Errorf("start_date and end_date are required for custom time range")
		}
		if query.EndDate.Before(*query.StartDate) {
			return fmt.Errorf("end_date must be after start_date")
		}
	}

	return nil
}

func (s *AnalyticsService) getOptimalGranularity(timeRange models.AnalyticsTimeRange) string {
	switch timeRange {
	case models.TimeRangeWeek:
		return "daily"
	case models.TimeRangeMonth:
		return "daily"
	case models.TimeRangeQuarter:
		return "weekly"
	case models.TimeRangeYear:
		return "monthly"
	default:
		return "daily"
	}
}

func (s *AnalyticsService) generateInsights(
	overview *models.SpendingOverviewResponse,
	categoryBreakdown *models.CategorySpendingResponse,
	monthlyComparison *models.MonthlyComparisonResponse,
	recurringPatterns *models.RecurringTransactionsResponse,
) []models.AnalyticsInsight {
	var insights []models.AnalyticsInsight
	now := time.Now()

	// High spending insight
	if overview.TotalExpenses > overview.AverageExpense*30 { // Arbitrary threshold
		insights = append(insights, models.AnalyticsInsight{
			Type:        "warning",
			Title:       "High Spending Alert",
			Description: fmt.Sprintf("Your expenses of ₹%.2f are significantly higher than your average. Consider reviewing your spending patterns.", overview.TotalExpenses),
			Actionable:  true,
			Priority:    4,
			CreatedAt:   now,
		})
	}

	// Uncategorized transactions insight
	if categoryBreakdown.Uncategorized.Amount > 0 && categoryBreakdown.Uncategorized.Percentage > 20 {
		insights = append(insights, models.AnalyticsInsight{
			Type:        "info",
			Title:       "Uncategorized Transactions",
			Description: fmt.Sprintf("%.1f%% of your expenses (₹%.2f) are uncategorized. Categorizing them will help you better understand your spending.", categoryBreakdown.Uncategorized.Percentage, categoryBreakdown.Uncategorized.Amount),
			Actionable:  true,
			Priority:    3,
			CreatedAt:   now,
		})
	}

	// Top category spending insight
	if len(categoryBreakdown.Categories) > 0 {
		topCategory := categoryBreakdown.Categories[0]
		if topCategory.Percentage > 40 {
			insights = append(insights, models.AnalyticsInsight{
				Type:        "warning",
				Title:       "Category Concentration",
				Description: fmt.Sprintf("%.1f%% of your spending is in %s (₹%.2f). Consider if this aligns with your budget goals.", topCategory.Percentage, topCategory.CategoryName, topCategory.Amount),
				Actionable:  true,
				Priority:    3,
				CreatedAt:   now,
			})
		}
	}

	// Monthly trend insight
	if len(monthlyComparison.Months) >= 2 {
		lastMonth := monthlyComparison.Months[len(monthlyComparison.Months)-1]
		if lastMonth.Change > 20 {
			insights = append(insights, models.AnalyticsInsight{
				Type:        "warning",
				Title:       "Spending Increase",
				Description: fmt.Sprintf("Your spending increased by %.1f%% (₹%.2f) compared to the previous month. Review what drove this increase.", lastMonth.Change, lastMonth.ChangeAmount),
				Actionable:  true,
				Priority:    4,
				CreatedAt:   now,
			})
		} else if lastMonth.Change < -20 {
			insights = append(insights, models.AnalyticsInsight{
				Type:        "success",
				Title:       "Great Savings!",
				Description: fmt.Sprintf("You reduced your spending by %.1f%% (₹%.2f) compared to the previous month. Keep up the good work!", math.Abs(lastMonth.Change), math.Abs(lastMonth.ChangeAmount)),
				Actionable:  false,
				Priority:    2,
				CreatedAt:   now,
			})
		}
	}

	// Recurring transactions insight
	if len(recurringPatterns.Patterns) > 0 {
		insights = append(insights, models.AnalyticsInsight{
			Type:        "info",
			Title:       "Recurring Expenses Detected",
			Description: fmt.Sprintf("Found %d recurring expense patterns totaling ₹%.2f. These might be subscriptions or regular bills you can optimize.", len(recurringPatterns.Patterns), recurringPatterns.TotalAmount),
			Actionable:  true,
			Priority:    2,
			CreatedAt:   now,
		})
	}

	// Income vs expenses insight
	if overview.TotalExpenses > overview.TotalIncome {
		deficit := overview.TotalExpenses - overview.TotalIncome
		insights = append(insights, models.AnalyticsInsight{
			Type:        "warning",
			Title:       "Spending Exceeds Income",
			Description: fmt.Sprintf("Your expenses (₹%.2f) exceed your income (₹%.2f) by ₹%.2f. Consider reducing expenses or increasing income.", overview.TotalExpenses, overview.TotalIncome, deficit),
			Actionable:  true,
			Priority:    5,
			CreatedAt:   now,
		})
	} else if overview.TotalIncome > overview.TotalExpenses {
		surplus := overview.TotalIncome - overview.TotalExpenses
		savingsRate := (surplus / overview.TotalIncome) * 100
		if savingsRate > 20 {
			insights = append(insights, models.AnalyticsInsight{
				Type:        "success",
				Title:       "Excellent Savings Rate",
				Description: fmt.Sprintf("You're saving %.1f%% of your income (₹%.2f). This is excellent financial discipline!", savingsRate, surplus),
				Actionable:  false,
				Priority:    1,
				CreatedAt:   now,
			})
		}
	}

	// Low transaction count insight
	if overview.TransactionCount < 5 {
		insights = append(insights, models.AnalyticsInsight{
			Type:        "tip",
			Title:       "Track More Transactions",
			Description: fmt.Sprintf("You have only %d transactions recorded. Tracking more of your expenses will provide better insights.", overview.TransactionCount),
			Actionable:  true,
			Priority:    1,
			CreatedAt:   now,
		})
	}

	return insights
}
