import type {
  PaginatedTransactionsResponse,
  Transaction,
} from "@/lib/models/transaction";
import { queryKeys } from "@/lib/query-client";
import { server } from "@/test/msw/server";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { HttpResponse, delay, http } from "msw";
import type { ReactNode } from "react";
import { toast } from "sonner";
import { afterAll, describe, expect, it, vi } from "vitest";

import {
  useCreateTransaction,
  useDeleteTransaction,
  useUpdateTransaction,
} from "./useTransactions";

const listKey = queryKeys.transactions({ page: 1 });

const existing: Transaction = {
  id: 1,
  date: "2026-08-01T00:00:00.000Z",
  name: "Coffee",
  description: null,
  amount: 250,
  category_ids: [],
  account_id: 1,
};

function setup() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  queryClient.setQueryData<PaginatedTransactionsResponse>(listKey, {
    transactions: [existing],
    total: 1,
    page: 1,
    page_size: 15,
  });

  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );

  return { queryClient, wrapper };
}

function listData(queryClient: QueryClient) {
  return queryClient.getQueryData<PaginatedTransactionsResponse>(listKey);
}

describe("transaction mutations", () => {
  const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});

  afterAll(() => {
    consoleError.mockRestore();
  });

  it("adds an optimistic row and reconciles it with the server result", async () => {
    server.use(
      http.post("*/api/v1/transaction", async ({ request }) => {
        const body = (await request.json()) as {
          name: string;
          amount: number;
        };
        await delay(50);
        return HttpResponse.json(
          {
            message: "Transaction created successfully",
            data: { ...existing, id: 99, name: body.name, amount: body.amount },
          },
          { status: 201 }
        );
      })
    );

    const { queryClient, wrapper } = setup();
    const { result } = renderHook(() => useCreateTransaction(), { wrapper });

    act(() => {
      result.current.mutate({
        name: "Lunch",
        amount: 12,
        date: "2026-09-01",
        category_ids: [],
        account_id: 1,
      });
    });

    await waitFor(() => {
      expect(listData(queryClient)?.transactions[0]?.name).toBe("Lunch");
      expect(listData(queryClient)?.total).toBe(2);
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(listData(queryClient)?.transactions).toEqual([
      expect.objectContaining({ id: 99, name: "Lunch", amount: 12 }),
      existing,
    ]);
    expect(listData(queryClient)?.total).toBe(2);
    expect(toast.success).toHaveBeenCalledWith(
      "Transaction created successfully"
    );
  });

  it("rolls the optimistic row back when creation fails", async () => {
    server.use(
      http.post("*/api/v1/transaction", () =>
        HttpResponse.json({ error: "invalid transaction" }, { status: 400 })
      )
    );

    const { queryClient, wrapper } = setup();
    const { result } = renderHook(() => useCreateTransaction(), { wrapper });

    act(() => {
      result.current.mutate({
        name: "Lunch",
        amount: 12,
        date: "2026-09-01",
        category_ids: [],
        account_id: 1,
      });
    });

    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(listData(queryClient)?.transactions).toEqual([existing]);
    expect(listData(queryClient)?.total).toBe(1);
    expect(toast.success).not.toHaveBeenCalled();
  });

  it("applies updates optimistically and keeps the server result", async () => {
    server.use(
      http.patch("*/api/v1/transaction/1", async ({ request }) => {
        const body = (await request.json()) as Partial<Transaction>;
        return HttpResponse.json({
          message: "Transaction updated successfully",
          data: { ...existing, ...body },
        });
      })
    );

    const { queryClient, wrapper } = setup();
    const { result } = renderHook(() => useUpdateTransaction(), { wrapper });

    act(() => {
      result.current.mutate({ id: 1, data: { name: "Flat White" } });
    });

    await waitFor(() =>
      expect(listData(queryClient)?.transactions[0]?.name).toBe("Flat White")
    );
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(listData(queryClient)?.transactions[0]?.name).toBe("Flat White");
    expect(toast.success).toHaveBeenCalledWith(
      "Transaction updated successfully"
    );
  });

  it("removes the deleted transaction from the cache", async () => {
    server.use(
      http.delete(
        "*/api/v1/transaction/1",
        () => new HttpResponse(null, { status: 204 })
      )
    );

    const { queryClient, wrapper } = setup();
    const { result } = renderHook(() => useDeleteTransaction(), { wrapper });

    act(() => {
      result.current.mutate(1);
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(listData(queryClient)?.transactions).toEqual([]);
  });
});
