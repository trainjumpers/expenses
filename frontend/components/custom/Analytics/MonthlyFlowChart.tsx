"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import type { InsightsMonthlyPoint } from "@/lib/models/analytics";
import { formatCurrency } from "@/lib/utils";
import { format } from "date-fns";
import { Bar, BarChart, ReferenceLine, XAxis, YAxis } from "recharts";

interface MonthlyFlowChartProps {
  monthly: InsightsMonthlyPoint[];
}

const chartConfig = {
  income: { label: "Income", color: "hsl(161, 64%, 36%)" },
  expenses: { label: "Expenses", color: "hsl(350, 70%, 48%)" },
};

export function MonthlyFlowChart({ monthly }: MonthlyFlowChartProps) {
  const data = monthly.map((point) => ({
    label: format(new Date(`${point.month}-01T00:00:00`), "MMM yy"),
    income: -point.income,
    expenses: point.expenses,
  }));

  return (
    <Card className="min-w-0 overflow-hidden rounded-none border-x-0 border-t-0 shadow-none">
      <CardHeader className="px-0 pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">
          Household cash flow
        </CardTitle>
        <p className="text-xs text-muted-foreground">
          Credits below zero, debits above.
        </p>
      </CardHeader>
      <CardContent className="px-0">
        {data.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            No household cash flow in this range.
          </p>
        ) : (
          <ChartContainer
            config={chartConfig}
            className="aspect-auto h-64 w-full"
          >
            <BarChart data={data} margin={{ top: 8, right: 4, left: 4 }}>
              <XAxis
                dataKey="label"
                axisLine={false}
                tickLine={false}
                minTickGap={16}
              />
              <YAxis hide />
              <ReferenceLine y={0} stroke="var(--border)" />
              <ChartTooltip
                content={
                  <ChartTooltipContent
                    formatter={(value, name) => [
                      formatCurrency(value as number),
                      name === "income" ? " Income" : " Expenses",
                    ]}
                  />
                }
              />
              <Bar dataKey="income" fill="var(--color-income)" />
              <Bar dataKey="expenses" fill="var(--color-expenses)" />
            </BarChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  );
}
