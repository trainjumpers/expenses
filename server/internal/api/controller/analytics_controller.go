package controller

import (
	"expenses/internal/config"
	"expenses/internal/models"
	"expenses/internal/service"
	"expenses/pkg/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AnalyticsController struct {
	*BaseController
	analyticsService service.AnalyticsServiceInterface
}

func NewAnalyticsController(cfg *config.Config, analyticsService service.AnalyticsServiceInterface) *AnalyticsController {
	return &AnalyticsController{
		BaseController:   NewBaseController(cfg),
		analyticsService: analyticsService,
	}
}

// GetSpendingOverview godoc
// @Summary Get spending overview
// @Description Get comprehensive spending overview including totals, averages, and transaction counts
// @Tags analytics
// @Accept json
// @Produce json
// @Param query body models.AnalyticsQuery true "Analytics query parameters"
// @Success 200 {object} models.SpendingOverviewResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /analytics/overview [post]
func (c *AnalyticsController) GetSpendingOverview(ctx *gin.Context) {
	var query models.AnalyticsQuery
	if err := c.BindJSON(ctx, &query); err != nil {
		return // BindJSON already sends the error response
	}

	logger.Debugf("Getting spending overview for user with query: %+v", query)

	overview, err := c.analyticsService.GetSpendingOverview(ctx, query)
	if err != nil {
		logger.Errorf("Failed to get spending overview: %v", err)
		c.HandleError(ctx, err)
		return
	}

	logger.Debugf("Successfully retrieved spending overview")
	c.SendSuccess(ctx, http.StatusOK, "Spending overview retrieved successfully", overview)
}

// GetCategorySpending godoc
// @Summary Get category spending breakdown
// @Description Get spending breakdown by categories with percentages
// @Tags analytics
// @Accept json
// @Produce json
// @Param query body models.AnalyticsQuery true "Analytics query parameters"
// @Success 200 {object} models.CategorySpendingResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /analytics/categories [post]
func (c *AnalyticsController) GetCategorySpending(ctx *gin.Context) {
	var query models.AnalyticsQuery
	if err := c.BindJSON(ctx, &query); err != nil {
		return
	}

	logger.Debugf("Getting category spending for user with query: %+v", query)

	categorySpending, err := c.analyticsService.GetCategorySpending(ctx, query)
	if err != nil {
		logger.Errorf("Failed to get category spending: %v", err)
		c.HandleError(ctx, err)
		return
	}

	logger.Debugf("Successfully retrieved category spending")
	c.SendSuccess(ctx, http.StatusOK, "Category spending retrieved successfully", categorySpending)
}

// GetSpendingTrends godoc
// @Summary Get spending trends
// @Description Get time-based spending trends with specified granularity
// @Tags analytics
// @Accept json
// @Produce json
// @Param query body models.AnalyticsQuery true "Analytics query parameters"
// @Param granularity query string false "Granularity (daily, weekly, monthly)" default(daily)
// @Success 200 {object} models.SpendingTrendsResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /analytics/trends [post]
func (c *AnalyticsController) GetSpendingTrends(ctx *gin.Context) {
	var query models.AnalyticsQuery
	if err := c.BindJSON(ctx, &query); err != nil {
		return
	}

	granularity := ctx.DefaultQuery("granularity", "daily")
	logger.Debugf("Getting spending trends for user with query: %+v, granularity: %s", query, granularity)

	trends, err := c.analyticsService.GetSpendingTrends(ctx, query, granularity)
	if err != nil {
		logger.Errorf("Failed to get spending trends: %v", err)
		c.HandleError(ctx, err)
		return
	}

	logger.Debugf("Successfully retrieved spending trends")
	c.SendSuccess(ctx, http.StatusOK, "Spending trends retrieved successfully", trends)
}

// GetAccountSpending godoc
// @Summary Get account spending breakdown
// @Description Get spending breakdown by accounts
// @Tags analytics
// @Accept json
// @Produce json
// @Param query body models.AnalyticsQuery true "Analytics query parameters"
// @Success 200 {object} models.AccountSpendingResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /analytics/accounts [post]
func (c *AnalyticsController) GetAccountSpending(ctx *gin.Context) {
	var query models.AnalyticsQuery
	if err := c.BindJSON(ctx, &query); err != nil {
		return
	}

	logger.Debugf("Getting account spending for user with query: %+v", query)

	accountSpending, err := c.analyticsService.GetAccountSpending(ctx, query)
	if err != nil {
		logger.Errorf("Failed to get account spending: %v", err)
		c.HandleError(ctx, err)
		return
	}

	logger.Debugf("Successfully retrieved account spending")
	c.SendSuccess(ctx, http.StatusOK, "Account spending retrieved successfully", accountSpending)
}

// GetTopTransactions godoc
// @Summary Get top transactions
// @Description Get highest spending transactions
// @Tags analytics
// @Accept json
// @Produce json
// @Param query body models.AnalyticsQuery true "Analytics query parameters"
// @Param limit query int false "Number of transactions to return" default(10)
// @Success 200 {object} models.TopTransactionsResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /analytics/top-transactions [post]
func (c *AnalyticsController) GetTopTransactions(ctx *gin.Context) {
	var query models.AnalyticsQuery
	if err := c.BindJSON(ctx, &query); err != nil {
		return
	}

	limitStr := ctx.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.SendError(ctx, http.StatusBadRequest, "Invalid limit parameter")
		return
	}

	logger.Debugf("Getting top transactions for user with query: %+v, limit: %d", query, limit)

	topTransactions, err := c.analyticsService.GetTopTransactions(ctx, query, limit)
	if err != nil {
		logger.Errorf("Failed to get top transactions: %v", err)
		c.HandleError(ctx, err)
		return
	}

	logger.Debugf("Successfully retrieved top transactions")
	c.SendSuccess(ctx, http.StatusOK, "Top transactions retrieved successfully", topTransactions)
}

// GetMonthlyComparison godoc
// @Summary Get monthly spending comparison
// @Description Get month-over-month spending comparison
// @Tags analytics
// @Accept json
// @Produce json
// @Param query body models.AnalyticsQuery true "Analytics query parameters"
// @Success 200 {object} models.MonthlyComparisonResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /analytics/monthly-comparison [post]
func (c *AnalyticsController) GetMonthlyComparison(ctx *gin.Context) {
	var query models.AnalyticsQuery
	if err := c.BindJSON(ctx, &query); err != nil {
		return
	}

	logger.Debugf("Getting monthly comparison for user with query: %+v", query)

	monthlyComparison, err := c.analyticsService.GetMonthlyComparison(ctx, query)
	if err != nil {
		logger.Errorf("Failed to get monthly comparison: %v", err)
		c.HandleError(ctx, err)
		return
	}

	logger.Debugf("Successfully retrieved monthly comparison")
	c.SendSuccess(ctx, http.StatusOK, "Monthly comparison retrieved successfully", monthlyComparison)
}

// GetRecurringTransactions godoc
// @Summary Get recurring transactions
// @Description Get detected recurring transaction patterns
// @Tags analytics
// @Accept json
// @Produce json
// @Param query body models.AnalyticsQuery true "Analytics query parameters"
// @Success 200 {object} models.RecurringTransactionsResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /analytics/recurring [post]
func (c *AnalyticsController) GetRecurringTransactions(ctx *gin.Context) {
	var query models.AnalyticsQuery
	if err := c.BindJSON(ctx, &query); err != nil {
		return
	}

	logger.Debugf("Getting recurring transactions for user with query: %+v", query)

	recurringTransactions, err := c.analyticsService.GetRecurringTransactions(ctx, query)
	if err != nil {
		logger.Errorf("Failed to get recurring transactions: %v", err)
		c.HandleError(ctx, err)
		return
	}

	logger.Debugf("Successfully retrieved recurring transactions")
	c.SendSuccess(ctx, http.StatusOK, "Recurring transactions retrieved successfully", recurringTransactions)
}

// GetAnalyticsSummary godoc
// @Summary Get comprehensive analytics summary
// @Description Get a comprehensive analytics summary including all major metrics
// @Tags analytics
// @Accept json
// @Produce json
// @Param query body models.AnalyticsQuery true "Analytics query parameters"
// @Success 200 {object} models.AnalyticsSummaryResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /analytics/summary [post]
func (c *AnalyticsController) GetAnalyticsSummary(ctx *gin.Context) {
	var query models.AnalyticsQuery
	if err := c.BindJSON(ctx, &query); err != nil {
		return
	}

	logger.Debugf("Getting analytics summary for user with query: %+v", query)

	summary, err := c.analyticsService.GetAnalyticsSummary(ctx, query)
	if err != nil {
		logger.Errorf("Failed to get analytics summary: %v", err)
		c.HandleError(ctx, err)
		return
	}

	logger.Debugf("Successfully retrieved analytics summary")
	c.SendSuccess(ctx, http.StatusOK, "Analytics summary retrieved successfully", summary)
}

// GetAnalyticsInsights godoc
// @Summary Get analytics insights
// @Description Get AI-generated insights based on spending patterns
// @Tags analytics
// @Accept json
// @Produce json
// @Param query body models.AnalyticsQuery true "Analytics query parameters"
// @Success 200 {object} models.AnalyticsInsightsResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /analytics/insights [post]
func (c *AnalyticsController) GetAnalyticsInsights(ctx *gin.Context) {
	var query models.AnalyticsQuery
	if err := c.BindJSON(ctx, &query); err != nil {
		return
	}

	logger.Debugf("Getting analytics insights for user with query: %+v", query)

	insights, err := c.analyticsService.GetAnalyticsInsights(ctx, query)
	if err != nil {
		logger.Errorf("Failed to get analytics insights: %v", err)
		c.HandleError(ctx, err)
		return
	}

	logger.Debugf("Successfully retrieved analytics insights")
	c.SendSuccess(ctx, http.StatusOK, "Analytics insights retrieved successfully", insights)
}
