"use client";

import { CategoryBreakdown } from "@/components/custom/Analytics/CategoryBreakdown";
import { InvestmentTable } from "@/components/custom/Analytics/InvestmentTable";
import { MonthlyFlowChart } from "@/components/custom/Analytics/MonthlyFlowChart";
import { NetWorth } from "@/components/custom/Dashboard/NetWorth";
import { useInsights } from "@/components/hooks/useAnalytics";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { DateRangePicker } from "@/components/ui/date-range-picker";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type {
  InsightsMonthlyPoint,
  InsightsTopExpense,
} from "@/lib/models/analytics";
import {
  cn,
  formatCurrency,
  formatPercentage,
  getTransactionColor,
} from "@/lib/utils";
import { format } from "date-fns";
import Link from "next/link";
import { useState } from "react";

interface KpiCardProps {
  label: string;
  value: string;
  valueClassName?: string;
}

function KpiCard({ label, value, valueClassName }: KpiCardProps) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">
          {label}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className={cn("text-2xl font-bold tabular-nums", valueClassName)}>
          {value}
        </div>
      </CardContent>
    </Card>
  );
}

function TopExpensesTable({ expenses }: { expenses: InsightsTopExpense[] }) {
  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle className="text-lg font-semibold text-muted-foreground">
          Top Expenses
        </CardTitle>
      </CardHeader>
      <CardContent>
        {expenses.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            No household expenses in this range.
          </p>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Payee</TableHead>
                <TableHead className="text-right">Count</TableHead>
                <TableHead className="text-right">Amount</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {expenses.map((expense) => (
                <TableRow key={expense.name}>
                  <TableCell>
                    <Link
                      href={`/transaction?search=${encodeURIComponent(expense.name)}`}
                      className="font-medium hover:underline"
                    >
                      {expense.name}
                    </Link>
                  </TableCell>
                  <TableCell className="text-right tabular-nums">
                    {expense.count}
                  </TableCell>
                  <TableCell
                    className={cn(
                      "text-right tabular-nums",
                      getTransactionColor(expense.amount)
                    )}
                  >
                    {formatCurrency(expense.amount)}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  );
}

function MonthlyDatasetTable({ monthly }: { monthly: InsightsMonthlyPoint[] }) {
  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle className="text-lg font-semibold text-muted-foreground">
          Monthly Dataset
        </CardTitle>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Month</TableHead>
              <TableHead className="text-right">Income</TableHead>
              <TableHead className="text-right">Expenses</TableHead>
              <TableHead className="text-right">Net</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {monthly.map((point) => (
              <TableRow key={point.month}>
                <TableCell>
                  {format(new Date(`${point.month}-01T00:00:00`), "MMM yyyy")}
                </TableCell>
                <TableCell className="text-right tabular-nums text-emerald-600 dark:text-emerald-400">
                  {formatCurrency(point.income)}
                </TableCell>
                <TableCell className="text-right tabular-nums text-rose-600 dark:text-rose-400">
                  {formatCurrency(point.expenses)}
                </TableCell>
                <TableCell
                  className={cn(
                    "text-right tabular-nums",
                    point.net < 0
                      ? "text-rose-600 dark:text-rose-400"
                      : "text-emerald-600 dark:text-emerald-400"
                  )}
                >
                  {formatCurrency(point.net)}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}

function LoadingSkeleton() {
  return (
    <div className="space-y-8">
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-3 xl:grid-cols-6">
        {Array.from({ length: 6 }).map((_, index) => (
          <Skeleton key={index} className="h-24 w-full" />
        ))}
      </div>
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <Skeleton className="h-72 w-full" />
        <Skeleton className="h-72 w-full" />
        <Skeleton className="h-72 w-full" />
        <Skeleton className="h-72 w-full" />
      </div>
    </div>
  );
}

export function AnalyticsView() {
  const [dateRange, setDateRange] = useState(() => {
    const now = new Date();
    return {
      from: new Date(now.getFullYear(), now.getMonth(), 1),
      to: now,
    };
  });

  const startDate = format(dateRange.from, "yyyy-MM-dd");
  const endDate = format(dateRange.to, "yyyy-MM-dd");
  const { data, isLoading, isError } = useInsights(startDate, endDate);

  return (
    <div className="px-4 py-6">
      <div className="mb-6 flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-foreground">Analytics</h1>
          <p className="text-sm text-muted-foreground">
            Household figures exclude Transfers and investment ledgers.
          </p>
        </div>
        <DateRangePicker
          onUpdate={(values) =>
            setDateRange((prev) => ({
              from: values.range.from || prev.from,
              to: values.range.to || prev.to,
            }))
          }
          initialDateFrom={startDate}
          initialDateTo={endDate}
          align="end"
          locale="en-GB"
          showCompare={false}
        />
      </div>

      {isError ? (
        <p className="text-sm text-muted-foreground">
          Could not load analytics insights.
        </p>
      ) : isLoading || !data ? (
        <LoadingSkeleton />
      ) : (
        <div className="space-y-8">
          <div className="grid grid-cols-2 gap-4 lg:grid-cols-3 xl:grid-cols-6">
            <KpiCard
              label="Net Worth"
              value={formatCurrency(data.summary.net_worth)}
            />
            <KpiCard
              label="Investments"
              value={formatCurrency(data.summary.investment_value)}
            />
            <KpiCard
              label="Banks"
              value={formatCurrency(data.summary.bank_value)}
            />
            <KpiCard
              label="Income"
              value={formatCurrency(data.summary.period_income)}
              valueClassName={getTransactionColor(-data.summary.period_income)}
            />
            <KpiCard
              label="Expenses"
              value={formatCurrency(data.summary.period_expenses)}
              valueClassName={getTransactionColor(data.summary.period_expenses)}
            />
            <KpiCard
              label="Savings Rate"
              value={formatPercentage(data.summary.savings_rate * 100)}
            />
          </div>

          <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
            <NetWorth dateRange={dateRange} onDateRangeChange={setDateRange} />
            <MonthlyFlowChart monthly={data.monthly} />
          </div>

          <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
            <CategoryBreakdown categories={data.categories} />
            <InvestmentTable investments={data.investments} />
          </div>

          <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
            <TopExpensesTable expenses={data.top_expenses} />
            <MonthlyDatasetTable monthly={data.monthly} />
          </div>
        </div>
      )}
    </div>
  );
}
