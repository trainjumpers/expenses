"use client";

import { useNetworthTimeSeries } from "@/components/hooks/useAnalytics";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import { Skeleton } from "@/components/ui/skeleton";
import { format, subDays } from "date-fns";
import { Line, LineChart, XAxis, YAxis } from "recharts";

interface ChartDataPoint {
  date: string;
  value: number;
  formattedDate: string;
}

const formatCurrency = (amount: number): string => {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "INR",
    minimumFractionDigits: 2,
  }).format(amount);
};

const formatPercentage = (percentage: number): string => {
  const sign = percentage > 0 ? "+" : "";
  return `${sign}${percentage.toFixed(1)}%`;
};

// Transform API data to chart format
const transformToChartData = (
  timeSeries: Array<{ date: string; networth: number }>
): ChartDataPoint[] => {
  return timeSeries.map((point) => ({
    date: format(new Date(point.date), "MMM dd"),
    value: point.networth,
    formattedDate: format(new Date(point.date), "MMM dd, yyyy"),
  }));
};

export function NetWorth() {
  // Calculate date range for the last 30 days
  const endDate = format(new Date(), "yyyy-MM-dd");
  const startDate = format(subDays(new Date(), 29), "yyyy-MM-dd");

  const { data: networthData, isLoading } = useNetworthTimeSeries(
    startDate,
    endDate
  );

  // Transform data for chart
  const chartData = networthData?.time_series
    ? transformToChartData(networthData.time_series)
    : [];

  // Get current and initial values
  const currentNetWorth = chartData[chartData.length - 1]?.value || 0;
  const initialNetWorth = networthData?.initial_balance || 0;

  // Calculate percentage change over the period
  const percentageChange =
    ((currentNetWorth - initialNetWorth) / Math.abs(initialNetWorth)) * 100;

  const absoluteChange = currentNetWorth - initialNetWorth;

  const chartStartDate = chartData[0]?.formattedDate || "";
  const chartEndDate = chartData[chartData.length - 1]?.formattedDate || "";

  if (isLoading) {
    return (
      <Card className="w-full">
        <CardHeader>
          <div className="flex items-center justify-between">
            <Skeleton className="h-6 w-20" />
            <Skeleton className="h-6 w-12" />
          </div>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div>
              <Skeleton className="h-10 w-48 mb-2" />
              <Skeleton className="h-5 w-64" />
            </div>
            <Skeleton className="h-32 w-full" />
            <div className="flex justify-between">
              <Skeleton className="h-4 w-20" />
              <Skeleton className="h-4 w-20" />
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="w-full">
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg font-semibold text-muted-foreground">
            Net Worth
          </CardTitle>
          <span className="text-sm text-muted-foreground bg-muted px-2 py-1 rounded">
            30D
          </span>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-6">
          {/* Net Worth Value and Change */}
          <div>
            <div className="text-3xl font-bold mb-2">
              {formatCurrency(currentNetWorth)}
            </div>
            <div
              className={`text-sm ${percentageChange >= 0 ? "text-green-600 dark:text-green-300" : "text-red-600 dark:text-red-300"}`}
            >
              {formatCurrency(absoluteChange)} (
              {formatPercentage(percentageChange)})
            </div>
          </div>

          {/* Chart */}
          <div className="h-32">
            <ChartContainer
              config={{
                netWorth: {
                  label: "Net Worth",
                  color: "hsl(142, 76%, 36%)",
                },
              }}
              className="h-full w-full"
            >
              <LineChart data={chartData}>
                <XAxis
                  dataKey="date"
                  axisLine={false}
                  tickLine={false}
                  tick={false}
                />
                <YAxis hide />
                <ChartTooltip
                  content={
                    <ChartTooltipContent
                      formatter={(value) => [
                        formatCurrency(value as number),
                        "Net Worth",
                      ]}
                      labelFormatter={(label, payload) => {
                        const data = payload?.[0]?.payload as ChartDataPoint;
                        return data?.formattedDate || label;
                      }}
                    />
                  }
                />
                <Line
                  type="monotone"
                  dataKey="value"
                  stroke="var(--color-netWorth)"
                  strokeWidth={2}
                  dot={false}
                />
              </LineChart>
            </ChartContainer>
          </div>

          {/* Date Range */}
          <div className="flex justify-between text-xs text-muted-foreground">
            <span>{chartStartDate}</span>
            <span>{chartEndDate}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
