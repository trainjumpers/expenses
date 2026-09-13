"use client";

import {
  AnalyticsSection,
  InlineNote,
  Money,
} from "@/components/custom/Analytics/AnalyticsShared";
import { CashFlowTrend } from "@/components/custom/Analytics/CashFlowTrend";
import { CategoryMovement } from "@/components/custom/Analytics/CategoryMovement";
import { DataHealth } from "@/components/custom/Analytics/DataHealth";
import { InvestmentTable } from "@/components/custom/Analytics/InvestmentTable";
import {
  CashBalanceHistory,
  NetWorthSnapshot,
} from "@/components/custom/Analytics/NetWorthAndCash";
import { SpendingHabits } from "@/components/custom/Analytics/SpendingHabits";
import { TopPayees } from "@/components/custom/Analytics/TopPayees";
import {
  useCashBalanceHistory,
  useInsights,
} from "@/components/hooks/useAnalytics";
import { Button } from "@/components/ui/button";
import { DateRangePicker } from "@/components/ui/date-range-picker";
import { Skeleton } from "@/components/ui/skeleton";
import type { InsightsSummary } from "@/lib/models/analytics";
import { cn, formatPercentage, getSurplusTone } from "@/lib/utils";
import { format } from "date-fns";
import Link from "next/link";
import { useState } from "react";

function defaultRange() {
  const to = new Date();
  return {
    from: new Date(to.getFullYear(), to.getMonth() - 11, 1),
    to,
  };
}

function Headline({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="min-w-0">
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className="mt-0.5 text-2xl font-semibold">{children}</dd>
    </div>
  );
}

function HeadlineStrip({ summary }: { summary: InsightsSummary }) {
  const hasActivity =
    summary.period_income !== 0 || summary.period_expenses !== 0;
  const netLabel = !hasActivity
    ? "Net"
    : summary.period_net < 0
      ? "Deficit"
      : "Surplus";

  return (
    <dl className="grid grid-cols-2 gap-x-6 gap-y-5 border-b pb-6 sm:grid-cols-4">
      <Headline label="Income">
        <Money value={summary.period_income} />
      </Headline>
      <Headline label="Spending">
        <Money value={summary.period_expenses} />
      </Headline>
      <Headline label={netLabel}>
        <span className={cn(getSurplusTone(summary.period_net))}>
          <Money value={summary.period_net} />
        </span>
      </Headline>
      <Headline label="Savings rate">
        {summary.period_income > 0
          ? formatPercentage(summary.savings_rate * 100)
          : "—"}
      </Headline>
    </dl>
  );
}

function LoadingSkeleton() {
  return (
    <div className="space-y-8">
      <div className="grid grid-cols-2 gap-x-6 gap-y-5 border-b pb-6 sm:grid-cols-4">
        {Array.from({ length: 4 }).map((_, index) => (
          <Skeleton key={index} className="h-12 w-full" />
        ))}
      </div>
      <div className="space-y-4 border-t pt-6">
        <Skeleton className="h-4 w-32" />
        <Skeleton className="h-64 w-full" />
      </div>
      <div className="grid gap-8 lg:grid-cols-2">
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
  const { data, isLoading, isError, refetch } = useInsights(startDate, endDate);
  const { data: cashHistory, isLoading: cashLoading } = useCashBalanceHistory(
    startDate,
    endDate
  );

  const hasActivity =
    data !== undefined &&
    (data.monthly.some((point) => point.income !== 0 || point.expenses !== 0) ||
      data.investments.length > 0 ||
      data.summary.net_worth !== 0);

  return (
    <div className="mx-auto w-full max-w-6xl px-4 py-6">
      <div className="mb-6 flex flex-wrap items-start justify-between gap-4">
        <div className="min-w-0">
          <h1 className="text-3xl font-bold text-foreground">Analytics</h1>
          <p className="mt-1 max-w-prose text-pretty text-sm text-muted-foreground">
            Household cash flow across bank and cash accounts, excluding
            transfers and investment ledgers.
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
        <AnalyticsSection title="Analytics unavailable">
          <InlineNote>Could not load analytics insights.</InlineNote>
          <Button
            variant="outline"
            size="sm"
            className="mt-4"
            onClick={() => refetch()}
          >
            Try again
          </Button>
        </AnalyticsSection>
      ) : isLoading || !data ? (
        <LoadingSkeleton />
      ) : !hasActivity ? (
        <AnalyticsSection title="No activity in this range">
          <InlineNote>
            Pick a wider range, or{" "}
            <Link href="/transaction" className="underline">
              import a statement
            </Link>{" "}
            to see cash flow and spending habits.
          </InlineNote>
        </AnalyticsSection>
      ) : (
        <div className="space-y-8">
          <HeadlineStrip summary={data.summary} />
          <CashFlowTrend
            monthly={data.monthly}
            trend={data.trend}
            movement={data.category_movement}
          />
          <div className="grid gap-8 lg:grid-cols-[1.3fr_1fr]">
            <CategoryMovement
              movement={data.category_movement}
              recentMonth={data.trend.recent_month}
              priorMonth={data.trend.prior_month}
            />
            <SpendingHabits
              summary={data.spending_summary}
              weekday={data.weekday_behavior}
            />
          </div>
          <div className="grid gap-8 lg:grid-cols-[1.3fr_1fr]">
            <TopPayees expenses={data.top_expenses} />
            <InvestmentTable investments={data.investments} />
          </div>
          <NetWorthSnapshot summary={data.summary} />
          <CashBalanceHistory history={cashHistory} isLoading={cashLoading} />
          <DataHealth
            summary={data.summary}
            confidence={data.data_confidence}
          />
        </div>
      )}
    </div>
  );
}
