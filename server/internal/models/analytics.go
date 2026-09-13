package models

import "time"

// AccountBalanceAnalytics represents the analytics data for a single account
type AccountBalanceAnalytics struct {
	AccountID          int64    `json:"account_id"`
	CurrentBalance     float64  `json:"current_balance"`
	BalanceOneMonthAgo float64  `json:"balance_one_month_ago"`
	CurrentValue       *float64 `json:"current_value"`
	PercentageIncrease *float64 `json:"percentage_increase"`
	Xirr               *float64 `json:"xirr"`
}

// AccountCashFlow represents a cash flow entry for XIRR calculations
// Amount should be negative for investments and positive for inflows
// Date is the transaction date
// Name is the transaction name, used to identify bookkeeping rows such as FD
// interest credits that must not enter the XIRR inputs.
// AccountID indicates which account the cash flow belongs to
// This is used internally by analytics services
// and is not part of API responses.
type AccountCashFlow struct {
	AccountID int64
	Amount    float64
	Date      time.Time
	Name      string
}

// AccountAnalyticsListResponse represents the complete analytics response
type AccountAnalyticsListResponse struct {
	AccountAnalytics []AccountBalanceAnalytics `json:"account_analytics"`
}

// CashBalanceDataPoint represents a single point in the cash-balance history.
// Investment ledgers are excluded because historical investment valuation is
// unavailable, so this is not a net-worth series.
type CashBalanceDataPoint struct {
	Date        string  `json:"date"`
	CashBalance float64 `json:"cash_balance"`
}

// CashBalanceHistoryResponse is the cash-only balance history over time.
type CashBalanceHistoryResponse struct {
	InitialBalance float64                `json:"initial_balance"`
	TotalIncome    float64                `json:"total_income"`
	TotalExpenses  float64                `json:"total_expenses"`
	TimeSeries     []CashBalanceDataPoint `json:"time_series"`
}

// CategoryAnalytics represents the category analytics for a given period
type CategoryAnalyticsResponse struct {
	CategoryTransactions []CategoryTransaction `json:"category_transactions"`
}

// CategoryTransaction represents the total transaction amount for a category
type CategoryTransaction struct {
	CategoryID   int64   `json:"category_id"`
	CategoryName string  `json:"category_name"`
	TotalAmount  float64 `json:"total_amount"`
}

// MonthlyAnalyticsResponse represents the monthly analytics response
type MonthlyAnalyticsResponse struct {
	TotalIncome   float64 `json:"total_income"`
	TotalExpenses float64 `json:"total_expenses"`
	TotalAmount   float64 `json:"total_amount"`
}

// InsightsSummary holds the headline figures of the insights response.
// Net worth, investment value and bank value are point in time; the period
// fields cover the requested date range only.
type InsightsSummary struct {
	NetWorth            float64 `json:"net_worth"`
	InvestmentValue     float64 `json:"investment_value"`
	BankValue           float64 `json:"bank_value"`
	PeriodIncome        float64 `json:"period_income"`
	PeriodExpenses      float64 `json:"period_expenses"`
	PeriodNet           float64 `json:"period_net"`
	SavingsRate         float64 `json:"savings_rate"`
	UncategorizedCount  int64   `json:"uncategorized_count"`
	UncategorizedAmount float64 `json:"uncategorized_amount"`
	RealizedInterest    float64 `json:"realized_interest"`
}

// InsightsMonthlyPoint is one month of household income and expenses
type InsightsMonthlyPoint struct {
	Month    string  `json:"month"`
	Income   float64 `json:"income"`
	Expenses float64 `json:"expenses"`
	Net      float64 `json:"net"`
}

// InsightsCategory is the signed household total for a category
type InsightsCategory struct {
	CategoryID   int64   `json:"category_id"`
	CategoryName string  `json:"category_name"`
	TotalAmount  float64 `json:"total_amount"`
}

// InsightsTopExpense is a payee aggregated from household debits
type InsightsTopExpense struct {
	Name    string  `json:"name"`
	Amount  float64 `json:"amount"`
	Count   int64   `json:"count"`
	Share   float64 `json:"share"`
	Average float64 `json:"average"`
}

// InsightsSpendingSummary describes expense behavior in the period
type InsightsSpendingSummary struct {
	ExpenseCount       int64   `json:"expense_count"`
	AverageTransaction float64 `json:"average_transaction"`
	MedianTransaction  float64 `json:"median_transaction"`
	LargestExpense     float64 `json:"largest_expense"`
	ActiveSpendingDays int64   `json:"active_spending_days"`
	NoSpendDays        int64   `json:"no_spend_days"`
}

// InsightsCategoryMonth is an expense-only category total for one month
type InsightsCategoryMonth struct {
	Month        string  `json:"month"`
	CategoryID   int64   `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Total        float64 `json:"total"`
}

// InsightsCategoryMovement compares expense-only category totals between the
// latest complete month and the month before it
type InsightsCategoryMovement struct {
	CategoryID   int64   `json:"category_id"`
	CategoryName string  `json:"category_name"`
	RecentTotal  float64 `json:"recent_total"`
	PriorTotal   float64 `json:"prior_total"`
	RecentShare  float64 `json:"recent_share"`
	PriorShare   float64 `json:"prior_share"`
	Change       float64 `json:"change"`
}

// InsightsWeekday is one weekday's expense behavior. ActiveDays is an
// intermediate value used to derive the average and is not serialized.
type InsightsWeekday struct {
	Weekday    int     `json:"weekday"`
	Total      float64 `json:"total"`
	Count      int64   `json:"count"`
	Average    float64 `json:"average"`
	Share      float64 `json:"share"`
	ActiveDays int64   `json:"-"`
}

// InsightsWeekdayBehavior groups weekday expenses
type InsightsWeekdayBehavior struct {
	Days         []InsightsWeekday `json:"days"`
	WeekendShare float64           `json:"weekend_share"`
}

// InsightsTrend describes spending over complete months
type InsightsTrend struct {
	RecentMonth               string  `json:"recent_month"`
	PriorMonth                string  `json:"prior_month"`
	RecentExpenses            float64 `json:"recent_expenses"`
	PriorExpenses             float64 `json:"prior_expenses"`
	Change                    float64 `json:"change"`
	TrailingThreeMonthAverage float64 `json:"trailing_three_month_average"`
}

// InsightsDataConfidence flags limitations that affect the metrics above
type InsightsDataConfidence struct {
	UncategorizedShare    float64  `json:"uncategorized_share"`
	MultiCategoryCount    int64    `json:"multi_category_count"`
	MultiCategoryShare    float64  `json:"multi_category_share"`
	LatestTransactionDate *string  `json:"latest_transaction_date"`
	StaleDays             int64    `json:"stale_days"`
	MultipleCurrencies    bool     `json:"multiple_currencies"`
	Currencies            []string `json:"currencies"`
}

// InsightsInvestment is a per-vehicle breakdown of the investment portfolio
type InsightsInvestment struct {
	AccountID          int64    `json:"account_id"`
	Name               string   `json:"name"`
	CurrentValue       float64  `json:"current_value"`
	Contributed        float64  `json:"contributed"`
	Distributed        float64  `json:"distributed"`
	RealizedInterest   float64  `json:"realized_interest"`
	Xirr               *float64 `json:"xirr"`
	PercentageIncrease *float64 `json:"percentage_increase"`
}

// AnalyticsInsightsResponse is the payload of GET /analytics/insights
type AnalyticsInsightsResponse struct {
	Summary          InsightsSummary            `json:"summary"`
	Monthly          []InsightsMonthlyPoint     `json:"monthly"`
	Categories       []InsightsCategory         `json:"categories"`
	TopExpenses      []InsightsTopExpense       `json:"top_expenses"`
	Investments      []InsightsInvestment       `json:"investments"`
	SpendingSummary  InsightsSpendingSummary    `json:"spending_summary"`
	CategoryMovement []InsightsCategoryMovement `json:"category_movement"`
	WeekdayBehavior  InsightsWeekdayBehavior    `json:"weekday_behavior"`
	Trend            InsightsTrend              `json:"trend"`
	DataConfidence   InsightsDataConfidence     `json:"data_confidence"`
}
