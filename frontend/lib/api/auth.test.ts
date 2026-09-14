import { server } from "@/test/msw/server";
import { HttpResponse, http } from "msw";
import { toast } from "sonner";
import { describe, expect, it } from "vitest";

import { login, logout, refresh, signup } from "./auth";

describe("auth api", () => {
  it("logs in with the email and password", async () => {
    const result = await login("test1@example.com", "password123");

    expect(result.user).toMatchObject({ email: "test1@example.com" });
  });

  it("explains a rejected login", async () => {
    const error: unknown = await login("test1@example.com", "wrong").catch(
      (e: unknown) => e
    );

    expect(toast.error).toHaveBeenCalledWith(
      "The email or password is incorrect",
      { id: "login-error" }
    );
    expect(error).toMatchObject({ status: 401 });
  });

  it("signs up a new account", async () => {
    const result = await signup("New User", "new@example.com", "password123");

    expect(result.user).toMatchObject({ email: "new@example.com" });
  });

  it("points an existing account at the login page", async () => {
    await expect(
      signup("Existing", "existing@example.com", "password123")
    ).rejects.toThrow("user already exists");

    expect(toast.info).toHaveBeenCalledWith(
      "Account already exists.",
      expect.objectContaining({ action: expect.any(Object) })
    );
  });

  it("returns the refresh response", async () => {
    const response = await refresh();

    expect(response.ok).toBe(true);
  });

  it("swallows logout failures", async () => {
    server.use(http.post("*/api/v1/logout", () => HttpResponse.error()));

    await expect(logout()).resolves.toBeUndefined();
  });
});
