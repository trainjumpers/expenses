"use client";

import { useNetworthTimeSeries } from "@/components/hooks/useAnalytics";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { formatCurrency } from "@/lib/utils";
import { format } from "date-fns";

interface IncomeExpenseSummaryProps {
  dateRange: {
    from: Date;
    to: Date;
  };
}

export function IncomeExpenseSummary({ dateRange }: IncomeExpenseSummaryProps) {
  const { data: networthData, isLoading } = useNetworthTimeSeries(
    format(dateRange.from, "yyyy-MM-dd"),
    format(dateRange.to, "yyyy-MM-dd")
  );

  if (isLoading) {
    return (
      <Card className="w-full h-full">
        <CardHeader className="pb-3">
          <Skeleton className="h-5 w-32" />
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 gap-3">
            <div className="text-center p-3 bg-green-50 dark:bg-green-950/20 rounded-lg">
              <Skeleton className="h-3 w-16 mx-auto mb-1" />
              <Skeleton className="h-5 w-20 mx-auto" />
            </div>
            <div className="text-center p-3 bg-red-50 dark:bg-red-950/20 rounded-lg">
              <Skeleton className="h-3 w-16 mx-auto mb-1" />
              <Skeleton className="h-5 w-20 mx-auto" />
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="w-full h-full">
      <CardHeader className="pb-3">
        <CardTitle className="text-sm font-medium text-muted-foreground">
          Income & Expenses
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-2 gap-3">
          {/* Income Card */}
          <div className="text-center p-3 bg-green-50 dark:bg-green-950/20 rounded-lg">
            <div className="text-xs text-muted-foreground mb-1">Income</div>
            <div className="text-base font-semibold text-green-600 dark:text-green-400">
              {formatCurrency(networthData?.total_income || 0)}
            </div>
          </div>

          {/* Expenses Card */}
          <div className="text-center p-3 bg-red-50 dark:bg-red-950/20 rounded-lg">
            <div className="text-xs text-muted-foreground mb-1">Expenses</div>
            <div className="text-base font-semibold text-red-600 dark:text-red-400">
              {formatCurrency(Math.abs(networthData?.total_expenses || 0))}
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
