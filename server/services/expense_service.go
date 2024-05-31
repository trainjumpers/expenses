package services

import (
	"expenses/models"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	logger "github.com/sirupsen/logrus"
)

type ExpenseService struct {
	db     *pgxpool.Pool
	schema string
}

func NewExpenseService(db *pgxpool.Pool) *ExpenseService {
	return &ExpenseService{
		db:     db,
		schema: os.Getenv("PGSCHEMA"), //unable to load as this is not inited anywhere in main, thus doesnt have access to env
	}
}

func (e *ExpenseService) GetExpensesByUserID(c *gin.Context, userID int64, schema string) []models.Expense {
	query := fmt.Sprintf(`SELECT * FROM %[1]s.expense WHERE id IN 
		(SELECT expense_id FROM %[1]s.expense_user_mapping WHERE user_id = $1)`,
		schema)

	rows, err := e.db.Query(c, query, userID)
	if err != nil {
		logger.Fatal(fmt.Errorf("error querying the database: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting expenses"})
		c.Abort()
		return nil
	}

	var expenses []models.Expense

	for rows.Next() {
		var expense models.Expense
		err := rows.Scan(&expense.ID, &expense.Amount, &expense.PayerID, &expense.Description, &expense.CreatedBy, &expense.CreatedAt)
		if err != nil {
			logger.Fatal(fmt.Errorf("error scanning the database output: %v", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing expenses"})
			c.Abort()
			return nil
		}
		expenses = append(expenses, expense)
	}

	return expenses
}

func (e *ExpenseService) CreateExpense(c *gin.Context, expense models.Expense, contributors []int64, contributions []float64, schema string) models.Expense {
	query := fmt.Sprintf(`
	WITH new_expense AS (
		INSERT INTO %[1]s.expense (
			amount, payer_id, description, created_by
			) VALUES (
			$1, $2, $3, $4
		) returning *
	), new_mappings AS (
		INSERT INTO %[1]s.expense_user_mapping (
			expense_id, user_id, amount
		) (SELECT id, unnest($5::bigint[]), unnest($6::numeric[]) from new_expense
	) RETURNING *)
	SELECT 
		ne.id AS expense_id, 
		ne.amount as total_amount, 
		ne.payer_id, 
		ne.description, 
		ne.created_by,
		ne.created_at 
	FROM new_expense ne LEFT JOIN new_mappings nm ON ne.id = nm.expense_id;
		`, schema)
	var addedExpense models.Expense

	logger.Info("Executing query to insert an expense: ", query)
	insert := e.db.QueryRow(c, query, expense.Amount, expense.PayerID, expense.Description, expense.CreatedBy, contributors, contributions)

	err := insert.Scan(&addedExpense.ID, &addedExpense.Amount, &addedExpense.PayerID, &addedExpense.Description, &addedExpense.CreatedBy, &addedExpense.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "fk_user") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Payer ID does not exist"})
			c.Abort()
			return models.Expense{}
		}
		logger.Error("Error inserting expense: ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting expense"})
		c.Abort()
		return models.Expense{}
	}

	return addedExpense
}
