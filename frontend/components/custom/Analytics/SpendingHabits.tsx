"use client";

import {
  AnalyticsSection,
  InlineNote,
  Money,
} from "@/components/custom/Analytics/AnalyticsShared";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type {
  InsightsSpendingSummary,
  InsightsWeekdayBehavior,
} from "@/lib/models/analytics";
import { formatPercentage } from "@/lib/utils";

const WEEKDAY_LABELS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];

function Stat({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="min-w-0">
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className="mt-0.5 truncate text-sm font-medium text-foreground">
        {children}
      </dd>
    </div>
  );
}

export function SpendingHabits({
  summary,
  weekday,
}: {
  summary: InsightsSpendingSummary;
  weekday: InsightsWeekdayBehavior;
}) {
  const byWeekday = new Map(weekday.days.map((day) => [day.weekday, day]));
  const ordered = WEEKDAY_LABELS.map((label, index) => {
    const weekdayNumber = (index + 1) % 7;
    const day = byWeekday.get(weekdayNumber);
    return {
      label,
      total: day?.total ?? 0,
      count: day?.count ?? 0,
      average: day?.average ?? 0,
      share: day?.share ?? 0,
    };
  });

  return (
    <AnalyticsSection
      title="Spending habits"
      description="How often you spend and how the week is distributed."
    >
      <div className="space-y-6">
        <dl className="grid grid-cols-2 gap-4 sm:grid-cols-3">
          <Stat label="Transactions">{summary.expense_count}</Stat>
          <Stat label="Average">
            <Money value={summary.average_transaction} />
          </Stat>
          <Stat label="Median">
            <Money value={summary.median_transaction} />
          </Stat>
          <Stat label="Largest">
            <Money value={summary.largest_expense} />
          </Stat>
          <Stat label="Spending days">{summary.active_spending_days}</Stat>
          <Stat label="No-spend days">{summary.no_spend_days}</Stat>
        </dl>

        {summary.expense_count === 0 ? (
          <InlineNote>No spending in this range.</InlineNote>
        ) : (
          <div>
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Day</TableHead>
                    <TableHead className="text-right">Transactions</TableHead>
                    <TableHead className="hidden text-right sm:table-cell">
                      Average
                    </TableHead>
                    <TableHead className="text-right">Total</TableHead>
                    <TableHead className="hidden text-right sm:table-cell">
                      Share
                    </TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {ordered.map((day) => (
                    <TableRow key={day.label}>
                      <TableCell className="font-medium">{day.label}</TableCell>
                      <TableCell className="text-right text-sm tabular-nums">
                        {day.count}
                      </TableCell>
                      <TableCell className="hidden text-right text-sm sm:table-cell">
                        <Money value={day.average} />
                      </TableCell>
                      <TableCell className="text-right text-sm">
                        <Money value={day.total} />
                      </TableCell>
                      <TableCell className="hidden text-right text-sm tabular-nums text-muted-foreground sm:table-cell">
                        {formatPercentage(day.share * 100)}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
            <p className="mt-3 text-xs text-muted-foreground">
              Weekends are {formatPercentage(weekday.weekend_share * 100)} of
              spending.
            </p>
          </div>
        )}
      </div>
    </AnalyticsSection>
  );
}
