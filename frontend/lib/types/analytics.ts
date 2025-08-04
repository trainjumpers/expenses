// Analytics types matching the backend models

export type AnalyticsTimeRange = 
  | 'week'
  | 'month'
  | 'quarter'
  | 'year'
  | 'custom'
  | 'all_time';

export interface AnalyticsQuery {
  time_range: AnalyticsTimeRange;
  start_date?: string;
  end_date?: string;
  account_ids?: number[];
  category_ids?: number[];
}

export interface SpendingOverviewResponse {
  total_expenses: number;
  total_income: number;
  net_amount: number;
  transaction_count: number;
  average_expense: number;
  average_income: number;
  period: string;
}

export interface CategorySpendingItem {
  category_id: number;
  category_name: string;
  category_icon?: string;
  amount: number;
  percentage: number;
  count: number;
}

export interface CategorySpendingResponse {
  categories: CategorySpendingItem[];
  uncategorized: CategorySpendingItem;
  total_amount: number;
  total_count: number;
}

export interface TimeSeriesDataPoint {
  date: string;
  amount: number;
  count: number;
  income: number;
  expenses: number;
}

export interface SpendingTrendsResponse {
  data_points: TimeSeriesDataPoint[];
  period: string;
  granularity: string;
}

export interface AccountSpendingItem {
  account_id: number;
  account_name: string;
  bank_name?: string;
  amount: number;
  percentage: number;
  count: number;
}

export interface AccountSpendingResponse {
  accounts: AccountSpendingItem[];
  total_amount: number;
  total_count: number;
}

export interface TopTransactionItem {
  id: number;
  name: string;
  amount: number;
  date: string;
  account_name: string;
  categories: string[];
}

export interface TopTransactionsResponse {
  top_expenses: TopTransactionItem[];
  top_income: TopTransactionItem[];
  limit: number;
}

export interface MonthlyComparisonItem {
  month: string;
  month_name: string;
  amount: number;
  count: number;
  change: number;
  change_amount: number;
}

export interface MonthlyComparisonResponse {
  months: MonthlyComparisonItem[];
  period: string;
  total_months: number;
}

export interface RecurringTransactionPattern {
  pattern: string;
  amount: number;
  frequency: string;
  next_expected_date?: string;
  confidence: number;
  transaction_ids: number[];
  count: number;
}

export interface RecurringTransactionsResponse {
  patterns: RecurringTransactionPattern[];
  total_amount: number;
  count: number;
}

export interface AnalyticsSummaryResponse {
  overview: SpendingOverviewResponse;
  category_breakdown: CategorySpendingResponse;
  account_breakdown: AccountSpendingResponse;
  top_transactions: TopTransactionsResponse;
  monthly_comparison: MonthlyComparisonResponse;
  recurring_patterns: RecurringTransactionsResponse;
  period: string;
  generated_at: string;
}

export interface AnalyticsInsight {
  type: 'warning' | 'info' | 'success' | 'tip';
  title: string;
  description: string;
  actionable: boolean;
  priority: number;
  created_at: string;
}

export interface AnalyticsInsightsResponse {
  insights: AnalyticsInsight[];
  count: number;
  generated_at: string;
}

// UI-specific types
export interface AnalyticsFilters {
  timeRange: AnalyticsTimeRange;
  startDate?: Date;
  endDate?: Date;
  selectedAccounts: number[];
  selectedCategories: number[];
}

export interface ChartDataPoint {
  date: string;
  value: number;
  label?: string;
  color?: string;
}

export interface PieChartData {
  name: string;
  value: number;
  percentage: number;
  color: string;
}

export interface BarChartData {
  name: string;
  value: number;
  change?: number;
  color?: string;
}

// Chart configuration types
export interface ChartConfig {
  colors: string[];
  showLegend: boolean;
  showTooltip: boolean;
  height: number;
  responsive: boolean;
}

export interface AnalyticsCardProps {
  title: string;
  value: string | number;
  change?: number;
  changeLabel?: string;
  icon?: React.ReactNode;
  loading?: boolean;
  error?: string;
}

// Analytics dashboard layout types
export interface DashboardLayout {
  overview: boolean;
  categoryBreakdown: boolean;
  spendingTrends: boolean;
  accountBreakdown: boolean;
  topTransactions: boolean;
  monthlyComparison: boolean;
  recurringPatterns: boolean;
  insights: boolean;
}

export interface AnalyticsPreferences {
  defaultTimeRange: AnalyticsTimeRange;
  defaultGranularity: 'daily' | 'weekly' | 'monthly';
  layout: DashboardLayout;
  chartColors: string[];
  currency: string;
  dateFormat: string;
}
