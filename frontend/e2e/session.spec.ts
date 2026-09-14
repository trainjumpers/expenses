import { expect, test } from "@playwright/test";

import { AUTH_STATE } from "./config";

test.describe("Session refresh", () => {
  test.use({ storageState: AUTH_STATE });

  test("recovers the session when the access token is gone", async ({
    page,
    context,
  }) => {
    await page.goto("/");
    await expect(
      page.getByRole("heading", { name: /welcome back/i })
    ).toBeVisible();

    await context.clearCookies({ name: "access_token" });
    expect(
      (await context.cookies()).some((cookie) => cookie.name === "access_token")
    ).toBe(false);

    await page.reload();

    await expect(
      page.getByRole("heading", { name: /welcome back/i })
    ).toBeVisible();
    expect(
      (await context.cookies()).some((cookie) => cookie.name === "access_token")
    ).toBe(true);
  });

  test("returns to login when the refresh token is gone as well", async ({
    page,
    context,
  }) => {
    await page.goto("/");
    await expect(
      page.getByRole("heading", { name: /welcome back/i })
    ).toBeVisible();

    await context.clearCookies();
    await page.reload();

    await expect(page).toHaveURL(/\/login/);
  });
});
