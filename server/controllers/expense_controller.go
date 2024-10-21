package controllers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	database "expenses/db"
	"expenses/entities"
	models "expenses/models"
	"expenses/services"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	logger "github.com/sirupsen/logrus"
)

type ExpenseController struct {
	expenseService *services.ExpenseService
}

func NewExpenseController(db *pgxpool.Pool) *ExpenseController {
	expenseService := services.NewExpenseService(db)
	return &ExpenseController{expenseService: expenseService}
}

// GetExpensesOfUser returns all expenses for a given user
func (e *ExpenseController) GetExpensesOfUser(c *gin.Context) {

	userID, err := strconv.ParseInt(c.Query("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

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
	var userID = c.GetInt64("userID")

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

	addedExpense, err := e.expenseService.CreateExpense(c, models.Expense{
		Amount:      expense.Amount,
		PayerID:     expense.PayerID,
		Description: expense.Description,
		CreatedBy:   userID,
	}, contributors, contributions)
	if err != nil {
		if strings.Contains(err.Error(), "fk_user") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Payer ID does not exist"})
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
	expenseID, err := strconv.ParseInt(expenseIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expense ID"})
		return
	}

	logger.Info("Recieved request to get an expense by ID: ", expenseID)

	var expense models.Expense

	expense, err = e.expenseService.GetExpenseByID(c, expenseID)
	if err != nil {
		logger.Error("Error getting expense: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting expense"})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": expense,
	})
}

func (e *ExpenseController) UpdateExpenseBasic(c *gin.Context) {
	expenseIDParam := c.Param("expenseID")
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

	updatedExpense, err := e.expenseService.UpdateExpenseBasicDetails(c, expenseInput, expenseID)
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

}

// DeleteExpense deletes an expense by ID
func (e *ExpenseController) DeleteExpense(c *gin.Context) {
	var schema = os.Getenv("PGSCHEMA")

	expenseID := c.Param("expenseID")
	logger.Info("Received request to delete an expense by ID: ", expenseID)

	query := fmt.Sprintf("DELETE FROM %s.expense WHERE id = $1;", schema)

	logger.Info("Executing query to delete an expense by ID: ", query)
	_, err := database.DbPool.Exec(c, query, expenseID)
	if err != nil {
		panic(err.Error())
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Expense deleted successfully!",
	})
}
