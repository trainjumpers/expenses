package controllers

import (
	"expenses/logger"
	"expenses/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type JobController struct {
	jobService *services.JobService
}

func NewJobController(db *pgxpool.Pool) *JobController {
	jobService := services.NewJobService(db)
	return &JobController{jobService: jobService}
}

func (j *JobController) GetJobStatus(c *gin.Context) {
	userID := c.GetInt64("authUserID")
	jobID, err := strconv.ParseInt(c.Param("jobID"), 10, 64)
	if err != nil {
		logger.Error("Error parsing job ID: ", jobID, " with error: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	jobStatus, err := j.jobService.GetJobStatus(c, userID, jobID)
	if err != nil {
		logger.Error("Error getting job status for user with ID: ", userID, " with error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting job status"})
		return
	}
	logger.Info("Successfully got job status for user with ID: ", userID)

	c.JSON(http.StatusOK, gin.H{
		"data": jobStatus,
	})
}
