package controllers

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	database "expenses/db"
	models "expenses/models"

	"github.com/gin-gonic/gin"
	logger "github.com/sirupsen/logrus"
)

type ExpenseController struct{}

func (e *ExpenseController) GetExpenses(c *gin.Context) {
	var schema = os.Getenv("PGSCHEMA")

	logger.Info("Recieved request to get all expenses")
	var expenses []models.Expense

	query := fmt.Sprintf("SELECT * FROM %s.expenses;", schema)

	logger.Info("Executing query to get all expenses: ", query)
	result, err := database.DbPool.Query(c, query)
	if err != nil {
		logger.Fatal(fmt.Errorf("error querying the database: %v", err))
	}

	for result.Next() {
		var expense models.Expense
		err := result.Scan(&expense.ID, &expense.Amount, &expense.PayerID, &expense.Description, &expense.CreatedBy, &expense.CreatedAt)
		if err != nil {
			panic(err.Error())
		}
		expenses = append(expenses, expense)
	}

	logger.Info("Number of expenses found: ", len(expenses))
	c.JSON(http.StatusOK, gin.H{
		"data": expenses,
	})
}

func (e *ExpenseController) CreateExpense(c *gin.Context) {
	var schema = os.Getenv("PGSCHEMA")
	var userID = c.GetInt64("userID")

	var expense models.ExpenseInput
	if err := c.ShouldBindJSON(&expense); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	logger.Info("Recieved request to create an expense with the following body: ", expense)

	query := fmt.Sprintf("INSERT INTO %s.expenses (amount, payer_id, description, created_by) VALUES ($1, $2, $3, $4) RETURNING *;", schema)
	var addedExpense models.Expense

	logger.Info("Executing query to insert an expense: ", query)
	insert := database.DbPool.QueryRow(c, query, expense.Amount, expense.PayerID, expense.Description, userID)

	err := insert.Scan(&addedExpense.ID, &addedExpense.Amount, &addedExpense.PayerID, &addedExpense.Description, &addedExpense.CreatedBy, &addedExpense.CreatedAt)
	if err != nil {
		// panic(err.Error())
		if strings.Contains(err.Error(), "fk_user") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Payer ID does not exist"})
			return
		}
		logger.Error("Error inserting expense: ", err.Error())
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Expense added successfully!",
		"data":    addedExpense,
	})
}

func (e *ExpenseController) GetExpenseByID(c *gin.Context) {
	var schema = os.Getenv("PGSCHEMA")

	expenseID := c.Param("expenseID")
	logger.Info("Recieved request to get an expense by ID: ", expenseID)

	var expense models.Expense

	query := fmt.Sprintf("SELECT * FROM %s.expenses WHERE id = $1;", schema)

	logger.Info("Executing query to get an expense by ID: ", query)
	result := database.DbPool.QueryRow(c, query, expenseID)

	err := result.Scan(&expense.ID, &expense.Amount, &expense.PayerID, &expense.Description, &expense.CreatedBy, &expense.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Expense not found"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": expense,
	})
}

func (e *ExpenseController) DeleteExpense(c *gin.Context) {
	var schema = os.Getenv("PGSCHEMA")

	expenseID := c.Param("expenseID")
	logger.Info("Recieved request to delete an expense by ID: ", expenseID)

	query := fmt.Sprintf("DELETE FROM %s.expenses WHERE id = $1;", schema)

	logger.Info("Executing query to delete an expense by ID: ", query)
	_, err := database.DbPool.Exec(c, query, expenseID)
	if err != nil {
		panic(err.Error())
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Expense deleted successfully!",
	})
}
