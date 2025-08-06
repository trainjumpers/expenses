import { useAccounts } from "@/components/hooks/useAccounts";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { AccountAnalyticsListResponse } from "@/lib/models/analytics";
import { formatCurrency } from "@/lib/utils";
import { ChevronRight } from "lucide-react";
import { useState } from "react";

interface AccountAnalyticsProps {
  data: AccountAnalyticsListResponse["account_analytics"];
}

// Color palette for different accounts
const accountColors = [
  "bg-purple-500",
  "bg-blue-600",
  "bg-gray-600",
  "bg-blue-400",
  "bg-pink-400",
  "bg-green-500",
  "bg-orange-500",
  "bg-indigo-500",
  "bg-teal-500",
  "bg-red-500",
];

export function AccountAnalytics({ data }: AccountAnalyticsProps) {
  const [expandedAccounts, setExpandedAccounts] = useState<Set<number>>(
    new Set()
  );
  const { data: accountsData } = useAccounts();

  if (!data || data.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Account Analytics</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8">
            <p className="text-muted-foreground">No account data found.</p>
            <p className="text-muted-foreground">
              Please add some accounts to see analytics.
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  // Calculate total balance
  const totalBalance = data.reduce(
    (sum, account) => sum + Math.abs(account.current_balance),
    0
  );
  // Calculate percentages and prepare data with account names
  const accountsWithPercentages = data
    .map((account, index) => {
      const absoluteBalance = Math.abs(account.current_balance);
      const percentage =
        totalBalance > 0 ? (absoluteBalance / totalBalance) * 100 : 0;

      // Find the account name from accounts data
      const accountInfo = accountsData?.find(
        (acc) => acc.id === account.account_id
      );
      const accountName = accountInfo?.name || `Account ${account.account_id}`;
      return {
        ...account,
        accountName,
        percentage,
        color: accountColors[index % accountColors.length],
      };
    })
    .sort((a, b) => b.percentage - a.percentage); // Sort by percentage descending

  const toggleAccountExpansion = (accountId: number) => {
    const newExpanded = new Set(expandedAccounts);
    if (newExpanded.has(accountId)) {
      newExpanded.delete(accountId);
    } else {
      newExpanded.add(accountId);
    }
    setExpandedAccounts(newExpanded);
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <span>Accounts</span>
          <span className="text-muted-foreground">•</span>
          <span>{formatCurrency(totalBalance)}</span>
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Horizontal Progress Bar */}
        <div className="space-y-4">
          <div className="h-4 bg-gray-200 rounded-full overflow-hidden">
            {accountsWithPercentages.map((account) => (
              <div
                key={account.account_id}
                className={`h-full ${account.color} inline-block`}
                style={{ width: `${account.percentage}%` }}
              />
            ))}
          </div>

          {/* Legend */}
          <div className="flex flex-wrap gap-4 text-sm">
            {accountsWithPercentages.map((account) => (
              <div key={account.account_id} className="flex items-center gap-2">
                <div className={`w-3 h-3 rounded-full ${account.color}`} />
                <span className="text-muted-foreground">
                  {account.accountName}:
                </span>
                <span className="font-medium">
                  {account.percentage.toFixed(1)}%
                </span>
              </div>
            ))}
          </div>
        </div>

        {/* Detailed Table */}
        <div className="border rounded-lg">
          <Table>
            <TableHeader>
              <TableRow className="border-b">
                <TableHead className="w-12"></TableHead>
                <TableHead className="text-left text-sm font-medium text-muted-foreground">
                  NAME
                </TableHead>
                <TableHead className="text-left text-sm font-medium text-muted-foreground">
                  WEIGHT
                </TableHead>
                <TableHead className="text-right text-sm font-medium text-muted-foreground">
                  VALUE
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {accountsWithPercentages.map((account) => (
                <TableRow key={account.account_id} className="border-b">
                  <TableCell className="w-12">
                    <button
                      onClick={() => toggleAccountExpansion(account.account_id)}
                      className="p-1 hover:bg-muted rounded transition-colors"
                    >
                      <ChevronRight
                        className={`h-4 w-4 transition-transform ${
                          expandedAccounts.has(account.account_id)
                            ? "rotate-90"
                            : ""
                        }`}
                      />
                    </button>
                  </TableCell>
                  <TableCell>
                    <span className="font-medium">{account.accountName}</span>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <div className="w-16 h-2 flex">
                        {Array.from({ length: 5 }).map((_, i) => (
                          <div
                            key={i}
                            className={`flex-1 h-full ${
                              i < Math.floor(account.percentage / 20)
                                ? account.color
                                : "bg-gray-200"
                            }`}
                            style={{
                              marginRight: i < 4 ? "1px" : "0",
                            }}
                          />
                        ))}
                      </div>
                      <span className="text-sm">
                        {account.percentage.toFixed(2)}%
                      </span>
                    </div>
                  </TableCell>
                  <TableCell className="text-right">
                    <span className="font-medium">
                      {formatCurrency(Math.abs(account.current_balance))}
                    </span>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  );
}
