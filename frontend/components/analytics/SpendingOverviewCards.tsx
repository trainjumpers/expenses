'use client';

import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import { TrendingUp, TrendingDown, DollarSign, CreditCard, ArrowUpDown, Receipt } from 'lucide-react';
import type { SpendingOverviewResponse } from '../../lib/types/analytics';

interface SpendingOverviewCardsProps {
  data?: SpendingOverviewResponse;
  isLoading?: boolean;
}

interface MetricCardProps {
  title: string;
  value: string;
  change?: number;
  changeLabel?: string;
  icon: React.ReactNode;
  isLoading?: boolean;
  color?: 'default' | 'success' | 'warning' | 'destructive';
}

function MetricCard({ 
  title, 
  value, 
  change, 
  changeLabel, 
  icon, 
  isLoading, 
  color = 'default' 
}: MetricCardProps) {
  const getChangeColor = (change?: number) => {
    if (!change) return 'text-muted-foreground';
    return change > 0 ? 'text-destructive' : 'text-green-600';
  };

  const getChangeIcon = (change?: number) => {
    if (!change) return null;
    return change > 0 ? (
      <TrendingUp className="h-3 w-3" />
    ) : (
      <TrendingDown className="h-3 w-3" />
    );
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">
            <div className="h-4 bg-muted rounded w-24 animate-pulse" />
          </CardTitle>
          <div className="h-4 w-4 bg-muted rounded animate-pulse" />
        </CardHeader>
        <CardContent>
          <div className="h-8 bg-muted rounded w-32 animate-pulse mb-2" />
          <div className="h-3 bg-muted rounded w-20 animate-pulse" />
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className={color !== 'default' ? `border-${color}` : ''}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        {icon}
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
        {change !== undefined && (
          <p className={`text-xs flex items-center gap-1 ${getChangeColor(change)}`}>
            {getChangeIcon(change)}
            {Math.abs(change).toFixed(1)}% {changeLabel || 'from last period'}
          </p>
        )}
      </CardContent>
    </Card>
  );
}

export function SpendingOverviewCards({ data, isLoading }: SpendingOverviewCardsProps) {
  const formatCurrency = (amount: number) => {
    return `₹${amount.toLocaleString('en-IN', { maximumFractionDigits: 0 })}`;
  };

  const formatPercentage = (value: number) => {
    return `${value.toFixed(1)}%`;
  };

  // Calculate savings rate
  const savingsRate = data ? 
    data.total_income > 0 ? ((data.total_income - data.total_expenses) / data.total_income) * 100 : 0 
    : 0;

  const metrics = [
    {
      title: 'Total Expenses',
      value: data ? formatCurrency(data.total_expenses) : '₹0',
      icon: <TrendingUp className="h-4 w-4 text-destructive" />,
      color: 'destructive' as const,
    },
    {
      title: 'Total Income',
      value: data ? formatCurrency(data.total_income) : '₹0',
      icon: <TrendingDown className="h-4 w-4 text-green-600" />,
      color: 'success' as const,
    },
    {
      title: 'Net Amount',
      value: data ? formatCurrency(data.net_amount) : '₹0',
      icon: <ArrowUpDown className="h-4 w-4 text-blue-600" />,
      color: data && data.net_amount < 0 ? 'destructive' : 'success',
    },
    {
      title: 'Transactions',
      value: data ? data.transaction_count.toString() : '0',
      icon: <Receipt className="h-4 w-4 text-muted-foreground" />,
    },
    {
      title: 'Avg. Expense',
      value: data ? formatCurrency(data.average_expense) : '₹0',
      icon: <DollarSign className="h-4 w-4 text-orange-600" />,
    },
    {
      title: 'Savings Rate',
      value: formatPercentage(savingsRate),
      icon: <CreditCard className="h-4 w-4 text-green-600" />,
      color: savingsRate > 20 ? 'success' : savingsRate > 10 ? 'warning' : 'destructive',
    },
  ];

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6">
      {metrics.map((metric, index) => (
        <MetricCard
          key={index}
          title={metric.title}
          value={metric.value}
          icon={metric.icon}
          isLoading={isLoading}
          color={metric.color}
        />
      ))}
    </div>
  );
}
