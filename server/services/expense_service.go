package services

import (
	"expenses/entities"
	"expenses/models"
	"expenses/utils"
	"expenses/validators"
	"fmt"
	"strconv"
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
func (e *ExpenseService) GetExpensesByUserID(c *gin.Context, userID int64) ([]models.ExpenseWithContribution, error) {
	query := fmt.Sprintf(`SELECT
		e.id,
		e.amount,
		e.payer_id,
		e.name,
		e.description,
		e.created_by,
		e.created_at,
		eum.amount AS user_amount
	FROM %[1]s.expense e
	LEFT JOIN %[1]s.expense_user_mapping eum ON e.id = eum.expense_id
	WHERE eum.user_id = $1;`, e.schema)

	rows, err := e.db.Query(c, query, userID)
	if err != nil {
		return []models.ExpenseWithContribution{}, err
	}

	var expenses []models.ExpenseWithContribution

	for rows.Next() {
		var expense models.ExpenseWithContribution
		err := rows.Scan(&expense.ID, &expense.Amount, &expense.PayerID, &expense.Name, &expense.Description, &expense.CreatedBy, &expense.CreatedAt, &expense.UserAmount)
		if err != nil {
			return []models.ExpenseWithContribution{}, err
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
			amount, payer_id, description, name, created_by
			) VALUES (
			$1, $2, $3, $4, $5
		) RETURNING *
	), new_mappings AS (
		INSERT INTO %[1]s.expense_user_mapping (
			expense_id, user_id, amount
		) (SELECT id, unnest($6::bigint[]), unnest($7::numeric[]) from new_expense
	) RETURNING *)
	SELECT 
		ne.id AS expense_id, 
		ne.amount as total_amount, 
		ne.payer_id,
		ne.name,
		ne.description, 
		ne.created_by,
		ne.created_at 
	FROM new_expense ne LEFT JOIN new_mappings nm ON ne.id = nm.expense_id;
		`, e.schema)
	var addedExpense models.Expense

	logger.Info("Executing query to insert an expense: ", query)
	insert := e.db.QueryRow(c, query, expense.Amount, expense.PayerID, expense.Description, expense.Name, expense.CreatedBy, contributors, contributions)

	err := insert.Scan(&addedExpense.ID, &addedExpense.Amount, &addedExpense.PayerID, &addedExpense.Name, &addedExpense.Description, &addedExpense.CreatedBy, &addedExpense.CreatedAt)
	if err != nil {
		return models.Expense{}, err
	}

	return addedExpense, nil
}

func (e *ExpenseService) GetExpenseByID(c *gin.Context, expenseID int64, userId int64) ([]models.ExpenseWithAllContributions, error) {
	var expenses []models.ExpenseWithAllContributions

	query := fmt.Sprintf(`
	SELECT e.*,
		eum.user_id AS contributor_id,
		eum.amount AS contribution,
		u.name AS contributor_name
	FROM (
		SELECT 
			id,
			amount,
			payer_id,
			name,
			description,
			created_by,
			created_at
		FROM %[1]s.expense
		WHERE id = $1
		AND id IN (
			SELECT expense_id FROM %[1]s.expense_user_mapping WHERE user_id = $2
		)
	) AS e
	LEFT JOIN %[1]s.expense_user_mapping eum ON e.id = eum.expense_id
	LEFT JOIN %[1]s.user u ON eum.user_id = u.id;
	`, e.schema)

	logger.Info("Executing query to get an expense by ID: ", query)
	rows, err := e.db.Query(c, query, expenseID, userId)
	if err != nil {
		return []models.ExpenseWithAllContributions{}, err
	}

	for rows.Next() {
		var expense models.ExpenseWithAllContributions
		err := rows.Scan(&expense.ID, &expense.Amount, &expense.PayerID, &expense.Name, &expense.Description, &expense.CreatedBy, &expense.CreatedAt, &expense.ContributorId, &expense.Contribution, &expense.ContributorName)
		if err != nil {
			return []models.ExpenseWithAllContributions{}, err
		}
		expenses = append(expenses, expense)
	}
	return expenses, nil
}

func (e *ExpenseService) UpdateExpenseBasicDetails(c *gin.Context, updatedFields entities.UpdateExpenseBasicInput, expenseID int64, userId int64) (models.Expense, error) {

	fields := map[string]interface{}{
		"description": updatedFields.Description,
		"payer_id":    updatedFields.PayerID,
		"name":        updatedFields.Name,
	}

	fieldsClause := ""
	argIndex := 1
	argValues := make([]interface{}, 0)
	for k, v := range fields {
		if v == "" || v == int64(0) {
			logger.Info("Skipping field: ", k)
			continue
		}

		fieldsClause += k + " = $" + strconv.FormatInt(int64(argIndex), 10) + ", "
		argIndex++
		argValues = append(argValues, v)
	}
	fieldsClause = strings.TrimSuffix(fieldsClause, ", ")
	if fieldsClause == "" {
		return models.Expense{}, fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf(`UPDATE %[1]s.expense 
	SET %[2]s WHERE id = $%[3]d AND ID IN (
		SELECT expense_id
		FROM %[1]s.expense_user_mapping 
		WHERE user_id = $%[4]d
	) RETURNING
		id,
		amount,
		payer_id,
		description,
		created_by,
		created_at,
		name 
	;`, e.schema, fieldsClause, argIndex, argIndex+1)

	logger.Info("Executing query to update an expense: ", query)

	var updatedExpense models.Expense
	logger.Info("arg ", argValues)

	update := e.db.QueryRow(c, query, append(argValues, expenseID, userId)...)
	err := update.Scan(&updatedExpense.ID, &updatedExpense.Amount, &updatedExpense.PayerID, &updatedExpense.Description, &updatedExpense.CreatedBy, &updatedExpense.CreatedAt, &updatedExpense.Name)
	if err != nil {
		logger.Error("Error updating expense into the db: ", err)
		return models.Expense{}, err
	}

	return updatedExpense, nil
}
func (e *ExpenseService) UpdateExpenseContributions(c *gin.Context, expenseID int64, userId int64, contributors []int64, contributions []float64) error {
	getQuery := fmt.Sprintf(`
	SELECT e.amount FROM %[1]s.expense e
	INNER JOIN %[1]s.expense_user_mapping eum ON e.id = eum.expense_id
	WHERE e.id = $1 AND eum.user_id = $2`, e.schema)
	
	deleteQuery := fmt.Sprintf(`
	DELETE FROM %[1]s.expense_user_mapping WHERE expense_id = $1;`, e.schema)

	insertQuery := fmt.Sprintf(`
	INSERT INTO %[1]s.expense_user_mapping (
		expense_id, user_id, amount
	) SELECT $1, unnest($2::bigint[]), unnest($3::numeric[])
	 RETURNING *;`, e.schema)

	logger.Info("Acquiring a db transaction to update expense contributions")
	tx, err := e.db.Begin(c)
	if err != nil {
		return err
	}
	defer tx.Rollback(c)
	logger.Info("Successfully acquired a db transaction to update expense contributions")

	logger.Info("Getting existing expense contributions with query: ", getQuery)
	expense := tx.QueryRow(c, getQuery, expenseID, userId)
	var amount float64
	err = expense.Scan(&amount)
	if err != nil {
		return err
	}
	logger.Info("Successfully got existing expense contributions")
	err = validators.ValidateContributions(contributions, amount)
	if err != nil {
		return err
	}
	
	logger.Info("Deleting existing expense contributions with query: ", deleteQuery)
	_, err = tx.Exec(c, deleteQuery, expenseID)
	if err != nil {
		return err
	}
	logger.Info("Successfully deleted existing expense contributions")
	logger.Info("Inserting new expense contributions with query: ", insertQuery)
	_, err = tx.Exec(c, insertQuery, expenseID, contributors, contributions)
	if err != nil {
		return err
	}
	logger.Info("Successfully inserted new expense contributions")
	err = tx.Commit(c)
	logger.Info("Successfully committed db transaction to update expense contributions")
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
