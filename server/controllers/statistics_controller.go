package controllers

import (
	logger "expenses/logger"
	"expenses/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StatisticsController struct {
	statisticsService *services.StatisticsService
}

func NewStatisticsController(db *pgxpool.Pool) *StatisticsController {
	statisticsService := services.NewStatisticsService(db)
	return &StatisticsController{statisticsService: statisticsService}
}

func (s *StatisticsController) GetSubcategoryBreakdown(c *gin.Context) {
	userID := c.GetInt64("authUserID")

	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		startTime = time.Now().AddDate(0, -1, 0) // Default to last month
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		endTime = time.Now() // Default to current time
	}

	logger.Info("Received request to get expense breakdown by subcategory for user: ", userID)
	logger.Info("Time range: ", startTime, " to ", endTime)

	breakdown, err := s.statisticsService.GetExpensesBySubcategory(c, userID, startTime, endTime)
	if err != nil {
		logger.Error("Error getting subcategory breakdown: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting subcategory breakdown"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": breakdown,
		"metadata": gin.H{
			"start_time": startTime,
			"end_time":   endTime,
		},
	})
}

func (s *StatisticsController) GetMonthlyTrend(c *gin.Context) {
	userID := c.GetInt64("authUserID")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	logger.Info("Received request to get monthly spending trends for user: ", userID)
	trends, err := s.statisticsService.GetMonthlyTrend(c, userID, startDate, endDate)
	if err != nil {
		logger.Error("Error getting monthly trends: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting monthly trends"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": trends,
	})
}

func (s *StatisticsController) GetDailyHeatmap(c *gin.Context) {
	userID := c.GetInt64("authUserID")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	logger.Info("Received request to get daily spending heatmap for user: ", userID)
	heatmap, err := s.statisticsService.GetDailySpendingHeatmap(c, userID, startDate, endDate)
	if err != nil {
		logger.Error("Error getting daily heatmap: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting daily heatmap"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": heatmap,
	})
}
