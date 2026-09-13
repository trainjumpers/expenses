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
import type { InsightsInvestment } from "@/lib/models/analytics";
import { formatCurrency, formatPercentage } from "@/lib/utils";
import Link from "next/link";

interface InvestmentTableProps {
  investments: InsightsInvestment[];
}

export function InvestmentTable({ investments }: InvestmentTableProps) {
  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle className="text-lg font-semibold text-muted-foreground">
          Investments
        </CardTitle>
      </CardHeader>
      <CardContent>
        {investments.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            No investment accounts with a current value.
          </p>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead className="text-right">Current</TableHead>
                <TableHead className="text-right">Contributed</TableHead>
                <TableHead className="text-right">Distributed</TableHead>
                <TableHead className="text-right">Interest</TableHead>
                <TableHead className="text-right">XIRR</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {investments.map((investment) => (
                <TableRow key={investment.account_id}>
                  <TableCell>
                    <Link
                      href={`/transaction?account_id=${investment.account_id}`}
                      className="font-medium hover:underline"
                    >
                      {investment.name}
                    </Link>
                  </TableCell>
                  <TableCell className="text-right tabular-nums">
                    {formatCurrency(investment.current_value)}
                  </TableCell>
                  <TableCell className="text-right tabular-nums">
                    {formatCurrency(investment.contributed)}
                  </TableCell>
                  <TableCell className="text-right tabular-nums">
                    {formatCurrency(investment.distributed)}
                  </TableCell>
                  <TableCell className="text-right tabular-nums">
                    {formatCurrency(investment.realized_interest)}
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
