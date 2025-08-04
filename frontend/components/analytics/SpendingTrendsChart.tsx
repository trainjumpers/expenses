'use client';

import React, { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Button } from '../ui/button';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, Area, AreaChart } from 'recharts';
import { TrendingUp, Calendar } from 'lucide-react';
import { useSpendingTrends } from '../../hooks/useAnalytics';
import type { AnalyticsQuery, TimeSeriesDataPoint } from '../../lib/types/analytics';

interface SpendingTrendsChartProps {
  query: AnalyticsQuery;
  isLoading?: boolean;
}

type GranularityOption = 'daily' | 'weekly' | 'monthly';

interface CustomTooltipProps {
  active?: boolean;
  payload?: Array<{
    name: string;
    value: number;
    color: string;
  }>;
  label?: string;
}

function CustomTooltip({ active, payload, label }: CustomTooltipProps) {
  if (active && payload && payload.length) {
    const date = new Date(label || '');
    const formattedDate = date.toLocaleDateString('en-IN', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });

    return (
      <div className="bg-background border rounded-lg shadow-lg p-3">
        <p className="font-medium mb-2">{formattedDate}</p>
        {payload.map((entry, index) => (
          <p key={index} className="text-sm" style={{ color: entry.color }}>
            {entry.name}: ₹{entry.value.toLocaleString()}
          </p>
        ))}
      </div>
    );
  }
  return null;
}

export function SpendingTrendsChart({ query, isLoading: parentLoading }: SpendingTrendsChartProps) {
  const [granularity, setGranularity] = useState<GranularityOption>('daily');
  const [chartType, setChartType] = useState<'line' | 'area'>('area');

  const { data, isLoading, error } = useSpendingTrends(query, granularity);

  const granularityOptions: { value: GranularityOption; label: string }[] = [
    { value: 'daily', label: 'Daily' },
    { value: 'weekly', label: 'Weekly' },
    { value: 'monthly', label: 'Monthly' },
  ];

  const chartData = data?.data_points.map(point => ({
    date: point.date,
    expenses: point.expenses,
    income: point.income,
    net: point.income - point.expenses,
    count: point.count,
  })) || [];

  const formatCurrency = (value: number) => {
    return `₹${value.toLocaleString('en-IN', { maximumFractionDigits: 0 })}`;
  };

  const formatXAxisLabel = (tickItem: string) => {
    const date = new Date(tickItem);
    switch (granularity) {
      case 'daily':
        return date.toLocaleDateString('en-IN', { month: 'short', day: 'numeric' });
      case 'weekly':
        return date.toLocaleDateString('en-IN', { month: 'short', day: 'numeric' });
      case 'monthly':
        return date.toLocaleDateString('en-IN', { month: 'short', year: '2-digit' });
      default:
        return date.toLocaleDateString('en-IN', { month: 'short', day: 'numeric' });
    }
  };

  if (parentLoading || isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <TrendingUp className="h-5 w-5" />
            Spending Trends
          </CardTitle>
          <CardDescription>
            Track your spending patterns over time
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="flex gap-2">
              {granularityOptions.map((option) => (
                <div key={option.value} className="h-8 w-16 bg-muted rounded animate-pulse" />
              ))}
            </div>
            <div className="h-64 bg-muted rounded animate-pulse" />
          </div>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <TrendingUp className="h-5 w-5" />
            Spending Trends
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col items-center justify-center h-64 text-muted-foreground">
            <TrendingUp className="h-12 w-12 mb-4" />
            <p>Failed to load spending trends</p>
            <p className="text-sm">{error.message}</p>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!data || chartData.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <TrendingUp className="h-5 w-5" />
            Spending Trends
          </CardTitle>
          <CardDescription>
            Track your spending patterns over time
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col items-center justify-center h-64 text-muted-foreground">
            <TrendingUp className="h-12 w-12 mb-4" />
            <p>No spending data available for the selected period</p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <TrendingUp className="h-5 w-5" />
              Spending Trends
            </CardTitle>
            <CardDescription>
              {data.granularity.charAt(0).toUpperCase() + data.granularity.slice(1)} view for {data.period}
            </CardDescription>
          </div>
          <div className="flex items-center gap-2">
            <div className="flex rounded-md border">
              {granularityOptions.map((option) => (
                <Button
                  key={option.value}
                  variant={granularity === option.value ? 'default' : 'ghost'}
                  size="sm"
                  onClick={() => setGranularity(option.value)}
                  className="rounded-none first:rounded-l-md last:rounded-r-md"
                >
                  {option.label}
                </Button>
              ))}
            </div>
            <div className="flex rounded-md border">
              <Button
                variant={chartType === 'area' ? 'default' : 'ghost'}
                size="sm"
                onClick={() => setChartType('area')}
                className="rounded-none rounded-l-md"
              >
                Area
              </Button>
              <Button
                variant={chartType === 'line' ? 'default' : 'ghost'}
                size="sm"
                onClick={() => setChartType('line')}
                className="rounded-none rounded-r-md"
              >
                Line
              </Button>
            </div>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className="h-80">
          <ResponsiveContainer width="100%" height="100%">
            {chartType === 'area' ? (
              <AreaChart data={chartData} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
                <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                <XAxis 
                  dataKey="date" 
                  tickFormatter={formatXAxisLabel}
                  className="text-xs"
                />
                <YAxis 
                  tickFormatter={formatCurrency}
                  className="text-xs"
                />
                <Tooltip content={<CustomTooltip />} />
                <Legend />
                <Area
                  type="monotone"
                  dataKey="expenses"
                  stackId="1"
                  stroke="#ef4444"
                  fill="#ef4444"
                  fillOpacity={0.6}
                  name="Expenses"
                />
                <Area
                  type="monotone"
                  dataKey="income"
                  stackId="2"
                  stroke="#22c55e"
                  fill="#22c55e"
                  fillOpacity={0.6}
                  name="Income"
                />
              </AreaChart>
            ) : (
              <LineChart data={chartData} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
                <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                <XAxis 
                  dataKey="date" 
                  tickFormatter={formatXAxisLabel}
                  className="text-xs"
                />
                <YAxis 
                  tickFormatter={formatCurrency}
                  className="text-xs"
                />
                <Tooltip content={<CustomTooltip />} />
                <Legend />
                <Line
                  type="monotone"
                  dataKey="expenses"
                  stroke="#ef4444"
                  strokeWidth={2}
                  dot={{ fill: '#ef4444', strokeWidth: 2, r: 4 }}
                  name="Expenses"
                />
                <Line
                  type="monotone"
                  dataKey="income"
                  stroke="#22c55e"
                  strokeWidth={2}
                  dot={{ fill: '#22c55e', strokeWidth: 2, r: 4 }}
                  name="Income"
                />
                <Line
                  type="monotone"
                  dataKey="net"
                  stroke="#3b82f6"
                  strokeWidth={2}
                  strokeDasharray="5 5"
                  dot={{ fill: '#3b82f6', strokeWidth: 2, r: 4 }}
                  name="Net"
                />
              </LineChart>
            )}
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  );
}
