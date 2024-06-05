package services

import (
	"expenses/models"
	"fmt"
	"os"

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

/*
GetExpensesByUserID returns all expenses of a given user

userID: ID of the user whose expenses are to be fetched

returns: List of expenses ([]models.Expense)
*/
func (e *ExpenseService) GetExpensesByUserID(c *gin.Context, userID int64) ([]models.Expense, error) {
	query := fmt.Sprintf(`SELECT * FROM %[1]s.expense WHERE id IN 
		(SELECT expense_id FROM %[1]s.expense_user_mapping WHERE user_id = $1)`,
		e.schema)

	rows, err := e.db.Query(c, query, userID)
	if err != nil {
		return []models.Expense{}, err
	}

	var expenses []models.Expense

	for rows.Next() {
		var expense models.Expense
		err := rows.Scan(&expense.ID, &expense.Amount, &expense.PayerID, &expense.Description, &expense.CreatedBy, &expense.CreatedAt)
		if err != nil {
			return []models.Expense{}, err
		}
		expenses = append(expenses, expense)
	}

	return expenses, nil
}

/*
CreateExpense creates a new expense in the expense table and adds contributions of users to the expense_user_mapping table

expense: Expense object containing the details of the expense to be created

contributors: List of user IDs contributing to the expense

contributions: List of amounts contributed by each user

returns: Expense object of the newly created expense
*/
func (e *ExpenseService) CreateExpense(c *gin.Context, expense models.Expense, contributors []int64, contributions []float64) (models.Expense, error) {
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
		`, e.schema)
	var addedExpense models.Expense

	logger.Info("Executing query to insert an expense: ", query)
	insert := e.db.QueryRow(c, query, expense.Amount, expense.PayerID, expense.Description, expense.CreatedBy, contributors, contributions)

	err := insert.Scan(&addedExpense.ID, &addedExpense.Amount, &addedExpense.PayerID, &addedExpense.Description, &addedExpense.CreatedBy, &addedExpense.CreatedAt)
	if err != nil {
		return models.Expense{}, err
	}

	return addedExpense, nil
}
