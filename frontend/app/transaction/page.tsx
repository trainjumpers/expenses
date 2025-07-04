"use client";

import Dashboard from "@/components/custom/Dashboard/Dashboard";
import { AddTransactionModal } from "@/components/custom/Modal/Transaction/AddTransactionModal";
import UpdateTransactionModal from "@/components/custom/Modal/Transaction/UpdateTransactionModal";
import { useAccounts } from "@/components/custom/Provider/AccountProvider";
import { useCategories } from "@/components/custom/Provider/CategoryProvider";
import TransactionFilters from "@/components/custom/Transaction/TransactionFilters";
import { TransactionsTable } from "@/components/custom/Transaction/TransactionsTable";
import { Button } from "@/components/ui/button";
import { getAllTransactions } from "@/lib/api/transaction";
import {
  PaginatedTransactionsResponse,
  Transaction,
  TransactionQueryParams,
} from "@/lib/models/transaction";
import { Pencil, Plus, Trash } from "lucide-react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useCallback, useEffect, useState } from "react";

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
  const { read: categories } = useCategories();
  const { read: accounts } = useAccounts();

  // Filter state
  const [filters, setFilters] =
    useState<TransactionFiltersState>(initialFilters);

  // Pagination and sorting state
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(15);
  const [sortBy, setSortBy] = useState<string>("date");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("desc");

  // Paginated response state
  const [paginated, setPaginated] =
    useState<PaginatedTransactionsResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedRows, setSelectedRows] = useState<Set<number>>(new Set());
  const [isAddTransactionModalOpen, setIsAddTransactionModalOpen] =
    useState(false);
  const [isUpdateTransactionModalOpen, setIsUpdateTransactionModalOpen] =
    useState(false);
  const [transactionToUpdate, setTransactionToUpdate] =
    useState<Transaction | null>(null);

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

  // Fetch transactions when URL changes
  useEffect(() => {
    fetchTransactions();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchParams]);

  const fetchTransactions = async () => {
    setLoading(true);
    setError(null);
    try {
      const params: TransactionQueryParams = {
        page: Number(searchParams.get("page")) || 1,
        page_size: pageSize,
        sort_by: searchParams.get("sort_by") || "date",
        sort_order:
          (searchParams.get("sort_order") as "asc" | "desc") || "desc",
        account_id: searchParams.get("account_id")
          ? Number(searchParams.get("account_id"))
          : undefined,
        category_id: searchParams.get("category_id")
          ? Number(searchParams.get("category_id"))
          : undefined,
        min_amount: searchParams.get("min_amount")
          ? Number(searchParams.get("min_amount"))
          : undefined,
        max_amount: searchParams.get("max_amount")
          ? Number(searchParams.get("max_amount"))
          : undefined,
        date_from: searchParams.get("date_from") || undefined,
        date_to: searchParams.get("date_to") || undefined,
        search: searchParams.get("search") || undefined,
      };
      const data = await getAllTransactions(params);
      setPaginated(data);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to fetch transactions"
      );
      setPaginated(null);
    } finally {
      setLoading(false);
    }
  };

  const optimisticallyUpdateTransaction = (updatedTx: Transaction) => {
    setPaginated((prev) => {
      if (!prev) return null;
      return {
        ...prev,
        transactions: prev.transactions.map((tx) =>
          tx.id === updatedTx.id ? updatedTx : tx
        ),
      };
    });
  };

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

  return (
    <Dashboard>
      <div className="flex justify-between items-center bg-card rounded-lg mb-4">
        <div className="flex justify-center items-center w-full">
          <div className="w-full">
            <TransactionFilters
              accounts={accounts()}
              categories={categories()}
              filters={filters}
              onFilterChange={handleFilterChange}
              onClear={handleClearFilters}
            />
          </div>
        </div>
        <div className="flex justify-end items-center gap-2 mr-4 mb-1">
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
              <Button variant="destructive" className="gap-2">
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
        loading={loading}
        error={error}
        currentPage={currentPage}
        setCurrentPage={setCurrentPage}
        total={paginated?.total || 0}
        pageSize={pageSize}
        sortBy={sortBy}
        sortOrder={sortOrder}
        setSortBy={setSortBy}
        setSortOrder={setSortOrder}
        onTransactionUpdate={optimisticallyUpdateTransaction}
      />
      <AddTransactionModal
        isOpen={isAddTransactionModalOpen}
        onOpenChange={setIsAddTransactionModalOpen}
        onTransactionAdded={fetchTransactions}
        isRefreshing={loading}
      />
      <UpdateTransactionModal
        isOpen={isUpdateTransactionModalOpen}
        onOpenChange={setIsUpdateTransactionModalOpen}
        transaction={transactionToUpdate}
        onTransactionUpdated={fetchTransactions}
        isRefreshing={loading}
      />
    </Dashboard>
  );
}
