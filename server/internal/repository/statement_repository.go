package repository

import (
	"context"
	"errors"
	"expenses/internal/config"
	"expenses/internal/database/helper"
	statementErrors "expenses/internal/errors"
	"expenses/internal/models"
	database "expenses/pkg/database/manager"
	"expenses/pkg/logger"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

type StatementRepositoryInterface interface {
	CreateStatement(ctx context.Context, input models.CreateStatementInput) (models.StatementResponse, error)
	CreateStatementTxn(ctx context.Context, statementId int64, transactionId int64) error
	CreateStatementTxns(ctx context.Context, statementId int64, transactionIds []int64) error
	UpdateStatementStatus(ctx context.Context, statementId int64, input models.UpdateStatementStatusInput) (models.StatementResponse, error)
	GetStatementByID(ctx context.Context, statementId int64, userId int64) (models.StatementResponse, error)
	ListStatementByUserId(ctx context.Context, userId int64, limit, offset int, query models.StatementListQuery) ([]models.StatementResponse, error)
	CountStatementsByUserId(ctx context.Context, userId int64, query models.StatementListQuery) (int, error)
}

type StatementRepository struct {
	db               database.DatabaseManager
	schema           string
	tableName        string
	mappingTableName string
}

func NewStatementRepository(db database.DatabaseManager, cfg *config.Config) StatementRepositoryInterface {
	return &StatementRepository{
		db:               db,
		schema:           cfg.DBSchema,
		tableName:        "statement",
		mappingTableName: "statement_transaction_mapping",
	}
}

func (r *StatementRepository) CreateStatement(ctx context.Context, input models.CreateStatementInput) (models.StatementResponse, error) {
	var statement models.StatementResponse
	query, values, ptrs, err := helper.CreateInsertQuery(&input, &statement, r.tableName, r.schema)
	if err != nil {
		return statement, statementErrors.NewStatementCreateError(err)
	}
	err = r.db.FetchOne(ctx, query, values...).Scan(ptrs...)
	if err != nil {
		if statementErrors.CheckForeignKey(err, "statement_account_id_fkey") {
			return statement, statementErrors.NewAccountNotFoundError(errors.New("account not found"))
		}
		return statement, statementErrors.NewStatementCreateError(err)
	}
	return statement, nil
}

func (r *StatementRepository) CreateStatementTxn(ctx context.Context, statementId int64, transactionId int64) error {
	query := fmt.Sprintf(`INSERT INTO %s.%s (statement_id, transaction_id) VALUES ($1, $2)`, r.schema, r.mappingTableName)
	_, err := r.db.ExecuteQuery(ctx, query, statementId, transactionId)
	if err != nil {
		return statementErrors.NewStatementCreateError(err)
	}
	return nil
}

func (r *StatementRepository) CreateStatementTxns(ctx context.Context, statementId int64, transactionIds []int64) error {
	if len(transactionIds) == 0 {
		return nil
	}

	const batchSize = 1000

	// Process in batches of 1000
	for batchStart := 0; batchStart < len(transactionIds); batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > len(transactionIds) {
			batchEnd = len(transactionIds)
		}

		batchTxIds := transactionIds[batchStart:batchEnd]

		// Build bulk insert using VALUES for this batch
		placeholders := make([]string, 0, len(batchTxIds))
		args := make([]interface{}, 0, len(batchTxIds)*2)
		argIndex := 1

		for _, txID := range batchTxIds {
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", argIndex, argIndex+1))
			args = append(args, statementId, txID)
			argIndex += 2
		}

		query := fmt.Sprintf(`INSERT INTO %s.%s (statement_id, transaction_id) VALUES %s`,
			r.schema, r.mappingTableName, strings.Join(placeholders, ", "))

		_, err := r.db.ExecuteQuery(ctx, query, args...)
		if err != nil {
			return statementErrors.NewStatementCreateError(err)
		}
	}

	return nil
}

func (r *StatementRepository) UpdateStatementStatus(ctx context.Context, statementId int64, input models.UpdateStatementStatusInput) (models.StatementResponse, error) {
	logger.Debugf("Updating statement %d with status %s", statementId, input.Status)
	fieldsClause, argValues, argIndex, err := helper.CreateUpdateParams(&input)
	if err != nil {
		return models.StatementResponse{}, statementErrors.NewStatementUpdateError(err)
	}
	var statement models.StatementResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&statement)
	if err != nil {
		return statement, statementErrors.NewStatementUpdateError(err)
	}
	query := fmt.Sprintf(`UPDATE %s.%s SET %s WHERE id = $%d RETURNING %s;`, r.schema, r.tableName, fieldsClause, argIndex, strings.Join(dbFields, ", "))
	argValues = append(argValues, statementId)
	err = r.db.FetchOne(ctx, query, argValues...).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return statement, statementErrors.NewStatementNotFoundError(err)
		}
		return statement, statementErrors.NewStatementUpdateError(err)
	}
	return statement, nil
}

func (r *StatementRepository) GetStatementByID(ctx context.Context, statementId int64, userId int64) (models.StatementResponse, error) {
	var statement models.StatementResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&statement)
	if err != nil {
		return statement, err
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s.%s
		WHERE id = $1 AND created_by = $2 AND deleted_at IS NULL`,
		strings.Join(dbFields, ", "), r.schema, r.tableName)
	err = r.db.FetchOne(ctx, query, statementId, userId).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return statement, statementErrors.NewStatementNotFoundError(err)
		}
		return statement, statementErrors.NewStatementNotFoundError(err)
	}
	return statement, nil
}

func (r *StatementRepository) ListStatementByUserId(ctx context.Context, userId int64, limit, offset int, query models.StatementListQuery) ([]models.StatementResponse, error) {
	statements := make([]models.StatementResponse, 0)
	var statement models.StatementResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&statement)
	if err != nil {
		return statements, err
	}

	whereClauses := []string{"created_by = $1", "deleted_at IS NULL"}
	argIndex := 2
	args := []interface{}{userId}

	if query.AccountId != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("account_id = $%d", argIndex))
		args = append(args, *query.AccountId)
		argIndex++
	}

	if query.DateFrom != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *query.DateFrom)
		argIndex++
	}

	if query.DateTo != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *query.DateTo)
		argIndex++
	}

	if query.Search != nil && *query.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("original_filename ILIKE $%d", argIndex))
		args = append(args, "%"+*query.Search+"%")
		argIndex++
	}

	sqlQuery := fmt.Sprintf(`
		SELECT %s
		FROM %s.%s
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`,
		strings.Join(dbFields, ", "), r.schema, r.tableName, strings.Join(whereClauses, " AND "), argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.FetchAll(ctx, sqlQuery, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return statements, nil
		}
		return statements, statementErrors.NewStatementGetError(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(ptrs...)
		if err != nil {
			return statements, statementErrors.NewStatementGetError(err)
		}
		statements = append(statements, statement)
	}
	return statements, nil
}

func (r *StatementRepository) CountStatementsByUserId(ctx context.Context, userId int64, query models.StatementListQuery) (int, error) {
	whereClauses := []string{"created_by = $1", "deleted_at IS NULL"}
	argIndex := 2
	args := []interface{}{userId}

	if query.AccountId != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("account_id = $%d", argIndex))
		args = append(args, *query.AccountId)
		argIndex++
	}

	if query.DateFrom != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *query.DateFrom)
		argIndex++
	}

	if query.DateTo != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *query.DateTo)
		argIndex++
	}

	if query.Search != nil && *query.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("original_filename ILIKE $%d", argIndex))
		args = append(args, "%"+*query.Search+"%")
		argIndex++
	}

	sqlQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s.%s WHERE %s`, r.schema, r.tableName, strings.Join(whereClauses, " AND "))
	var count int
	err := r.db.FetchOne(ctx, sqlQuery, args...).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
