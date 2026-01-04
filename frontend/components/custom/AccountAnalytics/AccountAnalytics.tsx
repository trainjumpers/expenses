import { AddAccountModal } from "@/components/custom/Modal/Accounts/AddAccountModal";
import { useAccounts } from "@/components/hooks/useAccounts";
import { useTransactions } from "@/components/hooks/useTransactions";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { AccountAnalyticsListResponse } from "@/lib/models/analytics";
import { formatCurrency, getTransactionColor } from "@/lib/utils";
import { format } from "date-fns";
import { ChevronRight, Plus, Wallet } from "lucide-react";
import { useTheme } from "next-themes";
import { Fragment, useEffect, useState } from "react";

interface AccountAnalyticsProps {
  data?: AccountAnalyticsListResponse["account_analytics"];
}

interface AccountTransactionsProps {
  accountId: number;
}

// Color palette for different accounts
const accountColors = [
  "bg-purple-500",
  "bg-blue-600",
  "bg-gray-600",
  "bg-blue-400",
  "bg-pink-400",
  "bg-green-500",
  "bg-orange-500",
  "bg-indigo-500",
  "bg-teal-500",
  "bg-red-500",
];

function AccountTransactions({ accountId }: AccountTransactionsProps) {
  const { theme } = useTheme();
  const { data, isLoading, error } = useTransactions({
    account_id: accountId,
    page: 1,
    page_size: 5,
    sort_by: "date",
    sort_order: "desc",
  });

  if (isLoading) {
    return (
      <TableRow>
        <TableCell colSpan={4} className="bg-muted/40">
          <div className="px-4 py-3 text-sm text-muted-foreground">
            Loading latest transactions...
          </div>
        </TableCell>
      </TableRow>
    );
  }

  if (error) {
    return (
      <TableRow>
        <TableCell colSpan={4} className="bg-muted/40">
          <div className="px-4 py-3 text-sm text-destructive">
            Failed to load transactions.
          </div>
        </TableCell>
      </TableRow>
    );
  }

  const transactions = data?.transactions ?? [];

  if (transactions.length === 0) {
    return (
      <TableRow>
        <TableCell colSpan={4} className="bg-muted/40">
          <div className="px-4 py-3 text-sm text-muted-foreground">
            No recent transactions for this account.
          </div>
        </TableCell>
      </TableRow>
    );
  }

  return (
    <TableRow>
      <TableCell colSpan={4} className="bg-muted/40">
        <div className="px-4 py-3">
          <div className="text-xs uppercase tracking-wide text-muted-foreground">
            Latest 5 transactions
          </div>
          <div className="mt-3 space-y-2">
            {transactions.map((transaction) => (
              <div
                key={transaction.id}
                className="flex items-center justify-between gap-3 rounded-md border border-border/60 bg-background px-3 py-2"
              >
                <div className="min-w-0">
                  <div className="text-sm font-medium text-foreground">
                    {transaction.name}
                  </div>
                  <div className="text-xs text-muted-foreground">
                    {format(new Date(transaction.date), "MMM d, yyyy")}
                  </div>
                </div>
                <div
                  className={`text-sm font-semibold ${getTransactionColor(
                    transaction.amount,
                    theme
                  )}`}
                >
                  {formatCurrency(Math.abs(transaction.amount))}
                </div>
              </div>
            ))}
          </div>
        </div>
      </TableCell>
    </TableRow>
  );
}

export function AccountAnalytics({ data }: AccountAnalyticsProps) {
  const [expandedAccounts, setExpandedAccounts] = useState<Set<number>>(
    new Set()
  );
  const [isAddAccountModalOpen, setIsAddAccountModalOpen] = useState(false);
  const { data: accountsData } = useAccounts();
  const [selectedAccountIds, setSelectedAccountIds] = useState<number[]>([]);
  const [draftSelectedIds, setDraftSelectedIds] = useState<number[]>([]);

  useEffect(() => {
    setDraftSelectedIds(selectedAccountIds);
  }, [selectedAccountIds]);

  if (!data || data.length === 0) {
    return (
      <>
        <Card className="h-full">
          <CardHeader>
            <CardTitle className="flex items-center justify-between">
              <span>Account Analytics</span>
              <Wallet className="h-5 w-5 text-muted-foreground" />
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col items-center justify-center py-12 space-y-6">
              <div className="rounded-full bg-muted p-6">
                <Wallet className="h-12 w-12 text-muted-foreground" />
              </div>
              <div className="text-center space-y-2">
                <h3 className="text-lg font-semibold">No accounts yet</h3>
                <p className="text-sm text-muted-foreground max-w-sm">
                  Start tracking your finances by adding your first account. You
                  can add bank accounts, credit cards, investments, and more.
                </p>
              </div>
              <Button
                onClick={() => setIsAddAccountModalOpen(true)}
                className="flex items-center gap-2"
              >
                <Plus className="h-4 w-4" />
                Add Your First Account
              </Button>
            </div>
          </CardContent>
        </Card>

        <AddAccountModal
          isOpen={isAddAccountModalOpen}
          onOpenChange={setIsAddAccountModalOpen}
          onAccountAdded={() => {
            // The account list will automatically refresh due to React Query
            setIsAddAccountModalOpen(false);
          }}
        />
      </>
    );
  }

  const hasAccountList = !!accountsData && accountsData.length > 0;
  const showFilter = hasAccountList;
  const allAccountIds = hasAccountList
    ? accountsData.map((account) => account.id)
    : [];
  const hasAllSelectedApplied =
    selectedAccountIds.length > 0 &&
    allAccountIds.every((accountId) => selectedAccountIds.includes(accountId));
  const hasAllSelectedDraft =
    draftSelectedIds.length > 0 &&
    allAccountIds.every((accountId) => draftSelectedIds.includes(accountId));
  const selectedAccountCount = selectedAccountIds.length;
  const triggerLabel =
    selectedAccountCount === 0 || hasAllSelectedApplied
      ? "All accounts"
      : `${selectedAccountCount} selected`;
  const isDirty =
    selectedAccountIds.length !== draftSelectedIds.length ||
    selectedAccountIds.some(
      (accountId) => !draftSelectedIds.includes(accountId)
    );

  const toggleAccountSelection = (accountId: number, checked: boolean) => {
    if (checked) {
      if (draftSelectedIds.includes(accountId)) {
        return;
      }
      setDraftSelectedIds([...draftSelectedIds, accountId]);
      return;
    }

    setDraftSelectedIds(draftSelectedIds.filter((id) => id !== accountId));
  };

  const toggleSelectAll = () => {
    if (hasAllSelectedDraft) {
      setDraftSelectedIds([]);
      return;
    }

    setDraftSelectedIds(allAccountIds);
  };

  const applyAccountFilter = () => {
    setSelectedAccountIds(draftSelectedIds);
  };

  const filteredData = selectedAccountIds.length
    ? data.filter((account) => selectedAccountIds.includes(account.account_id))
    : data;

  // Calculate percentages and prepare data with account names and initial balances
  const accountsWithBalances = filteredData.map((account, index) => {
    // Find the account info from accounts data
    const accountInfo = accountsData?.find(
      (acc) => acc.id === account.account_id
    );
    const accountName = accountInfo?.name || `Account ${account.account_id}`;
    const initialBalance = accountInfo?.balance || 0;

    // Calculate the actual balance including initial balance
    const actualBalance = account.current_balance + initialBalance;
    const absoluteBalance = Math.abs(actualBalance);

    return {
      ...account,
      accountName,
      initialBalance,
      actualBalance,
      absoluteBalance,
      color: accountColors[index % accountColors.length],
    };
  });

  // Calculate total balance from the actual balances
  const totalBalance = accountsWithBalances.reduce(
    (sum, account) => sum + account.absoluteBalance,
    0
  );

  // Calculate percentages and sort
  const accountsWithPercentages = accountsWithBalances
    .map((account) => ({
      ...account,
      percentage:
        totalBalance > 0 ? (account.absoluteBalance / totalBalance) * 100 : 0,
    }))
    .sort((a, b) => b.percentage - a.percentage); // Sort by percentage descending

  const toggleAccountExpansion = (accountId: number) => {
    const newExpanded = new Set(expandedAccounts);
    if (newExpanded.has(accountId)) {
      newExpanded.delete(accountId);
    } else {
      newExpanded.add(accountId);
    }
    setExpandedAccounts(newExpanded);
  };

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center justify-between gap-3">
            <div className="flex items-center gap-2">
              <span>Accounts</span>
              <span className="text-muted-foreground">•</span>
              <span>{formatCurrency(totalBalance)}</span>
            </div>
            <div className="flex items-center gap-2">
              {showFilter && (
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="outline" size="sm" className="h-8">
                      {triggerLabel}
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" className="w-56">
                    <div className="flex items-center justify-between px-2 py-1.5 text-xs text-muted-foreground">
                      <span>Accounts</span>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-6 px-2"
                        onClick={toggleSelectAll}
                      >
                        {hasAllSelectedDraft ? "Deselect all" : "Select all"}
                      </Button>
                    </div>
                    <DropdownMenuSeparator />
                    {accountsData?.map((account) => (
                      <DropdownMenuCheckboxItem
                        key={account.id}
                        checked={draftSelectedIds.includes(account.id)}
                        onCheckedChange={(checked) =>
                          toggleAccountSelection(account.id, Boolean(checked))
                        }
                        onSelect={(event) => event.preventDefault()}
                      >
                        {account.name}
                      </DropdownMenuCheckboxItem>
                    ))}
                    <DropdownMenuSeparator />
                    <div className="flex justify-end px-2 py-2">
                      <Button
                        size="sm"
                        onClick={applyAccountFilter}
                        disabled={!isDirty}
                      >
                        Apply
                      </Button>
                    </div>
                  </DropdownMenuContent>
                </DropdownMenu>
              )}
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setIsAddAccountModalOpen(true)}
                className="h-8 w-8 p-0"
              >
                <Plus className="h-4 w-4" />
              </Button>
            </div>
          </CardTitle>
        </CardHeader>

        <CardContent className="space-y-6">
          {/* Horizontal Progress Bar */}
          <div className="space-y-4">
            <div className="h-4 rounded-full overflow-hidden">
              {accountsWithPercentages.map((account) => (
                <div
                  key={account.account_id}
                  className={`h-full ${account.color} inline-block`}
                  style={{ width: `${account.percentage}%` }}
                />
              ))}
            </div>

            {/* Legend */}
            <div className="flex flex-wrap gap-4 text-sm">
              {accountsWithPercentages.map((account) => (
                <div
                  key={account.account_id}
                  className="flex items-center gap-2"
                >
                  <div className={`w-3 h-3 rounded-full ${account.color}`} />
                  <span className="text-muted-foreground">
                    {account.accountName}:
                  </span>
                  <span className="font-medium">
                    {account.percentage.toFixed(1)}%
                  </span>
                </div>
              ))}
            </div>
          </div>

          {/* Detailed Table */}
          <div className="border rounded-lg">
            <Table>
              <TableHeader>
                <TableRow className="border-b">
                  <TableHead className="w-12"></TableHead>
                  <TableHead className="text-left text-sm font-medium text-muted-foreground">
                    NAME
                  </TableHead>
                  <TableHead className="text-left text-sm font-medium text-muted-foreground">
                    WEIGHT
                  </TableHead>
                  <TableHead className="text-right text-sm font-medium text-muted-foreground">
                    VALUE
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {accountsWithPercentages.map((account) => {
                  const isExpanded = expandedAccounts.has(account.account_id);

                  return (
                    <Fragment key={account.account_id}>
                      <TableRow className="border-b">
                        <TableCell className="w-12">
                          <button
                            onClick={() =>
                              toggleAccountExpansion(account.account_id)
                            }
                            className="p-1 hover:bg-muted rounded transition-colors"
                          >
                            <ChevronRight
                              className={`h-4 w-4 transition-transform ${
                                isExpanded ? "rotate-90" : ""
                              }`}
                            />
                          </button>
                        </TableCell>
                        <TableCell>
                          <span className="font-medium">
                            {account.accountName}
                          </span>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <div className="w-16 h-2 flex">
                              {Array.from({ length: 5 }).map((_, i) => (
                                <div
                                  key={i}
                                  className={`flex-1 h-full ${
                                    i < Math.floor(account.percentage / 20)
                                      ? account.color
                                      : "bg-gray-200"
                                  }`}
                                  style={{
                                    marginRight: i < 4 ? "1px" : "0",
                                  }}
                                />
                              ))}
                            </div>
                            <span className="text-sm">
                              {account.percentage.toFixed(2)}%
                            </span>
                          </div>
                        </TableCell>
                        <TableCell className="text-right">
                          <span className="font-medium">
                            {formatCurrency(account.absoluteBalance)}
                          </span>
                        </TableCell>
                      </TableRow>
                      {isExpanded && (
                        <AccountTransactions accountId={account.account_id} />
                      )}
                    </Fragment>
                  );
                })}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      <AddAccountModal
        isOpen={isAddAccountModalOpen}
        onOpenChange={setIsAddAccountModalOpen}
        onAccountAdded={() => {
          // The account list will automatically refresh due to React Query
          setIsAddAccountModalOpen(false);
        }}
      />
    </>
  );
}
