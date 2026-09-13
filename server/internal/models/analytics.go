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

// NetworthDataPoint represents a single point in the networth time series
type NetworthDataPoint struct {
	Date     string  `json:"date"`
	Networth float64 `json:"networth"`
}

// NetworthTimeSeriesResponse represents the networth over time response
type NetworthTimeSeriesResponse struct {
	InitialBalance float64             `json:"initial_balance"`
	TotalIncome    float64             `json:"total_income"`
	TotalExpenses  float64             `json:"total_expenses"`
	TimeSeries     []NetworthDataPoint `json:"time_series"`
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
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Count  int64   `json:"count"`
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
	Summary     InsightsSummary        `json:"summary"`
	Monthly     []InsightsMonthlyPoint `json:"monthly"`
	Categories  []InsightsCategory     `json:"categories"`
	TopExpenses []InsightsTopExpense   `json:"top_expenses"`
	Investments []InsightsInvestment   `json:"investments"`
}
