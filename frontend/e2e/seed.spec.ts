import { expect, test } from "@playwright/test";

import { E2E_USER } from "./config";

test("seed", async ({ page }) => {
  await page.goto("/login");

  await page.getByPlaceholder("Email").fill(E2E_USER.email);
  await page.getByPlaceholder("Password").fill(E2E_USER.password);
  await page.getByRole("button", { name: "Sign In" }).click();

  await expect(page).toHaveURL("/");
});
