package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"expenses/entities"
	logger "expenses/logger"
	"expenses/mapper"
	models "expenses/models"
	"expenses/services"
	"expenses/utils"
	"expenses/validators"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExpenseController struct {
	expenseService *services.ExpenseService
}

func NewExpenseController(db *pgxpool.Pool) *ExpenseController {
	expenseService := services.NewExpenseService(db)
	return &ExpenseController{expenseService: expenseService}
}

// GetExpensesOfUser returns all expenses for a given user
// TODO: Add pagination
func (e *ExpenseController) GetExpensesOfUser(c *gin.Context) {
	userID := c.GetInt64("authUserID")

	logger.Info("Recieved request to get all expenses for user with ID: ", userID)

	expenses, err := e.expenseService.GetExpensesByUserID(c, userID)
	if err != nil {
		logger.Error("Error getting expenses: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting expenses"})
		c.Abort()
		return
	}

	logger.Info("Number of expenses found: ", len(expenses))
	c.JSON(http.StatusOK, gin.H{
		"data": expenses,
	})
}

// CreateExpense handles creation of a new expense
func (e *ExpenseController) CreateExpense(c *gin.Context) {
	var userID = c.GetInt64("authUserID")

	var expense entities.ExpenseInput
	if err := c.ShouldBindJSON(&expense); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	logger.Info("Recieved request to create an expense with the following body: ", expense)

	contributors := make([]int64, 0, len(expense.Contributions))
	contributions := make([]float64, 0, len(expense.Contributions))
	for k, v := range expense.Contributions {
		contributors = append(contributors, k)
		contributions = append(contributions, v)
	}

	logger.Info("Validating the expenses input")
	err := validators.ValidateContributions(contributions, expense.Amount)
	if err != nil {
		logger.Error("Error validating expense: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to validate expense", "reason": err.Error()})
		return
	}
	logger.Info("Expense input validated successfully")

	addedExpense, err := e.expenseService.CreateExpense(c, models.Expense{
		Amount:      expense.Amount,
		PayerID:     expense.PayerID,
		Description: expense.Description,
		Name:        expense.Name,
		CreatedBy:   userID,
	}, contributors, contributions)
	if err != nil {
		logger.Error("Error creating expense: ", err)
		if utils.CheckForeignKey(err, "expense", "user_id") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Payer ID does not exist"})
			c.Abort()
			return
		}
		if utils.CheckForeignKey(err, "expense_user_mapping", "user_id") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Contributors ID does not exist"})
			c.Abort()
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating expense"})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Expense added successfully!",
		"data":    addedExpense,
	})
}

func (e *ExpenseController) GetExpenseByID(c *gin.Context) {
	expenseIDParam := c.Param("expenseID")
	userId := c.GetInt64("authUserID")
	expenseID, err := strconv.ParseInt(expenseIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expense ID"})
		return
	}

	logger.Info("Recieved request to get an expense by ID: ", expenseID)

	expensesModel, err := e.expenseService.GetExpenseByID(c, expenseID, userId)
	if err != nil {
		logger.Error("Error getting expense: ", err)
		if strings.Contains(err.Error(), "no rows in result set") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Expense not found"})
			c.Abort()
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting expense"})
		c.Abort()
		return
	}

	expenses, err := mapper.ExpenseContributorToMapper(expensesModel)
	if err != nil {
		logger.Error("Error mapping expense: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error mapping expense", "reason": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": expenses,
	})
}

func (e *ExpenseController) UpdateExpenseBasic(c *gin.Context) {
	expenseIDParam := c.Param("expenseID")
	userId := c.GetInt64("authUserID")

	expenseID, err := strconv.ParseInt(expenseIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expense ID"})
		return
	}

	var expenseInput entities.UpdateExpenseBasicInput
	if err := c.ShouldBindJSON(&expenseInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	logger.Info("Received request to update an expense with the following body: ", expenseInput)

	updatedExpense, err := e.expenseService.UpdateExpenseBasicDetails(c, expenseInput, expenseID, userId)
	if err != nil {
		if strings.Contains(err.Error(), "fk_user") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Payer ID does not exist"})
			c.Abort()
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating expense"})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Expense updated successfully!",
		"data":    updatedExpense,
	})
}

func (e *ExpenseController) UpdateExpenseContributions(c *gin.Context) {
	expenseIDParam := c.Param("expenseID")
	userId := c.GetInt64("authUserID")
	expenseID, err := strconv.ParseInt(expenseIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expense ID"})
		return
	}
	var expenseInput entities.UpdateExpenseContributionsInput
	if err := c.ShouldBindJSON(&expenseInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	logger.Info("Received request to update an expense with the following body: ", expenseInput)

	contributors := make([]int64, 0, len(expenseInput.Contributions))
	contributions := make([]float64, 0, len(expenseInput.Contributions))

	for k, v := range expenseInput.Contributions {
		contributors = append(contributors, k)
		contributions = append(contributions, v)
	}
	logger.Info("Total number of contributors: ", len(contributors))

	err = e.expenseService.UpdateExpenseContributions(c, expenseID, userId, contributors, contributions)
	if err != nil {
		logger.Error("Error updating expense: ", err)
		if strings.Contains(err.Error(), "fk_user") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Payer ID does not exist"})
			c.Abort()
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating expense"})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Expense updated successfully!",
	})
}

// DeleteExpense deletes an expense by ID
func (e *ExpenseController) DeleteExpense(c *gin.Context) {
	expenseIDParam := c.Param("expenseID")
	expenseID, err := strconv.ParseInt(expenseIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expense ID"})
		return
	}

	logger.Info("Recieved request to delete an expense by ID: ", expenseID)
	err = e.expenseService.DeleteExpense(c, expenseID)
	if err != nil {
		logger.Error("Error deleting expense: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting expense"})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Expense deleted successfully!",
	})
}
