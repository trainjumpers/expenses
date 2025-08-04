'use client';

import React, { useMemo } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Badge } from '../ui/badge';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip, Legend } from 'recharts';
import { PieChart as PieChartIcon } from 'lucide-react';
import type { CategorySpendingResponse, CategorySpendingItem } from '../../lib/types/analytics';

interface CategoryBreakdownChartProps {
  data?: CategorySpendingResponse;
  isLoading?: boolean;
  showChart?: boolean;
  maxItems?: number;
}

// Color palette for categories
const COLORS = [
  '#8884d8', '#82ca9d', '#ffc658', '#ff7c7c', '#8dd1e1',
  '#d084d0', '#ffb347', '#87ceeb', '#dda0dd', '#98fb98',
  '#f0e68c', '#ff6347', '#40e0d0', '#ee82ee', '#90ee90'
];

interface CustomTooltipProps {
  active?: boolean;
  payload?: Array<{
    name: string;
    value: number;
    payload: {
      name: string;
      value: number;
      percentage: number;
      count: number;
    };
  }>;
}

function CustomTooltip({ active, payload }: CustomTooltipProps) {
  if (active && payload && payload.length) {
    const data = payload[0].payload;
    return (
      <div className="bg-background border rounded-lg shadow-lg p-3">
        <p className="font-medium">{data.name}</p>
        <p className="text-sm text-muted-foreground">
          Amount: ₹{data.value.toLocaleString()}
        </p>
        <p className="text-sm text-muted-foreground">
          Percentage: {data.percentage.toFixed(1)}%
        </p>
        <p className="text-sm text-muted-foreground">
          Transactions: {data.count}
        </p>
      </div>
    );
  }
  return null;
}

export function CategoryBreakdownChart({ 
  data, 
  isLoading, 
  showChart = true, 
  maxItems 
}: CategoryBreakdownChartProps) {
  const chartData = useMemo(() => {
    if (!data) return [];
    
    const categories = [...data.categories];
    if (data.uncategorized.amount > 0) {
      categories.push(data.uncategorized);
    }
    
    // Sort by amount and limit if maxItems is specified
    const sortedCategories = categories
      .sort((a, b) => b.amount - a.amount)
      .slice(0, maxItems);
    
    return sortedCategories.map((category, index) => ({
      name: category.category_name,
      value: category.amount,
      percentage: category.percentage,
      count: category.count,
      color: COLORS[index % COLORS.length],
    }));
  }, [data, maxItems]);

  const formatCurrency = (amount: number) => {
    return `₹${amount.toLocaleString('en-IN', { maximumFractionDigits: 0 })}`;
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <PieChartIcon className="h-5 w-5" />
            Category Breakdown
          </CardTitle>
          <CardDescription>
            Spending distribution by category
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {showChart && (
              <div className="h-64 flex items-center justify-center">
                <div className="h-32 w-32 bg-muted rounded-full animate-pulse" />
              </div>
            )}
            <div className="space-y-3">
              {[...Array(5)].map((_, i) => (
                <div key={i} className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="h-3 w-3 bg-muted rounded-full animate-pulse" />
                    <div className="h-4 bg-muted rounded w-24 animate-pulse" />
                  </div>
                  <div className="h-4 bg-muted rounded w-16 animate-pulse" />
                </div>
              ))}
            </div>
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
            <PieChartIcon className="h-5 w-5" />
            Category Breakdown
          </CardTitle>
          <CardDescription>
            Spending distribution by category
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col items-center justify-center h-64 text-muted-foreground">
            <PieChartIcon className="h-12 w-12 mb-4" />
            <p>No spending data available</p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <PieChartIcon className="h-5 w-5" />
          Category Breakdown
        </CardTitle>
        <CardDescription>
          Total spending: {formatCurrency(data.total_amount)} across {data.total_count} transactions
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {showChart && (
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={chartData}
                    cx="50%"
                    cy="50%"
                    innerRadius={60}
                    outerRadius={100}
                    paddingAngle={2}
                    dataKey="value"
                  >
                    {chartData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                  <Tooltip content={<CustomTooltip />} />
                  <Legend />
                </PieChart>
              </ResponsiveContainer>
            </div>
          )}
          
          <div className="space-y-3">
            {chartData.map((category, index) => (
              <div key={category.name} className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div 
                    className="h-3 w-3 rounded-full" 
                    style={{ backgroundColor: category.color }}
                  />
                  <div className="flex flex-col">
                    <span className="font-medium">{category.name}</span>
                    <span className="text-sm text-muted-foreground">
                      {category.count} transactions
                    </span>
                  </div>
                </div>
                <div className="text-right">
                  <div className="font-medium">{formatCurrency(category.value)}</div>
                  <Badge variant="secondary" className="text-xs">
                    {category.percentage.toFixed(1)}%
                  </Badge>
                </div>
              </div>
            ))}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
