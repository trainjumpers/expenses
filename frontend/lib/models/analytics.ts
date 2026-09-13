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

// NetworthDataPoint represents a single point in the networth time series
export interface NetworthDataPoint {
  date: string;
  networth: number;
}

// NetworthTimeSeriesResponse represents the networth over time response
export interface NetworthTimeSeriesResponse {
  initial_balance: number;
  total_income: number;
  total_expenses: number;
  time_series: NetworthDataPoint[];
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

// AnalyticsInsightsResponse is the payload of GET /analytics/insights
export interface AnalyticsInsightsResponse {
  summary: InsightsSummary;
  monthly: InsightsMonthlyPoint[];
  categories: InsightsCategory[];
  top_expenses: InsightsTopExpense[];
  investments: InsightsInvestment[];
}
