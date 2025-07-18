package repository

import (
	"errors"
	"expenses/internal/config"
	"expenses/internal/database/helper"
	database "expenses/internal/database/manager"
	customErrors "expenses/internal/errors"
	"expenses/internal/models"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type TransactionRepositoryInterface interface {
	CreateTransaction(c *gin.Context, transaction models.CreateBaseTransactionInput, categoryIds []int64) (models.TransactionResponse, error)
	GetTransactionById(c *gin.Context, transactionId int64, userId int64) (models.TransactionResponse, error)
	UpdateTransaction(c *gin.Context, transactionId int64, userId int64, input models.UpdateBaseTransactionInput) error
	DeleteTransaction(c *gin.Context, transactionId int64, userId int64) error
	ListTransactions(c *gin.Context, userId int64, query models.TransactionListQuery) (models.PaginatedTransactionsResponse, error)
	UpdateCategoryMapping(c *gin.Context, transactionId int64, userId int64, categoryIds []int64) error
}

type TransactionRepository struct {
	db                              database.DatabaseManager
	schema                          string
	tableName                       string
	transactionCategoryMappingTable string
}

func NewTransactionRepository(db database.DatabaseManager, cfg *config.Config) TransactionRepositoryInterface {
	return &TransactionRepository{
		db:                              db,
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
	var transactionResponse models.TransactionResponse
	var transaction models.TransactionBaseResponse

	err := r.db.WithTxn(c, func(tx pgx.Tx) error {
		query, values, ptrs, err := helper.CreateInsertQuery(&transactionInput, &transaction, r.tableName, r.schema)
		if err != nil {
			return err
		}
		err = tx.QueryRow(c, query, values...).Scan(ptrs...)
		if err != nil {
			if customErrors.CheckForeignKey(err, "idx_transaction_unique_composite") {
				return customErrors.NewTransactionAlreadyExistsError(err)
			}
			return err
		}

		if err := r.addMappings(c, tx, transaction.Id, categoryIds); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return models.TransactionResponse{}, err
	}

	transactionResponse = models.TransactionResponse{
		TransactionBaseResponse: transaction,
		CategoryIds:             categoryIds,
	}
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
	baseQuery := fmt.Sprintf(baseTransactionQuery, r.schema, r.tableName, r.schema, r.transactionCategoryMappingTable)
	query := baseQuery + ` WHERE t.id = $1 AND t.created_by = $2 AND t.deleted_at IS NULL GROUP BY t.id`
	row := r.db.FetchOne(c, query, transactionId, userId)
	resp, err := scanTransaction(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return resp, customErrors.NewTransactionNotFoundError(err)
		}
		return resp, err
	}
	return resp, nil
}

func (r *TransactionRepository) UpdateTransaction(c *gin.Context, transactionId int64, userId int64, transactionUpdate models.UpdateBaseTransactionInput) error {
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

	argValues = append(argValues, transactionId, userId)
	err = r.db.FetchOne(c, query, argValues...).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return customErrors.NewTransactionNotFoundError(err)
		}
		if customErrors.CheckForeignKey(err, "idx_transaction_unique_composite") {
			return customErrors.NewTransactionAlreadyExistsError(err)
		}
		return err
	}

	return nil
}

func (r *TransactionRepository) DeleteTransaction(c *gin.Context, transactionId int64, userId int64) error {
	query := fmt.Sprintf(`UPDATE %s.%s SET deleted_at = NOW() WHERE id = $1 AND created_by = $2 AND deleted_at IS NULL;`, r.schema, r.tableName)
	rowsAffected, err := r.db.ExecuteQuery(c, query, transactionId, userId)
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return customErrors.NewTransactionNotFoundError(fmt.Errorf("transaction with id %d not found", transactionId))
	}

	return nil
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
	err := r.db.WithTxn(c, func(tx pgx.Tx) error {
		return r.updateMapping(c, tx, r.transactionCategoryMappingTable, "transaction_id", "category_id", transactionId, categoryIds)
	})

	if err != nil {
		return err
	}

	return nil
}

// Helper to build WHERE clause and args for transaction queries
func (r *TransactionRepository) buildTransactionWhereClause(userId int64, q models.TransactionListQuery) (string, []any) {
	args := []any{userId}
	where := []string{"t.created_by = $1", "t.deleted_at IS NULL"}
	argIdx := len(args) + 1 // +1 because it is 1 based index

	if q.AccountId != nil {
		where = append(where, fmt.Sprintf("t.account_id = $%d", argIdx))
		args = append(args, *q.AccountId)
		argIdx++
	}
	if q.MinAmount != nil {
		where = append(where, fmt.Sprintf("t.amount >= $%d", argIdx))
		args = append(args, *q.MinAmount)
		argIdx++
	}
	if q.MaxAmount != nil {
		where = append(where, fmt.Sprintf("t.amount <= $%d", argIdx))
		args = append(args, *q.MaxAmount)
		argIdx++
	}
	if q.DateFrom != nil {
		where = append(where, fmt.Sprintf("t.date >= $%d", argIdx))
		args = append(args, *q.DateFrom)
		argIdx++
	}
	if q.DateTo != nil {
		where = append(where, fmt.Sprintf("t.date <= $%d", argIdx))
		args = append(args, *q.DateTo)
		argIdx++
	}
	if q.Search != nil && *q.Search != "" {
		where = append(where, fmt.Sprintf("(t.name ILIKE $%d OR t.description ILIKE $%d)", argIdx, argIdx+1))
		args = append(args, "%"+*q.Search+"%", "%"+*q.Search+"%")
		argIdx += 2
	}
	if q.CategoryId != nil {
		where = append(where, fmt.Sprintf("t.id IN (SELECT transaction_id FROM %s.%s WHERE category_id = $%d)", r.schema, r.transactionCategoryMappingTable, argIdx))
		args = append(args, *q.CategoryId)
		argIdx++
	}
	if q.StatementId != nil {
		where = append(where, fmt.Sprintf("t.id IN (SELECT transaction_id FROM %s.statement_transaction_mapping WHERE statement_id = $%d)", r.schema, argIdx))
		args = append(args, *q.StatementId)
		argIdx++
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = " WHERE " + strings.Join(where, " AND ")
	}
	return whereClause, args
}

// Helper to build the count query for transactions
func (r *TransactionRepository) buildTransactionCountQuery(q models.TransactionListQuery, whereClause string) string {
	countQuery := fmt.Sprintf("SELECT COUNT(DISTINCT t.id) FROM %s.%s t ", r.schema, r.tableName)
	countQuery += whereClause
	return countQuery
}

// Helper to build the data query for transactions
func (r *TransactionRepository) buildTransactionDataQuery(q models.TransactionListQuery, whereClause string, page, pageSize int) string {
	sortBy := "t.date"
	if q.SortBy != "" {
		switch q.SortBy {
		case "amount", "date", "name":
			sortBy = "t." + q.SortBy
		}
	}
	sortOrder := "DESC"
	if strings.ToLower(q.SortOrder) == "asc" {
		sortOrder = "ASC"
	}
	offset := (page - 1) * pageSize
	baseQuery := fmt.Sprintf(baseTransactionQuery, r.schema, r.tableName, r.schema, r.transactionCategoryMappingTable)
	dataQuery := baseQuery + whereClause + " GROUP BY t.id ORDER BY " + sortBy + " " + sortOrder + fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)
	return dataQuery
}

// ListTransactions returns paginated, sorted, and filtered transactions for a user
func (r *TransactionRepository) ListTransactions(c *gin.Context, userId int64, q models.TransactionListQuery) (models.PaginatedTransactionsResponse, error) {
	var resp models.PaginatedTransactionsResponse
	transactions := make([]models.TransactionResponse, 0)

	// Build WHERE clause and args
	whereClause, args := r.buildTransactionWhereClause(userId, q)
	// Count query
	countQuery := r.buildTransactionCountQuery(q, whereClause)
	var total int
	err := r.db.FetchOne(c, countQuery, args...).Scan(&total)
	if err != nil {
		return resp, err
	}

	// Data query
	dataQuery := r.buildTransactionDataQuery(q, whereClause, q.Page, q.PageSize)
	rows, err := r.db.FetchAll(c, dataQuery, args...)
	if err != nil {
		return resp, err
	}
	defer rows.Close()

	for rows.Next() {
		tx, err := scanTransaction(rows)
		if err != nil {
			return resp, err
		}
		transactions = append(transactions, tx)
	}

	resp = models.PaginatedTransactionsResponse{
		Transactions: transactions,
		Total:        total,
		Page:         q.Page,
		PageSize:     q.PageSize,
	}
	return resp, nil
}
