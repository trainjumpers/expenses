'use client';

import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { Calendar, TrendingUp, TrendingDown } from 'lucide-react';
import type { MonthlyComparisonResponse } from '../../lib/types/analytics';

interface MonthlyComparisonChartProps {
  data?: MonthlyComparisonResponse;
  isLoading?: boolean;
}

interface CustomTooltipProps {
  active?: boolean;
  payload?: Array<{
    value: number;
    payload: {
      month_name: string;
      amount: number;
      change: number;
      change_amount: number;
      count: number;
    };
  }>;
}

function CustomTooltip({ active, payload }: CustomTooltipProps) {
  if (active && payload && payload.length) {
    const data = payload[0].payload;
    return (
      <div className="bg-background border rounded-lg shadow-lg p-3">
        <p className="font-medium">{data.month_name}</p>
        <p className="text-sm text-muted-foreground">
          Amount: ₹{data.amount.toLocaleString()}
        </p>
        <p className="text-sm text-muted-foreground">
          Transactions: {data.count}
        </p>
        {data.change !== 0 && (
          <p className={`text-sm flex items-center gap-1 ${
            data.change > 0 ? 'text-destructive' : 'text-green-600'
          }`}>
            {data.change > 0 ? <TrendingUp className="h-3 w-3" /> : <TrendingDown className="h-3 w-3" />}
            {Math.abs(data.change).toFixed(1)}% vs previous month
          </p>
        )}
      </div>
    );
  }
  return null;
}

export function MonthlyComparisonChart({ data, isLoading }: MonthlyComparisonChartProps) {
  const formatCurrency = (value: number) => {
    return `₹${(value / 1000).toFixed(0)}K`;
  };

  const formatXAxisLabel = (tickItem: string) => {
    return tickItem.split(' ')[0]; // Show only month name
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Calendar className="h-5 w-5" />
            Monthly Comparison
          </CardTitle>
          <CardDescription>
            Month-over-month spending trends
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="h-64 bg-muted rounded animate-pulse" />
        </CardContent>
      </Card>
    );
  }

  if (!data || data.months.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Calendar className="h-5 w-5" />
            Monthly Comparison
          </CardTitle>
          <CardDescription>
            Month-over-month spending trends
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col items-center justify-center h-64 text-muted-foreground">
            <Calendar className="h-12 w-12 mb-4" />
            <p>No monthly data available</p>
          </div>
        </CardContent>
      </Card>
    );
  }

  const chartData = data.months.map(month => ({
    ...month,
    displayAmount: month.amount,
  }));

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Calendar className="h-5 w-5" />
          Monthly Comparison
        </CardTitle>
        <CardDescription>
          Spending trends across {data.total_months} months
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="h-64">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={chartData} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
              <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
              <XAxis 
                dataKey="month_name" 
                tickFormatter={formatXAxisLabel}
                className="text-xs"
              />
              <YAxis 
                tickFormatter={formatCurrency}
                className="text-xs"
              />
              <Tooltip content={<CustomTooltip />} />
              <Bar 
                dataKey="displayAmount" 
                fill="#8884d8" 
                radius={[4, 4, 0, 0]}
              />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  );
}
