import type { Statement } from "@/lib/models/statement";
import { server } from "@/test/msw/server";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import type { ReactNode } from "react";
import { toast } from "sonner";
import { describe, expect, it, vi } from "vitest";

import {
  usePreviewStatement,
  useStatement,
  useStatements,
  useUploadStatement,
} from "./useStatements";

const statement: Statement = {
  id: 5,
  account_id: 1,
  created_by: 1,
  original_filename: "statement.csv",
  file_type: "csv",
  status: "pending",
  created_at: "2026-09-01T00:00:00.000Z",
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

describe("statement queries", () => {
  it("loads a page of statements", async () => {
    const paginated = {
      statements: [statement],
      total: 1,
      page: 1,
      page_size: 10,
    };
    server.use(
      http.get("*/api/v1/statement", () =>
        HttpResponse.json({ message: "ok", data: paginated })
      )
    );

    const { wrapper } = setup();
    const { result } = renderHook(() => useStatements({ page: 1 }), {
      wrapper,
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(paginated);
  });

  it("loads a single statement", async () => {
    server.use(
      http.get("*/api/v1/statement/5", () =>
        HttpResponse.json({ message: "ok", data: statement })
      )
    );

    const { wrapper } = setup();
    const { result } = renderHook(() => useStatement(5), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(statement);
  });

  it("stays idle without an id", () => {
    const { wrapper } = setup();
    const { result } = renderHook(() => useStatement(0), { wrapper });

    expect(result.current.fetchStatus).toBe("idle");
  });
});

describe("statement mutations", () => {
  it("uploads a statement and refreshes the caches", async () => {
    server.use(
      http.post("*/api/v1/statement", () =>
        HttpResponse.json(
          { message: "Statement uploaded", data: statement },
          { status: 201 }
        )
      )
    );

    const { queryClient, wrapper } = setup();
    const invalidate = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useUploadStatement(), { wrapper });

    act(() => {
      result.current.mutate({
        account_id: 1,
        file: new File(["a,b"], "statement.csv"),
      });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(invalidate).toHaveBeenCalledWith({ queryKey: ["statements"] });
    expect(invalidate).toHaveBeenCalledWith({ queryKey: ["transactions"] });
    expect(toast.success).toHaveBeenCalledWith(
      "Statement uploaded successfully! Processing will begin shortly."
    );
  });

  it("previews a statement", async () => {
    server.use(
      http.post("*/api/v1/statement/preview", () =>
        HttpResponse.json({
          message: "ok",
          data: { headers: ["Date"], rows: [["2026-09-01"]] },
        })
      )
    );

    const { wrapper } = setup();
    const { result } = renderHook(() => usePreviewStatement(), { wrapper });

    act(() => {
      result.current.mutate({
        file: new File(["a"], "statement.csv"),
        skipRows: 0,
        rowSize: 5,
      });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual({
      headers: ["Date"],
      rows: [["2026-09-01"]],
    });
  });
});
