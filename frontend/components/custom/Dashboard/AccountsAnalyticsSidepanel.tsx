"use client";

import { AddAccountModal } from "@/components/custom/Modal/Accounts/AddAccountModal";
import { useAccounts } from "@/components/hooks/useAccounts";
import { useAccountAnalytics } from "@/components/hooks/useAnalytics";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Plus } from "lucide-react";
import { useState } from "react";

interface AccountData {
  id: number;
  name: string;
  balance: number;
  currency: string;
  percentageChange: number;
}

interface AccountsAnalyticsSidepanelProps {
  className?: string;
}

const formatCurrency = (amount: number, currency: string): string => {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: currency,
    minimumFractionDigits: 2,
  }).format(amount);
};

const formatPercentage = (percentage: number): string => {
  const sign = percentage > 0 ? "+" : "";
  return `${sign}${percentage.toFixed(1)}%`;
};

const calculatePercentageChange = (
  current: number,
  previous: number
): number => {
  if (previous === 0) return 0;
  return ((current - previous) / previous) * 100;
};

export function AccountsAnalyticsSidepanel({
  className,
}: AccountsAnalyticsSidepanelProps) {
  const { data: analyticsData, isLoading: analyticsLoading } =
    useAccountAnalytics();
  const { data: accountsData, isLoading: accountsLoading } = useAccounts();
  const [isAddAccountModalOpen, setIsAddAccountModalOpen] = useState(false);

  const isLoading = analyticsLoading || accountsLoading;

  // Combine analytics data with account names
  const accounts: AccountData[] =
    analyticsData?.account_analytics?.map((analytics) => {
      const account = accountsData?.find(
        (acc) => acc.id === analytics.account_id
      );
      const percentageChange = calculatePercentageChange(
        analytics.current_balance,
        analytics.balance_one_month_ago
      );

      return {
        id: analytics.account_id,
        name: account?.name || `Account ${analytics.account_id}`,
        currency: account?.currency || "INR",
        balance: analytics.current_balance,
        percentageChange,
      };
    }) || [];

  if (isLoading) {
    return (
      <Card className={`w-80 ${className}`}>
        <CardHeader className="pb-4">
          <div className="flex items-center justify-between">
            <Skeleton className="h-6 w-16" />
            <Skeleton className="h-8 w-8" />
          </div>
          <Skeleton className="h-4 w-20" />
        </CardHeader>
        <CardContent className="space-y-1">
          {Array.from({ length: 6 }).map((_, index) => (
            <div
              key={index}
              className="flex items-center justify-between py-3 px-2"
            >
              <div className="flex items-center space-x-3">
                <Skeleton className="h-4 w-4" />
                <Skeleton className="h-4 w-20" />
              </div>
              <div className="text-right space-y-1">
                <Skeleton className="h-4 w-16" />
                <Skeleton className="h-3 w-10" />
              </div>
            </div>
          ))}
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className={`w-80 ${className}`}>
      <CardHeader className="pb-4">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg font-semibold">Assets</CardTitle>
          <Button
            variant="ghost"
            size="sm"
            className="h-8 w-8 p-0"
            onClick={() => setIsAddAccountModalOpen(true)}
          >
            <Plus className="h-4 w-4" />
          </Button>
        </div>
      </CardHeader>

      <CardContent className="space-y-1">
        {accounts.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            <p className="text-sm">No accounts found</p>
          </div>
        ) : (
          accounts.map((account) => (
            <div
              key={account.id}
              className="flex items-center justify-between py-3 px-2 rounded-md hover:bg-muted/50 cursor-pointer group"
            >
              <div className="flex items-center space-x-3">
                <span className="font-medium text-sm">{account.name}</span>
              </div>

              <div className="text-right">
                <div className="font-semibold text-sm">
                  {formatCurrency(account.balance, account.currency)}
                </div>
                <div
                  className={`text-xs ${
                    account.percentageChange > 0
                      ? "text-green-600"
                      : account.percentageChange < 0
                        ? "text-red-600"
                        : "text-muted-foreground"
                  }`}
                >
                  {formatPercentage(account.percentageChange)}
                </div>
              </div>
            </div>
          ))
        )}
      </CardContent>

      <AddAccountModal
        isOpen={isAddAccountModalOpen}
        onOpenChange={setIsAddAccountModalOpen}
        onAccountAdded={() => {
          // The account list will automatically refresh due to React Query
          // No additional action needed
        }}
      />
    </Card>
  );
}
