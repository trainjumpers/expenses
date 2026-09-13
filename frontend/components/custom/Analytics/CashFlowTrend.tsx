"use client";

import {
  AnalyticsSection,
  InlineNote,
} from "@/components/custom/Analytics/AnalyticsShared";
import {
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import type {
  InsightsCategoryMovement,
  InsightsMonthlyPoint,
  InsightsTrend,
} from "@/lib/models/analytics";
import { formatPercentage, formatShortCurrency } from "@/lib/utils";
import { format } from "date-fns";
import { Bar, BarChart, ReferenceLine, XAxis, YAxis } from "recharts";

const chartConfig = {
  income: { label: "Income", color: "var(--chart-2)" },
  spending: { label: "Spending", color: "var(--chart-5)" },
};

function monthLabel(month: string) {
  return format(new Date(`${month}-01T00:00:00`), "MMM yy");
}

function monthName(month: string) {
  return format(new Date(`${month}-01T00:00:00`), "MMMM yyyy");
}

function changeDirection(change: number) {
  if (change > 0) return "up";
  if (change < 0) return "down";
  return "flat";
}

function TrendExplanation({
  trend,
  movement,
}: {
  trend: InsightsTrend;
  movement: InsightsCategoryMovement[];
}) {
  if (!trend.recent_month) {
    return (
      <InlineNote>
        Not enough complete months in this range to compare spending.
      </InlineNote>
    );
  }

  const hasComparison = trend.prior_month !== "";

  const drivers = hasComparison
    ? movement
        .filter((item) => item.change !== 0)
        .sort((a, b) => Math.abs(b.change) - Math.abs(a.change))
        .slice(0, 2)
    : [];

  const percentage =
    trend.prior_expenses > 0
      ? (trend.change / trend.prior_expenses) * 100
      : null;

  return (
    <div className="space-y-2">
      {hasComparison ? (
        <InlineNote>
          {monthName(trend.recent_month)} spending was{" "}
          <span className="font-medium text-foreground">
            {formatShortCurrency(trend.recent_expenses)}
          </span>
          , {changeDirection(trend.change)}{" "}
          <span className="font-medium text-foreground">
            {formatShortCurrency(Math.abs(trend.change))}
          </span>
          {percentage !== null
            ? ` (${formatPercentage(Math.abs(percentage))})`
            : ""}{" "}
          versus {monthName(trend.prior_month)}.
        </InlineNote>
      ) : (
        <InlineNote>
          {monthName(trend.recent_month)} spending was{" "}
          <span className="font-medium text-foreground">
            {formatShortCurrency(trend.recent_expenses)}
          </span>
          . The range does not include an earlier complete month to compare
          against.
        </InlineNote>
      )}
      {drivers.length > 0 ? (
        <InlineNote>
          Biggest moves:{" "}
          {drivers
            .map(
              (driver) =>
                `${driver.category_name} ${driver.change > 0 ? "up" : "down"} ${formatShortCurrency(Math.abs(driver.change))}`
            )
            .join(", ")}
          .
        </InlineNote>
      ) : null}
      {trend.trailing_three_month_average > 0 ? (
        <InlineNote>
          The three complete months ending {monthName(trend.recent_month)}{" "}
          average {formatShortCurrency(trend.trailing_three_month_average)} a
          month.
        </InlineNote>
      ) : null}
    </div>
  );
}

export function CashFlowTrend({
  monthly,
  trend,
  movement,
}: {
  monthly: InsightsMonthlyPoint[];
  trend: InsightsTrend;
  movement: InsightsCategoryMovement[];
}) {
  const data = monthly.map((point) => ({
    label: monthLabel(point.month),
    income: point.income,
    spending: -point.expenses,
  }));

  const summary = monthly
    .map(
      (point) =>
        `${monthName(point.month)}: income ${formatShortCurrency(point.income)}, spending ${formatShortCurrency(point.expenses)}`
    )
    .join("; ");

  return (
    <AnalyticsSection
      title="Cash flow"
      description="Income sits above the zero line, spending below it. The latest month may be partial."
    >
      <div className="grid gap-6 lg:grid-cols-[1.6fr_1fr] lg:items-start">
        <div className="min-w-0">
          {data.length === 0 ? (
            <InlineNote>No household cash flow in this range.</InlineNote>
          ) : (
            <ChartContainer
              config={chartConfig}
              className="aspect-auto h-64 w-full"
              role="img"
              aria-label="Monthly income and spending"
            >
              <BarChart data={data} margin={{ top: 8, right: 4, left: 4 }}>
                <XAxis
                  dataKey="label"
                  axisLine={false}
                  tickLine={false}
                  minTickGap={16}
                />
                <YAxis
                  axisLine={false}
                  tickLine={false}
                  width={56}
                  tickFormatter={(value: number) => formatShortCurrency(value)}
                />
                <ReferenceLine y={0} stroke="var(--border)" />
                <ChartTooltip
                  content={
                    <ChartTooltipContent
                      formatter={(value, name) => [
                        formatShortCurrency(Math.abs(value as number)),
                        name === "income" ? "Income" : "Spending",
                      ]}
                    />
                  }
                />
                <ChartLegend content={<ChartLegendContent />} />
                <Bar dataKey="income" fill="var(--color-income)" />
                <Bar dataKey="spending" fill="var(--color-spending)" />
              </BarChart>
            </ChartContainer>
          )}
          <p className="sr-only">{summary}</p>
        </div>
        <TrendExplanation trend={trend} movement={movement} />
      </div>
    </AnalyticsSection>
  );
}
