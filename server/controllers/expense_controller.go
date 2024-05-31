package controllers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

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

func (e *ExpenseController) GetExpensesOfUser(c *gin.Context) {
	var schema = os.Getenv("PGSCHEMA")

	userID, err := strconv.ParseInt(c.Query("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	logger.Info("Recieved request to get all expenses for user with ID: ", userID)

	expenses := e.expenseService.GetExpensesByUserID(c, userID, schema)

	logger.Info("Number of expenses found: ", len(expenses))
	c.JSON(http.StatusOK, gin.H{
		"data": expenses,
	})
}

func (e *ExpenseController) CreateExpense(c *gin.Context) {
	var schema = os.Getenv("PGSCHEMA")
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

	addedExpense := e.expenseService.CreateExpense(c, models.Expense{
		Amount:      expense.Amount,
		PayerID:     expense.PayerID,
		Description: expense.Description,
		CreatedBy:   userID,
	}, contributors, contributions, schema)

	c.JSON(http.StatusOK, gin.H{
		"message": "Expense added successfully!",
		"data":    addedExpense,
	})
}

// func (e *ExpenseController) GetExpenseByID(c *gin.Context) {
// 	var schema = os.Getenv("PGSCHEMA")

// 	expenseID := c.Param("expenseID")
// 	logger.Info("Recieved request to get an expense by ID: ", expenseID)

// 	var expense models.Expense

// 	query := fmt.Sprintf("SELECT * FROM %s.expense WHERE id = $1;", schema)

// 	logger.Info("Executing query to get an expense by ID: ", query)
// 	result := database.DbPool.QueryRow(c, query, expenseID)

// 	err := result.Scan(&expense.ID, &expense.Amount, &expense.PayerID, &expense.Description, &expense.CreatedBy, &expense.CreatedAt)
// 	if err != nil {
// 		if strings.Contains(err.Error(), "no rows in result set") {
// 			c.JSON(http.StatusNotFound, gin.H{"error": "Expense not found"})
// 			return
// 		}
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"data": expense,
// 	})
// }

func (e *ExpenseController) DeleteExpense(c *gin.Context) {
	var schema = os.Getenv("PGSCHEMA")

	expenseID := c.Param("expenseID")
	logger.Info("Recieved request to delete an expense by ID: ", expenseID)

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
