"use client";

import { AnalyticsSection } from "@/components/custom/Analytics/AnalyticsShared";
import type {
  InsightsDataConfidence,
  InsightsSummary,
} from "@/lib/models/analytics";
import { formatPercentage } from "@/lib/utils";
import { format } from "date-fns";
import Link from "next/link";

const STALE_DAYS = 30;

export function DataHealth({
  summary,
  confidence,
}: {
  summary: InsightsSummary;
  confidence: InsightsDataConfidence;
}) {
  const notes: React.ReactNode[] = [];

  if (summary.uncategorized_count > 0) {
    notes.push(
      <>
        {formatPercentage(confidence.uncategorized_share * 100)} of spending is
        uncategorized across {summary.uncategorized_count} transactions.{" "}
        <Link href="/transaction?uncategorized=true" className="underline">
          Review them
        </Link>
        .
      </>
    );
  }

  if (confidence.multi_category_count > 0) {
    notes.push(
      <>
        {confidence.multi_category_count} transactions (
        {formatPercentage(confidence.multi_category_share * 100)}) are split
        across several categories, so category totals allocate their amount
        evenly.
      </>
    );
  }

  if (
    confidence.stale_days >= STALE_DAYS &&
    confidence.latest_transaction_date
  ) {
    notes.push(
      <>
        The latest transaction is from{" "}
        {format(new Date(confidence.latest_transaction_date), "d MMM yyyy")} (
        {confidence.stale_days} days ago), so recent months may be incomplete.
      </>
    );
  }

  if (confidence.multiple_currencies) {
    notes.push(
      <>
        Accounts span {confidence.currencies.join(", ")}. Totals combine
        currencies without conversion.
      </>
    );
  }

  return (
    <AnalyticsSection
      title="Data health"
      description="Limitations that affect how precisely these figures should be read."
    >
      {notes.length === 0 ? (
        <p className="text-sm text-muted-foreground">
          No known data gaps in this range.
        </p>
      ) : (
        <ul className="space-y-2">
          {notes.map((note, index) => (
            <li
              key={index}
              className="text-pretty text-sm text-muted-foreground"
            >
              {note}
            </li>
          ))}
        </ul>
      )}
    </AnalyticsSection>
  );
}
