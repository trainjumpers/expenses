package repository

import (
	"context"
	"expenses/internal/config"
	"expenses/internal/models"
	database "expenses/pkg/database/manager"
	"fmt"
	"strings"
	"time"
)

type AnalyticsRepositoryInterface interface {
	GetBalance(ctx context.Context, userId int64, startDate *time.Time, endDate *time.Time) (map[int64]float64, error)
	GetNetworthTimeSeries(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (float64, float64, float64, []map[string]any, error)
	GetCategoryAnalytics(ctx context.Context, userId int64, startDate time.Time, endDate time.Time, categoryIds []int64) (*models.CategoryAnalyticsResponse, error)
	GetMonthlyAnalytics(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (*models.MonthlyAnalyticsResponse, error)
	GetAccountCashFlows(ctx context.Context, userId int64, accountIds []int64) ([]models.AccountCashFlow, error)
}

type AnalyticsRepository struct {
	db           database.DatabaseManager
	schema       string
	txnTableName string
}

func NewAnalyticsRepository(db database.DatabaseManager, cfg *config.Config) AnalyticsRepositoryInterface {
	return &AnalyticsRepository{
		db:           db,
		schema:       cfg.DBSchema,
		txnTableName: "transaction",
	}
}

// GetBalance calculates account balances within an optional date range
// startDate = nil, endDate = nil: All transactions (current balance)
// startDate = nil, endDate = oneMonthAgo: Balance up to one month ago
// Returns map[accountId]balance for efficient lookup
// Only returns accounts that have transaction data
func (r *AnalyticsRepository) GetBalance(ctx context.Context, userId int64, startDate *time.Time, endDate *time.Time) (map[int64]float64, error) {
	balances := make(map[int64]float64)

	query := fmt.Sprintf(`
		SELECT 
			account_id,
			COALESCE(SUM(amount), 0) * -1 as balance
		FROM %s.%s
		WHERE created_by = $1 
			AND deleted_at IS NULL
			AND ($2::DATE IS NULL OR date >= $2)
			AND ($3::DATE IS NULL OR date < $3)
		GROUP BY account_id`,
		r.schema, r.txnTableName)

	rows, err := r.db.FetchAll(ctx, query, userId, startDate, endDate)
	if err != nil {
		return balances, err
	}
	defer rows.Close()

	for rows.Next() {
		var accountId int64
		var balance float64
		err := rows.Scan(&accountId, &balance)
		if err != nil {
			return balances, err
		}
		balances[accountId] = balance
	}

	return balances, nil
}

// GetNetworthTimeSeries calculates the initial balance and daily networth changes
// Returns initial balance (sum of all transactions before startDate) and daily aggregated data
func (r *AnalyticsRepository) GetNetworthTimeSeries(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (float64, float64, float64, []map[string]any, error) {
	// First, get the initial balance (sum of all transactions before startDate)
	initialBalanceQuery := fmt.Sprintf(`
		SELECT COALESCE(SUM(amount), 0) * -1 as initial_balance
		FROM %s.%s
		WHERE created_by = $1
			AND deleted_at IS NULL
			AND date < $2`,
		r.schema, r.txnTableName)

	var initialBalance float64
	row := r.db.FetchOne(ctx, initialBalanceQuery, userId, startDate)
	err := row.Scan(&initialBalance)
	if err != nil {
		return 0, 0, 0, nil, err
	}

	// Get daily transaction sums within the date range
	timeSeriesQuery := fmt.Sprintf(`
		SELECT
			date,
			COALESCE(SUM(amount), 0) * -1 as daily_change,
			COALESCE(SUM(CASE WHEN amount > 0 THEN amount ELSE 0 END), 0) as total_expenses,
			COALESCE(SUM(CASE WHEN amount < 0 THEN amount * -1 ELSE 0 END), 0) as total_income
		FROM %s.%s
		WHERE created_by = $1
			AND deleted_at IS NULL
			AND date >= $2
			AND date <= $3
		GROUP BY date
		ORDER BY date`,
		r.schema, r.txnTableName)

	rows, err := r.db.FetchAll(ctx, timeSeriesQuery, userId, startDate, endDate)
	if err != nil {
		return initialBalance, 0, 0, nil, err
	}
	defer rows.Close()

	var timeSeries []map[string]any
	var totalIncome float64
	var totalExpenses float64
	for rows.Next() {
		var date time.Time
		var dailyChange float64
		var income float64
		var expense float64
		err := rows.Scan(&date, &dailyChange, &expense, &income)
		if err != nil {
			return initialBalance, 0, 0, nil, err
		}
		totalIncome += income
		totalExpenses += expense

		timeSeries = append(timeSeries, map[string]any{
			"date":         date.Format("2006-01-02"),
			"daily_change": dailyChange,
		})
	}

	return initialBalance, totalIncome, totalExpenses, timeSeries, nil
}

// GetCategoryAnalytics retrieves the category analytics for a given user and date range
func (r *AnalyticsRepository) GetCategoryAnalytics(ctx context.Context, userId int64, startDate time.Time, endDate time.Time, categoryIds []int64) (*models.CategoryAnalyticsResponse, error) {
	filterClause := ""
	args := []any{userId, startDate, endDate}

	if len(categoryIds) > 0 {
		includeUncategorized := false
		filteredIds := make([]int64, 0, len(categoryIds))
		for _, categoryID := range categoryIds {
			if categoryID == -1 {
				includeUncategorized = true
				continue
			}
			filteredIds = append(filteredIds, categoryID)
		}

		var conditions []string
		if len(filteredIds) > 0 {
			placeholders := make([]string, 0, len(filteredIds))
			for _, categoryID := range filteredIds {
				args = append(args, categoryID)
				placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)))
			}
			conditions = append(conditions, fmt.Sprintf("tcm.category_id IN (%s)", strings.Join(placeholders, ",")))
		}
		if includeUncategorized {
			conditions = append(conditions, "tcm.category_id IS NULL")
		}
		if len(conditions) > 0 {
			filterClause = fmt.Sprintf("AND (%s)", strings.Join(conditions, " OR "))
		}
	}

	query := fmt.Sprintf(`
        WITH user_transactions AS (
            SELECT
                t.amount,
                tcm.category_id
            FROM
                %s.transaction t
            LEFT JOIN
                %s.transaction_category_mapping tcm ON t.id = tcm.transaction_id
            WHERE
                t.created_by = $1
                AND t.deleted_at IS NULL
                AND t.date >= $2
                AND t.date <= $3
                %s
        )
        SELECT
            COALESCE(c.id, -1) AS category_id,
            COALESCE(c.name, 'Uncategorized') AS category_name,
            COALESCE(SUM(ut.amount), 0) AS total_amount
        FROM
            user_transactions ut
        LEFT JOIN
            %s.categories c ON ut.category_id = c.id AND c.created_by = $1
        GROUP BY
            c.id, c.name
        HAVING
            SUM(ut.amount) != 0;
    `, r.schema, r.schema, filterClause, r.schema)

	rows, err := r.db.FetchAll(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var analytics models.CategoryAnalyticsResponse
	analytics.CategoryTransactions = []models.CategoryTransaction{}

	for rows.Next() {
		var categoryTxn models.CategoryTransaction
		err := rows.Scan(
			&categoryTxn.CategoryID,
			&categoryTxn.CategoryName,
			&categoryTxn.TotalAmount,
		)
		if err != nil {
			return nil, err
		}
		analytics.CategoryTransactions = append(analytics.CategoryTransactions, categoryTxn)
	}

	return &analytics, nil
}

// GetMonthlyAnalytics retrieves income, expenses, and total amount for a specified date range
// Note: In our data model, expenses are stored as positive amounts and income as negative amounts
func (r *AnalyticsRepository) GetMonthlyAnalytics(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (*models.MonthlyAnalyticsResponse, error) {
	query := fmt.Sprintf(`
		SELECT 
			COALESCE(SUM(CASE WHEN amount > 0 THEN amount ELSE 0 END), 0) as total_expenses,
			COALESCE(SUM(CASE WHEN amount < 0 THEN amount * -1 ELSE 0 END), 0) as total_income
		FROM %s.%s
		WHERE created_by = $1 
			AND deleted_at IS NULL
			AND date >= $2 
			AND date <= $3`,
		r.schema, r.txnTableName)

	var totalExpenses, totalIncome float64
	row := r.db.FetchOne(ctx, query, userId, startDate, endDate)
	err := row.Scan(&totalExpenses, &totalIncome)
	if err != nil {
		return nil, err
	}
	totalAmount := totalIncome - totalExpenses

	return &models.MonthlyAnalyticsResponse{
		TotalIncome:   totalIncome,
		TotalExpenses: totalExpenses,
		TotalAmount:   totalAmount,
	}, nil
}

func (r *AnalyticsRepository) GetAccountCashFlows(ctx context.Context, userId int64, accountIds []int64) ([]models.AccountCashFlow, error) {
	if len(accountIds) == 0 {
		return []models.AccountCashFlow{}, nil
	}

	args := []any{userId}
	placeholders := make([]string, len(accountIds))
	for i, accountId := range accountIds {
		args = append(args, accountId)
		placeholders[i] = fmt.Sprintf("$%d", i+2)
	}

	query := fmt.Sprintf(`
		SELECT account_id, amount, date
		FROM %s.%s
		WHERE created_by = $1
			AND deleted_at IS NULL
			AND account_id IN (%s)
		ORDER BY account_id, date`,
		r.schema, r.txnTableName, strings.Join(placeholders, ", "))

	rows, err := r.db.FetchAll(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	flows := make([]models.AccountCashFlow, 0)
	for rows.Next() {
		var flow models.AccountCashFlow
		if err := rows.Scan(&flow.AccountID, &flow.Amount, &flow.Date); err != nil {
			return nil, err
		}
		flows = append(flows, flow)
	}

	return flows, nil
}
