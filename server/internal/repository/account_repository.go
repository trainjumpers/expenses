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

type AccountRepositoryInterface interface {
	CreateAccount(c *gin.Context, input models.CreateAccountInput) (models.AccountResponse, error)
	GetAccountById(c *gin.Context, accountId int64, userId int64) (models.AccountResponse, error)
	UpdateAccount(c *gin.Context, accountId int64, userId int64, input models.UpdateAccountInput) (models.AccountResponse, error)
	DeleteAccount(c *gin.Context, accountId int64, userId int64) error
	ListAccounts(c *gin.Context, userId int64) ([]models.AccountResponse, error)
}

type AccountRepository struct {
	db        *pgxpool.Pool
	schema    string
	tableName string
}

func NewAccountRepository(db *database.DatabaseManager, cfg *config.Config) *AccountRepository {
	return &AccountRepository{
		db:        db.GetPool(),
		schema:    cfg.DBSchema,
		tableName: "account",
	}
}

func (r *AccountRepository) CreateAccount(c *gin.Context, input models.CreateAccountInput) (models.AccountResponse, error) {
	var account models.AccountResponse
	query, values, ptrs, err := helper.CreateInsertQuery(&input, &account, r.tableName, r.schema)
	if err != nil {
		return account, err
	}
	logger.Debugf("Executing query to create account: %s", query)
	err = r.db.QueryRow(c, query, values...).Scan(ptrs...)
	if err != nil {
		return account, err
	}
	return account, nil
}

func (r *AccountRepository) GetAccountById(c *gin.Context, accountId int64, userId int64) (models.AccountResponse, error) {
	var account models.AccountResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&account)
	if err != nil {
		return account, err
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s.%s
		WHERE id = $1 AND created_by = $2`,
		strings.Join(dbFields, ", "), r.schema, r.tableName)
	logger.Debugf("Executing query to get account by ID: %s", query)
	err = r.db.QueryRow(c, query, accountId, userId).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return account, customErrors.NewAccountNotFoundError(err)
		}
		return account, err
	}
	return account, nil
}

func (r *AccountRepository) UpdateAccount(c *gin.Context, accountId int64, userId int64, input models.UpdateAccountInput) (models.AccountResponse, error) {
	fieldsClause, argValues, argIndex, err := helper.CreateUpdateParams(&input)
	if err != nil {
		return models.AccountResponse{}, err
	}
	var account models.AccountResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&account)
	if err != nil {
		return account, err
	}
	query := fmt.Sprintf(`UPDATE %s.%s SET %s WHERE id = $%d AND created_by = $%d RETURNING %s;`, r.schema, r.tableName, fieldsClause, argIndex, argIndex+1, strings.Join(dbFields, ", "))
	logger.Debugf("Executing query to update account: %s", query)
	argValues = append(argValues, accountId, userId)
	err = r.db.QueryRow(c, query, argValues...).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return account, customErrors.NewAccountNotFoundError(err)
		}
		return account, err
	}
	return account, nil
}

func (r *AccountRepository) DeleteAccount(c *gin.Context, accountId int64, userId int64) error {
	query := fmt.Sprintf(`
		DELETE FROM %s.%s
		WHERE id = $1 AND created_by = $2`,
		r.schema, r.tableName)
	logger.Debugf("Executing query to delete account: %s", query)
	_, err := r.db.Exec(c, query, accountId, userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return customErrors.NewAccountNotFoundError(err)
		}
		return err
	}
	return nil
}

func (r *AccountRepository) ListAccounts(c *gin.Context, userId int64) ([]models.AccountResponse, error) {
	accounts := make([]models.AccountResponse, 0)
	var account models.AccountResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&account)
	if err != nil {
		return accounts, err
	}
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s.%s
		WHERE created_by = $1
		ORDER BY created_at DESC`,
		strings.Join(dbFields, ", "), r.schema, r.tableName)
	logger.Debugf("Executing query to list accounts: %s", query)
	rows, err := r.db.Query(c, query, userId)
	if err != nil {
		return accounts, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(ptrs...)
		if err != nil {
			return accounts, err
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}
