import { apiRequest } from "@/lib/api/request";
import { API_BASE_URL } from "@/lib/constants/api";
import {
  AccountAnalyticsListResponse,
  NetworthTimeSeriesResponse,
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

export async function getNetworthTimeSeries(
  startDate: string,
  endDate: string,
  signal?: AbortSignal
): Promise<NetworthTimeSeriesResponse> {
  const params = new URLSearchParams({
    start_date: startDate,
    end_date: endDate,
  });

  return apiRequest<NetworthTimeSeriesResponse>(
    `${API_BASE_URL}/analytics/networth?${params.toString()}`,
    {
      credentials: "include",
      signal,
    },
    "analytics",
    [],
    "Failed to fetch networth time series"
  );
}
