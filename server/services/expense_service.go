package services

import (
	"expenses/entities"
	"expenses/models"
	"expenses/utils"
	"fmt"
	"net/http"
	"strings"

	logger "expenses/logger"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExpenseService struct {
	db     *pgxpool.Pool
	schema string
}

func NewExpenseService(db *pgxpool.Pool) *ExpenseService {
	return &ExpenseService{
		db:     db,
		schema: utils.GetPGSchema(), //unable to load as this is not inited anywhere in main, thus doesnt have access to env
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

func (e *ExpenseService) GetExpenseByID(c *gin.Context, expenseID int64) (models.Expense, error) {
	var expense models.Expense

	query := fmt.Sprintf("SELECT * FROM %s.expense WHERE id = $1;", e.schema)

	logger.Info("Executing query to get an expense by ID: ", query)
	result := e.db.QueryRow(c, query, expenseID)

	err := result.Scan(&expense.ID, &expense.Amount, &expense.PayerID, &expense.Description, &expense.CreatedBy, &expense.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Expense not found"})
			return models.Expense{}, err
		}
	}

	return expense, nil
}

func (e *ExpenseService) UpdateExpenseBasicDetails(c *gin.Context, updatedFields entities.UpdateExpenseBasicInput, expenseID int64) (models.Expense, error) {

	fieldClause := ""
	args := make([]interface{}, 0)
	argIndex := 1

	args = append(args, expenseID)
	argIndex++

	if updatedFields.Description != "" {
		fieldClause += fmt.Sprintf("description = $%d", argIndex)
		args = append(args, updatedFields.Description)
		argIndex++
	}
	if updatedFields.PayerID != 0 {
		if fieldClause != "" {
			fieldClause += ", "
		}
		fieldClause += fmt.Sprintf("payer_id = $%d", argIndex)
		args = append(args, updatedFields.PayerID)
		argIndex++
	}

	if fieldClause == "" {
		return models.Expense{}, fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf("UPDATE %[1]s.expense SET %[2]s WHERE id = $1 RETURNING *;", e.schema, fieldClause, argIndex)

	logger.Info("Executing query to update an expense: ", query)

	var updatedExpense models.Expense

	update := e.db.QueryRow(c, query, args...)
	err := update.Scan(&updatedExpense.ID, &updatedExpense.Amount, &updatedExpense.PayerID, &updatedExpense.Description, &updatedExpense.CreatedBy, &updatedExpense.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "fk_user") {
			return models.Expense{}, fmt.Errorf("payer ID does not exist")
		}
		logger.Error("Error updating expense into the db: ", err)
		return models.Expense{}, err
	}

	return updatedExpense, nil
}

func (e *ExpenseService) UpdateExpenseContributions(c *gin.Context, expenseID int64, contributors []int64, contributions []float64) error {
	query := fmt.Sprintf(`
	WITH updated_mappings AS (
		UPDATE %[1]s.expense_user_mapping SET amount = unnest($1::numeric[]) WHERE expense_id = $2 AND user_id = ANY($3::bigint[]) RETURNING *
	), updated_expense AS (
		UPDATE %[1]s.expense SET amount = $3 WHERE id = $2 RETURNING *
	)
	SELECT * FROM updated_mappings;
	`, e.schema)

	logger.Info("Executing query to update expense contributions: ", query)
	_, err := e.db.Exec(c, query, contributions, expenseID, contributors)
	if err != nil {
		return err
	}

	return nil
}

func (e *ExpenseService) DeleteExpense(c *gin.Context, expenseID int64) error {
	query := fmt.Sprintf("DELETE FROM %[1]s.expense WHERE id = $1;", e.schema)
	logger.Info("Executing query to delete an expense by ID: ", query)
	_, err := e.db.Exec(c, query, expenseID)
	if err != nil {
		return err
	}
	return nil
}