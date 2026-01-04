package repository

import (
	"context"
	"database/sql"
	"errors"
	"expenses/internal/config"
	"expenses/internal/database/helper"
	customErrors "expenses/internal/errors"
	"expenses/internal/models"
	database "expenses/pkg/database/manager"
	"expenses/pkg/utils"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type AccountRepositoryInterface interface {
	CreateAccount(ctx context.Context, input models.CreateAccountInput) (models.AccountResponse, error)
	GetAccountById(ctx context.Context, accountId int64, userId int64) (models.AccountResponse, error)
	UpdateAccount(ctx context.Context, accountId int64, userId int64, input models.UpdateAccountInput) (models.AccountResponse, error)
	DeleteAccount(ctx context.Context, accountId int64, userId int64) error
	ListAccounts(ctx context.Context, userId int64) ([]models.AccountResponse, error)
}

type AccountRepository struct {
	db                   database.DatabaseManager
	schema               string
	tableName            string
	investmentValueTable string
}

type accountDBInput struct {
	Name      string
	BankType  models.BankType
	Currency  string
	Balance   *float64
	CreatedBy int64
}

type accountDBUpdateInput struct {
	Name     string
	BankType models.BankType
	Currency string
	Balance  *float64
}

type accountDBOutput struct {
	Id        int64
	Name      string
	BankType  models.BankType
	Currency  string
	Balance   float64
	CreatedBy int64
}

func NewAccountRepository(db database.DatabaseManager, cfg *config.Config) AccountRepositoryInterface {
	return &AccountRepository{
		db:                   db,
		schema:               cfg.DBSchema,
		tableName:            "account",
		investmentValueTable: "investment_account_value",
	}
}

func (r *AccountRepository) CreateAccount(ctx context.Context, input models.CreateAccountInput) (models.AccountResponse, error) {
	var account models.AccountResponse
	var dbOutput accountDBOutput
	dbInput := accountDBInput{
		Name:      input.Name,
		BankType:  input.BankType,
		Currency:  input.Currency,
		Balance:   input.Balance,
		CreatedBy: input.CreatedBy,
	}

	query, values, ptrs, err := helper.CreateInsertQuery(&dbInput, &dbOutput, r.tableName, r.schema)
	if err != nil {
		return account, err
	}
	err = r.db.FetchOne(ctx, query, values...).Scan(ptrs...)
	if err != nil {
		return account, err
	}

	account = models.AccountResponse{
		Id:        dbOutput.Id,
		Name:      dbOutput.Name,
		BankType:  dbOutput.BankType,
		Currency:  dbOutput.Currency,
		Balance:   dbOutput.Balance,
		CreatedBy: dbOutput.CreatedBy,
	}

	if input.BankType == models.BankTypeInvestment && input.CurrentValue != nil {
		if err := r.upsertInvestmentCurrentValue(ctx, account.Id, *input.CurrentValue); err != nil {
			return account, err
		}
		account.CurrentValue = input.CurrentValue
	}

	return account, nil
}

func (r *AccountRepository) GetAccountById(ctx context.Context, accountId int64, userId int64) (models.AccountResponse, error) {
	var account models.AccountResponse
	var currentValue sql.NullFloat64

	query := fmt.Sprintf(`
		SELECT
			a.id,
			a.name,
			a.bank_type,
			a.currency,
			a.balance,
			a.created_by,
			iv.current_value
		FROM %s.%s a
		LEFT JOIN %s.%s iv ON iv.account_id = a.id
		WHERE a.id = $1 AND a.created_by = $2`,
		r.schema, r.tableName, r.schema, r.investmentValueTable)

	err := r.db.FetchOne(ctx, query, accountId, userId).Scan(
		&account.Id,
		&account.Name,
		&account.BankType,
		&account.Currency,
		&account.Balance,
		&account.CreatedBy,
		&currentValue,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return account, customErrors.NewAccountNotFoundError(err)
		}
		return account, err
	}
	if currentValue.Valid {
		account.CurrentValue = &currentValue.Float64
	}
	return account, nil
}

func (r *AccountRepository) UpdateAccount(ctx context.Context, accountId int64, userId int64, input models.UpdateAccountInput) (models.AccountResponse, error) {
	var account models.AccountResponse
	var dbInput accountDBUpdateInput
	utils.ConvertStruct(&input, &dbInput)

	fieldsClause, argValues, argIndex, err := helper.CreateUpdateParams(&dbInput)
	if err != nil {
		if !isNoFieldsToUpdateError(err) {
			return models.AccountResponse{}, err
		}
	} else {
		query := fmt.Sprintf(`UPDATE %s.%s SET %s WHERE id = $%d AND created_by = $%d RETURNING id;`, r.schema, r.tableName, fieldsClause, argIndex, argIndex+1)
		argValues = append(argValues, accountId, userId)
		var updatedId int64
		err = r.db.FetchOne(ctx, query, argValues...).Scan(&updatedId)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return account, customErrors.NewAccountNotFoundError(err)
			}
			return account, err
		}
	}

	account, err = r.GetAccountById(ctx, accountId, userId)
	if err != nil {
		return account, err
	}

	if account.BankType == models.BankTypeInvestment {
		if input.CurrentValue != nil {
			if err := r.upsertInvestmentCurrentValue(ctx, account.Id, *input.CurrentValue); err != nil {
				return account, err
			}
			account.CurrentValue = input.CurrentValue
		}
		return account, nil
	}

	if err := r.deleteInvestmentCurrentValue(ctx, account.Id); err != nil {
		return account, err
	}
	account.CurrentValue = nil

	return account, nil
}

func (r *AccountRepository) DeleteAccount(ctx context.Context, accountId int64, userId int64) error {
	query := fmt.Sprintf(`
		DELETE FROM %s.%s
		WHERE id = $1 AND created_by = $2`,
		r.schema, r.tableName)
	rowsAffected, err := r.db.ExecuteQuery(ctx, query, accountId, userId)
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

func (r *AccountRepository) ListAccounts(ctx context.Context, userId int64) ([]models.AccountResponse, error) {
	accounts := make([]models.AccountResponse, 0)
	query := fmt.Sprintf(`
		SELECT
			a.id,
			a.name,
			a.bank_type,
			a.currency,
			a.balance,
			a.created_by,
			iv.current_value
		FROM %s.%s a
		LEFT JOIN %s.%s iv ON iv.account_id = a.id
		WHERE a.created_by = $1
		ORDER BY a.created_at DESC`,
		r.schema, r.tableName, r.schema, r.investmentValueTable)
	rows, err := r.db.FetchAll(ctx, query, userId)
	if err != nil {
		return accounts, err
	}
	defer rows.Close()
	for rows.Next() {
		var account models.AccountResponse
		var currentValue sql.NullFloat64
		if err := rows.Scan(
			&account.Id,
			&account.Name,
			&account.BankType,
			&account.Currency,
			&account.Balance,
			&account.CreatedBy,
			&currentValue,
		); err != nil {
			return accounts, err
		}
		if currentValue.Valid {
			account.CurrentValue = &currentValue.Float64
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func (r *AccountRepository) upsertInvestmentCurrentValue(ctx context.Context, accountId int64, currentValue float64) error {
	query := fmt.Sprintf(`
		INSERT INTO %s.%s (account_id, current_value)
		VALUES ($1, $2)
		ON CONFLICT (account_id) DO UPDATE
		SET current_value = EXCLUDED.current_value,
			updated_at = CURRENT_TIMESTAMP`,
		r.schema, r.investmentValueTable)
	_, err := r.db.ExecuteQuery(ctx, query, accountId, currentValue)
	return err
}

func (r *AccountRepository) deleteInvestmentCurrentValue(ctx context.Context, accountId int64) error {
	query := fmt.Sprintf(`
		DELETE FROM %s.%s
		WHERE account_id = $1`,
		r.schema, r.investmentValueTable)
	_, err := r.db.ExecuteQuery(ctx, query, accountId)
	return err
}

func isNoFieldsToUpdateError(err error) bool {
	var authErr *customErrors.AuthError
	if errors.As(err, &authErr) {
		return authErr.ErrorType == "NoFieldsToUpdate"
	}
	return false
}
