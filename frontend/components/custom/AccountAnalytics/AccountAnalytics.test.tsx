import { testAccount } from "@/test/msw/handlers";
import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { describe, expect, it, vi } from "vitest";

import { AccountAnalytics } from "./AccountAnalytics";

vi.mock("@/components/ui/icon-picker", () => ({
  Icon: () => null,
  IconPicker: () => null,
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
    refresh: vi.fn(),
    back: vi.fn(),
    forward: vi.fn(),
    prefetch: vi.fn(),
  }),
  usePathname: () => "/",
  useSearchParams: () => new URLSearchParams(),
}));

const accountOne = {
  account_id: 1,
  current_balance: 1000,
  balance_one_month_ago: 900,
};

const accountTwo = {
  account_id: 2,
  current_balance: -500,
  balance_one_month_ago: -400,
};

const recentTransaction = {
  id: 9,
  date: "2026-09-01T00:00:00.000Z",
  name: "Lunch",
  description: null,
  amount: 120,
  category_ids: [],
  account_id: 1,
};

describe("AccountAnalytics", () => {
  it("invites the user to create the first account", async () => {
    const user = userEvent.setup();
    renderWithProviders(<AccountAnalytics />);

    expect(screen.getByText("No accounts yet")).toBeInTheDocument();
    await user.click(
      screen.getByRole("button", { name: /add your first account/i })
    );

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });

  it("renders the default rows and expands transactions", async () => {
    const user = userEvent.setup();
    server.use(
      http.get("*/api/v1/transaction", () =>
        HttpResponse.json({
          message: "ok",
          data: {
            transactions: [recentTransaction],
            total: 1,
            page: 1,
            page_size: 5,
          },
        })
      )
    );
    renderWithProviders(<AccountAnalytics data={[accountOne, accountTwo]} />);

    expect(await screen.findByText("HDFC Savings")).toBeInTheDocument();
    expect(screen.queryByText("Account 2")).not.toBeInTheDocument();
    expect(screen.getAllByText("₹2,000.00").length).toBeGreaterThan(0);

    await user.click(
      screen.getByRole("button", { name: "Expand HDFC Savings" })
    );

    expect(
      await screen.findByText("Latest 5 transactions")
    ).toBeInTheDocument();
    expect(await screen.findByText("Lunch")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "View all" })).toHaveAttribute(
      "href",
      "/transaction?account_id=1"
    );
  });

  it("reveals hidden accounts through the filter", async () => {
    const user = userEvent.setup();
    renderWithProviders(<AccountAnalytics data={[accountOne, accountTwo]} />);
    await screen.findByText("HDFC Savings");

    await user.click(screen.getByRole("button", { name: "Default accounts" }));
    await user.click(await screen.findByRole("button", { name: "Select all" }));
    await user.click(screen.getByRole("button", { name: "Apply" }));

    expect(
      within(screen.getByRole("table")).getByText("Account 2")
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "All accounts" })
    ).toBeInTheDocument();
  });

  it("explains when nothing is shown by default", () => {
    renderWithProviders(<AccountAnalytics data={[accountTwo]} />);

    expect(
      screen.getByText(
        /No accounts shown by default. Investment accounts and negative balances are hidden/
      )
    ).toBeInTheDocument();
  });

  it("shows investment value and xirr once selected", async () => {
    const user = userEvent.setup();
    server.use(
      http.get("*/api/v1/account", () =>
        HttpResponse.json({
          message: "ok",
          data: [
            {
              ...testAccount,
              id: 3,
              name: "Zerodha",
              bank_type: "investment",
            },
          ],
        })
      )
    );
    renderWithProviders(
      <AccountAnalytics
        data={[
          {
            account_id: 3,
            current_balance: 0,
            balance_one_month_ago: 0,
            current_value: 5000,
            xirr: 12.34,
          },
        ]}
      />
    );

    await user.click(screen.getByRole("button", { name: "Default accounts" }));
    await user.click(
      await screen.findByRole("menuitemcheckbox", { name: "Zerodha" })
    );
    await user.click(screen.getByRole("button", { name: "Apply" }));

    const table = screen.getByRole("table");
    expect(within(table).getByText("Zerodha")).toBeInTheDocument();
    expect(within(table).getByText("₹5,000.00")).toBeInTheDocument();
    expect(within(table).getByText("XIRR 12.3%")).toBeInTheDocument();
  });

  it("clears every account again after selecting all", async () => {
    const user = userEvent.setup();
    renderWithProviders(<AccountAnalytics data={[accountOne, accountTwo]} />);
    await screen.findByText("HDFC Savings");

    await user.click(screen.getByRole("button", { name: "Default accounts" }));
    await user.click(await screen.findByRole("button", { name: "Select all" }));
    await user.click(screen.getByRole("button", { name: "Apply" }));
    await screen.findByRole("button", { name: "All accounts" });

    await user.click(screen.getByRole("button", { name: "All accounts" }));
    await user.click(
      await screen.findByRole("button", { name: "Deselect all" })
    );
    await user.click(screen.getByRole("button", { name: "Apply" }));

    expect(
      await screen.findByRole("button", { name: "Default accounts" })
    ).toBeInTheDocument();
    expect(
      within(screen.getByRole("table")).queryByText("Account 2")
    ).not.toBeInTheDocument();
  });

  it("collapses an expanded account again", async () => {
    const user = userEvent.setup();
    renderWithProviders(<AccountAnalytics data={[accountOne]} />);
    await screen.findByText("HDFC Savings");

    await user.click(
      screen.getByRole("button", { name: "Expand HDFC Savings" })
    );
    await user.click(
      await screen.findByRole("button", { name: "Collapse HDFC Savings" })
    );

    expect(
      screen.getByRole("button", { name: "Expand HDFC Savings" })
    ).toBeInTheDocument();
  });

  it("reports a failed transaction lookup", async () => {
    const user = userEvent.setup();
    server.use(
      http.get("*/api/v1/transaction", () =>
        HttpResponse.json({ error: "boom" }, { status: 500 })
      )
    );
    renderWithProviders(<AccountAnalytics data={[accountOne]} />);
    await screen.findByText("HDFC Savings");

    await user.click(
      screen.getByRole("button", { name: "Expand HDFC Savings" })
    );

    expect(
      await screen.findByText("Failed to load transactions.")
    ).toBeInTheDocument();
  });

  it("closes the add dialog after creating an account", async () => {
    const user = userEvent.setup();
    server.use(
      http.post("*/api/v1/account", () =>
        HttpResponse.json(
          { data: { id: 9, name: "ICICI", bank_type: "icici" } },
          { status: 201 }
        )
      )
    );
    renderWithProviders(<AccountAnalytics data={[accountOne]} />);
    await screen.findByText("HDFC Savings");

    await user.click(screen.getByRole("button", { name: "Add account" }));
    await user.type(
      await screen.findByPlaceholderText("Enter account name"),
      "ICICI"
    );
    await user.click(screen.getAllByRole("combobox")[0]);
    await user.click(await screen.findByRole("option", { name: "ICICI Bank" }));
    await user.click(screen.getByRole("button", { name: "Add Account" }));

    await waitFor(() =>
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument()
    );
  });
});
