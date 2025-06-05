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
	CreateTransaction(c *gin.Context, input models.CreateTransactionInput) (models.TransactionResponse, error)
	GetTransactionById(c *gin.Context, transactionId int64, userId int64) (models.TransactionResponse, error)
	UpdateTransaction(c *gin.Context, transactionId int64, userId int64, input models.UpdateTransactionInput) (models.TransactionResponse, error)
	DeleteTransaction(c *gin.Context, transactionId int64, userId int64) error
	ListTransactions(c *gin.Context, userId int64) ([]models.TransactionResponse, error)
}

type TransactionRepository struct {
	db        *pgxpool.Pool
	schema    string
	tableName string
}

func NewTransactionRepository(db *database.DatabaseManager, cfg *config.Config) *TransactionRepository {
	return &TransactionRepository{
		db:        db.GetPool(),
		schema:    cfg.DBSchema,
		tableName: "transaction",
	}
}

func (r *TransactionRepository) CreateTransaction(c *gin.Context, input models.CreateTransactionInput) (models.TransactionResponse, error) {
	var transaction models.TransactionResponse
	query, values, ptrs, err := helper.CreateInsertQuery(&input, &transaction, r.tableName, r.schema)
	if err != nil {
		return transaction, err
	}
	logger.Info("Executing query to create transaction: ", query)
	err = r.db.QueryRow(c, query, values...).Scan(ptrs...)
	if err != nil {
		// Check for unique constraint violation on the composite index
		if customErrors.CheckForeignKey(err, "idx_transaction_unique_composite") {
			return transaction, customErrors.NewTransactionAlreadyExistsError(err)
		}
		return transaction, err
	}
	return transaction, nil
}

func (r *TransactionRepository) GetTransactionById(c *gin.Context, transactionId int64, userId int64) (models.TransactionResponse, error) {
	var transaction models.TransactionResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&transaction)
	if err != nil {
		return transaction, err
	}
	query := fmt.Sprintf(`SELECT %s FROM %s.%s WHERE id = $1 AND created_by = $2 AND deleted_at IS NULL;`, strings.Join(dbFields, ", "), r.schema, r.tableName)
	logger.Info("Executing query to get transaction by id: ", query)
	err = r.db.QueryRow(c, query, transactionId, userId).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return transaction, customErrors.NewTransactionNotFoundError(err)
		}
		return transaction, err
	}
	return transaction, nil
}

func (r *TransactionRepository) UpdateTransaction(c *gin.Context, transactionId int64, userId int64, input models.UpdateTransactionInput) (models.TransactionResponse, error) {
	fieldsClause, argValues, argIndex, err := helper.CreateUpdateParams(&input)
	if err != nil {
		return models.TransactionResponse{}, err
	}
	var transaction models.TransactionResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&transaction)
	if err != nil {
		return transaction, err
	}
	query := fmt.Sprintf(`UPDATE %s.%s SET %s WHERE id = $%d AND created_by = $%d AND deleted_at IS NULL RETURNING %s;`, r.schema, r.tableName, fieldsClause, argIndex, argIndex+1, strings.Join(dbFields, ", "))
	logger.Info("Executing query to update transaction: ", query)
	argValues = append(argValues, transactionId, userId)
	err = r.db.QueryRow(c, query, argValues...).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return transaction, customErrors.NewTransactionNotFoundError(err)
		}
		// Check for unique constraint violation on the composite index
		if customErrors.CheckForeignKey(err, "idx_transaction_unique_composite") {
			return transaction, customErrors.NewTransactionAlreadyExistsError(err)
		}
		return transaction, err
	}
	return transaction, nil
}

func (r *TransactionRepository) DeleteTransaction(c *gin.Context, transactionId int64, userId int64) error {
	query := fmt.Sprintf(`UPDATE %s.%s SET deleted_at = NOW() WHERE id = $1 AND created_by = $2 AND deleted_at IS NULL;`, r.schema, r.tableName)
	logger.Info("Executing query to delete transaction: ", query)
	_, err := r.db.Exec(c, query, transactionId, userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return customErrors.NewTransactionNotFoundError(err)
		}
		return err
	}
	return nil
}

func (r *TransactionRepository) ListTransactions(c *gin.Context, userId int64) ([]models.TransactionResponse, error) {
	_, dbFields, err := helper.GetDbFieldsFromObject(&models.TransactionResponse{})
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf(`SELECT %s FROM %s.%s WHERE created_by = $1 AND deleted_at IS NULL ORDER BY date DESC;`, strings.Join(dbFields, ", "), r.schema, r.tableName)
	logger.Info("Executing query to list transactions: ", query)
	rows, err := r.db.Query(c, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var transactions []models.TransactionResponse
	for rows.Next() {
		var transaction models.TransactionResponse
		ptrs, _, err := helper.GetDbFieldsFromObject(&transaction)
		if err != nil {
			return nil, err
		}
		err = rows.Scan(ptrs...)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
} 