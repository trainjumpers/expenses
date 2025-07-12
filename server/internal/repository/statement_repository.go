package repository

import (
	"errors"
	"expenses/internal/config"
	"expenses/internal/database/helper"
	database "expenses/internal/database/manager"
	statementErrors "expenses/internal/errors"
	"expenses/internal/models"
	"expenses/pkg/logger"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type StatementRepositoryInterface interface {
	CreateStatement(c *gin.Context, input models.CreateStatementInput) (models.StatementResponse, error)
	CreateStatementTxn(c *gin.Context, statementId int64, transactionId int64) error
	UpdateStatementStatus(c *gin.Context, statementId int64, input models.UpdateStatementStatusInput) (models.StatementResponse, error)
	GetStatementByID(c *gin.Context, statementId int64, userId int64) (models.StatementResponse, error)
	ListStatementByUserId(c *gin.Context, userId int64, limit, offset int) ([]models.StatementResponse, error)
	CountStatementsByUserId(c *gin.Context, userId int64) (int, error)
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

func (r *StatementRepository) CreateStatement(c *gin.Context, input models.CreateStatementInput) (models.StatementResponse, error) {
	var statement models.StatementResponse
	query, values, ptrs, err := helper.CreateInsertQuery(&input, &statement, r.tableName, r.schema)
	if err != nil {
		return statement, statementErrors.NewStatementCreateError(err)
	}
	err = r.db.FetchOne(c, query, values...).Scan(ptrs...)
	if err != nil {
		if statementErrors.CheckForeignKey(err, "statement_account_id_fkey") {
			return statement, statementErrors.NewAccountNotFoundError(errors.New("account not found"))
		}
		return statement, statementErrors.NewStatementCreateError(err)
	}
	return statement, nil
}

func (r *StatementRepository) CreateStatementTxn(c *gin.Context, statementId int64, transactionId int64) error {
	query := fmt.Sprintf(`INSERT INTO %s.%s (statement_id, transaction_id) VALUES ($1, $2)`, r.schema, r.mappingTableName)
	_, err := r.db.ExecuteQuery(c, query, statementId, transactionId)
	if err != nil {
		return statementErrors.NewStatementCreateError(err)
	}
	return nil
}

func (r *StatementRepository) UpdateStatementStatus(c *gin.Context, statementId int64, input models.UpdateStatementStatusInput) (models.StatementResponse, error) {
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
	err = r.db.FetchOne(c, query, argValues...).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return statement, statementErrors.NewStatementNotFoundError(err)
		}
		return statement, statementErrors.NewStatementUpdateError(err)
	}
	return statement, nil
}

func (r *StatementRepository) GetStatementByID(c *gin.Context, statementId int64, userId int64) (models.StatementResponse, error) {
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
	err = r.db.FetchOne(c, query, statementId, userId).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return statement, statementErrors.NewStatementNotFoundError(err)
		}
		return statement, statementErrors.NewStatementNotFoundError(err)
	}
	return statement, nil
}

func (r *StatementRepository) ListStatementByUserId(c *gin.Context, userId int64, limit, offset int) ([]models.StatementResponse, error) {
	statements := make([]models.StatementResponse, 0)
	var statement models.StatementResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&statement)
	if err != nil {
		return statements, err
	}
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s.%s
		WHERE created_by = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		strings.Join(dbFields, ", "), r.schema, r.tableName)
	rows, err := r.db.FetchAll(c, query, userId, limit, offset)
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

func (r *StatementRepository) CountStatementsByUserId(c *gin.Context, userId int64) (int, error) {
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s.%s WHERE created_by = $1 AND deleted_at IS NULL`, r.schema, r.tableName)
	var count int
	err := r.db.FetchOne(c, query, userId).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
