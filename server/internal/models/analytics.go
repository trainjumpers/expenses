package models

// AccountBalanceAnalytics represents the analytics data for a single account
type AccountBalanceAnalytics struct {
	AccountID          int64   `json:"account_id"`
	CurrentBalance     float64 `json:"current_balance"`
	BalanceOneMonthAgo float64 `json:"balance_one_month_ago"`
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
	TimeSeries     []NetworthDataPoint `json:"time_series"`
}
