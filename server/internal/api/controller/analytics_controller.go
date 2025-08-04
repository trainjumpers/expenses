package controller

import (
	"expenses/internal/config"
	"expenses/internal/service"
	"expenses/pkg/logger"
	"net/http"
	"time"

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

func (a *AnalyticsController) GetAccountAnalytics(ctx *gin.Context) {
	userId := a.GetAuthenticatedUserId(ctx)
	logger.Infof("Fetching account analytics for user %d", userId)

	analytics, err := a.analyticsService.GetAccountAnalytics(ctx, userId)
	if err != nil {
		logger.Errorf("Error getting account analytics: %v", err)
		a.HandleError(ctx, err)
		return
	}

	logger.Infof("Account analytics retrieved successfully for user %d", userId)
	a.SendSuccess(ctx, http.StatusOK, "Account analytics retrieved successfully", analytics)
}

func (a *AnalyticsController) GetNetworthTimeSeries(ctx *gin.Context) {
	userId := a.GetAuthenticatedUserId(ctx)
	logger.Infof("Fetching networth time series for user %d", userId)

	// Parse query parameters
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		a.SendError(ctx, http.StatusBadRequest, "start_date and end_date query parameters are required")
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		a.SendError(ctx, http.StatusBadRequest, "invalid start_date format, expected YYYY-MM-DD")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		a.SendError(ctx, http.StatusBadRequest, "invalid end_date format, expected YYYY-MM-DD")
		return
	}

	if startDate.After(endDate) {
		a.SendError(ctx, http.StatusBadRequest, "start_date cannot be after end_date")
		return
	}

	timeSeries, err := a.analyticsService.GetNetworthTimeSeries(ctx, userId, startDate, endDate)
	if err != nil {
		logger.Errorf("Error getting networth time series: %v", err)
		a.HandleError(ctx, err)
		return
	}

	logger.Infof("Networth time series retrieved successfully for user %d", userId)
	a.SendSuccess(ctx, http.StatusOK, "Networth time series retrieved successfully", timeSeries)
}
