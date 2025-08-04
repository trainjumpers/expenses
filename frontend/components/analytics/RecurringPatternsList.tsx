'use client';

import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Badge } from '../ui/badge';
import { Progress } from '../ui/progress';
import { Repeat, Clock } from 'lucide-react';
import type { RecurringTransactionsResponse } from '../../lib/types/analytics';

interface RecurringPatternsListProps {
  data?: RecurringTransactionsResponse;
  isLoading?: boolean;
}

export function RecurringPatternsList({ data, isLoading }: RecurringPatternsListProps) {
  const formatCurrency = (amount: number) => {
    return `₹${amount.toLocaleString('en-IN', { maximumFractionDigits: 0 })}`;
  };

  const getFrequencyColor = (frequency: string) => {
    switch (frequency) {
      case 'weekly':
        return 'bg-blue-500';
      case 'monthly':
        return 'bg-green-500';
      case 'irregular':
        return 'bg-orange-500';
      default:
        return 'bg-gray-500';
    }
  };

  const getConfidenceColor = (confidence: number) => {
    if (confidence >= 0.8) return 'text-green-600';
    if (confidence >= 0.6) return 'text-yellow-600';
    return 'text-red-600';
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Repeat className="h-5 w-5" />
            Recurring Patterns
          </CardTitle>
          <CardDescription>
            Detected recurring transaction patterns
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {[...Array(3)].map((_, i) => (
              <div key={i} className="p-4 border rounded">
                <div className="space-y-2">
                  <div className="h-4 bg-muted rounded w-48 animate-pulse" />
                  <div className="h-3 bg-muted rounded w-32 animate-pulse" />
                  <div className="h-2 bg-muted rounded w-full animate-pulse" />
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!data || data.patterns.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Repeat className="h-5 w-5" />
            Recurring Patterns
          </CardTitle>
          <CardDescription>
            Detected recurring transaction patterns
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col items-center justify-center h-32 text-muted-foreground">
            <Repeat className="h-8 w-8 mb-2" />
            <p>No recurring patterns detected</p>
            <p className="text-sm">Add more transactions to detect patterns</p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Repeat className="h-5 w-5" />
          Recurring Patterns
        </CardTitle>
        <CardDescription>
          {data.patterns.length} patterns detected • Total: {formatCurrency(data.total_amount)}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {data.patterns.map((pattern, index) => (
            <div key={index} className="p-4 border rounded hover:bg-muted/50 transition-colors">
              <div className="flex items-start justify-between mb-3">
                <div className="flex-1">
                  <h4 className="font-medium capitalize">{pattern.pattern}</h4>
                  <div className="flex items-center gap-2 mt-1">
                    <Badge variant="secondary" className="text-xs">
                      <Clock className="h-3 w-3 mr-1" />
                      {pattern.frequency}
                    </Badge>
                    <Badge variant="outline" className="text-xs">
                      {pattern.count} transactions
                    </Badge>
                  </div>
                </div>
                <div className="text-right">
                  <div className="font-semibold">{formatCurrency(pattern.amount)}</div>
                  <div className="text-sm text-muted-foreground">per occurrence</div>
                </div>
              </div>
              
              <div className="space-y-2">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground">Confidence</span>
                  <span className={`font-medium ${getConfidenceColor(pattern.confidence)}`}>
                    {(pattern.confidence * 100).toFixed(0)}%
                  </span>
                </div>
                <Progress 
                  value={pattern.confidence * 100} 
                  className="h-2"
                />
              </div>
              
              {pattern.next_expected_date && (
                <div className="mt-3 text-sm text-muted-foreground">
                  Next expected: {new Date(pattern.next_expected_date).toLocaleDateString('en-IN')}
                </div>
              )}
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
