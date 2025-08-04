package repository

import (
	"expenses/internal/config"
	database "expenses/internal/database/manager"
	"expenses/internal/models"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AnalyticsRepositoryInterface interface {
	GetSpendingOverview(ctx *gin.Context, query models.AnalyticsQuery) (*models.SpendingOverviewResponse, error)
	GetCategorySpending(ctx *gin.Context, query models.AnalyticsQuery) (*models.CategorySpendingResponse, error)
	GetSpendingTrends(ctx *gin.Context, query models.AnalyticsQuery, granularity string) (*models.SpendingTrendsResponse, error)
	GetAccountSpending(ctx *gin.Context, query models.AnalyticsQuery) (*models.AccountSpendingResponse, error)
	GetTopTransactions(ctx *gin.Context, query models.AnalyticsQuery, limit int) (*models.TopTransactionsResponse, error)
	GetMonthlyComparison(ctx *gin.Context, query models.AnalyticsQuery) (*models.MonthlyComparisonResponse, error)
	GetRecurringTransactions(ctx *gin.Context, query models.AnalyticsQuery) (*models.RecurringTransactionsResponse, error)
}

type AnalyticsRepository struct {
	db        database.DatabaseManager
	cfg       *config.Config
	schema    string
	tableName string
}

func NewAnalyticsRepository(db database.DatabaseManager, cfg *config.Config) AnalyticsRepositoryInterface {
	return &AnalyticsRepository{
		db:        db,
		cfg:       cfg,
		schema:    cfg.DBSchema,
		tableName: "transaction",
	}
}

// buildDateFilter builds the date filter condition and returns the WHERE clause and args
func (r *AnalyticsRepository) buildDateFilter(query models.AnalyticsQuery) (string, []interface{}, error) {
	var whereClause strings.Builder
	var args []interface{}
	argIndex := 1

	// Add created_by filter
	whereClause.WriteString(fmt.Sprintf("t.created_by = $%d", argIndex))
	args = append(args, query.CreatedBy)
	argIndex++

	// Add date filter based on time range
	var startDate, endDate time.Time
	now := time.Now()

	switch query.TimeRange {
	case models.TimeRangeWeek:
		startDate = now.AddDate(0, 0, -7)
		endDate = now
	case models.TimeRangeMonth:
		startDate = now.AddDate(0, -1, 0)
		endDate = now
	case models.TimeRangeQuarter:
		startDate = now.AddDate(0, -3, 0)
		endDate = now
	case models.TimeRangeYear:
		startDate = now.AddDate(-1, 0, 0)
		endDate = now
	case models.TimeRangeCustom:
		if query.StartDate == nil || query.EndDate == nil {
			return "", nil, fmt.Errorf("start_date and end_date are required for custom time range")
		}
		startDate = *query.StartDate
		endDate = *query.EndDate
	case models.TimeRangeAllTime:
		// No date filter for all time
	default:
		return "", nil, fmt.Errorf("invalid time range: %s", query.TimeRange)
	}

	if query.TimeRange != models.TimeRangeAllTime {
		whereClause.WriteString(fmt.Sprintf(" AND t.date >= $%d AND t.date <= $%d", argIndex, argIndex+1))
		args = append(args, startDate, endDate)
		argIndex += 2
	}

	// Add account filter
	if len(query.AccountIds) > 0 {
		placeholders := make([]string, len(query.AccountIds))
		for i, accountId := range query.AccountIds {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, accountId)
			argIndex++
		}
		whereClause.WriteString(fmt.Sprintf(" AND t.account_id IN (%s)", strings.Join(placeholders, ",")))
	}

	// Add category filter
	if len(query.CategoryIds) > 0 {
		placeholders := make([]string, len(query.CategoryIds))
		for i, categoryId := range query.CategoryIds {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, categoryId)
			argIndex++
		}
		whereClause.WriteString(fmt.Sprintf(" AND EXISTS (SELECT 1 FROM %s.transaction_category_mapping tc WHERE tc.transaction_id = t.id AND tc.category_id IN (%s))", r.schema, strings.Join(placeholders, ",")))
	}

	return whereClause.String(), args, nil
}

func (r *AnalyticsRepository) GetSpendingOverview(ctx *gin.Context, query models.AnalyticsQuery) (*models.SpendingOverviewResponse, error) {
	whereClause, args, err := r.buildDateFilter(query)
	if err != nil {
		return nil, err
	}

	sqlQuery := fmt.Sprintf(`
		SELECT 
			COALESCE(SUM(CASE WHEN amount < 0 THEN ABS(amount) ELSE 0 END), 0) as total_expenses,
			COALESCE(SUM(CASE WHEN amount > 0 THEN amount ELSE 0 END), 0) as total_income,
			COALESCE(SUM(amount), 0) as net_amount,
			COUNT(*) as transaction_count,
			COALESCE(AVG(CASE WHEN amount < 0 THEN ABS(amount) ELSE NULL END), 0) as average_expense,
			COALESCE(AVG(CASE WHEN amount > 0 THEN amount ELSE NULL END), 0) as average_income
		FROM %s.%s t
		WHERE %s
	`, r.schema, r.tableName, whereClause)

	var overview models.SpendingOverviewResponse
	row := r.db.FetchOne(ctx.Request.Context(), sqlQuery, args...)
	err = row.Scan(
		&overview.TotalExpenses,
		&overview.TotalIncome,
		&overview.NetAmount,
		&overview.TransactionCount,
		&overview.AverageExpense,
		&overview.AverageIncome,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get spending overview: %w", err)
	}

	overview.Period = string(query.TimeRange)
	return &overview, nil
}

func (r *AnalyticsRepository) GetCategorySpending(ctx *gin.Context, query models.AnalyticsQuery) (*models.CategorySpendingResponse, error) {
	whereClause, args, err := r.buildDateFilter(query)
	if err != nil {
		return nil, err
	}

	// Get categorized transactions
	categorizedQuery := fmt.Sprintf(`
		SELECT 
			c.id,
			c.name,
			c.icon,
			COALESCE(SUM(CASE WHEN t.amount < 0 THEN ABS(t.amount) ELSE 0 END), 0) as amount,
			COUNT(t.id) as count
		FROM %s.categories c
		INNER JOIN %s.transaction_category_mapping tc ON c.id = tc.category_id
		INNER JOIN %s.%s t ON tc.transaction_id = t.id
		WHERE %s AND t.amount < 0
		GROUP BY c.id, c.name, c.icon
		ORDER BY amount DESC
	`, r.schema, r.schema, r.schema, r.tableName, whereClause)

	rows, err := r.db.FetchAll(ctx.Request.Context(), categorizedQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get category spending: %w", err)
	}
	defer rows.Close()

	// Initialize empty slice to avoid null in JSON
	categories := make([]models.CategorySpendingItem, 0)
	var totalCategorizedAmount float64
	var totalCategorizedCount int

	for rows.Next() {
		var item models.CategorySpendingItem
		err := rows.Scan(&item.CategoryId, &item.CategoryName, &item.CategoryIcon, &item.Amount, &item.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category spending: %w", err)
		}
		categories = append(categories, item)
		totalCategorizedAmount += item.Amount
		totalCategorizedCount += item.Count
	}

	// Get uncategorized transactions
	uncategorizedQuery := fmt.Sprintf(`
		SELECT 
			COALESCE(SUM(CASE WHEN amount < 0 THEN ABS(amount) ELSE 0 END), 0) as amount,
			COUNT(*) as count
		FROM %s.%s t
		WHERE %s AND t.amount < 0 AND NOT EXISTS (
			SELECT 1 FROM %s.transaction_category_mapping tc WHERE tc.transaction_id = t.id
		)
	`, r.schema, r.tableName, whereClause, r.schema)

	var uncategorized models.CategorySpendingItem
	uncategorized.CategoryName = "Uncategorized"
	row := r.db.FetchOne(ctx.Request.Context(), uncategorizedQuery, args...)
	err = row.Scan(&uncategorized.Amount, &uncategorized.Count)
	if err != nil {
		return nil, fmt.Errorf("failed to get uncategorized spending: %w", err)
	}

	totalAmount := totalCategorizedAmount + uncategorized.Amount
	totalCount := totalCategorizedCount + uncategorized.Count

	// Calculate percentages
	for i := range categories {
		if totalAmount > 0 {
			categories[i].Percentage = (categories[i].Amount / totalAmount) * 100
		}
	}
	if totalAmount > 0 {
		uncategorized.Percentage = (uncategorized.Amount / totalAmount) * 100
	}

	return &models.CategorySpendingResponse{
		Categories:    categories,
		Uncategorized: uncategorized,
		TotalAmount:   totalAmount,
		TotalCount:    totalCount,
	}, nil
}

func (r *AnalyticsRepository) GetSpendingTrends(ctx *gin.Context, query models.AnalyticsQuery, granularity string) (*models.SpendingTrendsResponse, error) {
	whereClause, args, err := r.buildDateFilter(query)
	if err != nil {
		return nil, err
	}

	var dateTrunc string
	switch granularity {
	case "daily":
		dateTrunc = "day"
	case "weekly":
		dateTrunc = "week"
	case "monthly":
		dateTrunc = "month"
	default:
		dateTrunc = "day"
	}

	trendsQuery := fmt.Sprintf(`
		SELECT 
			DATE_TRUNC('%s', t.date) as period_date,
			COALESCE(SUM(CASE WHEN amount < 0 THEN ABS(amount) ELSE 0 END), 0) as expenses,
			COALESCE(SUM(CASE WHEN amount > 0 THEN amount ELSE 0 END), 0) as income,
			COALESCE(SUM(amount), 0) as net_amount,
			COUNT(*) as count
		FROM %s.%s t
		WHERE %s
		GROUP BY DATE_TRUNC('%s', t.date)
		ORDER BY period_date
	`, dateTrunc, r.schema, r.tableName, whereClause, dateTrunc)

	rows, err := r.db.FetchAll(ctx.Request.Context(), trendsQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get spending trends: %w", err)
	}
	defer rows.Close()

	// Initialize empty slice to avoid null in JSON
	dataPoints := make([]models.TimeSeriesDataPoint, 0)
	for rows.Next() {
		var point models.TimeSeriesDataPoint
		err := rows.Scan(&point.Date, &point.Expenses, &point.Income, &point.Amount, &point.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan spending trends: %w", err)
		}
		dataPoints = append(dataPoints, point)
	}

	return &models.SpendingTrendsResponse{
		DataPoints:  dataPoints,
		Period:      string(query.TimeRange),
		Granularity: granularity,
	}, nil
}

func (r *AnalyticsRepository) GetAccountSpending(ctx *gin.Context, query models.AnalyticsQuery) (*models.AccountSpendingResponse, error) {
	whereClause, args, err := r.buildDateFilter(query)
	if err != nil {
		return nil, err
	}

	accountQuery := fmt.Sprintf(`
		SELECT 
			a.id,
			a.name,
			a.bank_type,
			COALESCE(SUM(CASE WHEN t.amount < 0 THEN ABS(t.amount) ELSE 0 END), 0) as amount,
			COUNT(t.id) as count
		FROM %s.account a
		INNER JOIN %s.%s t ON a.id = t.account_id
		WHERE %s AND t.amount < 0
		GROUP BY a.id, a.name, a.bank_type
		ORDER BY amount DESC
	`, r.schema, r.schema, r.tableName, whereClause)

	rows, err := r.db.FetchAll(ctx.Request.Context(), accountQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get account spending: %w", err)
	}
	defer rows.Close()

	// Initialize empty slice to avoid null in JSON
	accounts := make([]models.AccountSpendingItem, 0)
	var totalAmount float64
	var totalCount int

	for rows.Next() {
		var item models.AccountSpendingItem
		err := rows.Scan(&item.AccountId, &item.AccountName, &item.BankType, &item.Amount, &item.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account spending: %w", err)
		}
		accounts = append(accounts, item)
		totalAmount += item.Amount
		totalCount += item.Count
	}

	// Calculate percentages
	for i := range accounts {
		if totalAmount > 0 {
			accounts[i].Percentage = (accounts[i].Amount / totalAmount) * 100
		}
	}

	return &models.AccountSpendingResponse{
		Accounts:    accounts,
		TotalAmount: totalAmount,
		TotalCount:  totalCount,
	}, nil
}

func (r *AnalyticsRepository) GetTopTransactions(ctx *gin.Context, query models.AnalyticsQuery, limit int) (*models.TopTransactionsResponse, error) {
	whereClause, args, err := r.buildDateFilter(query)
	if err != nil {
		return nil, err
	}

	// Get top expenses
	expensesQuery := fmt.Sprintf(`
		SELECT 
			t.id,
			t.name,
			ABS(t.amount) as amount,
			t.date,
			a.name as account_name,
			COALESCE(
				ARRAY_AGG(c.name ORDER BY c.name) FILTER (WHERE c.name IS NOT NULL),
				ARRAY[]::text[]
			) as categories
		FROM %s.%s t
		INNER JOIN %s.account a ON t.account_id = a.id
		LEFT JOIN %s.transaction_category_mapping tc ON t.id = tc.transaction_id
		LEFT JOIN %s.categories c ON tc.category_id = c.id
		WHERE %s AND t.amount < 0
		GROUP BY t.id, t.name, t.amount, t.date, a.name
		ORDER BY ABS(t.amount) DESC
		LIMIT $%d
	`, r.schema, r.tableName, r.schema, r.schema, r.schema, whereClause, len(args)+1)

	args = append(args, limit)
	expenseRows, err := r.db.FetchAll(ctx.Request.Context(), expensesQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get top expenses: %w", err)
	}
	defer expenseRows.Close()

	// Initialize empty slices to avoid null in JSON
	topExpenses := make([]models.TopTransactionItem, 0)
	for expenseRows.Next() {
		var item models.TopTransactionItem
		err := expenseRows.Scan(&item.Id, &item.Name, &item.Amount, &item.Date, &item.AccountName, &item.Categories)
		if err != nil {
			return nil, fmt.Errorf("failed to scan top expense: %w", err)
		}
		topExpenses = append(topExpenses, item)
	}

	// Get top income
	incomeQuery := fmt.Sprintf(`
		SELECT 
			t.id,
			t.name,
			t.amount,
			t.date,
			a.name as account_name,
			COALESCE(
				ARRAY_AGG(c.name ORDER BY c.name) FILTER (WHERE c.name IS NOT NULL),
				ARRAY[]::text[]
			) as categories
		FROM %s.%s t
		INNER JOIN %s.account a ON t.account_id = a.id
		LEFT JOIN %s.transaction_category_mapping tc ON t.id = tc.transaction_id
		LEFT JOIN %s.categories c ON tc.category_id = c.id
		WHERE %s AND t.amount > 0
		GROUP BY t.id, t.name, t.amount, t.date, a.name
		ORDER BY t.amount DESC
		LIMIT $%d
	`, r.schema, r.tableName, r.schema, r.schema, r.schema, whereClause, len(args))

	incomeRows, err := r.db.FetchAll(ctx.Request.Context(), incomeQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get top income: %w", err)
	}
	defer incomeRows.Close()

	topIncome := make([]models.TopTransactionItem, 0)
	for incomeRows.Next() {
		var item models.TopTransactionItem
		err := incomeRows.Scan(&item.Id, &item.Name, &item.Amount, &item.Date, &item.AccountName, &item.Categories)
		if err != nil {
			return nil, fmt.Errorf("failed to scan top income: %w", err)
		}
		topIncome = append(topIncome, item)
	}

	return &models.TopTransactionsResponse{
		TopExpenses: topExpenses,
		TopIncome:   topIncome,
		Limit:       limit,
	}, nil
}

func (r *AnalyticsRepository) GetMonthlyComparison(ctx *gin.Context, query models.AnalyticsQuery) (*models.MonthlyComparisonResponse, error) {
	whereClause, args, err := r.buildDateFilter(query)
	if err != nil {
		return nil, err
	}

	monthlyQuery := fmt.Sprintf(`
		SELECT 
			DATE_TRUNC('month', t.date) as month,
			COALESCE(SUM(CASE WHEN amount < 0 THEN ABS(amount) ELSE 0 END), 0) as amount,
			COUNT(*) as count
		FROM %s.%s t
		WHERE %s
		GROUP BY DATE_TRUNC('month', t.date)
		ORDER BY month
	`, r.schema, r.tableName, whereClause)

	rows, err := r.db.FetchAll(ctx.Request.Context(), monthlyQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly comparison: %w", err)
	}
	defer rows.Close()

	// Initialize empty slice to avoid null in JSON
	months := make([]models.MonthlyComparisonItem, 0)
	var previousAmount float64

	for rows.Next() {
		var item models.MonthlyComparisonItem
		err := rows.Scan(&item.Month, &item.Amount, &item.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan monthly comparison: %w", err)
		}

		item.MonthName = item.Month.Format("January 2006")
		
		// Calculate change from previous month
		if len(months) > 0 && previousAmount > 0 {
			item.ChangeAmount = item.Amount - previousAmount
			item.Change = (item.ChangeAmount / previousAmount) * 100
		}

		months = append(months, item)
		previousAmount = item.Amount
	}

	return &models.MonthlyComparisonResponse{
		Months:      months,
		Period:      string(query.TimeRange),
		TotalMonths: len(months),
	}, nil
}

func (r *AnalyticsRepository) GetRecurringTransactions(ctx *gin.Context, query models.AnalyticsQuery) (*models.RecurringTransactionsResponse, error) {
	whereClause, args, err := r.buildDateFilter(query)
	if err != nil {
		return nil, err
	}

	// Simple recurring pattern detection based on similar amounts and names
	recurringQuery := fmt.Sprintf(`
		WITH similar_transactions AS (
			SELECT 
				LOWER(TRIM(name)) as pattern,
				ROUND(ABS(amount)::numeric, 2) as rounded_amount,
				COUNT(*) as frequency,
				AVG(ABS(amount)) as avg_amount,
				ARRAY_AGG(id ORDER BY date) as transaction_ids,
				MIN(date) as first_date,
				MAX(date) as last_date
			FROM %s.%s t
			WHERE %s AND amount < 0
			GROUP BY LOWER(TRIM(name)), ROUND(ABS(amount)::numeric, 2)
			HAVING COUNT(*) >= 3
		)
		SELECT 
			pattern,
			avg_amount,
			frequency,
			transaction_ids,
			CASE 
				WHEN frequency >= 12 THEN 'monthly'
				WHEN frequency >= 4 THEN 'weekly'
				ELSE 'irregular'
			END as frequency_type,
			CASE 
				WHEN frequency >= 6 THEN 0.9
				WHEN frequency >= 4 THEN 0.7
				ELSE 0.5
			END as confidence
		FROM similar_transactions
		ORDER BY avg_amount DESC, frequency DESC
		LIMIT 20
	`, r.schema, r.tableName, whereClause)

	rows, err := r.db.FetchAll(ctx.Request.Context(), recurringQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get recurring transactions: %w", err)
	}
	defer rows.Close()

	// Initialize empty slice to avoid null in JSON
	patterns := make([]models.RecurringTransactionPattern, 0)
	var totalAmount float64
	var totalCount int

	for rows.Next() {
		var pattern models.RecurringTransactionPattern
		var frequencyType string
		err := rows.Scan(
			&pattern.Pattern,
			&pattern.Amount,
			&pattern.Count,
			&pattern.TransactionIds,
			&frequencyType,
			&pattern.Confidence,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recurring pattern: %w", err)
		}

		pattern.Frequency = frequencyType
		patterns = append(patterns, pattern)
		totalAmount += pattern.Amount * float64(pattern.Count)
		totalCount += pattern.Count
	}

	return &models.RecurringTransactionsResponse{
		Patterns:    patterns,
		TotalAmount: totalAmount,
		Count:       totalCount,
	}, nil
}
