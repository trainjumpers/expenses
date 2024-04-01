package expense

import (
	"fmt"
	"net/http"
	"os"

	database "expenses/db"
	models "expenses/models"

	"github.com/gin-gonic/gin"
	logger "github.com/sirupsen/logrus"
)

func GetExpenses(c *gin.Context) {
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
		err := result.Scan(&expense.ID, &expense.Amount, &expense.PayerID, &expense.Description)
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

func CreateExpense(c *gin.Context) {
	var schema = os.Getenv("PGSCHEMA")

	var expense models.ExpenseInput
	if err := c.ShouldBindJSON(&expense); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	logger.Info("Recieved request to create an expense with the following body: ", expense)

	query := fmt.Sprintf("INSERT INTO %s.expenses (amount, payer_id, description) VALUES ($1, $2, $3) RETURNING *;", schema)
	var addedExpense models.Expense

	logger.Info("Executing query to insert an expense: ", query)
	insert := database.DbPool.QueryRow(c, query, expense.Amount, expense.PayerID, expense.Description)

	err := insert.Scan(&addedExpense.ID, &addedExpense.Amount, &addedExpense.PayerID, &addedExpense.Description)
	if err != nil {
		panic(err.Error())
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Expense added successfully!",
		"data":    addedExpense,
	})
}
