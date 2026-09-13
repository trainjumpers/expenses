"use client";

import {
  getAccountAnalytics,
  getCashBalanceHistory,
  getInsights,
  getMonthlyAnalytics,
} from "@/lib/api/analytics";
import type {
  AccountAnalyticsListResponse,
  AnalyticsInsightsResponse,
  CashBalanceHistoryResponse,
  MonthlyAnalyticsResponse,
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

export function useCashBalanceHistory(startDate: string, endDate: string) {
  return useQuery<CashBalanceHistoryResponse>({
    queryKey: queryKeys.analytics.cashBalanceHistory(startDate, endDate),
    queryFn: ({ signal }) => getCashBalanceHistory(startDate, endDate, signal),
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

export function useInsights(startDate: string, endDate: string) {
  return useQuery<AnalyticsInsightsResponse>({
    queryKey: queryKeys.analytics.insights(startDate, endDate),
    queryFn: ({ signal }) => getInsights(startDate, endDate, signal),
    staleTime: 5 * 60 * 1000, // 5 minutes
    enabled: !!startDate && !!endDate, // Only run query when dates are provided
  });
}
