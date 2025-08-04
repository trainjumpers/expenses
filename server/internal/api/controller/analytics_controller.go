package controller

import (
	"expenses/internal/config"
	"expenses/internal/service"
	"expenses/pkg/logger"
	"net/http"

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
