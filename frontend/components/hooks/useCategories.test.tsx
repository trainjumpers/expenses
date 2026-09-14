import type { Category } from "@/lib/models/category";
import { queryKeys } from "@/lib/query-client";
import { server } from "@/test/msw/server";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { HttpResponse, delay, http } from "msw";
import type { ReactNode } from "react";
import { toast } from "sonner";
import { afterAll, describe, expect, it, vi } from "vitest";

import {
  useCreateCategory,
  useDeleteCategory,
  useUpdateCategory,
} from "./useCategories";

const existing: Category = { id: 1, name: "Food", created_by: 1 };

function setup() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  queryClient.setQueryData<Category[]>(queryKeys.categories, [existing]);

  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );

  return { queryClient, wrapper };
}

function categories(queryClient: QueryClient) {
  return queryClient.getQueryData<Category[]>(queryKeys.categories);
}

describe("category mutations", () => {
  const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});

  afterAll(() => {
    consoleError.mockRestore();
  });

  it("adds an optimistic category and reconciles it with the server row", async () => {
    server.use(
      http.post("*/api/v1/category", async ({ request }) => {
        const body = (await request.json()) as { name: string };
        await delay(50);
        return HttpResponse.json(
          {
            message: "Category created successfully",
            data: { ...existing, id: 9, name: body.name },
          },
          { status: 201 }
        );
      })
    );

    const { queryClient, wrapper } = setup();
    const { result } = renderHook(() => useCreateCategory(), { wrapper });

    act(() => {
      result.current.mutate({ name: "Travel" });
    });

    await waitFor(() => {
      expect(categories(queryClient)).toHaveLength(2);
      expect(categories(queryClient)?.[1]?.name).toBe("Travel");
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(categories(queryClient)).toEqual([
      existing,
      expect.objectContaining({ id: 9, name: "Travel" }),
    ]);
    expect(toast.success).toHaveBeenCalledWith("Category created successfully");
  });

  it("rolls the optimistic category back on failure", async () => {
    server.use(
      http.post("*/api/v1/category", () =>
        HttpResponse.json({ error: "duplicate" }, { status: 409 })
      )
    );

    const { queryClient, wrapper } = setup();
    const { result } = renderHook(() => useCreateCategory(), { wrapper });

    act(() => {
      result.current.mutate({ name: "Travel" });
    });

    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(categories(queryClient)).toEqual([existing]);
    expect(consoleError).toHaveBeenCalled();
  });

  it("applies category updates optimistically", async () => {
    server.use(
      http.patch("*/api/v1/category/1", async ({ request }) => {
        const body = (await request.json()) as Partial<Category>;
        return HttpResponse.json({
          message: "Category updated successfully",
          data: { ...existing, ...body },
        });
      })
    );

    const { queryClient, wrapper } = setup();
    const { result } = renderHook(() => useUpdateCategory(), { wrapper });

    act(() => {
      result.current.mutate({ id: 1, data: { name: "Groceries" } });
    });

    await waitFor(() =>
      expect(categories(queryClient)?.[0]?.name).toBe("Groceries")
    );
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(toast.success).toHaveBeenCalledWith("Category updated successfully");
  });

  it("removes the deleted category", async () => {
    server.use(
      http.delete(
        "*/api/v1/category/1",
        () => new HttpResponse(null, { status: 204 })
      )
    );

    const { queryClient, wrapper } = setup();
    const { result } = renderHook(() => useDeleteCategory(), { wrapper });

    act(() => {
      result.current.mutate(1);
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(categories(queryClient)).toEqual([]);
    expect(toast.success).toHaveBeenCalledWith("Category deleted successfully");
  });
});
