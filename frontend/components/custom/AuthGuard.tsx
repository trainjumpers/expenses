"use client";

import { PUBLIC_ROUTES, useSession } from "@/components/hooks/useSession";
import { usePathname } from "next/navigation";
import { ReactNode } from "react";

import DashboardSkeleton from "./Dashboard/DashboardSkeleton";

interface AuthGuardProps {
  children: ReactNode;
}

export function AuthGuard({ children }: AuthGuardProps) {
  const { isAuthenticated, isLoading } = useSession();
  const pathname = usePathname();
  const isPublicRoute = PUBLIC_ROUTES.includes(pathname);
  if (!isPublicRoute && isLoading) {
    return <DashboardSkeleton />;
  }
  if (isPublicRoute) {
    return <>{children}</>;
  }
  if (!isAuthenticated) {
    return <DashboardSkeleton />;
  }

  return <>{children}</>;
}
