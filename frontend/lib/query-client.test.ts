import { describe, expect, it } from "vitest";

import { queryClient } from "./query-client";

const retry = queryClient.getDefaultOptions().queries?.retry as (
  failureCount: number,
  error: unknown
) => boolean;

describe("query client retry policy", () => {
  it("does not retry client errors except timeouts and rate limits", () => {
    expect(retry(0, { status: 400 })).toBe(false);
    expect(retry(0, { status: 404 })).toBe(false);
    expect(retry(0, { status: 408 })).toBe(true);
    expect(retry(0, { status: 429 })).toBe(true);
  });

  it("retries server and network errors up to three times", () => {
    expect(retry(0, { status: 500 })).toBe(true);
    expect(retry(2, { status: 500 })).toBe(true);
    expect(retry(3, { status: 500 })).toBe(false);
    expect(retry(0, new Error("network down"))).toBe(true);
  });
});
