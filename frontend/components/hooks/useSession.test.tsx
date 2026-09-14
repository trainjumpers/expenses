import { queryKeys } from "@/lib/query-client";
import { testUser } from "@/test/msw/handlers";
import { server } from "@/test/msw/server";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import type { ReactNode } from "react";
import { afterAll, beforeEach, describe, expect, it, vi } from "vitest";

import { useSession } from "./useSession";

const mockPush = vi.fn();
const mockPathname = vi.fn(() => "/");

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: mockPush,
    replace: vi.fn(),
    refresh: vi.fn(),
    back: vi.fn(),
    forward: vi.fn(),
    prefetch: vi.fn(),
  }),
  usePathname: () => mockPathname(),
  useSearchParams: () => new URLSearchParams(),
}));

function setup() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
  return { queryClient, wrapper };
}

describe("useSession", () => {
  const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});

  beforeEach(() => {
    mockPathname.mockReturnValue("/");
  });

  afterAll(() => {
    consoleError.mockRestore();
  });

  it("reports an authenticated session when the user endpoint succeeds", async () => {
    const { wrapper } = setup();
    const { result } = renderHook(() => useSession(), { wrapper });

    await waitFor(() => expect(result.current.isAuthenticated).toBe(true));
    expect(mockPush).not.toHaveBeenCalled();
  });

  it("refreshes an expired session and stays authenticated", async () => {
    let userCalls = 0;
    server.use(
      http.get("*/api/v1/user", () => {
        userCalls += 1;
        if (userCalls === 1) {
          return HttpResponse.json({ error: "token expired" }, { status: 401 });
        }
        return HttpResponse.json({
          message: "User retrieved successfully",
          data: testUser,
        });
      })
    );

    const { wrapper } = setup();
    const { result } = renderHook(() => useSession(), { wrapper });

    await waitFor(() => expect(result.current.isAuthenticated).toBe(true));
    expect(userCalls).toBe(2);
    expect(mockPush).not.toHaveBeenCalled();
  });

  it("redirects to login and clears the cache when the session is invalid", async () => {
    server.use(
      http.get("*/api/v1/user", () =>
        HttpResponse.json({ error: "unauthorized" }, { status: 401 })
      ),
      http.post("*/api/v1/refresh", () =>
        HttpResponse.json({ error: "invalid token" }, { status: 401 })
      )
    );

    const { queryClient, wrapper } = setup();
    const clearSpy = vi.spyOn(queryClient, "clear");
    const { result } = renderHook(() => useSession(), { wrapper });

    await waitFor(() => expect(mockPush).toHaveBeenCalledWith("/login"));
    expect(clearSpy).toHaveBeenCalled();
    expect(result.current.isAuthenticated).toBe(false);
  });

  it("skips the session check on public routes", async () => {
    mockPathname.mockReturnValue("/login");
    let userCalls = 0;
    server.use(
      http.get("*/api/v1/user", () => {
        userCalls += 1;
        return HttpResponse.json({
          message: "User retrieved successfully",
          data: testUser,
        });
      })
    );

    const { wrapper } = setup();
    const { result } = renderHook(() => useSession(), { wrapper });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(userCalls).toBe(0);
    expect(result.current.isAuthenticated).toBe(false);
  });

  it("redirects authenticated users away from public routes", async () => {
    mockPathname.mockReturnValue("/login");
    const { queryClient, wrapper } = setup();
    queryClient.setQueryData(queryKeys.session, {
      isValid: true,
      needsRefresh: false,
    });

    renderHook(() => useSession(), { wrapper });

    await waitFor(() => expect(mockPush).toHaveBeenCalledWith("/"));
  });

  it("treats a server failure as an invalid session", async () => {
    server.use(
      http.get("*/api/v1/user", () =>
        HttpResponse.json({ error: "boom" }, { status: 500 })
      )
    );

    const { wrapper } = setup();
    const { result } = renderHook(() => useSession(), { wrapper });

    await waitFor(() => expect(mockPush).toHaveBeenCalledWith("/login"));
    expect(result.current.isAuthenticated).toBe(false);
  });

  it("logs a failing refresh attempt", async () => {
    server.use(
      http.get("*/api/v1/user", () =>
        HttpResponse.json({ error: "expired" }, { status: 401 })
      ),
      http.post("*/api/v1/refresh", () => HttpResponse.error())
    );

    const { wrapper } = setup();
    const { result } = renderHook(() => useSession(), { wrapper });

    await waitFor(() =>
      expect(consoleError).toHaveBeenCalledWith(
        "Token refresh failed:",
        expect.any(Error)
      )
    );
    expect(result.current.isAuthenticated).toBe(false);
  });

  it("logs a throwing session check", async () => {
    server.use(http.get("*/api/v1/user", () => HttpResponse.error()));

    const { wrapper } = setup();
    const { result } = renderHook(() => useSession(), { wrapper });

    await waitFor(() =>
      expect(consoleError).toHaveBeenCalledWith(
        "Session check failed:",
        expect.any(Error)
      )
    );
    expect(result.current.isAuthenticated).toBe(false);
  });
});
