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
import { Bar, BarChart, XAxis, YAxis } from "recharts";

interface MonthlyFlowChartProps {
  monthly: InsightsMonthlyPoint[];
}

const chartConfig = {
  income: { label: "Income", color: "hsl(152, 60%, 40%)" },
  expenses: { label: "Expenses", color: "hsl(350, 70%, 50%)" },
};

export function MonthlyFlowChart({ monthly }: MonthlyFlowChartProps) {
  const data = monthly.map((point) => ({
    ...point,
    label: format(new Date(`${point.month}-01T00:00:00`), "MMM yy"),
  }));

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle className="text-lg font-semibold text-muted-foreground">
          Household Cash Flow
        </CardTitle>
      </CardHeader>
      <CardContent>
        <ChartContainer config={chartConfig} className="h-64 w-full">
          <BarChart data={data}>
            <XAxis
              dataKey="label"
              axisLine={false}
              tickLine={false}
              minTickGap={16}
            />
            <YAxis hide />
            <ChartTooltip
              content={
                <ChartTooltipContent
                  formatter={(value) => formatCurrency(value as number)}
                />
              }
            />
            <Bar
              dataKey="income"
              fill="var(--color-income)"
              radius={[4, 4, 0, 0]}
            />
            <Bar
              dataKey="expenses"
              fill="var(--color-expenses)"
              radius={[4, 4, 0, 0]}
            />
          </BarChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
}
