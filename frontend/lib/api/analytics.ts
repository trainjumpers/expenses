import { apiRequest } from "@/lib/api/request";
import { API_BASE_URL } from "@/lib/constants/api";
import type {
  AccountAnalyticsListResponse,
  AnalyticsInsightsResponse,
  CashBalanceHistoryResponse,
  CategoryAnalyticsResponse,
  MonthlyAnalyticsResponse,
} from "@/lib/models/analytics";

export async function getAccountAnalytics(
  signal?: AbortSignal
): Promise<AccountAnalyticsListResponse> {
  return apiRequest<AccountAnalyticsListResponse>(
    `${API_BASE_URL}/analytics/account`,
    {
      credentials: "include",
      signal,
    },
    "analytics",
    [],
    "Failed to fetch account analytics"
  );
}

export async function getCashBalanceHistory(
  startDate: string,
  endDate: string,
  signal?: AbortSignal
): Promise<CashBalanceHistoryResponse> {
  const params = new URLSearchParams({
    start_date: startDate,
    end_date: endDate,
  });

  return apiRequest<CashBalanceHistoryResponse>(
    `${API_BASE_URL}/analytics/cash-balance?${params.toString()}`,
    {
      credentials: "include",
      signal,
    },
    "analytics",
    [],
    "Failed to fetch cash balance history"
  );
}

export async function getCategoryAnalytics(
  startDate: string,
  endDate: string,
  categoryIds?: number[],
  signal?: AbortSignal
): Promise<CategoryAnalyticsResponse> {
  const params = new URLSearchParams({
    start_date: startDate,
    end_date: endDate,
  });

  if (categoryIds && categoryIds.length > 0) {
    params.set("category_ids", categoryIds.join(","));
  }

  return apiRequest<CategoryAnalyticsResponse>(
    `${API_BASE_URL}/analytics/category?${params.toString()}`,
    {
      credentials: "include",
      signal,
    },
    "analytics",
    [],
    "Failed to fetch category analytics"
  );
}

export async function getMonthlyAnalytics(
  startDate: string,
  endDate: string,
  signal?: AbortSignal
): Promise<MonthlyAnalyticsResponse> {
  const params = new URLSearchParams({
    start_date: startDate,
    end_date: endDate,
  });

  return apiRequest<MonthlyAnalyticsResponse>(
    `${API_BASE_URL}/analytics/monthly?${params.toString()}`,
    {
      credentials: "include",
      signal,
    },
    "analytics",
    [],
    "Failed to fetch monthly analytics"
  );
}

export async function getInsights(
  startDate: string,
  endDate: string,
  signal?: AbortSignal
): Promise<AnalyticsInsightsResponse> {
  const params = new URLSearchParams({
    start_date: startDate,
    end_date: endDate,
  });

  return apiRequest<AnalyticsInsightsResponse>(
    `${API_BASE_URL}/analytics/insights?${params.toString()}`,
    {
      credentials: "include",
      signal,
    },
    "analytics",
    [],
    "Failed to fetch analytics insights"
  );
}
