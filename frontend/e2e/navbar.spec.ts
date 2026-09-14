import { expect, test } from "@playwright/test";

import { AUTH_STATE, E2E_USER } from "./config";

test.use({ storageState: AUTH_STATE });

test.describe("Navbar", () => {
  test("switches the theme from the menu", async ({ page }) => {
    await page.goto("/");

    await page.getByRole("button", { name: "Toggle theme" }).click();
    await page.getByRole("menuitem", { name: "Dark" }).click();
    await expect(page.locator("html")).toHaveClass(/dark/);

    await page.getByRole("button", { name: "Toggle theme" }).click();
    await page.getByRole("menuitem", { name: "Light" }).click();
    await expect(page.locator("html")).not.toHaveClass(/dark/);
  });

  test("shows the signed-in account in the profile dialog", async ({
    page,
  }) => {
    await page.goto("/");

    await page.getByRole("button", { name: "ETU" }).click();
    await page.getByRole("button", { name: "Profile" }).click();

    await expect(
      page.getByRole("heading", { name: "Edit Profile" })
    ).toBeVisible();
    await expect(page.getByRole("textbox", { name: "Name" })).toHaveValue(
      "E2E Test User"
    );
    await expect(page.getByRole("textbox", { name: "Email" })).toHaveValue(
      E2E_USER.email
    );
  });
});
