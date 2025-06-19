"use client";

import {
  createAccount,
  deleteAccount,
  listAccounts,
  updateAccount,
} from "@/lib/api/account";
import { Account, CreateAccountInput } from "@/lib/models/account";
import { createResource } from "@/lib/utils/suspense";
import React, {
  ReactNode,
  createContext,
  useContext,
  useEffect,
  useState,
} from "react";

export type AccountResource = {
  read: () => Account[];
  refresh: () => void;
  create: (account: CreateAccountInput) => Promise<Account>;
  update: (
    id: number,
    account: Partial<CreateAccountInput>
  ) => Promise<Account>;
  delete: (id: number) => Promise<void>;
};

const AccountContext = createContext<AccountResource | null>(null);

export const AccountProvider = ({ children }: { children: ReactNode }) => {
  const [abortController, setAbortController] =
    useState<AbortController | null>(null);
  const [resource, setResource] = useState(() => {
    const controller = new AbortController();
    setAbortController(controller);
    return createResource<Account[]>(listAccounts, controller.signal);
  });

  const refresh = () => {
    if (abortController) {
      abortController.abort();
    }
    const controller = new AbortController();
    setAbortController(controller);
    const newResource = createResource<Account[]>(
      listAccounts,
      controller.signal
    );
    setResource(newResource);
  };

  const create = async (account: CreateAccountInput) => {
    try {
      const newAccount = await createAccount(account);
      return newAccount;
    } finally {
      refresh();
    }
  };
  const update = async (id: number, account: Partial<CreateAccountInput>) => {
    try {
      const updated = await updateAccount(id, account);
      return updated;
    } finally {
      refresh();
    }
  };
  const del = async (id: number) => {
    try {
      await deleteAccount(id);
    } finally {
      refresh();
    }
  };

  useEffect(() => {
    refresh();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const read = () => {
    if (!resource) throw new Error("Resource not found");
    const result = resource.read();
    if (!result) return [];
    return result;
  };

  const value: AccountResource = {
    read,
    refresh,
    create,
    update,
    delete: del,
  };

  return (
    <AccountContext.Provider value={value}>{children}</AccountContext.Provider>
  );
};

export function useAccounts() {
  const ctx = useContext(AccountContext);
  if (!ctx)
    throw new Error("useAccounts must be used within an AccountProvider");
  return ctx;
}
