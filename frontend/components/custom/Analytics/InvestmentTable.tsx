"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
import { formatPercentage, formatShortCurrency } from "@/lib/utils";
import Link from "next/link";

interface InvestmentTableProps {
  investments: InsightsInvestment[];
}

export function InvestmentTable({ investments }: InvestmentTableProps) {
  const rows = [...investments].sort(
    (a, b) => b.current_value - a.current_value
  );

  return (
    <Card className="min-w-0 overflow-hidden rounded-none border-x-0 border-t-0 shadow-none">
      <CardHeader className="px-0 pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">
          Investments
        </CardTitle>
      </CardHeader>
      <CardContent className="px-0">
        {rows.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            No investment accounts with a current value.
          </p>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead className="text-right">Current</TableHead>
                <TableHead className="hidden text-right sm:table-cell">
                  In
                </TableHead>
                <TableHead className="hidden text-right md:table-cell">
                  Out
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
                  <TableCell className="text-right tabular-nums">
                    {formatShortCurrency(investment.current_value)}
                  </TableCell>
                  <TableCell className="hidden text-right tabular-nums sm:table-cell">
                    {formatShortCurrency(investment.contributed)}
                  </TableCell>
                  <TableCell className="hidden text-right tabular-nums md:table-cell">
                    {formatShortCurrency(investment.distributed)}
                  </TableCell>
                  <TableCell className="text-right tabular-nums">
                    {investment.xirr === null || investment.xirr === undefined
                      ? "-"
                      : formatPercentage(investment.xirr)}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  );
}
