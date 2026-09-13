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
import type { InsightsInvestment } from "@/lib/models/analytics";
import { formatPercentage } from "@/lib/utils";
import Link from "next/link";

export function InvestmentTable({
  investments,
}: {
  investments: InsightsInvestment[];
}) {
  const rows = [...investments].sort(
    (a, b) => b.current_value - a.current_value
  );

  return (
    <AnalyticsSection
      title="Investments"
      description="Current value is the latest recorded valuation. XIRR is the money-weighted return."
    >
      {rows.length === 0 ? (
        <InlineNote>No investment accounts with a current value.</InlineNote>
      ) : (
        <div className="overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead className="text-right">Current</TableHead>
                <TableHead className="hidden text-right sm:table-cell">
                  Invested
                </TableHead>
                <TableHead className="hidden text-right md:table-cell">
                  Withdrawn
                </TableHead>
                <TableHead className="text-right">XIRR</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {rows.map((investment) => (
                <TableRow key={investment.account_id}>
                  <TableCell className="max-w-[10rem]">
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Link
                          href={`/transaction?account_id=${investment.account_id}`}
                          className="block truncate font-medium hover:underline"
                        >
                          {investment.name}
                        </Link>
                      </TooltipTrigger>
                      <TooltipContent>{investment.name}</TooltipContent>
                    </Tooltip>
                  </TableCell>
                  <TableCell className="text-right text-sm font-medium">
                    <Money value={investment.current_value} />
                  </TableCell>
                  <TableCell className="hidden text-right text-sm sm:table-cell">
                    <Money value={investment.contributed} />
                  </TableCell>
                  <TableCell className="hidden text-right text-sm md:table-cell">
                    <Money value={investment.distributed} />
                  </TableCell>
                  <TableCell className="text-right text-sm tabular-nums">
                    {investment.xirr === null || investment.xirr === undefined
                      ? "-"
                      : formatPercentage(investment.xirr)}
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
