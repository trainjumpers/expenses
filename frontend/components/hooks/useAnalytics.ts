"use client";

import { getAccountAnalytics } from "@/lib/api/analytics";
import { AccountAnalyticsListResponse } from "@/lib/models/analytics";
import { queryKeys } from "@/lib/query-client";
import { useQuery } from "@tanstack/react-query";

export function useAccountAnalytics() {
  return useQuery<AccountAnalyticsListResponse>({
    queryKey: queryKeys.analytics.accountAnalytics,
    queryFn: ({ signal }) => getAccountAnalytics(signal),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}