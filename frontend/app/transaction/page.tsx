"use client";

import Dashboard from "@/components/custom/Dashboard/Dashboard";
import { ImportStatementModal } from "@/components/custom/Modal/Statement/ImportStatementModal";
import { AddTransactionModal } from "@/components/custom/Modal/Transaction/AddTransactionModal";
import UpdateTransactionModal from "@/components/custom/Modal/Transaction/UpdateTransactionModal";
import TransactionFilters from "@/components/custom/Transaction/TransactionFilters";
import { TransactionsTable } from "@/components/custom/Transaction/TransactionsTable";
import { useAccounts } from "@/components/hooks/useAccounts";
import { useCategories } from "@/components/hooks/useCategories";
import {
  useDeleteTransaction,
  useTransactions,
} from "@/components/hooks/useTransactions";
import { Button } from "@/components/ui/button";
import { Transaction, TransactionQueryParams } from "@/lib/models/transaction";
import { Pencil, Plus, Trash, Upload } from "lucide-react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";
import { toast } from "sonner";

export interface TransactionFiltersState {
  accountId: number | undefined;
  categoryId: number | undefined;
  minAmount: number | undefined;
  maxAmount: number | undefined;
  dateFrom: string | undefined;
  dateTo: string | undefined;
  search: string;
}

const initialFilters: TransactionFiltersState = {
  accountId: undefined,
  categoryId: undefined,
  minAmount: undefined,
  maxAmount: undefined,
  dateFrom: undefined,
  dateTo: undefined,
  search: "",
};

export default function TransactionPage() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();
  const { data: categories = [] } = useCategories();
  const { data: accounts = [] } = useAccounts();

  // Filter state
  const [filters, setFilters] =
    useState<TransactionFiltersState>(initialFilters);

  // Pagination and sorting state
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(15);
  const [sortBy, setSortBy] = useState<string>("date");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("desc");

  // Modal state
  const [selectedRows, setSelectedRows] = useState<Set<number>>(new Set());
  const [isAddTransactionModalOpen, setIsAddTransactionModalOpen] =
    useState(false);
  const [isUpdateTransactionModalOpen, setIsUpdateTransactionModalOpen] =
    useState(false);
  const [isImportStatementModalOpen, setIsImportStatementModalOpen] =
    useState(false);
  const [transactionToUpdate, setTransactionToUpdate] =
    useState<Transaction | null>(null);

  const deleteTransactionMutation = useDeleteTransaction();

  // Build query params for transactions
  const transactionParams: TransactionQueryParams = useMemo(
    () => ({
      page: currentPage,
      page_size: pageSize,
      sort_by: sortBy,
      sort_order: sortOrder,
      account_id: filters.accountId,
      category_id: filters.categoryId,
      min_amount: filters.minAmount,
      max_amount: filters.maxAmount,
      date_from: filters.dateFrom,
      date_to: filters.dateTo,
      search: filters.search || undefined,
    }),
    [currentPage, pageSize, sortBy, sortOrder, filters]
  );

  // Get transactions using React Query
  const {
    data: paginated,
    isLoading: transactionsLoading,
    error: transactionsError,
  } = useTransactions(transactionParams);

  // Parse initial state from URL
  useEffect(() => {
    const getNum = (key: string) => {
      const val = searchParams.get(key);
      return val ? Number(val) : undefined;
    };
    setCurrentPage(getNum("page") || 1);
    setSortBy(searchParams.get("sort_by") || "date");
    setSortOrder((searchParams.get("sort_order") as "asc" | "desc") || "desc");
    setFilters({
      accountId: getNum("account_id"),
      categoryId: getNum("category_id"),
      minAmount: getNum("min_amount"),
      maxAmount: getNum("max_amount"),
      dateFrom: searchParams.get("date_from") || undefined,
      dateTo: searchParams.get("date_to") || undefined,
      search: searchParams.get("search") || "",
    });
  }, [searchParams]);

  // Update URL when state changes
  const updateUrl = useCallback(() => {
    const params = new URLSearchParams();
    if (currentPage > 1) params.set("page", String(currentPage));
    if (sortBy) params.set("sort_by", sortBy);
    if (sortOrder) params.set("sort_order", sortOrder);
    if (filters.accountId) params.set("account_id", String(filters.accountId));
    if (filters.categoryId)
      params.set("category_id", String(filters.categoryId));
    if (filters.minAmount !== undefined)
      params.set("min_amount", String(filters.minAmount));
    if (filters.maxAmount !== undefined)
      params.set("max_amount", String(filters.maxAmount));
    if (filters.dateFrom) params.set("date_from", filters.dateFrom);
    if (filters.dateTo) params.set("date_to", filters.dateTo);
    if (filters.search) params.set("search", filters.search);
    router.replace(`${pathname}?${params.toString()}`, { scroll: false });
  }, [currentPage, sortBy, sortOrder, filters, router, pathname]);

  // When any filter/sort/page changes, update URL
  useEffect(() => {
    updateUrl();
  }, [currentPage, sortBy, sortOrder, filters, updateUrl]);

  // Handlers for filter changes
  const handleFilterChange = (newFilters: Partial<TransactionFiltersState>) => {
    setFilters((prev) => ({ ...prev, ...newFilters }));
    setCurrentPage(1);
  };

  const handleClearFilters = () => {
    setFilters(initialFilters);
    setCurrentPage(1);
  };

  const handleUpdateClick = () => {
    if (selectedRows.size === 1 && paginated) {
      const id = Array.from(selectedRows)[0];
      const tx = paginated.transactions.find((t) => t.id === id) || null;
      setTransactionToUpdate(tx);
      setIsUpdateTransactionModalOpen(true);
    }
  };

  const handleAddClick = () => {
    setIsAddTransactionModalOpen(true);
  };

  const handleImportClick = () => {
    setIsImportStatementModalOpen(true);
  };

  const handleDeleteClick = async () => {
    if (selectedRows.size === 0) return;
    const ids = Array.from(selectedRows);
    let successCount = 0;
    for (const id of ids) {
      await new Promise((resolve) => {
        deleteTransactionMutation.mutate(id, {
          onSuccess: () => {
            successCount++;
            resolve(null);
          },
          onError: () => {
            console.error("Failed to delete transaction");
            resolve(null);
          },
        });
      });
    }
    if (successCount > 0) {
      toast.success(
        successCount === 1
          ? "Transaction deleted"
          : `${successCount} transactions deleted`
      );
      setSelectedRows(new Set());
    }
  };

  return (
    <Dashboard>
      <div className="flex justify-between items-center bg-card rounded-lg mb-4">
        <div className="flex justify-center items-center w-full">
          <div className="w-full">
            <TransactionFilters
              accounts={accounts}
              categories={categories}
              filters={filters}
              onFilterChange={handleFilterChange}
              onClear={handleClearFilters}
            />
          </div>
        </div>
        <div className="flex justify-end items-center gap-2 mr-4 mb-1">
          <Button variant="outline" onClick={handleImportClick}>
            <Upload className="h-4 w-4" />
            Import
          </Button>
          <Button onClick={handleAddClick}>
            <Plus className="h-4 w-4" />
            Add Transaction
          </Button>
          {selectedRows.size > 0 && (
            <>
              {selectedRows.size === 1 && (
                <Button
                  variant="default"
                  className="gap-2"
                  onClick={handleUpdateClick}
                >
                  <Pencil className="w-4 h-4" />
                  Update
                </Button>
              )}
              <Button
                variant="destructive"
                className="gap-2"
                onClick={handleDeleteClick}
                disabled={deleteTransactionMutation.status === "pending"}
              >
                <Trash className="w-4 h-4" />
                Delete
              </Button>
            </>
          )}
        </div>
      </div>

      <TransactionsTable
        selectedRows={selectedRows}
        setSelectedRows={setSelectedRows}
        transactions={paginated?.transactions || []}
        loading={transactionsLoading}
        error={transactionsError?.message || null}
        currentPage={currentPage}
        setCurrentPage={setCurrentPage}
        total={paginated?.total || 0}
        pageSize={pageSize}
        sortBy={sortBy}
        sortOrder={sortOrder}
        setSortBy={setSortBy}
        setSortOrder={setSortOrder}
      />

      <AddTransactionModal
        isOpen={isAddTransactionModalOpen}
        onOpenChange={setIsAddTransactionModalOpen}
      />

      <UpdateTransactionModal
        isOpen={isUpdateTransactionModalOpen}
        onOpenChange={setIsUpdateTransactionModalOpen}
        transaction={transactionToUpdate}
      />

      <ImportStatementModal
        isOpen={isImportStatementModalOpen}
        onOpenChange={setIsImportStatementModalOpen}
      />
    </Dashboard>
  );
}
