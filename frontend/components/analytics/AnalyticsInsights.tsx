'use client';

import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Badge } from '../ui/badge';
import { AlertCircle, Info, CheckCircle, Lightbulb } from 'lucide-react';
import type { AnalyticsInsightsResponse, AnalyticsInsight } from '../../lib/types/analytics';

interface AnalyticsInsightsProps {
  data?: AnalyticsInsightsResponse;
  isLoading?: boolean;
  maxItems?: number;
  showHeader?: boolean;
}

function InsightIcon({ type }: { type: AnalyticsInsight['type'] }) {
  switch (type) {
    case 'warning':
      return <AlertCircle className="h-4 w-4 text-orange-500" />;
    case 'info':
      return <Info className="h-4 w-4 text-blue-500" />;
    case 'success':
      return <CheckCircle className="h-4 w-4 text-green-500" />;
    case 'tip':
      return <Lightbulb className="h-4 w-4 text-yellow-500" />;
    default:
      return <Info className="h-4 w-4 text-gray-500" />;
  }
}

function InsightCard({ insight }: { insight: AnalyticsInsight }) {
  const getBorderColor = (type: AnalyticsInsight['type']) => {
    switch (type) {
      case 'warning':
        return 'border-l-orange-500';
      case 'info':
        return 'border-l-blue-500';
      case 'success':
        return 'border-l-green-500';
      case 'tip':
        return 'border-l-yellow-500';
      default:
        return 'border-l-gray-500';
    }
  };

  const getPriorityBadge = (priority: number) => {
    if (priority >= 4) return { variant: 'destructive' as const, label: 'High' };
    if (priority >= 3) return { variant: 'default' as const, label: 'Medium' };
    return { variant: 'secondary' as const, label: 'Low' };
  };

  const priorityBadge = getPriorityBadge(insight.priority);

  return (
    <div className={`p-4 border-l-4 bg-muted/30 rounded-r ${getBorderColor(insight.type)}`}>
      <div className="flex items-start gap-3">
        <InsightIcon type={insight.type} />
        <div className="flex-1 space-y-2">
          <div className="flex items-center justify-between">
            <h4 className="font-medium">{insight.title}</h4>
            <div className="flex items-center gap-2">
              <Badge variant={priorityBadge.variant} className="text-xs">
                {priorityBadge.label}
              </Badge>
              {insight.actionable && (
                <Badge variant="outline" className="text-xs">
                  Actionable
                </Badge>
              )}
            </div>
          </div>
          <p className="text-sm text-muted-foreground">{insight.description}</p>
          <div className="text-xs text-muted-foreground">
            {new Date(insight.created_at).toLocaleDateString('en-IN', {
              month: 'short',
              day: 'numeric',
              hour: '2-digit',
              minute: '2-digit',
            })}
          </div>
        </div>
      </div>
    </div>
  );
}

export function AnalyticsInsights({ 
  data, 
  isLoading, 
  maxItems, 
  showHeader = true 
}: AnalyticsInsightsProps) {
  if (isLoading) {
    return (
      <div className="space-y-4">
        {showHeader && (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Lightbulb className="h-5 w-5" />
                Analytics Insights
              </CardTitle>
              <CardDescription>
                AI-generated insights based on your spending patterns
              </CardDescription>
            </CardHeader>
          </Card>
        )}
        <div className="space-y-3">
          {[...Array(3)].map((_, i) => (
            <div key={i} className="p-4 border-l-4 border-l-muted bg-muted/30 rounded-r">
              <div className="space-y-2">
                <div className="h-4 bg-muted rounded w-48 animate-pulse" />
                <div className="h-3 bg-muted rounded w-full animate-pulse" />
                <div className="h-3 bg-muted rounded w-3/4 animate-pulse" />
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (!data || data.insights.length === 0) {
    return (
      <div className="space-y-4">
        {showHeader && (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Lightbulb className="h-5 w-5" />
                Analytics Insights
              </CardTitle>
              <CardDescription>
                AI-generated insights based on your spending patterns
              </CardDescription>
            </CardHeader>
          </Card>
        )}
        <Card>
          <CardContent className="flex flex-col items-center justify-center h-32 text-muted-foreground">
            <Lightbulb className="h-8 w-8 mb-2" />
            <p>No insights available</p>
            <p className="text-sm">Add more transactions to generate insights</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Sort insights by priority (highest first) and limit if specified
  const sortedInsights = [...data.insights]
    .sort((a, b) => b.priority - a.priority)
    .slice(0, maxItems);

  return (
    <div className="space-y-4">
      {showHeader && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Lightbulb className="h-5 w-5" />
              Analytics Insights
            </CardTitle>
            <CardDescription>
              {data.count} insights generated • Last updated: {' '}
              {new Date(data.generated_at).toLocaleDateString('en-IN', {
                month: 'short',
                day: 'numeric',
                hour: '2-digit',
                minute: '2-digit',
              })}
            </CardDescription>
          </CardHeader>
        </Card>
      )}
      
      <div className="space-y-3">
        {sortedInsights.map((insight, index) => (
          <InsightCard key={index} insight={insight} />
        ))}
      </div>
    </div>
  );
}
