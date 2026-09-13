// AccountBalanceAnalytics represents the analytics data for a single account
export interface AccountBalanceAnalytics {
  account_id: number;
  current_balance: number;
  balance_one_month_ago: number;
  current_value?: number | null;
  percentage_increase?: number | null;
  xirr?: number | null;
}

// AccountAnalyticsListResponse represents the complete analytics response
export interface AccountAnalyticsListResponse {
  account_analytics: AccountBalanceAnalytics[];
}

// CashBalanceDataPoint is one day of the cash-only balance history
export interface CashBalanceDataPoint {
  date: string;
  cash_balance: number;
}

// CashBalanceHistoryResponse is the cash-only balance history over time
export interface CashBalanceHistoryResponse {
  initial_balance: number;
  total_income: number;
  total_expenses: number;
  time_series: CashBalanceDataPoint[];
}

// CategoryAnalyticsResponse represents the category analytics for a given period
export interface CategoryAnalyticsResponse {
  category_transactions: CategoryTransaction[];
}

// CategoryTransaction represents the total transaction amount for a category
export interface CategoryTransaction {
  category_id: number;
  category_name: string;
  total_amount: number;
}

// MonthlyAnalyticsResponse represents the monthly analytics response
export interface MonthlyAnalyticsResponse {
  total_income: number;
  total_expenses: number;
  total_amount: number;
}

// InsightsSummary holds the headline figures of the insights response
export interface InsightsSummary {
  net_worth: number;
  investment_value: number;
  bank_value: number;
  period_income: number;
  period_expenses: number;
  period_net: number;
  savings_rate: number;
  uncategorized_count: number;
  uncategorized_amount: number;
  realized_interest: number;
}

// InsightsMonthlyPoint is one month of household income and expenses
export interface InsightsMonthlyPoint {
  month: string;
  income: number;
  expenses: number;
  net: number;
}

// InsightsCategory is the signed household total for a category
export interface InsightsCategory {
  category_id: number;
  category_name: string;
  total_amount: number;
}

// InsightsTopExpense is a payee aggregated from household debits
export interface InsightsTopExpense {
  name: string;
  amount: number;
  count: number;
  share: number;
  average: number;
}

// InsightsInvestment is a per-vehicle breakdown of the investment portfolio
export interface InsightsInvestment {
  account_id: number;
  name: string;
  current_value: number;
  contributed: number;
  distributed: number;
  realized_interest: number;
  xirr?: number | null;
  percentage_increase?: number | null;
}

// InsightsSpendingSummary describes expense behavior in the requested range
export interface InsightsSpendingSummary {
  expense_count: number;
  average_transaction: number;
  median_transaction: number;
  largest_expense: number;
  active_spending_days: number;
  no_spend_days: number;
}

// InsightsCategoryMovement compares expense-only category totals between the
// latest complete month and the month before it
export interface InsightsCategoryMovement {
  category_id: number;
  category_name: string;
  recent_total: number;
  prior_total: number;
  recent_share: number;
  prior_share: number;
  change: number;
}

// InsightsWeekday is one weekday's expense behavior
export interface InsightsWeekday {
  weekday: number;
  total: number;
  count: number;
  average: number;
  share: number;
}

// InsightsWeekdayBehavior groups weekday spending
export interface InsightsWeekdayBehavior {
  days: InsightsWeekday[];
  weekend_share: number;
}

// InsightsTrend describes the spend trajectory over complete months
export interface InsightsTrend {
  recent_month: string;
  prior_month: string;
  recent_expenses: number;
  prior_expenses: number;
  change: number;
  trailing_three_month_average: number;
}

// InsightsDataConfidence flags limitations that affect the metrics above
export interface InsightsDataConfidence {
  uncategorized_share: number;
  multi_category_count: number;
  multi_category_share: number;
  latest_transaction_date?: string | null;
  stale_days: number;
  multiple_currencies: boolean;
  currencies: string[];
}

// AnalyticsInsightsResponse is the payload of GET /analytics/insights
export interface AnalyticsInsightsResponse {
  summary: InsightsSummary;
  monthly: InsightsMonthlyPoint[];
  categories: InsightsCategory[];
  top_expenses: InsightsTopExpense[];
  investments: InsightsInvestment[];
  spending_summary: InsightsSpendingSummary;
  category_movement: InsightsCategoryMovement[];
  weekday_behavior: InsightsWeekdayBehavior;
  trend: InsightsTrend;
  data_confidence: InsightsDataConfidence;
}
