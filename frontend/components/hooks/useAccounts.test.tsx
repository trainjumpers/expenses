import type { Account } from "@/lib/models/account";
import { queryKeys } from "@/lib/query-client";
import { server } from "@/test/msw/server";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { HttpResponse, delay, http } from "msw";
import type { ReactNode } from "react";
import { toast } from "sonner";
import { afterAll, describe, expect, it, vi } from "vitest";

import {
  useAccount,
  useCreateAccount,
  useDeleteAccount,
  useUpdateAccount,
} from "./useAccounts";

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
    refresh: vi.fn(),
    back: vi.fn(),
    forward: vi.fn(),
    prefetch: vi.fn(),
  }),
  usePathname: () => "/",
  useSearchParams: () => new URLSearchParams(),
}));

const existing: Account = {
  id: 1,
  name: "HDFC Savings",
  bank_type: "hdfc",
  currency: "inr",
  created_by: 1,
};

function setup() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  queryClient.setQueryData<Account[]>(queryKeys.accounts, [existing]);

  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );

  return { queryClient, wrapper };
}

function accounts(queryClient: QueryClient) {
  return queryClient.getQueryData<Account[]>(queryKeys.accounts);
}

describe("account mutations", () => {
  const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});

  afterAll(() => {
    consoleError.mockRestore();
  });

  it("adds an optimistic account and reconciles it with the server row", async () => {
    server.use(
      http.post("*/api/v1/account", async ({ request }) => {
        const body = (await request.json()) as { name: string };
        await delay(50);
        return HttpResponse.json(
          {
            message: "Account created successfully",
            data: { ...existing, id: 7, name: body.name },
          },
          { status: 201 }
        );
      })
    );

    const { queryClient, wrapper } = setup();
    const { result } = renderHook(() => useCreateAccount(), { wrapper });

    act(() => {
      result.current.mutate({
        name: "ICICI",
        bank_type: "icici",
        currency: "inr",
      });
    });

    await waitFor(() => expect(accounts(queryClient)).toHaveLength(2));

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(accounts(queryClient)).toEqual([
      existing,
      expect.objectContaining({ id: 7, name: "ICICI" }),
    ]);
    expect(toast.success).toHaveBeenCalledWith("Account created successfully");
  });

  it("rolls the optimistic account back on failure", async () => {
    server.use(
      http.post("*/api/v1/account", () =>
        HttpResponse.json({ error: "duplicate" }, { status: 409 })
      )
    );

    const { queryClient, wrapper } = setup();
    const { result } = renderHook(() => useCreateAccount(), { wrapper });

    act(() => {
      result.current.mutate({
        name: "ICICI",
        bank_type: "icici",
        currency: "inr",
      });
    });

    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(accounts(queryClient)).toEqual([existing]);
    expect(consoleError).toHaveBeenCalled();
  });

  it("applies account updates optimistically", async () => {
    server.use(
      http.patch("*/api/v1/account/1", async ({ request }) => {
        const body = (await request.json()) as Partial<Account>;
        return HttpResponse.json({
          message: "Account updated successfully",
          data: { ...existing, ...body },
        });
      })
    );

    const { queryClient, wrapper } = setup();
    const { result } = renderHook(() => useUpdateAccount(), { wrapper });

    act(() => {
      result.current.mutate({ id: 1, data: { name: "HDFC Salary" } });
    });

    await waitFor(() =>
      expect(accounts(queryClient)?.[0]?.name).toBe("HDFC Salary")
    );
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(toast.success).toHaveBeenCalledWith("Account updated successfully");
  });

  it("removes the deleted account", async () => {
    server.use(
      http.delete(
        "*/api/v1/account/1",
        () => new HttpResponse(null, { status: 204 })
      )
    );

    const { queryClient, wrapper } = setup();
    const { result } = renderHook(() => useDeleteAccount(), { wrapper });

    act(() => {
      result.current.mutate(1);
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(accounts(queryClient)).toEqual([]);
    expect(toast.success).toHaveBeenCalledWith("Account deleted successfully");
  });

  it("explains a blocked deletion and restores the account", async () => {
    server.use(
      http.delete("*/api/v1/account/1", () =>
        HttpResponse.json(
          { message: "cannot delete account with existing transactions" },
          { status: 409 }
        )
      )
    );

    const { queryClient, wrapper } = setup();
    const { result } = renderHook(() => useDeleteAccount(), { wrapper });

    act(() => {
      result.current.mutate(1);
    });

    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(toast.error).toHaveBeenCalledWith(
      "Cannot delete account with existing transactions",
      { duration: 2000 }
    );
    expect(accounts(queryClient)).toEqual([existing]);
  });

  it("resolves a cached account by id", async () => {
    const { wrapper } = setup();
    const { result } = renderHook(() => useAccount(1), { wrapper });

    await waitFor(() => expect(result.current.data?.name).toBe("HDFC Savings"));
  });

  it("fails for an unknown account id", async () => {
    const { wrapper } = setup();
    const { result } = renderHook(() => useAccount(99), { wrapper });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toEqual(new Error("Account not found"));
  });

  it("rolls account updates back on failure", async () => {
    server.use(
      http.patch("*/api/v1/account/1", () =>
        HttpResponse.json({ error: "nope" }, { status: 500 })
      )
    );
    const { queryClient, wrapper } = setup();
    const { result } = renderHook(() => useUpdateAccount(), { wrapper });

    act(() => {
      result.current.mutate({ id: 1, data: { name: "Renamed" } });
    });

    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(accounts(queryClient)).toEqual([existing]);
    expect(consoleError).toHaveBeenCalledWith("nope");
  });
});
