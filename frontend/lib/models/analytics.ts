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