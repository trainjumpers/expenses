import { queryClient, queryKeys } from "@/lib/query-client";
import { afterAll, describe, expect, it, vi } from "vitest";

type Retry = (failureCount: number, error: unknown) => boolean;

function retry() {
  return queryClient.getDefaultOptions().queries!.retry as Retry;
}

describe("query client defaults", () => {
  it("keeps client errors out of the retry loop", () => {
    expect(retry()(0, { status: 404 })).toBe(false);
    expect(retry()(0, { status: 401 })).toBe(false);
  });

  it("still retries timeouts, throttling and unknown failures", () => {
    expect(retry()(0, { status: 408 })).toBe(true);
    expect(retry()(0, { status: 429 })).toBe(true);
    expect(retry()(0, { status: 500 })).toBe(true);
    expect(retry()(0, new Error("network"))).toBe(true);
    expect(retry()(3, new Error("network"))).toBe(false);
  });
});

describe("mutation errors", () => {
  const consoleWarn = vi.spyOn(console, "warn").mockImplementation(() => {});
  const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});

  afterAll(() => {
    consoleWarn.mockRestore();
    consoleError.mockRestore();
  });

  it("clears the session on a 401", () => {
    const clear = vi.spyOn(queryClient, "clear");
    const onError = queryClient.getDefaultOptions().mutations!
      .onError as (error: unknown) => void;

    onError({ status: 401 });

    expect(consoleWarn).toHaveBeenCalledWith(
      "401 error in mutation, session may have expired"
    );
    expect(clear).toHaveBeenCalled();
  });

  it("logs every other mutation failure", () => {
    const onError = queryClient.getDefaultOptions().mutations!
      .onError as (error: unknown) => void;

    onError(new Error("nope"));

    expect(consoleError).toHaveBeenCalledWith(
      "Mutation error:",
      expect.any(Error)
    );
  });
});

describe("queryKeys", () => {
  it("builds parameterised keys", () => {
    expect(queryKeys.transactions()).toEqual(["transactions"]);
    expect(queryKeys.transactions({ page: 1 })).toEqual([
      "transactions",
      { page: 1 },
    ]);
    expect(queryKeys.transaction(7)).toEqual(["transactions", 7]);
    expect(queryKeys.account(2)).toEqual(["accounts", 2]);
    expect(queryKeys.category(3)).toEqual(["categories", 3]);
    expect(queryKeys.rule(4)).toEqual(["rules", 4]);
    expect(queryKeys.analytics.categoryAnalytics("2026-09-01", "2026-09-30")).toEqual(
      ["analytics", "category", "2026-09-01", "2026-09-30"]
    );
    expect(
      queryKeys.analytics.monthlyAnalytics("2026-09-01", "2026-09-30")
    ).toEqual(["analytics", "monthly", "2026-09-01", "2026-09-30"]);
    expect(
      queryKeys.analytics.cashBalanceHistory("2026-09-01", "2026-09-30")
    ).toEqual(["analytics", "cash-balance", "2026-09-01", "2026-09-30"]);
    expect(queryKeys.analytics.insights("2026-09-01", "2026-09-30")).toEqual([
      "analytics",
      "insights",
      "2026-09-01",
      "2026-09-30",
    ]);
  });
});
