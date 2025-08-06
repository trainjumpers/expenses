"use client";

import {
  getAccountAnalytics,
  getMonthlyAnalytics,
  getNetworthTimeSeries,
} from "@/lib/api/analytics";
import {
  AccountAnalyticsListResponse,
  MonthlyAnalyticsResponse,
  NetworthTimeSeriesResponse,
} from "@/lib/models/analytics";
import { queryKeys } from "@/lib/query-client";
import { useQuery } from "@tanstack/react-query";

export function useAccountAnalytics() {
  return useQuery<AccountAnalyticsListResponse>({
    queryKey: queryKeys.analytics.accountAnalytics,
    queryFn: ({ signal }) => getAccountAnalytics(signal),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

export function useNetworthTimeSeries(startDate: string, endDate: string) {
  return useQuery<NetworthTimeSeriesResponse>({
    queryKey: queryKeys.analytics.networthTimeSeries(startDate, endDate),
    queryFn: ({ signal }) => getNetworthTimeSeries(startDate, endDate, signal),
    staleTime: 5 * 60 * 1000, // 5 minutes
    enabled: !!startDate && !!endDate, // Only run query when dates are provided
  });
}

export function useMonthlyAnalytics(startDate: string, endDate: string) {
  return useQuery<MonthlyAnalyticsResponse>({
    queryKey: queryKeys.analytics.monthlyAnalytics(startDate, endDate),
    queryFn: ({ signal }) => getMonthlyAnalytics(startDate, endDate, signal),
    staleTime: 5 * 60 * 1000, // 5 minutes
    enabled: !!startDate && !!endDate, // Only run query when dates are provided
  });
}
