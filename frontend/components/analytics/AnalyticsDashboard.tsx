"use client";

import {
  AlertCircle,
  Calendar,
  CreditCard,
  Lightbulb,
  PieChart,
} from "lucide-react";
import React, { useMemo, useState } from "react";

import { useDashboardData } from "../../hooks/useAnalytics";
import type {
  AnalyticsQuery,
  AnalyticsTimeRange,
} from "../../lib/types/analytics";
import { Button } from "../ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../ui/tabs";
import { AnalyticsFilters } from "./AnalyticsFilters";
import { AnalyticsInsights } from "./AnalyticsInsights";
import { CategoryBreakdownChart } from "./CategoryBreakdownChart";
import { MonthlyComparisonChart } from "./MonthlyComparisonChart";
import { RecurringPatternsList } from "./RecurringPatternsList";
import { SpendingOverviewCards } from "./SpendingOverviewCards";
import { SpendingTrendsChart } from "./SpendingTrendsChart";
import { TopTransactionsList } from "./TopTransactionsList";

interface AnalyticsDashboardProps {
  className?: string;
}

const TIME_RANGE_OPTIONS: { value: AnalyticsTimeRange; label: string }[] = [
  { value: "week", label: "Last 7 days" },
  { value: "month", label: "Last 30 days" },
  { value: "quarter", label: "Last 3 months" },
  { value: "year", label: "Last 12 months" },
  { value: "all_time", label: "All time" },
];

export function AnalyticsDashboard({ className }: AnalyticsDashboardProps) {
  const [timeRange, setTimeRange] = useState<AnalyticsTimeRange>("month");
  const [selectedAccounts, setSelectedAccounts] = useState<number[]>([]);
  const [selectedCategories, setSelectedCategories] = useState<number[]>([]);
  const [activeTab, setActiveTab] = useState("overview");

  const analyticsQuery: AnalyticsQuery = useMemo(
    () => ({
      time_range: timeRange,
      account_ids: selectedAccounts.length > 0 ? selectedAccounts : undefined,
      category_ids:
        selectedCategories.length > 0 ? selectedCategories : undefined,
    }),
    [timeRange, selectedAccounts, selectedCategories]
  );

  const { data, isLoading, error } = useDashboardData(analyticsQuery);

  const currentTimeRangeLabel =
    TIME_RANGE_OPTIONS.find((option) => option.value === timeRange)?.label ||
    "Selected period";

  if (error) {
    return (
      <div className="flex items-center justify-center h-64">
        <Card className="w-full max-w-md">
          <CardContent className="flex flex-col items-center justify-center p-6">
            <AlertCircle className="h-12 w-12 text-destructive mb-4" />
            <h3 className="text-lg font-semibold mb-2">
              Failed to load analytics
            </h3>
            <p className="text-sm text-muted-foreground text-center mb-4">
              {error.message ||
                "An error occurred while loading your analytics data."}
            </p>
            <Button onClick={() => window.location.reload()} variant="outline">
              Try Again
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Analytics</h1>
          <p className="text-muted-foreground">
            Insights and trends for {currentTimeRangeLabel.toLowerCase()}
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Select
            value={timeRange}
            onValueChange={(value: AnalyticsTimeRange) => setTimeRange(value)}
          >
            <SelectTrigger className="w-[180px]">
              <Calendar className="h-4 w-4 mr-2" />
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {TIME_RANGE_OPTIONS.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Filters */}
      <AnalyticsFilters
        selectedAccounts={selectedAccounts}
        selectedCategories={selectedCategories}
        onAccountsChange={setSelectedAccounts}
        onCategoriesChange={setSelectedCategories}
      />

      {/* Main Content */}
      <Tabs
        value={activeTab}
        onValueChange={setActiveTab}
        className="space-y-6"
      >
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="trends">Trends</TabsTrigger>
          <TabsTrigger value="breakdown">Breakdown</TabsTrigger>
          <TabsTrigger value="insights">Insights</TabsTrigger>
        </TabsList>

        {/* Overview Tab */}
        <TabsContent value="overview" className="space-y-6">
          {/* Overview Cards */}
          <SpendingOverviewCards
            data={data?.summary.overview}
            isLoading={isLoading}
          />

          {/* Quick Stats Grid */}
          <div className="grid gap-6 md:grid-cols-2">
            {/* Top Categories */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <PieChart className="h-5 w-5" />
                  Top Categories
                </CardTitle>
                <CardDescription>
                  Your highest spending categories
                </CardDescription>
              </CardHeader>
              <CardContent>
                <CategoryBreakdownChart
                  data={data?.summary.category_breakdown}
                  isLoading={isLoading}
                  showChart={false}
                  maxItems={5}
                />
              </CardContent>
            </Card>

            {/* Recent Insights */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Lightbulb className="h-5 w-5" />
                  Key Insights
                </CardTitle>
                <CardDescription>
                  AI-generated spending insights
                </CardDescription>
              </CardHeader>
              <CardContent>
                <AnalyticsInsights
                  data={data?.insights}
                  isLoading={isLoading}
                  maxItems={3}
                  showHeader={false}
                />
              </CardContent>
            </Card>
          </div>

          {/* Top Transactions */}
          <TopTransactionsList
            data={data?.summary.top_transactions}
            isLoading={isLoading}
          />
        </TabsContent>

        {/* Trends Tab */}
        <TabsContent value="trends" className="space-y-6">
          <div className="grid gap-6">
            <SpendingTrendsChart query={analyticsQuery} isLoading={isLoading} />
            <MonthlyComparisonChart
              data={data?.summary.monthly_comparison}
              isLoading={isLoading}
            />
          </div>
        </TabsContent>

        {/* Breakdown Tab */}
        <TabsContent value="breakdown" className="space-y-6">
          <div className="grid gap-6 lg:grid-cols-2">
            <CategoryBreakdownChart
              data={data?.summary.category_breakdown}
              isLoading={isLoading}
            />
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <CreditCard className="h-5 w-5" />
                  Account Breakdown
                </CardTitle>
                <CardDescription>
                  Spending distribution across your accounts
                </CardDescription>
              </CardHeader>
              <CardContent>
                {isLoading ? (
                  <div className="space-y-3">
                    {[...Array(3)].map((_, i) => (
                      <div
                        key={i}
                        className="flex items-center justify-between"
                      >
                        <div className="h-4 bg-muted rounded w-32 animate-pulse" />
                        <div className="h-4 bg-muted rounded w-20 animate-pulse" />
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="space-y-3">
                    {data?.summary.account_breakdown.accounts.map((account) => (
                      <div
                        key={account.account_id}
                        className="flex items-center justify-between"
                      >
                        <div className="flex flex-col">
                          <span className="font-medium">
                            {account.account_name}
                          </span>
                          {account.bank_name && (
                            <span className="text-sm text-muted-foreground">
                              {account.bank_name}
                            </span>
                          )}
                        </div>
                        <div className="text-right">
                          <div className="font-medium">
                            ₹{account.amount.toLocaleString()}
                          </div>
                          <div className="text-sm text-muted-foreground">
                            {account.percentage.toFixed(1)}%
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </div>

          <RecurringPatternsList
            data={data?.summary.recurring_patterns}
            isLoading={isLoading}
          />
        </TabsContent>

        {/* Insights Tab */}
        <TabsContent value="insights" className="space-y-6">
          <AnalyticsInsights data={data?.insights} isLoading={isLoading} />
        </TabsContent>
      </Tabs>
    </div>
  );
}
