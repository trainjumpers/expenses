import { expect, test } from "@playwright/test";

import { API_BASE_URL, BASE_URL } from "./config";

test.describe("Signup", () => {
  test("creates an account and lands on the dashboard", async ({ page }) => {
    const email = `e2e-signup-${Date.now()}@neurospend.test`;

    await page.goto("/signup");

    await page.getByPlaceholder("Name").fill("E2E Signup");
    await page.getByPlaceholder("Email").fill(email);
    await page.getByPlaceholder("Password").fill("password123");
    await page.getByRole("button", { name: "Sign Up" }).click();

    await expect(page).toHaveURL(`${BASE_URL}/`);

    // Keep local dev databases from accumulating throwaway users.
    const cleanup = await page.request.delete(`${API_BASE_URL}/api/v1/user`);
    expect(cleanup.status()).toBe(204);
  });
});
