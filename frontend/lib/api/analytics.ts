import { apiRequest } from "@/lib/api/request";
import { API_BASE_URL } from "@/lib/constants/api";
import { AccountAnalyticsListResponse } from "@/lib/models/analytics";

export async function getAccountAnalytics(signal?: AbortSignal): Promise<AccountAnalyticsListResponse> {
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