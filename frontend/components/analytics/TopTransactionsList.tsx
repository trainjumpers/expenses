'use client';

import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Badge } from '../ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../ui/tabs';
import { TrendingUp, TrendingDown, Receipt } from 'lucide-react';
import type { TopTransactionsResponse } from '../../lib/types/analytics';

interface TopTransactionsListProps {
  data?: TopTransactionsResponse;
  isLoading?: boolean;
}

export function TopTransactionsList({ data, isLoading }: TopTransactionsListProps) {
  const formatCurrency = (amount: number) => {
    return `₹${amount.toLocaleString('en-IN', { maximumFractionDigits: 0 })}`;
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-IN', {
      month: 'short',
      day: 'numeric',
    });
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Receipt className="h-5 w-5" />
            Top Transactions
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {[...Array(5)].map((_, i) => (
              <div key={i} className="flex items-center justify-between p-3 border rounded">
                <div className="space-y-2">
                  <div className="h-4 bg-muted rounded w-32 animate-pulse" />
                  <div className="h-3 bg-muted rounded w-24 animate-pulse" />
                </div>
                <div className="h-6 bg-muted rounded w-20 animate-pulse" />
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!data) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Receipt className="h-5 w-5" />
            Top Transactions
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col items-center justify-center h-32 text-muted-foreground">
            <Receipt className="h-8 w-8 mb-2" />
            <p>No transaction data available</p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Receipt className="h-5 w-5" />
          Top Transactions
        </CardTitle>
        <CardDescription>
          Highest expenses and income transactions
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Tabs defaultValue="expenses" className="w-full">
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="expenses" className="flex items-center gap-2">
              <TrendingUp className="h-4 w-4" />
              Top Expenses
            </TabsTrigger>
            <TabsTrigger value="income" className="flex items-center gap-2">
              <TrendingDown className="h-4 w-4" />
              Top Income
            </TabsTrigger>
          </TabsList>
          
          <TabsContent value="expenses" className="space-y-3 mt-4">
            {data.top_expenses.length === 0 ? (
              <p className="text-muted-foreground text-center py-4">No expense transactions found</p>
            ) : (
              data.top_expenses.map((transaction) => (
                <div key={transaction.id} className="flex items-center justify-between p-3 border rounded hover:bg-muted/50 transition-colors">
                  <div className="flex-1">
                    <div className="font-medium">{transaction.name}</div>
                    <div className="text-sm text-muted-foreground">
                      {transaction.account_name} • {formatDate(transaction.date)}
                    </div>
                    {transaction.categories.length > 0 && (
                      <div className="flex gap-1 mt-1">
                        {transaction.categories.slice(0, 2).map((category) => (
                          <Badge key={category} variant="secondary" className="text-xs">
                            {category}
                          </Badge>
                        ))}
                        {transaction.categories.length > 2 && (
                          <Badge variant="secondary" className="text-xs">
                            +{transaction.categories.length - 2}
                          </Badge>
                        )}
                      </div>
                    )}
                  </div>
                  <div className="text-right">
                    <div className="font-semibold text-destructive">
                      {formatCurrency(transaction.amount)}
                    </div>
                  </div>
                </div>
              ))
            )}
          </TabsContent>
          
          <TabsContent value="income" className="space-y-3 mt-4">
            {data.top_income.length === 0 ? (
              <p className="text-muted-foreground text-center py-4">No income transactions found</p>
            ) : (
              data.top_income.map((transaction) => (
                <div key={transaction.id} className="flex items-center justify-between p-3 border rounded hover:bg-muted/50 transition-colors">
                  <div className="flex-1">
                    <div className="font-medium">{transaction.name}</div>
                    <div className="text-sm text-muted-foreground">
                      {transaction.account_name} • {formatDate(transaction.date)}
                    </div>
                    {transaction.categories.length > 0 && (
                      <div className="flex gap-1 mt-1">
                        {transaction.categories.slice(0, 2).map((category) => (
                          <Badge key={category} variant="secondary" className="text-xs">
                            {category}
                          </Badge>
                        ))}
                        {transaction.categories.length > 2 && (
                          <Badge variant="secondary" className="text-xs">
                            +{transaction.categories.length - 2}
                          </Badge>
                        )}
                      </div>
                    )}
                  </div>
                  <div className="text-right">
                    <div className="font-semibold text-green-600">
                      {formatCurrency(transaction.amount)}
                    </div>
                  </div>
                </div>
              ))
            )}
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  );
}
