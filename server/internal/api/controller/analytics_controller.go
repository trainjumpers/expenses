package controller

import (
	"errors"
	"expenses/internal/config"
	"expenses/internal/service"
	"expenses/pkg/logger"
	"net/http"
	"strconv"
	"strings"

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

	startDate, endDate, err := a.ParseDateRange(ctx)
	if err != nil {
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

func (a *AnalyticsController) GetCategoryAnalytics(ctx *gin.Context) {
	userId := a.GetAuthenticatedUserId(ctx)
	logger.Infof("Fetching category analytics for user %d", userId)

	startDate, endDate, err := a.ParseDateRange(ctx)
	if err != nil {
		return
	}

	categoryIds, err := parseCategoryIds(ctx.Query("category_ids"))
	if err != nil {
		a.SendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	analytics, err := a.analyticsService.GetCategoryAnalytics(ctx, userId, startDate, endDate, categoryIds)
	if err != nil {
		logger.Errorf("Error getting category analytics: %v", err)
		a.HandleError(ctx, err)
		return
	}

	logger.Infof("Category analytics retrieved successfully for user %d", userId)
	a.SendSuccess(ctx, http.StatusOK, "Category analytics retrieved successfully", analytics)
}

func (a *AnalyticsController) GetMonthlyAnalytics(ctx *gin.Context) {
	userId := a.GetAuthenticatedUserId(ctx)
	logger.Infof("Fetching monthly analytics for user %d", userId)

	startDate, endDate, err := a.ParseDateRange(ctx)
	if err != nil {
		return
	}

	analytics, err := a.analyticsService.GetMonthlyAnalytics(ctx, userId, startDate, endDate)
	if err != nil {
		logger.Errorf("Error getting monthly analytics: %v", err)
		a.HandleError(ctx, err)
		return
	}

	logger.Infof("Monthly analytics retrieved successfully for user %d", userId)
	a.SendSuccess(ctx, http.StatusOK, "Monthly analytics retrieved successfully", analytics)
}

func parseCategoryIds(raw string) ([]int64, error) {
	if raw == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	categoryIds := make([]int64, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			return nil, errors.New("invalid category_ids format, expected comma-separated integers")
		}
		categoryID, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			return nil, errors.New("invalid category_ids format, expected comma-separated integers")
		}
		categoryIds = append(categoryIds, categoryID)
	}

	return categoryIds, nil
}
