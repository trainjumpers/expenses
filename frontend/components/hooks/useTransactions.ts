"use client";

import {
  createTransaction,
  deleteTransaction,
  getAllTransactions,
  updateTransaction,
} from "@/lib/api/transaction";
import type {
  CreateTransaction,
  PaginatedTransactionsResponse,
  Transaction,
  TransactionQueryParams,
} from "@/lib/models/transaction";
import { queryKeys } from "@/lib/query-client";
import type { ApiErrorType} from "@/lib/types/errors";
import { getErrorMessage } from "@/lib/types/errors";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { useSession } from "./useSession";

export function useTransactions(params: TransactionQueryParams = {}) {
  const { isAuthenticated } = useSession();

  return useQuery({
    queryKey: queryKeys.transactions(params as Record<string, unknown>),
    queryFn: () => getAllTransactions(params),
    enabled: isAuthenticated,
    staleTime: 2 * 60 * 1000,
    placeholderData: (previousData) => previousData,
  });
}

export function useTransaction(id: number) {
  const queryClient = useQueryClient();

  return useQuery({
    queryKey: queryKeys.transaction(id),
    queryFn: () => {
      const allTransactionsQueries =
        queryClient.getQueriesData<PaginatedTransactionsResponse>({
          queryKey: ["transactions"],
        });

      for (const [, data] of allTransactionsQueries) {
        if (data?.transactions) {
          const transaction = data.transactions.find((t) => t.id === id);
          if (transaction) return transaction;
        }
      }

      throw new Error("Transaction not found");
    },
    enabled: !!id,
  });
}

export function useCreateTransaction() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (transactionData: CreateTransaction) =>
      createTransaction(transactionData),
    onMutate: async (newTransaction) => {
      await queryClient.cancelQueries({ queryKey: ["transactions"] });

      const optimisticTransaction: Transaction = {
        id: Date.now(),
        date: newTransaction.date,
        name: newTransaction.name,
        description: newTransaction.description || null,
        amount: newTransaction.amount,
        category_ids: newTransaction.category_ids,
        account_id: newTransaction.account_id,
      };

      queryClient.setQueriesData<PaginatedTransactionsResponse>(
        { queryKey: ["transactions"] },
        (old) => {
          if (!old) return old;
          return {
            ...old,
            transactions: [optimisticTransaction, ...old.transactions],
            total: old.total + 1,
          };
        }
      );

      return { optimisticTransaction };
    },
    onError: (error: ApiErrorType, variables, context) => {
      if (context?.optimisticTransaction) {
        queryClient.setQueriesData<PaginatedTransactionsResponse>(
          { queryKey: ["transactions"] },
          (old) => {
            if (!old) return old;
            return {
              ...old,
              transactions: Array.isArray(old.transactions)
                ? old.transactions.filter(
                    (t) => t.id !== context.optimisticTransaction.id
                  )
                : [],
              total: old.total - 1,
            };
          }
        );
      }
      const message = getErrorMessage(error);
      console.error(message || "Failed to create transaction");
    },
    onSuccess: (newTransaction) => {
      queryClient.setQueriesData<PaginatedTransactionsResponse>(
        { queryKey: ["transactions"] },
        (old) => {
          if (!old) return old;
          return {
            ...old,
            transactions: old.transactions.map((t) =>
              t.id === newTransaction.id ? newTransaction : t
            ),
          };
        }
      );

      queryClient.invalidateQueries({ queryKey: ["transactions"] });

      toast.success("Transaction created successfully");
    },
  });
}

export function useUpdateTransaction() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: number;
      data: Partial<CreateTransaction>;
    }) => updateTransaction(id, data),
    onMutate: async ({ id, data }) => {
      await queryClient.cancelQueries({ queryKey: ["transactions"] });

      queryClient.setQueriesData<PaginatedTransactionsResponse>(
        { queryKey: ["transactions"] },
        (old) => {
          if (!old) return old;
          return {
            ...old,
            transactions: Array.isArray(old.transactions)
              ? old.transactions.map((transaction) =>
                  transaction.id === id
                    ? { ...transaction, ...data }
                    : transaction
                )
              : [],
          };
        }
      );

      queryClient.setQueryData(
        queryKeys.transaction(id),
        (old: Transaction | undefined) => {
          if (!old) return old;
          return { ...old, ...data };
        }
      );
    },
    onError: (error: ApiErrorType, { id }) => {
      queryClient.invalidateQueries({ queryKey: ["transactions"] });
      queryClient.invalidateQueries({ queryKey: queryKeys.transaction(id) });
      const message = getErrorMessage(error);
      console.error(message || "Failed to update transaction");
    },
    onSuccess: (updatedTransaction) => {
      queryClient.setQueriesData<PaginatedTransactionsResponse>(
        { queryKey: ["transactions"] },
        (old) => {
          if (!old) return old;
          return {
            ...old,
            transactions: Array.isArray(old.transactions)
              ? old.transactions.map((t) =>
                  t.id === updatedTransaction.id ? updatedTransaction : t
                )
              : [],
          };
        }
      );

      queryClient.setQueryData(
        queryKeys.transaction(updatedTransaction.id),
        updatedTransaction
      );

      toast.success("Transaction updated successfully");
    },
  });
}

export function useDeleteTransaction() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteTransaction(id),
    onSuccess: (_, id) => {
      // Remove the transaction from all cached pages
      queryClient.setQueriesData<PaginatedTransactionsResponse>(
        { queryKey: ["transactions"] },
        (old) => {
          if (!old) return old;
          return {
            ...old,
            transactions: Array.isArray(old.transactions)
              ? old.transactions.filter((t) => t.id !== id)
              : [],
          };
        }
      );
    },
  });
}
