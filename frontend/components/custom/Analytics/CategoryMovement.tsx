"use client";

import {
  AnalyticsSection,
  InlineNote,
  Money,
} from "@/components/custom/Analytics/AnalyticsShared";
import type { InsightsCategoryMovement } from "@/lib/models/analytics";
import { cn, formatPercentage, getSurplusTone } from "@/lib/utils";
import { format } from "date-fns";
import Link from "next/link";

function monthName(month: string) {
  if (!month) return "the prior period";
  return format(new Date(`${month}-01T00:00:00`), "MMMM");
}

function CategoryChange({ change }: { change: number }) {
  if (change === 0) {
    return <span className="text-xs text-muted-foreground">no change</span>;
  }

  const increased = change > 0;
  return (
    <span
      className={cn(
        "whitespace-nowrap text-sm tabular-nums",
        getSurplusTone(-change)
      )}
    >
      <span aria-hidden="true">{increased ? "▲" : "▼"}</span>{" "}
      <span className="sr-only">{increased ? "up " : "down "}</span>
      <Money value={Math.abs(change)} />
    </span>
  );
}

export function CategoryMovement({
  movement,
  recentMonth,
  priorMonth,
}: {
  movement: InsightsCategoryMovement[];
  recentMonth: string;
  priorMonth: string;
}) {
  const rows = movement
    .filter((item) => item.recent_total !== 0 || item.prior_total !== 0)
    .sort((a, b) => Math.abs(b.change) - Math.abs(a.change))
    .slice(0, 8);

  return (
    <AnalyticsSection
      title="Category movement"
      description={`Expense totals for ${monthName(recentMonth)} compared with ${monthName(priorMonth)}.`}
    >
      {rows.length === 0 ? (
        <InlineNote>No categorized spending to compare yet.</InlineNote>
      ) : (
        <ul className="space-y-4">
          {rows.map((item) => (
            <li key={item.category_id} className="space-y-1.5">
              <div className="flex items-baseline justify-between gap-3">
                <Link
                  href={
                    item.category_id === -1
                      ? "/transaction?uncategorized=true"
                      : `/transaction?category_id=${item.category_id}`
                  }
                  className="min-w-0 truncate text-sm font-medium text-foreground hover:underline"
                  title={item.category_name}
                >
                  {item.category_name}
                </Link>
                <span className="flex items-baseline gap-3">
                  <Money value={item.recent_total} className="text-sm" />
                  <CategoryChange change={item.change} />
                </span>
              </div>
              <div className="h-1.5 w-full overflow-hidden rounded-full bg-muted">
                <div
                  className="h-full rounded-full bg-(--chart-3)"
                  style={{
                    width: `${Math.min(item.recent_share * 100, 100)}%`,
                  }}
                />
              </div>
              <p className="text-xs text-muted-foreground">
                {formatPercentage(item.recent_share * 100)} of recent spending
                {item.prior_share > 0
                  ? `, ${formatPercentage(item.prior_share * 100)} prior`
                  : ""}
              </p>
            </li>
          ))}
        </ul>
      )}
    </AnalyticsSection>
  );
}
