"use client";

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
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import type {
  AnalyticsInsightsResponse,
  InsightsMonthlyPoint,
  InsightsTopExpense,
} from "@/lib/models/analytics";
import {
  cn,
  formatCurrency,
  formatPercentage,
  formatShortCurrency,
  getTransactionColor,
} from "@/lib/utils";
import { format } from "date-fns";
import Link from "next/link";
import { type ReactNode, useState } from "react";

function defaultRange() {
  const to = new Date();
  return {
    from: new Date(to.getFullYear(), to.getMonth() - 11, 1),
    to,
  };
}

function FlowStat({ label, amount }: { label: string; amount: number }) {
  const formatted = formatCurrency(amount);
  return (
    <div className="min-w-0">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div
        className={cn(
          "truncate text-xl font-semibold tabular-nums",
          getTransactionColor(amount)
        )}
        title={formatted}
      >
        {formatShortCurrency(amount)}
      </div>
    </div>
  );
}

function TopPayees({ expenses }: { expenses: InsightsTopExpense[] }) {
  const rows = expenses.slice(0, 8);

  return (
    <Card className="min-w-0 overflow-hidden rounded-none border-x-0 border-t-0 shadow-none">
      <CardHeader className="px-0 pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">
          Top payees
        </CardTitle>
      </CardHeader>
      <CardContent className="px-0">
        {rows.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            No household expenses in this range.
          </p>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Payee</TableHead>
                <TableHead className="text-right">N</TableHead>
                <TableHead className="text-right">Amount</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {rows.map((expense) => (
                <TableRow key={expense.name}>
                  <TableCell className="max-w-[12rem]">
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Link
                          href={`/transaction?search=${encodeURIComponent(expense.name)}`}
                          className="block truncate font-medium hover:underline"
                        >
                          {expense.name}
                        </Link>
                      </TooltipTrigger>
                      <TooltipContent className="max-w-xs">
                        {expense.name}
                      </TooltipContent>
                    </Tooltip>
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
                    {formatShortCurrency(expense.amount)}
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

function completeMonths(monthly: InsightsMonthlyPoint[], endDate: Date) {
  const currentMonth = format(endDate, "yyyy-MM");
  return monthly.filter(
    (point) =>
      point.month < currentMonth && (point.income !== 0 || point.expenses !== 0)
  );
}

function RunRate({
  data,
  endDate,
}: {
  data: AnalyticsInsightsResponse;
  endDate: Date;
}) {
  const history = completeMonths(data.monthly, endDate);
  const trailing = history.slice(-3);
  const prior = history.at(-2);
  const latest = history.at(-1);

  const forecastIncome =
    trailing.reduce((sum, point) => sum + point.income, 0) /
    Math.max(trailing.length, 1);
  const forecastExpenses =
    trailing.reduce((sum, point) => sum + point.expenses, 0) /
    Math.max(trailing.length, 1);
  const mom =
    latest && prior && prior.expenses !== 0
      ? (latest.expenses - prior.expenses) / prior.expenses
      : null;

  const lines: { key: string; body: ReactNode }[] = [];

  if (trailing.length > 0) {
    lines.push({
      key: "run-rate",
      body: (
        <>
          Next month run-rate{" "}
          <span className={getTransactionColor(-forecastIncome)}>
            {formatShortCurrency(-forecastIncome)}
          </span>{" "}
          in,{" "}
          <span className={getTransactionColor(forecastExpenses)}>
            {formatShortCurrency(forecastExpenses)}
          </span>{" "}
          out, net{" "}
          <span
            className={getTransactionColor(forecastExpenses - forecastIncome)}
          >
            {formatShortCurrency(forecastExpenses - forecastIncome)}
          </span>
          .
        </>
      ),
    });
  }

  if (mom !== null && latest && prior) {
    lines.push({
      key: "mom",
      body: (
        <>
          {format(new Date(`${latest.month}-01T00:00:00`), "MMM yyyy")} spend{" "}
          {formatPercentage(mom * 100)} vs{" "}
          {format(new Date(`${prior.month}-01T00:00:00`), "MMM")}.
        </>
      ),
    });
  }

  if (data.summary.uncategorized_count > 0) {
    lines.push({
      key: "uncat",
      body: (
        <Link
          href="/transaction?uncategorized=true"
          className="hover:underline"
        >
          {data.summary.uncategorized_count} uncategorized transactions
        </Link>
      ),
    });
  }

  if (data.summary.realized_interest !== 0) {
    lines.push({
      key: "interest",
      body: (
        <>
          Realized interest{" "}
          <span
            className={getTransactionColor(-data.summary.realized_interest)}
          >
            {formatShortCurrency(-data.summary.realized_interest)}
          </span>
        </>
      ),
    });
  }

  if (lines.length === 0) {
    return null;
  }

  return (
    <Card className="min-w-0 overflow-hidden rounded-none border-x-0 border-t-0 shadow-none">
      <CardHeader className="px-0 pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">
          Run-rate
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-2 px-0 text-sm">
        {lines.map((line) => (
          <p key={line.key} className="text-pretty">
            {line.body}
          </p>
        ))}
      </CardContent>
    </Card>
  );
}

function LoadingSkeleton() {
  return (
    <div className="space-y-9">
      <div className="grid grid-cols-2 gap-4 border-b py-4 sm:grid-cols-4">
        {Array.from({ length: 4 }).map((_, index) => (
          <Skeleton key={index} className="h-12 w-full" />
        ))}
      </div>
      <div className="grid grid-cols-1 gap-9 lg:grid-cols-2">
        <Skeleton className="h-64 w-full" />
        <Skeleton className="h-64 w-full" />
      </div>
    </div>
  );
}

export function AnalyticsView() {
  const [dateRange, setDateRange] = useState(defaultRange);

  const startDate = format(dateRange.from, "yyyy-MM-dd");
  const endDate = format(dateRange.to, "yyyy-MM-dd");
  const { data, isLoading, isError } = useInsights(startDate, endDate);

  const income = data ? -data.summary.period_income : 0;
  const expenses = data ? data.summary.period_expenses : 0;
  const net = data ? -data.summary.period_net : 0;

  return (
    <div className="px-4 py-6">
      <div className="mb-6 flex flex-wrap items-start justify-between gap-4">
        <div className="min-w-0">
          <h1 className="text-3xl font-bold text-foreground">Analytics</h1>
          <p className="text-sm text-muted-foreground">
            Household figures exclude Transfers and investment ledgers. Credits
            are negative, debits are positive.
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
        <div className="space-y-9">
          <div className="grid grid-cols-2 gap-4 border-b py-4 sm:grid-cols-4">
            <FlowStat label="Income" amount={income} />
            <FlowStat label="Expenses" amount={expenses} />
            <FlowStat label="Net" amount={net} />
            <div className="min-w-0">
              <div className="text-xs text-muted-foreground">Savings rate</div>
              <div
                className="truncate text-xl font-semibold tabular-nums"
                title={formatPercentage(data.summary.savings_rate * 100)}
              >
                {formatPercentage(data.summary.savings_rate * 100)}
              </div>
            </div>
          </div>

          <div className="grid grid-cols-1 gap-9 lg:grid-cols-2">
            <div className="min-w-0">
              <NetWorth
                dateRange={dateRange}
                showDatePicker={false}
                className="rounded-none border-x-0 border-t-0 py-4 shadow-none [&_[data-slot=card-content]]:px-0 [&_[data-slot=card-header]]:px-0"
              />
            </div>
            <MonthlyFlowChart monthly={data.monthly} />
          </div>

          <div className="grid grid-cols-1 gap-9 lg:grid-cols-2">
            <InvestmentTable investments={data.investments} />
            <TopPayees expenses={data.top_expenses} />
          </div>

          <RunRate data={data} endDate={dateRange.to} />
        </div>
      )}
    </div>
  );
}
