"use client";

import { useCashBalanceHistory } from "@/components/hooks/useAnalytics";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import { DateRangePicker } from "@/components/ui/date-range-picker";
import { Skeleton } from "@/components/ui/skeleton";
import {
  cn,
  formatCurrency,
  formatShortCurrency,
  transformToChartData,
} from "@/lib/utils";
import { format } from "date-fns";
import { Line, LineChart, XAxis, YAxis } from "recharts";

interface ChartDataPoint {
  date: string;
  value: number;
  formattedDate: string;
}

interface NetWorthProps {
  dateRange: {
    from: Date;
    to: Date;
  };
  onDateRangeChange?: (dateRange: { from: Date; to: Date }) => void;
  showDatePicker?: boolean;
  className?: string;
}

export function NetWorth({
  dateRange,
  onDateRangeChange,
  showDatePicker = true,
  className,
}: NetWorthProps) {
  const { data: history, isLoading } = useCashBalanceHistory(
    format(dateRange.from, "yyyy-MM-dd"),
    format(dateRange.to, "yyyy-MM-dd")
  );

  const chartData = history?.time_series
    ? transformToChartData(history.time_series)
    : [];

  const currentBalance = chartData[chartData.length - 1]?.value ?? 0;
  const initialBalance = history?.initial_balance ?? 0;
  const absoluteChange = currentBalance - initialBalance;

  const chartStartDate = chartData[0]?.formattedDate ?? "";
  const chartEndDate = chartData[chartData.length - 1]?.formattedDate ?? "";

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

  if (isLoading) {
    return (
      <Card className={cn("w-full", className)}>
        <CardHeader>
          <div className="flex items-center justify-between">
            <Skeleton className="h-6 w-28" />
            {showDatePicker ? <Skeleton className="h-6 w-12" /> : null}
          </div>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div>
              <Skeleton className="mb-2 h-10 w-48" />
              <Skeleton className="h-5 w-40" />
            </div>
            <Skeleton className="h-24 w-full" />
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className={cn("w-full", className)}>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg font-semibold text-muted-foreground">
            Cash balance
          </CardTitle>
          {showDatePicker && onDateRangeChange ? (
            <DateRangePicker
              onUpdate={(values) =>
                onDateRangeChange({
                  from: values.range.from || dateRange.from,
                  to: values.range.to || dateRange.to,
                })
              }
              initialDateFrom={format(dateRange.from, "yyyy-MM-dd")}
              initialDateTo={format(dateRange.to, "yyyy-MM-dd")}
              align="start"
              locale="en-GB"
              showCompare={false}
            />
          ) : null}
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-6">
          <div>
            <div className="mb-2 text-3xl font-bold tabular-nums">
              {formatCurrency(currentBalance)}
            </div>
            <div className="text-sm text-muted-foreground">
              {formatShortCurrency(absoluteChange)} across this range. Bank and
              cash accounts only.
            </div>
          </div>

          <div className="h-24">
            <ChartContainer
              config={{
                balance: {
                  label: "Cash balance",
                  color: "var(--chart-1)",
                },
              }}
              className="aspect-auto h-full w-full"
              role="img"
              aria-label="Cash balance over time"
            >
              <LineChart data={chartData}>
                <XAxis
                  dataKey="date"
                  axisLine={false}
                  tickLine={false}
                  tick={false}
                />
                <YAxis hide domain={yDomain} />
                <ChartTooltip
                  content={
                    <ChartTooltipContent
                      formatter={(value) => [
                        formatCurrency(value as number),
                        "Cash balance",
                      ]}
                      labelFormatter={(_, payload) => {
                        const point = payload?.[0]?.payload as
                          ChartDataPoint | undefined;
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

          <div className="flex justify-between text-xs text-muted-foreground">
            <span>{chartStartDate}</span>
            <span>{chartEndDate}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
