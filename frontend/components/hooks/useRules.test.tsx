import type { Rule } from "@/lib/models/rule";
import { ConditionLogic } from "@/lib/models/rule";
import { server } from "@/test/msw/server";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import type { ReactNode } from "react";
import { toast } from "sonner";
import { describe, expect, it, vi } from "vitest";

import {
  useCreateRule,
  useDeleteRule,
  useExecuteRules,
  useRules,
  useUpdateRule,
} from "./useRules";

const rule: Rule = {
  id: 5,
  name: "Coffee rule",
  description: "Matches coffee",
  condition_logic: ConditionLogic.AND,
  effective_from: new Date(0).toISOString(),
  created_by: 1,
};

const input = {
  rule: {
    name: "Coffee rule",
    description: "Matches coffee",
    condition_logic: ConditionLogic.AND,
    effective_from: new Date(0).toISOString(),
  },
  actions: [{ action_type: "category" as const, action_value: "1" }],
  conditions: [
    {
      condition_type: "name" as const,
      condition_operator: "contains" as const,
      condition_value: "coffee",
    },
  ],
};

function setup() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
  return { queryClient, wrapper };
}

describe("rule hooks", () => {
  it("loads a rule page with the provided query", async () => {
    let url = "";
    server.use(
      http.get("*/api/v1/rule", ({ request }) => {
        url = request.url;
        return HttpResponse.json({
          data: { rules: [rule], total: 1, page: 1, page_size: 5 },
        });
      })
    );

    const { wrapper } = setup();
    const { result } = renderHook(
      () => useRules({ page: 1, page_size: 5, search: "coffee" }),
      { wrapper }
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    const params = new URL(url).searchParams;
    expect(params.get("page")).toBe("1");
    expect(params.get("page_size")).toBe("5");
    expect(params.get("search")).toBe("coffee");
    expect(result.current.data?.rules).toEqual([rule]);
  });

  it("creates a rule and refreshes the list", async () => {
    server.use(
      http.post("*/api/v1/rule", () =>
        HttpResponse.json({ data: rule }, { status: 201 })
      )
    );

    const { queryClient, wrapper } = setup();
    const invalidate = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useCreateRule(), { wrapper });

    act(() => {
      result.current.mutate(input);
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(invalidate).toHaveBeenCalledWith({ queryKey: ["rules"] });
  });

  it("updates a rule with the id from the payload", async () => {
    let path = "";
    server.use(
      http.patch("*/api/v1/rule/5", ({ request }) => {
        path = new URL(request.url).pathname;
        return HttpResponse.json({ data: rule });
      })
    );

    const { queryClient, wrapper } = setup();
    const invalidate = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useUpdateRule(), { wrapper });

    act(() => {
      result.current.mutate({ id: 5, input: { name: "Renamed rule" } });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(path).toBe("/api/v1/rule/5");
    expect(invalidate).toHaveBeenCalledWith({ queryKey: ["rules"] });
  });

  it("deletes a rule and refreshes the list", async () => {
    let deleted: number | null = null;
    server.use(
      http.delete("*/api/v1/rule/5", () => {
        deleted = 5;
        return new HttpResponse(null, { status: 204 });
      })
    );

    const { queryClient, wrapper } = setup();
    const invalidate = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useDeleteRule(), { wrapper });

    act(() => {
      result.current.mutate(5);
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(deleted).toBe(5);
    expect(invalidate).toHaveBeenCalledWith({ queryKey: ["rules"] });
  });

  it("announces a rule execution", async () => {
    server.use(
      http.post("*/api/v1/rule/execute", () =>
        HttpResponse.json({
          data: { modified: [], processed_transactions: [] },
        })
      )
    );

    const { wrapper } = setup();
    const { result } = renderHook(() => useExecuteRules(), { wrapper });

    act(() => {
      result.current.mutate({ transaction_ids: [1, 2] });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(toast.success).toHaveBeenCalledWith(
      "Rule execution started in the background."
    );
  });

  it("surfaces the failure reason when executing rules", async () => {
    server.use(
      http.post("*/api/v1/rule/execute", () =>
        HttpResponse.json(
          { error: "Rules are already running" },
          { status: 409 }
        )
      )
    );

    const { wrapper } = setup();
    const { result } = renderHook(() => useExecuteRules(), { wrapper });

    act(() => {
      result.current.mutate(undefined);
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(toast.error).toHaveBeenCalledWith("Rules are already running");
  });
});
