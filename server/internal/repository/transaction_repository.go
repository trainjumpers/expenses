package repository

import (
	"errors"
	"expenses/internal/config"
	"expenses/internal/database/helper"
	database "expenses/internal/database/postgres"
	customErrors "expenses/internal/errors"
	"expenses/internal/models"
	"expenses/pkg/logger"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepositoryInterface interface {
	CreateTransaction(c *gin.Context, transaction models.CreateBaseTransactionInput, categoryIds []int64) (models.TransactionResponse, error)
	GetTransactionById(c *gin.Context, transactionId int64, userId int64) (models.TransactionResponse, error)
	UpdateTransaction(c *gin.Context, transactionId int64, userId int64, input models.UpdateBaseTransactionInput) error
	DeleteTransaction(c *gin.Context, transactionId int64, userId int64) error
	ListTransactions(c *gin.Context, userId int64) ([]models.TransactionResponse, error)
	UpdateCategoryMapping(c *gin.Context, transactionId int64, userId int64, categoryIds []int64) error
}

type TransactionRepository struct {
	db                              *pgxpool.Pool
	schema                          string
	tableName                       string
	transactionCategoryMappingTable string
}

func NewTransactionRepository(db *database.DatabaseManager, cfg *config.Config) *TransactionRepository {
	return &TransactionRepository{
		db:                              db.GetPool(),
		schema:                          cfg.DBSchema,
		tableName:                       "transaction",
		transactionCategoryMappingTable: "transaction_category_mapping",
	}
}

var baseTransactionQuery = `
	SELECT t.id, t.name, t.description, t.amount, t.date, t.created_by, t.account_id,
		COALESCE(array_agg(DISTINCT tcm.category_id) FILTER (WHERE tcm.category_id IS NOT NULL), '{}') AS category_ids
	FROM %s.%s t
	LEFT JOIN %s.%s tcm ON t.id = tcm.transaction_id
`

func (r *TransactionRepository) CreateTransaction(c *gin.Context, transactionInput models.CreateBaseTransactionInput, categoryIds []int64) (models.TransactionResponse, error) {
	logger.Debugf("Creating transaction for user %d", transactionInput.CreatedBy)

	tx, err := r.db.Begin(c)
	if err != nil {
		return models.TransactionResponse{}, err
	}
	defer tx.Rollback(c)

	var transaction models.TransactionBaseResponse
	query, values, ptrs, err := helper.CreateInsertQuery(&transactionInput, &transaction, r.tableName, r.schema)
	if err != nil {
		return models.TransactionResponse{}, err
	}
	logger.Debugf("Executing query to create transaction: %s", query)
	err = tx.QueryRow(c, query, values...).Scan(ptrs...)
	if err != nil {
		if customErrors.CheckForeignKey(err, "idx_transaction_unique_composite") {
			return models.TransactionResponse{}, customErrors.NewTransactionAlreadyExistsError(err)
		}
		return models.TransactionResponse{}, err
	}

	if err := r.addMappings(c, tx, transaction.Id, categoryIds); err != nil {
		return models.TransactionResponse{}, err
	}

	if err := tx.Commit(c); err != nil {
		return models.TransactionResponse{}, err
	}

	logger.Debugf("Transaction created successfully with ID %d", transaction.Id)
	transactionResponse := models.TransactionResponse{
		TransactionBaseResponse: transaction,
		CategoryIds:             categoryIds,
	}
	logger.Debugf("Returning transaction response: %+v", transactionResponse)
	return transactionResponse, nil
}

func (r *TransactionRepository) addMappings(c *gin.Context, tx pgx.Tx, transactionId int64, categoryIds []int64) error {
	if err := r.updateMapping(c, tx, r.transactionCategoryMappingTable, "transaction_id", "category_id", transactionId, categoryIds); err != nil {
		if customErrors.CheckForeignKey(err, "fk_category") {
			return customErrors.NewCategoryNotFoundError(err)
		}
		return err
	}
	return nil
}

func scanTransaction(row pgx.Row) (models.TransactionResponse, error) {
	var resp models.TransactionResponse
	err := row.Scan(
		&resp.Id, &resp.Name, &resp.Description, &resp.Amount, &resp.Date, &resp.CreatedBy,
		&resp.AccountId, &resp.CategoryIds,
	)
	return resp, err
}

func (r *TransactionRepository) GetTransactionById(c *gin.Context, transactionId int64, userId int64) (models.TransactionResponse, error) {
	logger.Debugf("Fetching transaction ID %d for user %d", transactionId, userId)

	baseQuery := fmt.Sprintf(baseTransactionQuery, r.schema, r.tableName, r.schema, r.transactionCategoryMappingTable)
	query := baseQuery + ` WHERE t.id = $1 AND t.created_by = $2 AND t.deleted_at IS NULL GROUP BY t.id`
	logger.Debugf("Executing optimized query to get transaction by id: %s", query)
	row := r.db.QueryRow(c, query, transactionId, userId)
	resp, err := scanTransaction(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return resp, customErrors.NewTransactionNotFoundError(err)
		}
		return resp, err
	}
	logger.Debugf("Transaction retrieved successfully: %s", resp.Name)
	return resp, nil
}

func (r *TransactionRepository) UpdateTransaction(c *gin.Context, transactionId int64, userId int64, transactionUpdate models.UpdateBaseTransactionInput) error {
	logger.Debugf("Updating transaction ID %d for user %d", transactionId, userId)

	fieldsClause, argValues, argIndex, err := helper.CreateUpdateParams(&transactionUpdate)
	if err != nil {
		return err
	}
	var transaction models.TransactionBaseResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&transaction)
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`
	UPDATE %s.%s SET %s 
	WHERE id = $%d AND 
	created_by = $%d AND 
	deleted_at IS NULL 
	RETURNING %s;`,
		r.schema, r.tableName, fieldsClause, argIndex, argIndex+1, strings.Join(dbFields, ", "))

	logger.Debugf("Executing query to update transaction: %s", query)
	argValues = append(argValues, transactionId, userId)
	err = r.db.QueryRow(c, query, argValues...).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return customErrors.NewTransactionNotFoundError(err)
		}
		if customErrors.CheckForeignKey(err, "idx_transaction_unique_composite") {
			return customErrors.NewTransactionAlreadyExistsError(err)
		}
		return err
	}

	logger.Debugf("Transaction updated successfully: %s", transaction.Name)
	return nil
}

func (r *TransactionRepository) DeleteTransaction(c *gin.Context, transactionId int64, userId int64) error {
	logger.Debugf("Deleting transaction ID %d for user %d", transactionId, userId)

	query := fmt.Sprintf(`UPDATE %s.%s SET deleted_at = NOW() WHERE id = $1 AND created_by = $2 AND deleted_at IS NULL;`, r.schema, r.tableName)
	logger.Debugf("Executing query to delete transaction: %s", query)
	result, err := r.db.Exec(c, query, transactionId, userId)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return customErrors.NewTransactionNotFoundError(fmt.Errorf("transaction with id %d not found", transactionId))
	}

	logger.Debugf("Transaction deleted successfully with ID %d", transactionId)
	return nil
}

func (r *TransactionRepository) ListTransactions(c *gin.Context, userId int64) ([]models.TransactionResponse, error) {
	logger.Debugf("Fetching transactions for user %d", userId)

	baseQuery := fmt.Sprintf(baseTransactionQuery, r.schema, r.tableName, r.schema, r.transactionCategoryMappingTable)
	query := baseQuery + ` WHERE t.created_by = $1 AND t.deleted_at IS NULL GROUP BY t.id ORDER BY t.date DESC`
	logger.Debugf("Executing optimized query to list transactions: %s", query)
	rows, err := r.db.Query(c, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.TransactionResponse
	for rows.Next() {
		resp, err := scanTransaction(rows)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, resp)
	}
	logger.Debugf("Found %d transactions for user %d", len(transactions), userId)
	return transactions, nil
}

func (r *TransactionRepository) updateMapping(c *gin.Context, tx pgx.Tx, mappingTable, transactionColumn, idColumn string, transactionId int64, ids []int64) error {
	// Clear existing mappings
	_, err := tx.Exec(c, fmt.Sprintf(`DELETE FROM %s.%s WHERE %s = $1;`, r.schema, mappingTable, transactionColumn), transactionId)
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		return nil
	}

	// Prepare the insert statement
	query := fmt.Sprintf(`INSERT INTO %s.%s (%s, %s) VALUES ($1, $2) ON CONFLICT DO NOTHING;`, r.schema, mappingTable, idColumn, transactionColumn)
	batch := &pgx.Batch{}
	for _, id := range ids {
		batch.Queue(query, id, transactionId)
	}

	results := tx.SendBatch(c, batch)
	defer results.Close()

	for i := 0; i < len(ids); i++ {
		_, err := results.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *TransactionRepository) UpdateCategoryMapping(c *gin.Context, transactionId int64, userId int64, categoryIds []int64) error {
	logger.Debugf("Updating category mapping for transaction ID %d for user %d", transactionId, userId)

	tx, err := r.db.Begin(c)
	if err != nil {
		return err
	}
	defer tx.Rollback(c) // Rollback is a no-op if the transaction has been committed

	err = r.updateMapping(c, tx, r.transactionCategoryMappingTable, "transaction_id", "category_id", transactionId, categoryIds)
	if err != nil {
		return err
	}
	logger.Debugf("Category mapping updated successfully for transaction ID %d", transactionId)
	return tx.Commit(c)
}
