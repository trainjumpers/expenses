package controllers

import (
	"encoding/json"
	"expenses/logger"
	"expenses/mapper"
	"expenses/models"
	"expenses/parser"
	"expenses/services"
	"expenses/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StatementController struct {
	expenseService *services.ExpenseService
	jobService     *services.JobService
}

func NewStatementController(db *pgxpool.Pool) *StatementController {
	expenseService := services.NewExpenseService(db)
	jobService := services.NewJobService(db)
	return &StatementController{expenseService: expenseService, jobService: jobService}
}

func (s *StatementController) ParseStatement(c *gin.Context) {
	userID := c.GetInt64("authUserID")
	jobName := "Parsing statement.csv"
	metadata := json.RawMessage(`{"file_name": "statement.csv"}`)
	jobStatus, err := s.jobService.CreateJob(c, userID, jobName, metadata)
	if err != nil {
		logger.Error("Error creating job: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating job", "reason": err.Error()})
		c.Abort()
		return
	}
	logger.Info("Successfully created job of id: ", jobStatus.ID, "for user with ID: ", userID)

	go func() {
		jobStatus, err := s.jobService.UpdateJobStatus(c, userID, jobStatus.ID, models.JobStatusRunning)
		if err != nil {
			logger.Error("Error updating job status: ", err)
			return
		}
		records := utils.ReadCSV("statement.csv")
		statements, err := parser.ParseBankStatement(records)
		if err != nil {
			logger.Error("Error parsing statement: ", err)
			s.jobService.UpdateJobStatus(c, userID, jobStatus.ID, models.JobStatusFailed)
			return
		}
		logger.Info("Successfully parsed statement of length: ", len(statements), " for user with ID: ", userID)
		expenses, err := mapper.StatementExpenseMapper(statements, userID)
		if err != nil {
			logger.Error("Error mapping statement to expenses: ", err)
			s.jobService.UpdateJobStatus(c, userID, jobStatus.ID, models.JobStatusFailed)
			return
		}

		logger.Info("Successfully mapped statement to expenses of length: ", len(expenses), " for user with ID: ", userID)

		_, err = s.expenseService.CreateMultipleExpenses(c.Copy(), expenses)
		if err != nil {
			logger.Error("Error creating expense in background: ", err)
			s.jobService.UpdateJobStatus(c, userID, jobStatus.ID, models.JobStatusFailed)
			return
		}

		logger.Info("Successfully created expenses of length: ", len(expenses), " for user with ID: ", userID)
		jobStatus, err = s.jobService.UpdateJobStatus(c, userID, jobStatus.ID, models.JobStatusComplete)
		if err != nil {
			logger.Error("Error updating job status: ", err)
			s.jobService.UpdateJobStatus(c, userID, jobStatus.ID, models.JobStatusFailed)
			logger.Info("Updated job status to failed for job with ID: ", jobStatus.ID, " for user with ID: ", userID)
			return
		}
		logger.Info("Successfully updated job status for job with ID: ", jobStatus.ID, " for user with ID: ", userID)
	}()
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Processing statement in background",
		"status":  jobStatus.Status,
		"id":      jobStatus.ID,
	})
}
