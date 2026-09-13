"use client";

import {
  AnalyticsSection,
  InlineNote,
  Money,
} from "@/components/custom/Analytics/AnalyticsShared";
import { Button } from "@/components/ui/button";
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import { Skeleton } from "@/components/ui/skeleton";
import type {
  CashBalanceHistoryResponse,
  InsightsSummary,
} from "@/lib/models/analytics";
import {
  formatCurrency,
  formatShortCurrency,
  transformToChartData,
} from "@/lib/utils";
import { Line, LineChart, XAxis, YAxis } from "recharts";

export function NetWorthSnapshot({ summary }: { summary: InsightsSummary }) {
  return (
    <AnalyticsSection
      title="Current net worth"
      description="Point-in-time value of every account. Investment accounts use their latest recorded value."
    >
      <p className="text-2xl font-semibold text-foreground">
        <Money value={summary.net_worth} full />
      </p>
      <dl className="mt-4 grid grid-cols-2 gap-4">
        <div>
          <dt className="text-xs text-muted-foreground">Bank and cash</dt>
          <dd className="mt-0.5 text-sm font-medium">
            <Money value={summary.bank_value} />
          </dd>
        </div>
        <div>
          <dt className="text-xs text-muted-foreground">Investments</dt>
          <dd className="mt-0.5 text-sm font-medium">
            <Money value={summary.investment_value} />
          </dd>
        </div>
      </dl>
    </AnalyticsSection>
  );
}

export function CashBalanceHistory({
  history,
  isLoading,
  isError,
  onRetry,
}: {
  history?: CashBalanceHistoryResponse;
  isLoading: boolean;
  isError: boolean;
  onRetry: () => void;
}) {
  const chartData = history ? transformToChartData(history.time_series) : [];
  const current = chartData[chartData.length - 1]?.value ?? 0;
  const summary = chartData
    .map((point) => `${point.formattedDate}: ${formatCurrency(point.value)}`)
    .join("; ");

  const yDomain = (() => {
    const values = chartData.map((point) => point.value);
    if (values.length === 0) return ["dataMin", "dataMax"] as [string, string];
    const min = Math.min(...values);
    const max = Math.max(...values);
    if (min === max) {
      const buffer = Math.max(Math.abs(min) * 0.05, 1);
      return [min - buffer, max + buffer] as [number, number];
    }
    const padding = (max - min) * 0.05;
    return [min - padding, max + padding] as [number, number];
  })();

  return (
    <AnalyticsSection
      title="Cash balance history"
      description="Balance across bank and cash accounts only. Investment ledgers are excluded, so this is not comparable with current net worth."
    >
      {isError ? (
        <div>
          <InlineNote>Could not load cash balance history.</InlineNote>
          <Button
            variant="outline"
            size="sm"
            className="mt-3"
            onClick={onRetry}
          >
            Try again
          </Button>
        </div>
      ) : isLoading ? (
        <Skeleton className="h-40 w-full" />
      ) : chartData.length === 0 ? (
        <InlineNote>No cash activity in this range.</InlineNote>
      ) : (
        <>
          <p className="text-2xl font-semibold text-foreground">
            <Money value={current} full />
          </p>
          <div className="mt-4">
            <ChartContainer
              config={{
                balance: { label: "Cash balance", color: "var(--chart-1)" },
              }}
              className="aspect-auto h-40 w-full"
              role="img"
              aria-label="Cash balance over time"
            >
              <LineChart data={chartData}>
                <XAxis
                  dataKey="date"
                  axisLine={false}
                  tickLine={false}
                  minTickGap={24}
                />
                <YAxis
                  axisLine={false}
                  tickLine={false}
                  width={56}
                  domain={yDomain}
                  tickFormatter={(value: number) => formatShortCurrency(value)}
                />
                <ChartTooltip
                  content={
                    <ChartTooltipContent
                      formatter={(value) => [
                        formatCurrency(value as number),
                        "Cash balance",
                      ]}
                      labelFormatter={(_, payload) => {
                        const point = payload?.[0]?.payload as
                          { formattedDate?: string } | undefined;
                        return point?.formattedDate ?? "";
                      }}
                    />
                  }
                />
                <Line
                  type="monotone"
                  dataKey="value"
                  stroke="var(--color-balance)"
                  strokeWidth={2}
                  dot={false}
                />
              </LineChart>
            </ChartContainer>
          </div>
          <p className="sr-only">{summary}</p>
        </>
      )}
    </AnalyticsSection>
  );
}
