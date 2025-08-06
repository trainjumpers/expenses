"use client";

import { useMonthlyAnalytics } from "@/components/hooks/useAnalytics";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { formatCurrency } from "@/lib/utils";
import {
  ArrowRightLeftIcon,
  CalendarIcon,
  TrendingDownIcon,
  TrendingUpIcon,
} from "lucide-react";
import { useRef, useState } from "react";

interface MonthlyData {
  startDate: string;
  endDate: string;
  label: string;
  description: string;
  data: {
    total_income: number;
    total_expenses: number;
    total_amount: number;
  } | null;
  isLoading: boolean;
}

// Helper function to get date ranges for different time periods
function getDateRanges() {
  const now = new Date();
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());

  // Format date as YYYY-MM-DD
  const formatDate = (date: Date) => date.toISOString().split("T")[0];

  return [
    {
      label: "This Month",
      description: "Current month",
      startDate: formatDate(new Date(now.getFullYear(), now.getMonth(), 1)),
      endDate: formatDate(today),
    },
    {
      label: "Last Month",
      description: "Previous month",
      startDate: formatDate(new Date(now.getFullYear(), now.getMonth() - 1, 1)),
      endDate: formatDate(new Date(now.getFullYear(), now.getMonth(), 0)), // Last day of previous month
    },
    {
      label: "3 Months",
      description: "Last 3 months",
      startDate: formatDate(new Date(now.getFullYear(), now.getMonth() - 3, 1)),
      endDate: formatDate(today),
    },
    {
      label: "6 Months",
      description: "Last 6 months",
      startDate: formatDate(new Date(now.getFullYear(), now.getMonth() - 6, 1)),
      endDate: formatDate(today),
    },
    {
      label: "1 Year",
      description: "Last 12 months",
      startDate: formatDate(
        new Date(now.getFullYear(), now.getMonth() - 12, 1)
      ),
      endDate: formatDate(today),
    },
    {
      label: "All Time",
      description: "All transactions",
      startDate: "2000-01-01", // Far past date to include all transactions
      endDate: formatDate(today),
    },
  ];
}

export function MonthlyAnalyticsCard() {
  const [currentIndex, setCurrentIndex] = useState(0); // Start with "This Month" (index 0)
  const scrollRef = useRef<HTMLDivElement>(null);

  // Get date ranges for all time periods
  const dateRanges = getDateRanges();

  // Fetch data for all monthly options using individual hooks
  const monthlyAnalytics = [
    useMonthlyAnalytics(dateRanges[0].startDate, dateRanges[0].endDate),
    useMonthlyAnalytics(dateRanges[1].startDate, dateRanges[1].endDate),
    useMonthlyAnalytics(dateRanges[2].startDate, dateRanges[2].endDate),
    useMonthlyAnalytics(dateRanges[3].startDate, dateRanges[3].endDate),
    useMonthlyAnalytics(dateRanges[4].startDate, dateRanges[4].endDate),
    useMonthlyAnalytics(dateRanges[5].startDate, dateRanges[5].endDate),
  ];

  // Combine data with date ranges
  const monthlyData: MonthlyData[] = dateRanges.map((dateRange, index) => ({
    startDate: dateRange.startDate,
    endDate: dateRange.endDate,
    label: dateRange.label,
    description: dateRange.description,
    data: monthlyAnalytics[index].data || null,
    isLoading: monthlyAnalytics[index].isLoading,
  }));

  const currentData = monthlyData[currentIndex];

  const handleScroll = () => {
    if (!scrollRef.current) return;

    const { scrollLeft, clientWidth } = scrollRef.current;
    const itemWidth = clientWidth;
    const newIndex = Math.round(scrollLeft / itemWidth);

    if (
      newIndex !== currentIndex &&
      newIndex >= 0 &&
      newIndex < dateRanges.length
    ) {
      setCurrentIndex(newIndex);
    }
  };

  const scrollToIndex = (index: number) => {
    if (!scrollRef.current) return;

    const itemWidth = scrollRef.current.clientWidth;
    scrollRef.current.scrollTo({
      left: index * itemWidth,
      behavior: "smooth",
    });
    setCurrentIndex(index);
  };

  if (currentData?.isLoading) {
    return (
      <Card className="w-full">
        <CardHeader className="pb-3">
          <CardTitle className="text-sm font-medium flex items-center gap-2">
            <CalendarIcon className="h-4 w-4" />
            Monthly Analytics
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <Skeleton className="h-4 w-16" />
          <div className="space-y-2">
            <Skeleton className="h-3 w-full" />
            <Skeleton className="h-3 w-full" />
            <Skeleton className="h-3 w-3/4" />
          </div>
          <div className="flex justify-center gap-1">
            {[...Array(6)].map((_, i) => (
              <Skeleton key={i} className="h-1.5 w-1.5 rounded-full" />
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="w-full">
      <CardHeader className="pb-3">
        <CardTitle className="text-sm font-medium flex items-center gap-2">
          <CalendarIcon className="h-4 w-4" />
          Monthly Analytics
        </CardTitle>
      </CardHeader>

      <CardContent className="space-y-4">
        {/* Scrollable Content */}
        <div
          ref={scrollRef}
          className="flex overflow-x-auto scrollbar-hide snap-x snap-mandatory"
          onScroll={handleScroll}
          style={{ scrollbarWidth: "none", msOverflowStyle: "none" }}
        >
          {monthlyData.map((item) => (
            <div
              key={`${item.startDate}-${item.endDate}`}
              className="min-w-full snap-center space-y-3"
            >
              {/* Period Label */}
              <div className="text-center">
                <span className="text-lg font-semibold text-foreground">
                  {item.label}
                </span>
                <p className="text-xs text-muted-foreground">
                  {item.description}
                </p>
              </div>

              {/* Analytics Data */}
              {item.data ? (
                <div className="space-y-2">
                  {/* Total Income */}
                  <div className="flex items-center justify-between p-2 bg-green-50 dark:bg-green-950/20 rounded-lg">
                    <div className="flex items-center gap-2">
                      <TrendingUpIcon className="h-3 w-3 text-green-600" />
                      <span className="text-xs font-medium text-green-700 dark:text-green-400">
                        Income
                      </span>
                    </div>
                    <span className="text-xs font-semibold text-green-800 dark:text-green-300">
                      {formatCurrency(item.data.total_income)}
                    </span>
                  </div>

                  {/* Total Expenses */}
                  <div className="flex items-center justify-between p-2 bg-red-50 dark:bg-red-950/20 rounded-lg">
                    <div className="flex items-center gap-2">
                      <TrendingDownIcon className="h-3 w-3 text-red-600" />
                      <span className="text-xs font-medium text-red-700 dark:text-red-400">
                        Expenses
                      </span>
                    </div>
                    <span className="text-xs font-semibold text-red-800 dark:text-red-300">
                      {formatCurrency(item.data.total_expenses)}
                    </span>
                  </div>

                  {/* Net Amount */}
                  <div className="flex items-center justify-between p-2 bg-blue-50 dark:bg-blue-950/20 rounded-lg">
                    <div className="flex items-center gap-2">
                      <ArrowRightLeftIcon className="h-3 w-3 text-blue-600" />
                      <span className="text-xs font-medium text-blue-700 dark:text-blue-400">
                        Net Flow
                      </span>
                    </div>
                    <span className="text-xs font-semibold text-blue-800 dark:text-blue-300">
                      {formatCurrency(item.data.total_amount)}
                    </span>
                  </div>
                </div>
              ) : (
                <div className="text-center py-4">
                  <p className="text-xs text-muted-foreground">
                    No data available
                  </p>
                </div>
              )}
            </div>
          ))}
        </div>

        {/* Dots Indicator */}
        <div className="flex justify-center gap-1">
          {dateRanges.map((_, index) => (
            <button
              key={index}
              className={`h-1.5 w-1.5 rounded-full transition-all duration-200 ${
                index === currentIndex
                  ? "bg-primary w-4"
                  : "bg-muted-foreground/30 hover:bg-muted-foreground/50"
              }`}
              onClick={() => scrollToIndex(index)}
              aria-label={`View ${dateRanges[index].label} analytics`}
            />
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
