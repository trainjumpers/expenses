package repository

import (
	"context"
	"expenses/internal/config"
	database "expenses/pkg/database/manager"
	"fmt"
	"time"
)

type AnalyticsRepositoryInterface interface {
	GetBalance(ctx context.Context, userId int64, startDate *time.Time, endDate *time.Time) (map[int64]float64, error)
	GetNetworthTimeSeries(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (float64, []map[string]any, error)
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
func (r *AnalyticsRepository) GetNetworthTimeSeries(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (float64, []map[string]any, error) {
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
		return 0, nil, err
	}

	// Get daily transaction sums within the date range
	timeSeriesQuery := fmt.Sprintf(`
		SELECT 
			date,
			COALESCE(SUM(amount), 0) * -1 as daily_change
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
		return initialBalance, nil, err
	}
	defer rows.Close()

	var timeSeries []map[string]any
	for rows.Next() {
		var date time.Time
		var dailyChange float64
		err := rows.Scan(&date, &dailyChange)
		if err != nil {
			return initialBalance, nil, err
		}

		timeSeries = append(timeSeries, map[string]any{
			"date":         date.Format("2006-01-02"),
			"daily_change": dailyChange,
		})
	}

	return initialBalance, timeSeries, nil
}
