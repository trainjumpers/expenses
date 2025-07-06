"use client";

import {
  createAccount,
  deleteAccount,
  listAccounts,
  updateAccount,
} from "@/lib/api/account";
import { Account, CreateAccountInput } from "@/lib/models/account";
import { queryKeys } from "@/lib/query-client";
import { ApiErrorType, getErrorMessage } from "@/lib/types/errors";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { useSession } from "./useSession";

export function useAccounts() {
  const { isAuthenticated } = useSession();

  return useQuery({
    queryKey: queryKeys.accounts,
    queryFn: () => listAccounts(),
    enabled: isAuthenticated,
    staleTime: 5 * 60 * 1000,
  });
}

export function useAccount(id: number) {
  const { data: accounts } = useAccounts();

  return useQuery({
    queryKey: queryKeys.account(id),
    queryFn: () => {
      const account = accounts?.find((account: Account) => account.id === id);
      if (!account) throw new Error("Account not found");
      return account;
    },
    enabled: !!id && !!accounts,
  });
}

export function useCreateAccount() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (accountData: CreateAccountInput) => createAccount(accountData),
    onMutate: async (newAccount) => {
      await queryClient.cancelQueries({ queryKey: queryKeys.accounts });
      const previousAccounts = queryClient.getQueryData<Account[]>(
        queryKeys.accounts
      );
      const tempId = Date.now();
      if (previousAccounts) {
        const optimisticAccount: Account = {
          id: tempId,
          created_by: 0,
          ...newAccount,
        };
        queryClient.setQueryData<Account[]>(queryKeys.accounts, [
          ...previousAccounts,
          optimisticAccount,
        ]);
        return { previousAccounts, tempId };
      }
      return { previousAccounts, tempId };
    },
    onError: (error: ApiErrorType, variables, context) => {
      // Rollback on error
      if (context?.previousAccounts) {
        queryClient.setQueryData(queryKeys.accounts, context.previousAccounts);
      }
      const message = getErrorMessage(error);
      console.error(message || "Failed to create account");
    },
    onSuccess: (newAccount, variables, context) => {
      queryClient.setQueryData<Account[]>(queryKeys.accounts, (old) => {
        if (!old) return [newAccount];
        const withoutOptimistic = old.filter(
          (account) => account.id !== context?.tempId
        );
        return [...withoutOptimistic, newAccount];
      });

      // Invalidate related queries
      queryClient.invalidateQueries({ queryKey: ["transactions"] });
      toast.success("Account created successfully");
    },
  });
}

// Update account mutation
export function useUpdateAccount() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: number;
      data: Partial<CreateAccountInput>;
    }) => updateAccount(id, data),
    onMutate: async ({ id, data }) => {
      await queryClient.cancelQueries({ queryKey: queryKeys.accounts });

      const previousAccounts = queryClient.getQueryData<Account[]>(
        queryKeys.accounts
      );

      if (previousAccounts) {
        queryClient.setQueryData<Account[]>(
          queryKeys.accounts,
          previousAccounts.map((account) =>
            account.id === id ? { ...account, ...data } : account
          )
        );
      }

      return { previousAccounts };
    },
    onError: (error: ApiErrorType, _, context) => {
      if (context?.previousAccounts) {
        queryClient.setQueryData(queryKeys.accounts, context.previousAccounts);
      }
      const message = getErrorMessage(error);
      console.error(message);
    },
    onSuccess: (updatedAccount) => {
      queryClient.setQueryData<Account[]>(queryKeys.accounts, (old) => {
        if (!old) return [updatedAccount];
        return old.map((account) =>
          account.id === updatedAccount.id ? updatedAccount : account
        );
      });
      queryClient.invalidateQueries({ queryKey: ["transactions"] });
      toast.success("Account updated successfully");
    },
  });
}

// Delete account mutation
export function useDeleteAccount() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => deleteAccount(id),
    onMutate: async (id) => {
      await queryClient.cancelQueries({ queryKey: queryKeys.accounts });

      const previousAccounts = queryClient.getQueryData<Account[]>(
        queryKeys.accounts
      );

      if (previousAccounts) {
        queryClient.setQueryData<Account[]>(
          queryKeys.accounts,
          previousAccounts.filter((account) => account.id !== id)
        );
      }

      return { previousAccounts };
    },
    onError: (error: ApiErrorType, variables, context) => {
      if (context?.previousAccounts) {
        queryClient.setQueryData(queryKeys.accounts, context.previousAccounts);
      }
      const message = getErrorMessage(error);
      console.error(message || "Failed to delete account");
    },
    onSuccess: (_, deletedId) => {
      queryClient.setQueryData<Account[]>(queryKeys.accounts, (old) => {
        if (!old) return [];
        return old.filter((account) => account.id !== deletedId);
      });

      queryClient.invalidateQueries({ queryKey: ["transactions"] });

      toast.success("Account deleted successfully");
    },
  });
}
