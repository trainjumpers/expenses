import { useQuery, UseQueryOptions } from '@tanstack/react-query';
import {
  getSpendingOverview,
  getCategorySpending,
  getSpendingTrends,
  getAccountSpending,
  getTopTransactions,
  getMonthlyComparison,
  getRecurringTransactions,
  getAnalyticsSummary,
  getAnalyticsInsights,
  getDashboardData,
  getCategoryAnalytics,
  getAccountAnalytics,
} from '../lib/api/analytics';
import type {
  AnalyticsQuery,
  SpendingOverviewResponse,
  CategorySpendingResponse,
  SpendingTrendsResponse,
  AccountSpendingResponse,
  TopTransactionsResponse,
  MonthlyComparisonResponse,
  RecurringTransactionsResponse,
  AnalyticsSummaryResponse,
  AnalyticsInsightsResponse,
} from '../lib/types/analytics';

// Query keys for React Query caching
export const ANALYTICS_QUERY_KEYS = {
  all: ['analytics'] as const,
  overview: (query: AnalyticsQuery) => ['analytics', 'overview', query] as const,
  categories: (query: AnalyticsQuery) => ['analytics', 'categories', query] as const,
  trends: (query: AnalyticsQuery, granularity: string) => 
    ['analytics', 'trends', query, granularity] as const,
  accounts: (query: AnalyticsQuery) => ['analytics', 'accounts', query] as const,
  topTransactions: (query: AnalyticsQuery, limit: number) => 
    ['analytics', 'top-transactions', query, limit] as const,
  monthlyComparison: (query: AnalyticsQuery) => 
    ['analytics', 'monthly-comparison', query] as const,
  recurring: (query: AnalyticsQuery) => ['analytics', 'recurring', query] as const,
  summary: (query: AnalyticsQuery) => ['analytics', 'summary', query] as const,
  insights: (query: AnalyticsQuery) => ['analytics', 'insights', query] as const,
  dashboard: (query: AnalyticsQuery) => ['analytics', 'dashboard', query] as const,
  categoryDetail: (categoryId: number, timeRange: string) => 
    ['analytics', 'category', categoryId, timeRange] as const,
  accountDetail: (accountId: number, timeRange: string) => 
    ['analytics', 'account', accountId, timeRange] as const,
} as const;

// Default query options for analytics
const defaultOptions = {
  staleTime: 5 * 60 * 1000, // 5 minutes
  gcTime: 10 * 60 * 1000, // 10 minutes
  refetchOnWindowFocus: false,
  retry: 2,
} as const;

/**
 * Hook for spending overview analytics
 */
export function useSpendingOverview(
  query: AnalyticsQuery,
  options?: UseQueryOptions<SpendingOverviewResponse>
) {
  return useQuery({
    queryKey: ANALYTICS_QUERY_KEYS.overview(query),
    queryFn: () => getSpendingOverview(query),
    ...defaultOptions,
    ...options,
  });
}

/**
 * Hook for category spending breakdown
 */
export function useCategorySpending(
  query: AnalyticsQuery,
  options?: UseQueryOptions<CategorySpendingResponse>
) {
  return useQuery({
    queryKey: ANALYTICS_QUERY_KEYS.categories(query),
    queryFn: () => getCategorySpending(query),
    ...defaultOptions,
    ...options,
  });
}

/**
 * Hook for spending trends with granularity
 */
export function useSpendingTrends(
  query: AnalyticsQuery,
  granularity: 'daily' | 'weekly' | 'monthly' = 'daily',
  options?: UseQueryOptions<SpendingTrendsResponse>
) {
  return useQuery({
    queryKey: ANALYTICS_QUERY_KEYS.trends(query, granularity),
    queryFn: () => getSpendingTrends(query, granularity),
    ...defaultOptions,
    ...options,
  });
}

/**
 * Hook for account spending breakdown
 */
export function useAccountSpending(
  query: AnalyticsQuery,
  options?: UseQueryOptions<AccountSpendingResponse>
) {
  return useQuery({
    queryKey: ANALYTICS_QUERY_KEYS.accounts(query),
    queryFn: () => getAccountSpending(query),
    ...defaultOptions,
    ...options,
  });
}

/**
 * Hook for top transactions
 */
export function useTopTransactions(
  query: AnalyticsQuery,
  limit: number = 10,
  options?: UseQueryOptions<TopTransactionsResponse>
) {
  return useQuery({
    queryKey: ANALYTICS_QUERY_KEYS.topTransactions(query, limit),
    queryFn: () => getTopTransactions(query, limit),
    ...defaultOptions,
    ...options,
  });
}

/**
 * Hook for monthly spending comparison
 */
export function useMonthlyComparison(
  query: AnalyticsQuery,
  options?: UseQueryOptions<MonthlyComparisonResponse>
) {
  return useQuery({
    queryKey: ANALYTICS_QUERY_KEYS.monthlyComparison(query),
    queryFn: () => getMonthlyComparison(query),
    ...defaultOptions,
    ...options,
  });
}

/**
 * Hook for recurring transaction patterns
 */
export function useRecurringTransactions(
  query: AnalyticsQuery,
  options?: UseQueryOptions<RecurringTransactionsResponse>
) {
  return useQuery({
    queryKey: ANALYTICS_QUERY_KEYS.recurring(query),
    queryFn: () => getRecurringTransactions(query),
    ...defaultOptions,
    ...options,
  });
}

/**
 * Hook for comprehensive analytics summary
 */
export function useAnalyticsSummary(
  query: AnalyticsQuery,
  options?: UseQueryOptions<AnalyticsSummaryResponse>
) {
  return useQuery({
    queryKey: ANALYTICS_QUERY_KEYS.summary(query),
    queryFn: () => getAnalyticsSummary(query),
    ...defaultOptions,
    ...options,
  });
}

/**
 * Hook for AI-generated analytics insights
 */
export function useAnalyticsInsights(
  query: AnalyticsQuery,
  options?: UseQueryOptions<AnalyticsInsightsResponse>
) {
  return useQuery({
    queryKey: ANALYTICS_QUERY_KEYS.insights(query),
    queryFn: () => getAnalyticsInsights(query),
    ...defaultOptions,
    ...options,
  });
}

/**
 * Hook for dashboard data (summary + insights)
 */
export function useDashboardData(
  query: AnalyticsQuery,
  options?: UseQueryOptions<{
    summary: AnalyticsSummaryResponse;
    insights: AnalyticsInsightsResponse;
  }>
) {
  return useQuery({
    queryKey: ANALYTICS_QUERY_KEYS.dashboard(query),
    queryFn: () => getDashboardData(query),
    ...defaultOptions,
    ...options,
  });
}

/**
 * Hook for detailed category analytics
 */
export function useCategoryAnalytics(
  categoryId: number,
  timeRange: AnalyticsQuery['time_range'] = 'month',
  options?: UseQueryOptions<{
    overview: SpendingOverviewResponse;
    trends: SpendingTrendsResponse;
    topTransactions: TopTransactionsResponse;
  }>
) {
  return useQuery({
    queryKey: ANALYTICS_QUERY_KEYS.categoryDetail(categoryId, timeRange),
    queryFn: () => getCategoryAnalytics(categoryId, timeRange),
    ...defaultOptions,
    ...options,
  });
}

/**
 * Hook for detailed account analytics
 */
export function useAccountAnalytics(
  accountId: number,
  timeRange: AnalyticsQuery['time_range'] = 'month',
  options?: UseQueryOptions<{
    overview: SpendingOverviewResponse;
    categories: CategorySpendingResponse;
    trends: SpendingTrendsResponse;
  }>
) {
  return useQuery({
    queryKey: ANALYTICS_QUERY_KEYS.accountDetail(accountId, timeRange),
    queryFn: () => getAccountAnalytics(accountId, timeRange),
    ...defaultOptions,
    ...options,
  });
}

// Utility hooks for common patterns

/**
 * Hook that provides loading state for multiple analytics queries
 */
export function useAnalyticsLoading(queries: Array<{ isLoading: boolean }>) {
  return queries.some(query => query.isLoading);
}

/**
 * Hook that provides error state for multiple analytics queries
 */
export function useAnalyticsError(queries: Array<{ error: Error | null }>) {
  return queries.find(query => query.error)?.error || null;
}

/**
 * Hook for analytics with automatic refetch on time range change
 */
export function useAnalyticsWithRefetch(
  query: AnalyticsQuery,
  enabled: boolean = true
) {
  const overview = useSpendingOverview(query, { enabled });
  const categories = useCategorySpending(query, { enabled });
  const trends = useSpendingTrends(query, 'daily', { enabled });
  const accounts = useAccountSpending(query, { enabled });

  const isLoading = useAnalyticsLoading([overview, categories, trends, accounts]);
  const error = useAnalyticsError([overview, categories, trends, accounts]);

  return {
    overview: overview.data,
    categories: categories.data,
    trends: trends.data,
    accounts: accounts.data,
    isLoading,
    error,
    refetch: () => {
      overview.refetch();
      categories.refetch();
      trends.refetch();
      accounts.refetch();
    },
  };
}

/**
 * Hook for real-time analytics updates
 */
export function useRealTimeAnalytics(
  query: AnalyticsQuery,
  intervalMs: number = 30000 // 30 seconds
) {
  return useAnalyticsSummary(query, {
    refetchInterval: intervalMs,
    refetchIntervalInBackground: false,
  });
}
