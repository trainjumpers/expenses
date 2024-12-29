package controllers

import (
	"expenses/logger"
	"expenses/mapper"
	"expenses/parser"
	"expenses/services"
	"expenses/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StatementController struct {
	expenseService *services.ExpenseService
}

func NewStatementController(db *pgxpool.Pool) *StatementController {
	expenseService := services.NewExpenseService(db)
	return &StatementController{expenseService: expenseService}
}

func (s *StatementController) ParseStatement(c *gin.Context) {
	userID := c.GetInt64("authUserID")
	records := utils.ReadCSV("statement.csv")
	statements, err := parser.ParseBankStatement(records)
	if err != nil {
		logger.Error("Error parsing statement: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing statement", "reason": err.Error()})
		return
	}
	logger.Info("Successfully parsed statement of length: ", len(statements), " for user with ID: ", userID)
	expenses, err := mapper.StatementExpenseMapper(statements, userID)
	if err != nil {
		logger.Error("Error mapping statement to expenses: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error mapping statement to expenses", "reason": err.Error()})
		c.Abort()
		return
	}

	logger.Info("Successfully mapped statement to expenses of length: ", len(expenses), " for user with ID: ", userID)
	addedExpense, err := s.expenseService.CreateMultipleExpenses(c, expenses)
	if err != nil {
		logger.Error("Error creating expense: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating expense", "reason": err.Error()})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Expense added successfully!",
		"data":    addedExpense,
	})
}
