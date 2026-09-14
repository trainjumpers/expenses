import { expect, test as setup } from "@playwright/test";

import { AUTH_STATE, E2E_USER } from "./config";

setup("sign in and save the session", async ({ page }) => {
  await page.goto("/login");
  await page.getByPlaceholder("Email").fill(E2E_USER.email);
  await page.getByPlaceholder("Password").fill(E2E_USER.password);
  await page.getByRole("button", { name: "Sign In" }).click();

  await expect(page).toHaveURL("/");
  await page.context().storageState({ path: AUTH_STATE });
});
