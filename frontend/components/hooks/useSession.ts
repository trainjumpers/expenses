"use client";

import { refresh as refreshApi } from "@/lib/api/auth";
import { checkUser } from "@/lib/api/user";
import { queryKeys } from "@/lib/query-client";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { usePathname, useRouter } from "next/navigation";
import { useEffect } from "react";

export const PUBLIC_ROUTES = ["/login", "/signup"];

export function useSession() {
  const pathname = usePathname();
  const router = useRouter();
  const queryClient = useQueryClient();
  const isPublicRoute = PUBLIC_ROUTES.includes(pathname);

  const sessionQuery = useQuery({
    queryKey: queryKeys.session,
    queryFn: async () => {
      try {
        const response = await checkUser();

        if (response && response.ok) {
          return { isValid: true, needsRefresh: false };
        }

        if (response && response.status === 401) {
          try {
            const refreshResponse = await refreshApi();

            if (refreshResponse.ok) {
              const verifyResponse = await checkUser();
              if (verifyResponse && verifyResponse.ok) {
                return { isValid: true, needsRefresh: false };
              }
            }
          } catch (refreshError) {
            console.error("Token refresh failed:", refreshError);
          }

          return { isValid: false, needsRefresh: false };
        }

        return { isValid: false, needsRefresh: false };
      } catch (error) {
        console.error("Session check failed:", error);
        return { isValid: false, needsRefresh: false };
      }
    },
    enabled: !isPublicRoute,
    staleTime: 5 * 60 * 1000,
    retry: false,
    refetchOnWindowFocus: true,
    refetchInterval: 15 * 60 * 1000,
  });

  useEffect(() => {
    if (isPublicRoute) return;

    if (
      sessionQuery.data?.isValid === false &&
      !sessionQuery.isLoading &&
      !sessionQuery.isFetching
    ) {
      console.error("User not authenticated, redirecting to login");
      queryClient.clear();
      router.push("/login");
    }
  }, [
    sessionQuery.data?.isValid,
    sessionQuery.isLoading,
    sessionQuery.isFetching,
    isPublicRoute,
    queryClient,
    router,
  ]);

  useEffect(() => {
    if (isPublicRoute && sessionQuery.data?.isValid === true) {
      router.push("/");
    }
  }, [isPublicRoute, sessionQuery.data?.isValid, router]);

  return {
    isAuthenticated: sessionQuery.data?.isValid ?? false,
    isLoading: sessionQuery.isLoading || sessionQuery.isFetching,
    error: sessionQuery.error,
    refetch: sessionQuery.refetch,
  };
}
