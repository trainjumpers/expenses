import { apiRequest } from "@/lib/api/request";
import { API_BASE_URL } from "@/lib/constants/api";

import type {
  AccountSpendingResponse,
  AnalyticsInsightsResponse,
  AnalyticsQuery,
  AnalyticsSummaryResponse,
  CategorySpendingResponse,
  MonthlyComparisonResponse,
  RecurringTransactionsResponse,
  SpendingOverviewResponse,
  SpendingTrendsResponse,
  TopTransactionsResponse,
} from "../types/analytics";

/**
 * Get spending overview analytics
 */
export async function getSpendingOverview(
  query: AnalyticsQuery
): Promise<SpendingOverviewResponse> {
  return apiRequest<SpendingOverviewResponse>(
    `${API_BASE_URL}/analytics/overview`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify(query),
    },
    "analytics overview"
  );
}

/**
 * Get category spending breakdown
 */
export async function getCategorySpending(
  query: AnalyticsQuery
): Promise<CategorySpendingResponse> {
  return apiRequest<CategorySpendingResponse>(
    `${API_BASE_URL}/analytics/categories`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify(query),
    },
    "category spending"
  );
}

/**
 * Get spending trends with granularity
 */
export async function getSpendingTrends(
  query: AnalyticsQuery,
  granularity: "daily" | "weekly" | "monthly" = "daily"
): Promise<SpendingTrendsResponse> {
  const url = `${API_BASE_URL}/analytics/trends?granularity=${granularity}`;
  return apiRequest<SpendingTrendsResponse>(
    url,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify(query),
    },
    "spending trends"
  );
}

/**
 * Get account spending breakdown
 */
export async function getAccountSpending(
  query: AnalyticsQuery
): Promise<AccountSpendingResponse> {
  return apiRequest<AccountSpendingResponse>(
    `${API_BASE_URL}/analytics/accounts`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify(query),
    },
    "account spending"
  );
}

/**
 * Get top transactions (highest expenses and income)
 */
export async function getTopTransactions(
  query: AnalyticsQuery,
  limit: number = 10
): Promise<TopTransactionsResponse> {
  const url = `${API_BASE_URL}/analytics/top-transactions?limit=${limit}`;
  return apiRequest<TopTransactionsResponse>(
    url,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify(query),
    },
    "top transactions"
  );
}

/**
 * Get monthly spending comparison
 */
export async function getMonthlyComparison(
  query: AnalyticsQuery
): Promise<MonthlyComparisonResponse> {
  return apiRequest<MonthlyComparisonResponse>(
    `${API_BASE_URL}/analytics/monthly-comparison`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify(query),
    },
    "monthly comparison"
  );
}

/**
 * Get recurring transaction patterns
 */
export async function getRecurringTransactions(
  query: AnalyticsQuery
): Promise<RecurringTransactionsResponse> {
  return apiRequest<RecurringTransactionsResponse>(
    `${API_BASE_URL}/analytics/recurring`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify(query),
    },
    "recurring transactions"
  );
}

/**
 * Get comprehensive analytics summary
 */
export async function getAnalyticsSummary(
  query: AnalyticsQuery
): Promise<AnalyticsSummaryResponse> {
  return apiRequest<AnalyticsSummaryResponse>(
    `${API_BASE_URL}/analytics/summary`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify(query),
    },
    "analytics summary"
  );
}

/**
 * Get AI-generated analytics insights
 */
export async function getAnalyticsInsights(
  query: AnalyticsQuery
): Promise<AnalyticsInsightsResponse> {
  return apiRequest<AnalyticsInsightsResponse>(
    `${API_BASE_URL}/analytics/insights`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify(query),
    },
    "analytics insights"
  );
}

// Utility functions for common analytics operations

/**
 * Get analytics data for a specific time range
 */
export async function getAnalyticsForTimeRange(
  timeRange: AnalyticsQuery["time_range"],
  startDate?: string,
  endDate?: string,
  accountIds?: number[],
  categoryIds?: number[]
) {
  const query: AnalyticsQuery = {
    time_range: timeRange,
    start_date: startDate,
    end_date: endDate,
    account_ids: accountIds,
    category_ids: categoryIds,
  };

  const [overview, categories, trends, accounts] = await Promise.all([
    getSpendingOverview(query),
    getCategorySpending(query),
    getSpendingTrends(query),
    getAccountSpending(query),
  ]);

  return {
    overview,
    categories,
    trends,
    accounts,
  };
}

/**
 * Get dashboard data (summary + insights)
 */
export async function getDashboardData(query: AnalyticsQuery) {
  const [summary, insights] = await Promise.all([
    getAnalyticsSummary(query),
    getAnalyticsInsights(query),
  ]);
  console.log(summary, insights);

  return {
    summary,
    insights,
  };
}

/**
 * Get detailed analytics for a specific category
 */
export async function getCategoryAnalytics(
  categoryId: number,
  timeRange: AnalyticsQuery["time_range"] = "month"
) {
  const query: AnalyticsQuery = {
    time_range: timeRange,
    category_ids: [categoryId],
  };

  const [overview, trends, topTransactions] = await Promise.all([
    getSpendingOverview(query),
    getSpendingTrends(query),
    getTopTransactions(query, 5),
  ]);

  return {
    overview,
    trends,
    topTransactions,
  };
}

/**
 * Get detailed analytics for a specific account
 */
export async function getAccountAnalytics(
  accountId: number,
  timeRange: AnalyticsQuery["time_range"] = "month"
) {
  const query: AnalyticsQuery = {
    time_range: timeRange,
    account_ids: [accountId],
  };

  const [overview, categories, trends] = await Promise.all([
    getSpendingOverview(query),
    getCategorySpending(query),
    getSpendingTrends(query),
  ]);

  return {
    overview,
    categories,
    trends,
  };
}
