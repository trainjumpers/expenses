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
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import type { InsightsTopExpense } from "@/lib/models/analytics";
import { formatPercentage } from "@/lib/utils";
import Link from "next/link";

export function TopPayees({ expenses }: { expenses: InsightsTopExpense[] }) {
  const rows = expenses.slice(0, 8);

  return (
    <AnalyticsSection
      title="Top payees"
      description="Grouped by the payee name saved on each transaction."
    >
      {rows.length === 0 ? (
        <InlineNote>No household expenses in this range.</InlineNote>
      ) : (
        <div className="overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Payee</TableHead>
                <TableHead className="text-right">Share</TableHead>
                <TableHead className="text-right">Average</TableHead>
                <TableHead className="text-right">Total</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {rows.map((expense) => (
                <TableRow key={expense.name}>
                  <TableCell className="max-w-[12rem]">
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Link
                          href={`/transaction?search=${encodeURIComponent(expense.name)}`}
                          className="block truncate font-medium hover:underline"
                        >
                          {expense.name}
                        </Link>
                      </TooltipTrigger>
                      <TooltipContent className="max-w-xs">
                        {expense.name}
                      </TooltipContent>
                    </Tooltip>
                  </TableCell>
                  <TableCell className="text-right text-sm tabular-nums text-muted-foreground">
                    {formatPercentage(expense.share * 100)}
                  </TableCell>
                  <TableCell className="text-right text-sm">
                    <Money value={expense.average} />
                  </TableCell>
                  <TableCell className="text-right text-sm font-medium">
                    <Money value={expense.amount} />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}
    </AnalyticsSection>
  );
}
