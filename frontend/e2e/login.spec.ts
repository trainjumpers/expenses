import { expect, test } from "@playwright/test";

import { E2E_USER } from "./config";

test.describe("Login", () => {
  test("signs in with valid credentials", async ({ page }) => {
    await page.goto("/login");

    await page.getByPlaceholder("Email").fill(E2E_USER.email);
    await page.getByPlaceholder("Password").fill(E2E_USER.password);
    await page.getByRole("button", { name: "Sign In" }).click();

    await expect(page.getByText("Welcome back!")).toBeVisible();
    await expect(page).toHaveURL("/");
  });

  test("rejects invalid credentials", async ({ page }) => {
    await page.goto("/login");

    await page.getByPlaceholder("Email").fill(E2E_USER.email);
    await page.getByPlaceholder("Password").fill("wrong-password");
    await page.getByRole("button", { name: "Sign In" }).click();

    await expect(
      page.getByText("The email or password is incorrect")
    ).toBeVisible();
    await expect(page).toHaveURL("/login");
  });
});
