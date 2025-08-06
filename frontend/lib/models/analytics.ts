// AccountBalanceAnalytics represents the analytics data for a single account
export interface AccountBalanceAnalytics {
  account_id: number;
  current_balance: number;
  balance_one_month_ago: number;
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
