import { testUser } from "@/test/msw/handlers";
import { server } from "@/test/msw/server";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import type { ReactNode } from "react";
import { toast } from "sonner";
import { afterAll, describe, expect, it, vi } from "vitest";

import {
  useLogin,
  useLogout,
  useSignup,
  useUpdatePassword,
  useUpdateUser,
} from "./useUser";

function setup() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
  return { queryClient, wrapper };
}

describe("user mutations", () => {
  const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});

  afterAll(() => {
    consoleError.mockRestore();
  });

  it("logs in and seeds the user cache", async () => {
    const { queryClient, wrapper } = setup();
    const invalidate = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useLogin(), { wrapper });

    act(() => {
      result.current.mutate({
        email: testUser.email,
        password: "password123",
      });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(queryClient.getQueryData(["user"])).toEqual({ user: testUser });
    expect(invalidate).toHaveBeenCalledWith({ queryKey: ["session"] });
    expect(invalidate).toHaveBeenCalledWith({ queryKey: ["user"] });
  });

  it("logs the reason for a rejected login", async () => {
    const { wrapper } = setup();
    const { result } = renderHook(() => useLogin(), { wrapper });

    act(() => {
      result.current.mutate({ email: "wrong@example.com", password: "nope" });
    });

    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(consoleError).toHaveBeenCalledWith("invalid credentials");
  });

  it("signs up and announces the new account", async () => {
    const { queryClient, wrapper } = setup();
    const { result } = renderHook(() => useSignup(), { wrapper });

    act(() => {
      result.current.mutate({
        name: "New User",
        email: "new@example.com",
        password: "password123",
      });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(queryClient.getQueryData(["user"])).toEqual({
      user: { id: 1, name: "New User", email: "new@example.com" },
    });
    expect(toast.success).toHaveBeenCalledWith("Account created successfully!");
  });

  it("points an existing account at the login page", async () => {
    const { wrapper } = setup();
    const { result } = renderHook(() => useSignup(), { wrapper });

    act(() => {
      result.current.mutate({
        name: "Existing",
        email: "existing@example.com",
        password: "password123",
      });
    });

    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(toast.info).toHaveBeenCalledWith(
      "Account already exists.",
      expect.objectContaining({ action: expect.any(Object) })
    );
  });

  it("clears the cache on logout", async () => {
    const { queryClient, wrapper } = setup();
    const clear = vi.spyOn(queryClient, "clear");
    const { result } = renderHook(() => useLogout(), { wrapper });

    act(() => {
      result.current.mutate();
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(clear).toHaveBeenCalled();
    expect(toast.success).toHaveBeenCalledWith("Logged out successfully");
  });

  it("clears the cache even when the logout call fails", async () => {
    const { queryClient, wrapper } = setup();
    const clear = vi.spyOn(queryClient, "clear");
    server.use(
      http.post("*/api/v1/logout", () => HttpResponse.error(), { once: true })
    );
    const { result } = renderHook(() => useLogout(), { wrapper });

    act(() => {
      result.current.mutate();
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(clear).toHaveBeenCalled();
    expect(consoleError).toHaveBeenCalledWith(
      "Logout API call failed:",
      expect.anything()
    );
  });

  it("updates the cached profile", async () => {
    const { queryClient, wrapper } = setup();
    server.use(
      http.patch("*/api/v1/user", () =>
        HttpResponse.json({ data: { ...testUser, name: "Renamed" } })
      )
    );
    const { result } = renderHook(() => useUpdateUser(), { wrapper });

    act(() => {
      result.current.mutate({ name: "Renamed" });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(queryClient.getQueryData(["user"])).toEqual({
      ...testUser,
      name: "Renamed",
    });
    expect(toast.success).toHaveBeenCalledWith("Profile updated successfully!");
  });

  it("updates the password", async () => {
    const { wrapper } = setup();
    server.use(
      http.post("*/api/v1/user/password", () =>
        HttpResponse.json({ message: "Password updated" })
      )
    );
    const { result } = renderHook(() => useUpdatePassword(), { wrapper });

    act(() => {
      result.current.mutate({
        currentPassword: "password123",
        newPassword: "password456",
      });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(toast.success).toHaveBeenCalledWith(
      "Password updated successfully!"
    );
  });

  it("keeps the failure reason of a password update", async () => {
    const { wrapper } = setup();
    server.use(
      http.post("*/api/v1/user/password", () =>
        HttpResponse.json(
          { error: "current password is wrong" },
          { status: 400 }
        )
      )
    );
    const { result } = renderHook(() => useUpdatePassword(), { wrapper });

    act(() => {
      result.current.mutate({
        currentPassword: "nope",
        newPassword: "password456",
      });
    });

    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(consoleError).toHaveBeenCalledWith("current password is wrong");
  });
});
