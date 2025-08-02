package repository

import (
	"errors"
	"expenses/internal/config"
	"expenses/internal/database/helper"
	customErrors "expenses/internal/errors"
	"expenses/internal/models"
	database "expenses/pkg/database/manager"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type AccountRepositoryInterface interface {
	CreateAccount(c *gin.Context, input models.CreateAccountInput) (models.AccountResponse, error)
	GetAccountById(c *gin.Context, accountId int64, userId int64) (models.AccountResponse, error)
	UpdateAccount(c *gin.Context, accountId int64, userId int64, input models.UpdateAccountInput) (models.AccountResponse, error)
	DeleteAccount(c *gin.Context, accountId int64, userId int64) error
	ListAccounts(c *gin.Context, userId int64) ([]models.AccountResponse, error)
}

type AccountRepository struct {
	db        database.DatabaseManager
	schema    string
	tableName string
}

func NewAccountRepository(db database.DatabaseManager, cfg *config.Config) AccountRepositoryInterface {
	return &AccountRepository{
		db:        db,
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
	err = r.db.FetchOne(c, query, values...).Scan(ptrs...)
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
	err = r.db.FetchOne(c, query, accountId, userId).Scan(ptrs...)
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
	argValues = append(argValues, accountId, userId)
	err = r.db.FetchOne(c, query, argValues...).Scan(ptrs...)
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
	rowsAffected, err := r.db.ExecuteQuery(c, query, accountId, userId)
	if err != nil {
		if customErrors.CheckForeignKey(err, "fk_transaction_account_id") {
			return customErrors.NewAccountHasTransactionsError(err)
		}
		return err
	}
	if rowsAffected == 0 {
		return customErrors.NewAccountNotFoundError(errors.New("account not found or not owned by user"))
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
	rows, err := r.db.FetchAll(c, query, userId)
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
