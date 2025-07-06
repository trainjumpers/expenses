"use client";

import { useAccounts } from "@/components/hooks/useAccounts";
import { useCategories } from "@/components/hooks/useCategories";
import { useUpdateTransaction } from "@/components/hooks/useTransactions";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Icon, IconName } from "@/components/ui/icon-picker";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Transaction } from "@/lib/models/transaction";
import { ArrowDown, ArrowUp, ArrowUpDown } from "lucide-react";
import { useState } from "react";

import TablePagination from "./TablePagination";
import TransactionTableRow from "./TransactionTableRow";
import { TransactionsTableSkeleton } from "./TransactionsTableSkeleton";

interface TransactionsTableProps {
  selectedRows: Set<number>;
  setSelectedRows: React.Dispatch<React.SetStateAction<Set<number>>>;
  transactions: Transaction[];
  loading: boolean;
  error: string | null;
  currentPage: number;
  setCurrentPage: (page: number) => void;
  total: number;
  pageSize: number;
  sortBy: string;
  sortOrder: "asc" | "desc";
  setSortBy: (key: string) => void;
  setSortOrder: (order: "asc" | "desc") => void;
}

export function TransactionsTable({
  selectedRows,
  setSelectedRows,
  transactions,
  loading,
  error,
  currentPage,
  setCurrentPage,
  total,
  pageSize,
  sortBy,
  sortOrder,
  setSortBy,
  setSortOrder,
}: TransactionsTableProps) {
  const [editing, setEditing] = useState<{
    id: number;
    field: "category" | "account" | null;
  }>({ id: -1, field: null });

  const { data: categories = [] } = useCategories();
  const { data: accounts = [] } = useAccounts();
  const updateTransactionMutation = useUpdateTransaction();

  const totalPages = Math.ceil(total / pageSize);
  const currentData = transactions;

  const handleSort = (key: keyof Transaction) => {
    if (sortBy === key) {
      setSortOrder(sortOrder === "asc" ? "desc" : "asc");
    } else {
      setSortBy(key);
      setSortOrder("asc");
    }
  };

  const getSortIcon = (key: keyof Transaction) => {
    if (sortBy !== key) {
      return <ArrowUpDown className="h-4 w-4" />;
    }
    return sortOrder === "asc" ? (
      <ArrowUp className="h-4 w-4" />
    ) : (
      <ArrowDown className="h-4 w-4" />
    );
  };

  const toggleRowSelection = (id: number) => {
    setSelectedRows((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(id)) {
        newSet.delete(id);
      } else {
        newSet.add(id);
      }
      return newSet;
    });
  };

  const toggleAllRows = () => {
    if (selectedRows.size === currentData.length) {
      setSelectedRows(new Set());
    } else {
      setSelectedRows(new Set(currentData.map((row: Transaction) => row.id)));
    }
  };

  const getAccountName = (id: number) => {
    return accounts.find((acc) => acc.id === id)?.name || "-";
  };

  const renderCategoryPills = (ids: number[]) => {
    return categories
      .filter((cat) => ids.includes(cat.id))
      .map((cat) => (
        <span
          key={cat.id}
          className="inline-flex items-center gap-1 bg-muted px-2 py-1 rounded-full text-xs font-medium mr-1 border border-border"
          style={{ verticalAlign: "middle" }}
        >
          {cat.icon && (
            <Icon name={cat.icon as IconName} className="w-4 h-4 mr-1" />
          )}
          {cat.name}
        </span>
      ));
  };

  // Update helper using React Query mutation
  const handleUpdate = async (
    originalTransaction: Transaction,
    updatedTransaction: Transaction
  ) => {
    const diff: Partial<Pick<Transaction, "account_id" | "category_ids">> = {};
    if (originalTransaction.account_id !== updatedTransaction.account_id) {
      diff.account_id = updatedTransaction.account_id;
    }
    if (
      JSON.stringify(originalTransaction.category_ids.sort()) !==
      JSON.stringify(updatedTransaction.category_ids.sort())
    ) {
      diff.category_ids = updatedTransaction.category_ids;
    }

    if (Object.keys(diff).length === 0) return;

    updateTransactionMutation.mutate({ id: updatedTransaction.id, data: diff });
  };

  if (loading) {
    return <TransactionsTableSkeleton />;
  }

  if (error) {
    return (
      <div className="w-full h-[700px] flex items-center justify-center bg-card">
        <div className="text-destructive">{error}</div>
      </div>
    );
  }

  return (
    <div className="w-full">
      <div className="border border-border bg-card rounded-t-md">
        <div className="h-[700px] flex flex-col">
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-muted/50">
                <TableHead className="w-[50px] py-4 px-6">
                  <Checkbox
                    checked={
                      selectedRows.size === currentData.length &&
                      currentData.length > 0
                    }
                    onCheckedChange={toggleAllRows}
                    aria-label="Select all"
                    className="translate-y-[2px]"
                  />
                </TableHead>
                <TableHead className="text-muted-foreground py-4 px-6 text-center">
                  <Button
                    variant="ghost"
                    onClick={() => handleSort("name")}
                    className="flex items-center justify-center gap-1 hover:bg-muted w-full"
                  >
                    Name
                    {getSortIcon("name")}
                  </Button>
                </TableHead>
                <TableHead className="text-muted-foreground py-4 px-6 text-center">
                  <Button
                    variant="ghost"
                    onClick={() => handleSort("description")}
                    className="flex items-center justify-center gap-1 hover:bg-muted w-full"
                  >
                    Description
                    {getSortIcon("description")}
                  </Button>
                </TableHead>
                <TableHead className="text-muted-foreground py-4 px-6 text-center">
                  Category
                </TableHead>
                <TableHead className="text-right text-muted-foreground py-4 px-6">
                  <Button
                    variant="ghost"
                    onClick={() => handleSort("amount")}
                    className="flex items-center gap-1 ml-auto hover:bg-muted"
                  >
                    Amount
                    {getSortIcon("amount")}
                  </Button>
                </TableHead>
                <TableHead className="text-muted-foreground py-4 px-6 text-center">
                  Account
                </TableHead>
                <TableHead className="text-muted-foreground py-4 px-6 text-center">
                  <Button
                    variant="ghost"
                    onClick={() => handleSort("date")}
                    className="flex items-center justify-center gap-1 hover:bg-muted w-full"
                  >
                    Date
                    {getSortIcon("date")}
                  </Button>
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {currentData.length > 0 ? (
                currentData.map((transaction: Transaction) => (
                  <TransactionTableRow
                    key={transaction.id}
                    transaction={transaction}
                    selected={selectedRows.has(transaction.id)}
                    onSelect={() => toggleRowSelection(transaction.id)}
                    accounts={accounts}
                    categories={categories}
                    editing={editing}
                    setEditing={setEditing}
                    handleUpdate={handleUpdate}
                    renderCategoryPills={renderCategoryPills}
                    getAccountName={getAccountName}
                  />
                ))
              ) : (
                <TableRow>
                  <TableCell
                    colSpan={7}
                    className="h-[500px] text-center text-muted-foreground"
                  >
                    No transactions found
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>
      <TablePagination
        currentPage={currentPage}
        totalPages={totalPages}
        setCurrentPage={setCurrentPage}
      />
    </div>
  );
}
