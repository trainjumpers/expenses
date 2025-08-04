import { QueryClient } from "@tanstack/react-query";

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5 minutes
      gcTime: 10 * 60 * 1000, // 10 minutes (formerly cacheTime)
      retry: (failureCount, error: unknown) => {
        const errorStatus = (error as { status?: number })?.status;
        if (
          errorStatus &&
          errorStatus >= 400 &&
          errorStatus < 500 &&
          ![408, 429].includes(errorStatus)
        ) {
          return false;
        }
        return failureCount < 3;
      },
      refetchOnWindowFocus: false,
      refetchOnReconnect: true,
    },
    mutations: {
      onError: (error: unknown) => {
        const errorStatus = (error as { status?: number })?.status;
        if (errorStatus === 401) {
          console.log("401 error in mutation, session may have expired");
          queryClient.clear();
          if (typeof window !== "undefined") {
            window.location.href = "/login";
          }
          return;
        }
        console.error("Mutation error:", error);
      },
    },
  },
});

export const queryKeys = {
  user: ["user"] as const,
  accounts: ["accounts"] as const,
  account: (id: number) => ["accounts", id] as const,
  categories: ["categories"] as const,
  category: (id: number) => ["categories", id] as const,
  transactions: (params?: Record<string, unknown>) =>
    params ? (["transactions", params] as const) : (["transactions"] as const),
  transaction: (id: number) => ["transactions", id] as const,
  rules: ["rules"] as const,
  rule: (id: number) => ["rules", id] as const,
  session: ["session"] as const,
  analytics: {
    accountAnalytics: ["analytics", "account"] as const,
    networthTimeSeries: (startDate: string, endDate: string) =>
      ["analytics", "networth", startDate, endDate] as const,
  },
} as const;
