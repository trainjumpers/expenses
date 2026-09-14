import { expect, test } from "@playwright/test";

import { AUTH_STATE } from "./config";

test.use({ storageState: AUTH_STATE });

test.describe("Transactions", () => {
  test("shows an empty table with pagination disabled", async ({ page }) => {
    await page.goto("/transaction");

    await expect(page.getByText("No transactions found")).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Previous", exact: true })
    ).toBeDisabled();
    await expect(
      page.getByRole("button", { name: "Next", exact: true })
    ).toBeDisabled();
  });

  test("applies amount filters to the url", async ({ page }) => {
    await page.goto("/transaction");

    await page.getByRole("button", { name: "Filter" }).click();
    await page.getByRole("button", { name: "Amount" }).click();
    await page.getByRole("spinbutton").nth(0).fill("100");
    await page.getByRole("spinbutton").nth(1).fill("5000");
    await page.getByRole("button", { name: "Apply" }).click();

    await expect(page).toHaveURL(/min_amount=100/);
    await expect(page).toHaveURL(/max_amount=5000/);
  });

  test("sorts by a column and syncs the url", async ({ page }) => {
    await page.goto("/transaction");

    await page.getByRole("button", { name: "Name" }).click();

    await expect(page).toHaveURL(/sort_by=name/);
    await expect(page).toHaveURL(/sort_order=asc/);
  });

  test("keeps the dialog open when the transaction has no account", async ({
    page,
  }) => {
    await page.goto("/transaction");

    await page.getByRole("button", { name: "Add Transaction" }).click();
    const dialog = page.getByRole("dialog", { name: "Add New Transaction" });
    await dialog.getByPlaceholder("Enter transaction name").fill("No Account");
    await dialog.getByRole("spinbutton").fill("100");
    await dialog.getByRole("button", { name: "Add", exact: true }).click();

    await expect(
      page.getByText("Please check your input and try again")
    ).toBeVisible();
    await expect(dialog).toBeVisible();
  });
});
