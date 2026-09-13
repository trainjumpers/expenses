"use client";

import {
  AnalyticsSection,
  InlineNote,
  Money,
} from "@/components/custom/Analytics/AnalyticsShared";
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
  const maxTotal = Math.max(...ordered.map((day) => day.total), 0);

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
            <ul className="space-y-2">
              {ordered.map((day) => {
                const share = day.share;
                return (
                  <li
                    key={day.label}
                    className="grid grid-cols-[2.5rem_1fr_auto] items-center gap-3"
                  >
                    <span className="text-xs font-medium text-muted-foreground">
                      {day.label}
                    </span>
                    <span
                      className="h-2 overflow-hidden rounded-full bg-muted"
                      title={`${day.label}: ${formatPercentage(share * 100)} of spending`}
                    >
                      <span
                        className="block h-full rounded-full bg-(--chart-4)"
                        style={{
                          width:
                            maxTotal > 0
                              ? `${(day.total / maxTotal) * 100}%`
                              : "0%",
                        }}
                      />
                    </span>
                    <span className="text-xs tabular-nums text-muted-foreground">
                      {formatPercentage(share * 100)}
                    </span>
                  </li>
                );
              })}
            </ul>
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
