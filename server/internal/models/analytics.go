package models

import (
	"time"
)

// AnalyticsTimeRange represents different time ranges for analytics
type AnalyticsTimeRange string

const (
	TimeRangeWeek     AnalyticsTimeRange = "week"
	TimeRangeMonth    AnalyticsTimeRange = "month"
	TimeRangeQuarter  AnalyticsTimeRange = "quarter"
	TimeRangeYear     AnalyticsTimeRange = "year"
	TimeRangeCustom   AnalyticsTimeRange = "custom"
	TimeRangeAllTime  AnalyticsTimeRange = "all_time"
)

// AnalyticsQuery holds query parameters for analytics requests
type AnalyticsQuery struct {
	TimeRange   AnalyticsTimeRange `json:"time_range" binding:"required"`
	StartDate   *time.Time         `json:"start_date"`
	EndDate     *time.Time         `json:"end_date"`
	AccountIds  []int64            `json:"account_ids"`
	CategoryIds []int64            `json:"category_ids"`
	CreatedBy   int64              `json:"created_by" binding:"required"`
}

// SpendingOverviewResponse provides high-level spending metrics
type SpendingOverviewResponse struct {
	TotalExpenses    float64 `json:"total_expenses"`
	TotalIncome      float64 `json:"total_income"`
	NetAmount        float64 `json:"net_amount"`
	TransactionCount int     `json:"transaction_count"`
	AverageExpense   float64 `json:"average_expense"`
	AverageIncome    float64 `json:"average_income"`
	Period           string  `json:"period"`
}

// CategorySpendingItem represents spending data for a single category
type CategorySpendingItem struct {
	CategoryId   int64   `json:"category_id"`
	CategoryName string  `json:"category_name"`
	CategoryIcon *string `json:"category_icon"`
	Amount       float64 `json:"amount"`
	Percentage   float64 `json:"percentage"`
	Count        int     `json:"count"`
}

// CategorySpendingResponse provides category-wise spending breakdown
type CategorySpendingResponse struct {
	Categories     []CategorySpendingItem `json:"categories"`
	Uncategorized  CategorySpendingItem   `json:"uncategorized"`
	TotalAmount    float64                `json:"total_amount"`
	TotalCount     int                    `json:"total_count"`
}

// TimeSeriesDataPoint represents a single point in time series data
type TimeSeriesDataPoint struct {
	Date     time.Time `json:"date"`
	Amount   float64   `json:"amount"`
	Count    int       `json:"count"`
	Income   float64   `json:"income"`
	Expenses float64   `json:"expenses"`
}

// SpendingTrendsResponse provides time-based spending trends
type SpendingTrendsResponse struct {
	DataPoints []TimeSeriesDataPoint `json:"data_points"`
	Period     string                `json:"period"`
	Granularity string               `json:"granularity"` // daily, weekly, monthly
}

// AccountSpendingItem represents spending data for a single account
type AccountSpendingItem struct {
	AccountId   int64   `json:"account_id"`
	AccountName string  `json:"account_name"`
	BankType    *string `json:"bank_type"`
	Amount      float64 `json:"amount"`
	Percentage  float64 `json:"percentage"`
	Count       int     `json:"count"`
}

// AccountSpendingResponse provides account-wise spending breakdown
type AccountSpendingResponse struct {
	Accounts    []AccountSpendingItem `json:"accounts"`
	TotalAmount float64               `json:"total_amount"`
	TotalCount  int                   `json:"total_count"`
}

// TopTransactionItem represents a high-value transaction
type TopTransactionItem struct {
	Id          int64     `json:"id"`
	Name        string    `json:"name"`
	Amount      float64   `json:"amount"`
	Date        time.Time `json:"date"`
	AccountName string    `json:"account_name"`
	Categories  []string  `json:"categories"`
}

// TopTransactionsResponse provides highest spending transactions
type TopTransactionsResponse struct {
	TopExpenses []TopTransactionItem `json:"top_expenses"`
	TopIncome   []TopTransactionItem `json:"top_income"`
	Limit       int                  `json:"limit"`
}

// MonthlyComparisonItem represents spending comparison for a month
type MonthlyComparisonItem struct {
	Month       time.Time `json:"month"`
	MonthName   string    `json:"month_name"`
	Amount      float64   `json:"amount"`
	Count       int       `json:"count"`
	Change      float64   `json:"change"`       // percentage change from previous month
	ChangeAmount float64  `json:"change_amount"` // absolute change from previous month
}

// MonthlyComparisonResponse provides month-over-month spending comparison
type MonthlyComparisonResponse struct {
	Months      []MonthlyComparisonItem `json:"months"`
	Period      string                  `json:"period"`
	TotalMonths int                     `json:"total_months"`
}

// BudgetAnalysisItem represents budget vs actual spending for a category
type BudgetAnalysisItem struct {
	CategoryId     int64   `json:"category_id"`
	CategoryName   string  `json:"category_name"`
	BudgetAmount   float64 `json:"budget_amount"`
	ActualAmount   float64 `json:"actual_amount"`
	Difference     float64 `json:"difference"`
	PercentageUsed float64 `json:"percentage_used"`
	Status         string  `json:"status"` // "under", "over", "on_track"
}

// BudgetAnalysisResponse provides budget vs actual analysis
type BudgetAnalysisResponse struct {
	Categories      []BudgetAnalysisItem `json:"categories"`
	TotalBudget     float64              `json:"total_budget"`
	TotalActual     float64              `json:"total_actual"`
	TotalDifference float64              `json:"total_difference"`
	OverallStatus   string               `json:"overall_status"`
}

// RecurringTransactionPattern represents a detected recurring transaction
type RecurringTransactionPattern struct {
	Pattern         string    `json:"pattern"`
	Amount          float64   `json:"amount"`
	Frequency       string    `json:"frequency"` // weekly, monthly, quarterly
	NextExpectedDate *time.Time `json:"next_expected_date"`
	Confidence      float64   `json:"confidence"`
	TransactionIds  []int64   `json:"transaction_ids"`
	Count           int       `json:"count"`
}

// RecurringTransactionsResponse provides detected recurring transaction patterns
type RecurringTransactionsResponse struct {
	Patterns    []RecurringTransactionPattern `json:"patterns"`
	TotalAmount float64                       `json:"total_amount"`
	Count       int                           `json:"count"`
}

// AnalyticsSummaryResponse provides a comprehensive analytics summary
type AnalyticsSummaryResponse struct {
	Overview            SpendingOverviewResponse      `json:"overview"`
	CategoryBreakdown   CategorySpendingResponse      `json:"category_breakdown"`
	AccountBreakdown    AccountSpendingResponse       `json:"account_breakdown"`
	TopTransactions     TopTransactionsResponse       `json:"top_transactions"`
	MonthlyComparison   MonthlyComparisonResponse     `json:"monthly_comparison"`
	RecurringPatterns   RecurringTransactionsResponse `json:"recurring_patterns"`
	Period              string                        `json:"period"`
	GeneratedAt         time.Time                     `json:"generated_at"`
}

// AnalyticsInsight represents an AI-generated insight
type AnalyticsInsight struct {
	Type        string    `json:"type"`        // "warning", "info", "success", "tip"
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Actionable  bool      `json:"actionable"`
	Priority    int       `json:"priority"`   // 1-5, 5 being highest
	CreatedAt   time.Time `json:"created_at"`
}

// AnalyticsInsightsResponse provides AI-generated insights
type AnalyticsInsightsResponse struct {
	Insights    []AnalyticsInsight `json:"insights"`
	Count       int                `json:"count"`
	GeneratedAt time.Time          `json:"generated_at"`
}
