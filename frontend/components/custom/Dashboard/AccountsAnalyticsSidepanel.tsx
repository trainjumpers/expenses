"use client";

import { AddAccountModal } from "@/components/custom/Modal/Accounts/AddAccountModal";
import { useAccounts } from "@/components/hooks/useAccounts";
import { useAccountAnalytics } from "@/components/hooks/useAnalytics";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  formatCurrency,
  formatPercentage,
  formatShortCurrency,
  getTransactionColor,
} from "@/lib/utils";
import { Plus } from "lucide-react";
import { useState } from "react";

import { MonthlyAnalyticsCard } from "./MonthlyAnalyticsCard";

interface AccountData {
  id: number;
  name: string;
  balance: number;
  txnBalance: number;
  currency: string;
  percentageChange: number;
  bankType?: string;
  isInvestment: boolean;
  xirr: number;
}

interface AccountsAnalyticsSidepanelProps {
  className?: string;
}

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
      const isInvestment =
        account?.bank_type === "investment" &&
        analytics.current_value !== null &&
        analytics.current_value !== undefined;
      const percentageChange = isInvestment
        ? (-1 * (analytics.percentage_increase ?? 0))
        : calculatePercentageChange(
            analytics.current_balance,
            analytics.balance_one_month_ago
          );

      return {
        id: analytics.account_id,
        name: account?.name || `Account ${analytics.account_id}`,
        currency: account?.currency || "INR",
        balance: isInvestment
          ? Number(analytics.current_value)
          : analytics.current_balance + (account?.balance || 0),
        txnBalance: analytics.current_balance + (account?.balance || 0),
        percentageChange,
        bankType: account?.bank_type || "others",
        isInvestment,
        xirr: analytics.xirr ?? 0,
      };
    }) || [];

  console.log("Accounts Data:", accounts);

  if (isLoading) {
    return (
      <div className={`w-80 flex flex-col h-full ${className}`}>
        {/* Monthly Analytics Card at the top */}
        <div className="shrink-0 mb-4">
          <MonthlyAnalyticsCard />
        </div>

        <Card className="flex-1 flex flex-col">
          <CardHeader className="pb-4 shrink-0">
            <div className="flex items-center justify-between">
              <Skeleton className="h-6 w-16" />
              <Skeleton className="h-8 w-8" />
            </div>
            <Skeleton className="h-4 w-20" />
          </CardHeader>
          <CardContent className="flex-1 overflow-y-auto space-y-1">
            <div>
              <Skeleton className="h-3 w-24 mb-2" />
              {Array.from({ length: 3 }).map((_, index) => (
                <div
                  key={`inv-${index}`}
                  className="flex items-center justify-between py-3 px-2"
                >
                  <div className="flex items-center space-x-3">
                    <Skeleton className="h-4 w-4" />
                    <Skeleton className="h-4 w-28" />
                  </div>
                  <div className="text-right space-y-1">
                    <Skeleton className="h-4 w-16" />
                    <Skeleton className="h-3 w-10" />
                  </div>
                </div>
              ))}
            </div>

            <div>
              <Skeleton className="h-3 w-16 mb-2 mt-2" />
              {Array.from({ length: 3 }).map((_, index) => (
                <div
                  key={`bank-${index}`}
                  className="flex items-center justify-between py-3 px-2"
                >
                  <div className="flex items-center space-x-3">
                    <Skeleton className="h-4 w-4" />
                    <Skeleton className="h-4 w-28" />
                  </div>
                  <div className="text-right space-y-1">
                    <Skeleton className="h-4 w-16" />
                    <Skeleton className="h-3 w-10" />
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className={`w-80 flex flex-col h-full ${className}`}>
      {/* Monthly Analytics Card at the top */}
      <div className="shrink-0 mb-4">
        <MonthlyAnalyticsCard />
      </div>

      <Card className="flex-1 flex flex-col">
        <CardHeader className="pb-4 shrink-0">
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

        <CardContent className="flex-1 overflow-y-auto space-y-1">
          <TooltipProvider>
            {accounts.length === 0 ? (
              <div className="text-center py-8 space-y-4">
                <div className="text-muted-foreground space-y-2">
                  <p className="text-sm font-medium">No accounts yet</p>
                  <p className="text-xs">Add an account to start tracking</p>
                </div>
                <Button
                  onClick={() => setIsAddAccountModalOpen(true)}
                  size="sm"
                  className="w-full"
                >
                  <Plus className="h-3 w-3 mr-2" />
                  Add Account
                </Button>
              </div>
            ) : (
              (() => {
                const investments = accounts.filter(
                  (a) => a.bankType === "investment"
                );
                const banks = accounts.filter(
                  (a) => a.bankType !== "investment"
                );

                return (
                  <div className="space-y-3">
                    {investments.length > 0 && (
                      <div>
                        <div className="text-xs font-medium text-muted-foreground mb-1">
                          Investments
                        </div>
                        <div className="space-y-1">
                          {investments.map((account) => (
                            <div
                              key={account.id}
                              className="flex items-center justify-between py-3 px-2 rounded-md hover:bg-muted/50 cursor-pointer group"
                            >
                              <div className="flex items-center space-x-3">
                                <span className="font-medium text-sm">
                                  {account.name}
                                </span>
                              </div>

                              <div className="text-right">
                                <div className="font-semibold text-sm">
                                  <Tooltip skipProvider>
                                    <TooltipTrigger asChild>
                                      <span>
                                        {formatShortCurrency(
                                          account.balance,
                                          account.currency
                                        )}
                                      </span>
                                    </TooltipTrigger>
                                    <TooltipContent side="top">
                                      {formatCurrency(
                                        account.balance,
                                        account.currency
                                      )}
                                    </TooltipContent>
                                  </Tooltip>
                                  <span className="text-xs text-muted-foreground ml-2">
                                    •{" "}
                                    <Tooltip skipProvider>
                                      <TooltipTrigger asChild>
                                        <span>
                                          {formatShortCurrency(
                                            account.txnBalance,
                                            account.currency
                                          )}
                                        </span>
                                      </TooltipTrigger>
                                      <TooltipContent side="top">
                                        {formatCurrency(
                                          account.txnBalance,
                                          account.currency
                                        )}
                                      </TooltipContent>
                                    </Tooltip>
                                  </span>
                                </div>
                                <div
                                  className={`text-xs ${getTransactionColor(
                                    -1 * account.percentageChange
                                  )}`}
                                >
                                  {formatPercentage(account.percentageChange)}
                                  <span className="mx-1">•</span>
                                  <span className="text-xs font-medium">
                                    XIRR
                                  </span>
                                  <span className="ml-1">
                                    {formatPercentage(account.xirr)}
                                  </span>
                                </div>
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}

                    {banks.length > 0 && (
                      <div>
                        <div className="text-xs font-medium text-muted-foreground mb-1">
                          Banks
                        </div>
                        <div className="space-y-1">
                          {banks.map((account) => (
                            <div
                              key={account.id}
                              className="flex items-center justify-between py-3 px-2 rounded-md hover:bg-muted/50 cursor-pointer group"
                            >
                              <div className="flex items-center space-x-3">
                                <span className="font-medium text-sm">
                                  {account.name}
                                </span>
                              </div>

                              <div className="text-right">
                                <div className="font-semibold text-sm">
                                  <Tooltip skipProvider>
                                    <TooltipTrigger asChild>
                                      <span>
                                        {formatShortCurrency(
                                          account.balance,
                                          account.currency
                                        )}
                                      </span>
                                    </TooltipTrigger>
                                    <TooltipContent side="top">
                                      {formatCurrency(
                                        account.balance,
                                        account.currency
                                      )}
                                    </TooltipContent>
                                  </Tooltip>
                                </div>
                                <div
                                  className={`text-xs ${getTransactionColor(
                                    -1 * account.percentageChange
                                  )}`}
                                >
                                  {formatPercentage(account.percentageChange)}
                                </div>
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                );
              })()
            )}
          </TooltipProvider>
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
    </div>
  );
}
