package repository

import (
	"context"
	"database/sql"
	"expenses/internal/config"
	"expenses/internal/models"
	database "expenses/pkg/database/manager"
	"fmt"
	"strings"
	"time"
)

type AnalyticsRepositoryInterface interface {
	GetBalance(ctx context.Context, userId int64, startDate *time.Time, endDate *time.Time) (map[int64]float64, error)
	GetCashBalanceHistory(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (float64, float64, float64, []map[string]any, error)
	GetCategoryAnalytics(ctx context.Context, userId int64, startDate time.Time, endDate time.Time, categoryIds []int64) (*models.CategoryAnalyticsResponse, error)
	GetMonthlyAnalytics(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (*models.MonthlyAnalyticsResponse, error)
	GetAccountCashFlows(ctx context.Context, userId int64, accountIds []int64) ([]models.AccountCashFlow, error)

	GetInsightsMonthly(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) ([]models.InsightsMonthlyPoint, error)
	GetInsightsCategories(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) ([]models.InsightsCategory, error)
	GetInsightsTopExpenses(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) ([]models.InsightsTopExpense, error)
	GetInsightsUncategorized(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (int64, float64, error)
	GetInsightsSpendingSummary(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (models.InsightsSpendingSummary, error)
	GetInsightsCategoryMonths(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) ([]models.InsightsCategoryMonth, error)
	GetInsightsWeekday(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) ([]models.InsightsWeekday, error)
	GetInsightsLatestTransactionDate(ctx context.Context, userId int64) (*time.Time, error)
	GetInsightsMultiCategoryCount(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (int64, error)
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

// GetCashBalanceHistory calculates the initial balance and daily cash-balance
// changes. Investment account ledgers are excluded because historical
// investment valuation is unavailable.
// Returns initial balance (sum of all transactions before startDate) and daily aggregated data
func (r *AnalyticsRepository) GetCashBalanceHistory(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (float64, float64, float64, []map[string]any, error) {
	// First, get the initial balance (sum of all transactions before startDate)
	initialBalanceQuery := fmt.Sprintf(`
		SELECT COALESCE(SUM(t.amount), 0) * -1 as initial_balance
		FROM %s.%s t
		JOIN %s.account a ON a.id = t.account_id
		WHERE t.created_by = $1
			AND t.deleted_at IS NULL
			AND t.date < $2
			AND a.bank_type <> 'investment'`,
		r.schema, r.txnTableName, r.schema)

	var initialBalance float64
	row := r.db.FetchOne(ctx, initialBalanceQuery, userId, startDate)
	err := row.Scan(&initialBalance)
	if err != nil {
		return 0, 0, 0, nil, err
	}

	// Get daily transaction sums within the date range
	timeSeriesQuery := fmt.Sprintf(`
		SELECT
			t.date,
			COALESCE(SUM(t.amount), 0) * -1 as daily_change,
			COALESCE(SUM(CASE WHEN t.amount > 0 THEN t.amount ELSE 0 END), 0) as total_expenses,
			COALESCE(SUM(CASE WHEN t.amount < 0 THEN t.amount * -1 ELSE 0 END), 0) as total_income
		FROM %s.%s t
		JOIN %s.account a ON a.id = t.account_id
		WHERE t.created_by = $1
			AND t.deleted_at IS NULL
			AND t.date >= $2
			AND t.date <= $3
			AND a.bank_type <> 'investment'
		GROUP BY t.date
		ORDER BY t.date`,
		r.schema, r.txnTableName, r.schema)

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
// under the household scope. A transaction mapped to several categories has its
// amount shared equally between them so category totals stay additive.
func (r *AnalyticsRepository) GetCategoryAnalytics(ctx context.Context, userId int64, startDate time.Time, endDate time.Time, categoryIds []int64) (*models.CategoryAnalyticsResponse, error) {
	whereClause := ""
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
			conditions = append(conditions, fmt.Sprintf("s.category_id IN (%s)", strings.Join(placeholders, ",")))
		}
		if includeUncategorized {
			conditions = append(conditions, "s.category_id IS NULL")
		}
		if len(conditions) > 0 {
			whereClause = fmt.Sprintf("WHERE (%s)", strings.Join(conditions, " OR "))
		}
	}

	allocation := categoryAllocationExpression()
	query := fmt.Sprintf(`
        WITH %s
        SELECT
            COALESCE(c.id, -1) AS category_id,
            COALESCE(c.name, 'Uncategorized') AS category_name,
            COALESCE(SUM(%s), 0) AS total_amount
        FROM
            scoped s
        JOIN
            txn_counts tc ON tc.txn_id = s.txn_id
        LEFT JOIN
            %s.categories c ON s.category_id = c.id AND c.created_by = $1
        %s
        GROUP BY
            c.id, c.name
        HAVING
            SUM(%s) != 0;
    `, r.categoryAllocationCTE(), allocation, r.schema, whereClause, allocation)

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
// under the household scope: non-investment accounts and no transfer-category transactions.
// Note: In our data model, expenses are stored as positive amounts and income as negative amounts
func (r *AnalyticsRepository) GetMonthlyAnalytics(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (*models.MonthlyAnalyticsResponse, error) {
	query := fmt.Sprintf(`
		SELECT 
			COALESCE(SUM(CASE WHEN t.amount > 0 THEN t.amount ELSE 0 END), 0) as total_expenses,
			COALESCE(SUM(CASE WHEN t.amount < 0 THEN t.amount * -1 ELSE 0 END), 0) as total_income
		FROM %s
		WHERE t.created_by = $1 
			AND t.date >= $2 
			AND t.date <= $3
			AND %s`,
		r.householdFrom(), r.householdPredicate())

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
		SELECT account_id, amount, date, name
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
		if err := rows.Scan(&flow.AccountID, &flow.Amount, &flow.Date, &flow.Name); err != nil {
			return nil, err
		}
		flows = append(flows, flow)
	}

	return flows, nil
}

// transferCategoriesPredicate matches both the canonical "Transfers" name and
// the legacy singular "Transfer" spelling.
func (r *AnalyticsRepository) transferCategoriesPredicate() string {
	names := models.TransferCategoryNames()
	quoted := make([]string, 0, len(names))
	for _, name := range names {
		quoted = append(quoted, fmt.Sprintf("'%s'", name))
	}
	return fmt.Sprintf("LOWER(c.name) IN (%s)", strings.Join(quoted, ", "))
}

// householdPredicate is the single definition of a household transaction:
// non-deleted, on a non-investment account, and not mapped to a transfer
// category. It is shared by insights, monthly, and category analytics so the
// definition cannot drift between them.
func (r *AnalyticsRepository) householdPredicate() string {
	return fmt.Sprintf(`
			t.deleted_at IS NULL
			AND a.bank_type <> 'investment'
			AND NOT EXISTS (
				SELECT 1
				FROM %s.transaction_category_mapping tcm
				JOIN %s.categories c ON c.id = tcm.category_id
				WHERE tcm.transaction_id = t.id
					AND %s
			)`, r.schema, r.schema, r.transferCategoriesPredicate())
}

func (r *AnalyticsRepository) householdFrom() string {
	return fmt.Sprintf("%s.transaction t JOIN %s.account a ON a.id = t.account_id", r.schema, r.schema)
}

// categoryAllocationCTE expands every household transaction into one row per
// mapped category. The transaction amount is later shared equally across those
// categories, so a plain mapping join cannot double count a multi-category row.
// txn_counts holds the number of mapped categories per transaction.
func (r *AnalyticsRepository) categoryAllocationCTE() string {
	return fmt.Sprintf(`
		scoped AS (
			SELECT t.id AS txn_id, t.amount, t.date, tcm.category_id
			FROM %s
			LEFT JOIN %s.transaction_category_mapping tcm ON tcm.transaction_id = t.id
			WHERE t.created_by = $1
				AND t.date >= $2
				AND t.date <= $3
				AND %s
		),
		txn_counts AS (
			SELECT txn_id, COUNT(category_id) AS mapped_count
			FROM scoped
			GROUP BY txn_id
		)`,
		r.householdFrom(), r.schema, r.householdPredicate())
}

func categoryAllocationExpression() string {
	return `CASE WHEN s.category_id IS NULL THEN s.amount ELSE s.amount / NULLIF(tc.mapped_count, 0) END`
}

func (r *AnalyticsRepository) GetInsightsMonthly(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) ([]models.InsightsMonthlyPoint, error) {
	query := fmt.Sprintf(`
		SELECT
			to_char(date_trunc('month', t.date), 'YYYY-MM') AS month,
			COALESCE(SUM(CASE WHEN t.amount < 0 THEN t.amount * -1 ELSE 0 END), 0) AS income,
			COALESCE(SUM(CASE WHEN t.amount > 0 THEN t.amount ELSE 0 END), 0) AS expenses
		FROM %s
		WHERE t.created_by = $1
			AND t.date >= $2
			AND t.date <= $3
			AND %s
		GROUP BY month
		ORDER BY month`,
		r.householdFrom(), r.householdPredicate())

	rows, err := r.db.FetchAll(ctx, query, userId, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	points := make([]models.InsightsMonthlyPoint, 0)
	for rows.Next() {
		var point models.InsightsMonthlyPoint
		if err := rows.Scan(&point.Month, &point.Income, &point.Expenses); err != nil {
			return nil, err
		}
		point.Net = point.Income - point.Expenses
		points = append(points, point)
	}

	return points, nil
}

func (r *AnalyticsRepository) GetInsightsCategories(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) ([]models.InsightsCategory, error) {
	allocation := categoryAllocationExpression()
	query := fmt.Sprintf(`
		WITH %s
		SELECT
			COALESCE(c.id, -1) AS category_id,
			COALESCE(c.name, 'Uncategorized') AS category_name,
			COALESCE(SUM(%s), 0) AS total_amount
		FROM scoped s
		JOIN txn_counts tc ON tc.txn_id = s.txn_id
		LEFT JOIN %s.categories c ON c.id = s.category_id
		GROUP BY c.id, c.name
		HAVING SUM(%s) != 0
		ORDER BY ABS(SUM(%s)) DESC`,
		r.categoryAllocationCTE(), allocation, r.schema, allocation, allocation)

	rows, err := r.db.FetchAll(ctx, query, userId, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]models.InsightsCategory, 0)
	for rows.Next() {
		var category models.InsightsCategory
		if err := rows.Scan(&category.CategoryID, &category.CategoryName, &category.TotalAmount); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (r *AnalyticsRepository) GetInsightsTopExpenses(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) ([]models.InsightsTopExpense, error) {
	query := fmt.Sprintf(`
		SELECT t.name, COALESCE(SUM(t.amount), 0) AS amount, COUNT(*) AS count
		FROM %s
		WHERE t.created_by = $1
			AND t.date >= $2
			AND t.date <= $3
			AND t.amount > 0
			AND %s
		GROUP BY t.name
		ORDER BY amount DESC
		LIMIT 15`,
		r.householdFrom(), r.householdPredicate())

	rows, err := r.db.FetchAll(ctx, query, userId, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	expenses := make([]models.InsightsTopExpense, 0)
	for rows.Next() {
		var expense models.InsightsTopExpense
		if err := rows.Scan(&expense.Name, &expense.Amount, &expense.Count); err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}

	return expenses, nil
}

func (r *AnalyticsRepository) GetInsightsUncategorized(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (int64, float64, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*), COALESCE(SUM(t.amount), 0)
		FROM %s
		WHERE t.created_by = $1
			AND t.date >= $2
			AND t.date <= $3
			AND t.amount > 0
			AND NOT EXISTS (
				SELECT 1
				FROM %s.transaction_category_mapping tcm
				WHERE tcm.transaction_id = t.id
			)
			AND %s`,
		r.householdFrom(), r.schema, r.householdPredicate())

	var count int64
	var amount float64
	row := r.db.FetchOne(ctx, query, userId, startDate, endDate)
	if err := row.Scan(&count, &amount); err != nil {
		return 0, 0, err
	}

	return count, amount, nil
}

func (r *AnalyticsRepository) GetInsightsSpendingSummary(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (models.InsightsSpendingSummary, error) {
	query := fmt.Sprintf(`
		SELECT
			COUNT(*),
			COALESCE(AVG(t.amount), 0),
			COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY t.amount), 0),
			COALESCE(MAX(t.amount), 0),
			COUNT(DISTINCT t.date)
		FROM %s
		WHERE t.created_by = $1
			AND t.date >= $2
			AND t.date <= $3
			AND t.amount > 0
			AND %s`,
		r.householdFrom(), r.householdPredicate())

	var summary models.InsightsSpendingSummary
	row := r.db.FetchOne(ctx, query, userId, startDate, endDate)
	if err := row.Scan(
		&summary.ExpenseCount,
		&summary.AverageTransaction,
		&summary.MedianTransaction,
		&summary.LargestExpense,
		&summary.ActiveSpendingDays,
	); err != nil {
		return models.InsightsSpendingSummary{}, err
	}

	return summary, nil
}

// GetInsightsCategoryMonths returns expense-only category totals per month,
// with multi-category transactions allocated across their categories.
func (r *AnalyticsRepository) GetInsightsCategoryMonths(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) ([]models.InsightsCategoryMonth, error) {
	allocation := categoryAllocationExpression()
	query := fmt.Sprintf(`
		WITH %s
		SELECT
			to_char(date_trunc('month', s.date), 'YYYY-MM') AS month,
			COALESCE(c.id, -1) AS category_id,
			COALESCE(c.name, 'Uncategorized') AS category_name,
			COALESCE(SUM(%s), 0) AS total
		FROM scoped s
		JOIN txn_counts tc ON tc.txn_id = s.txn_id
		LEFT JOIN %s.categories c ON c.id = s.category_id
		WHERE s.amount > 0
		GROUP BY date_trunc('month', s.date), c.id, c.name
		HAVING SUM(%s) != 0
		ORDER BY month, total DESC`,
		r.categoryAllocationCTE(), allocation, r.schema, allocation)

	rows, err := r.db.FetchAll(ctx, query, userId, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	months := make([]models.InsightsCategoryMonth, 0)
	for rows.Next() {
		var month models.InsightsCategoryMonth
		if err := rows.Scan(
			&month.Month,
			&month.CategoryID,
			&month.CategoryName,
			&month.Total,
		); err != nil {
			return nil, err
		}
		months = append(months, month)
	}

	return months, nil
}

func (r *AnalyticsRepository) GetInsightsWeekday(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) ([]models.InsightsWeekday, error) {
	query := fmt.Sprintf(`
		SELECT
			EXTRACT(DOW FROM t.date)::int AS weekday,
			COALESCE(SUM(t.amount), 0),
			COUNT(*),
			COUNT(DISTINCT t.date)
		FROM %s
		WHERE t.created_by = $1
			AND t.date >= $2
			AND t.date <= $3
			AND t.amount > 0
			AND %s
		GROUP BY 1
		ORDER BY 1`,
		r.householdFrom(), r.householdPredicate())

	rows, err := r.db.FetchAll(ctx, query, userId, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	days := make([]models.InsightsWeekday, 0)
	for rows.Next() {
		var day models.InsightsWeekday
		if err := rows.Scan(&day.Weekday, &day.Total, &day.Count, &day.ActiveDays); err != nil {
			return nil, err
		}
		days = append(days, day)
	}

	return days, nil
}

// GetInsightsLatestTransactionDate returns the newest transaction date across
// all of the user's accounts, including investment ledgers, so staleness
// reflects the whole dataset.
func (r *AnalyticsRepository) GetInsightsLatestTransactionDate(ctx context.Context, userId int64) (*time.Time, error) {
	query := fmt.Sprintf(`
		SELECT MAX(date)
		FROM %s.%s
		WHERE created_by = $1
			AND deleted_at IS NULL`,
		r.schema, r.txnTableName)

	var latest sql.NullTime
	row := r.db.FetchOne(ctx, query, userId)
	if err := row.Scan(&latest); err != nil {
		return nil, err
	}
	if !latest.Valid {
		return nil, nil
	}

	return &latest.Time, nil
}

func (r *AnalyticsRepository) GetInsightsMultiCategoryCount(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) (int64, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
		WHERE t.created_by = $1
			AND t.date >= $2
			AND t.date <= $3
			AND t.amount > 0
			AND %s
			AND (
				SELECT COUNT(*)
				FROM %s.transaction_category_mapping tcm
				WHERE tcm.transaction_id = t.id
			) > 1`,
		r.householdFrom(), r.householdPredicate(), r.schema)

	var count int64
	row := r.db.FetchOne(ctx, query, userId, startDate, endDate)
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}
