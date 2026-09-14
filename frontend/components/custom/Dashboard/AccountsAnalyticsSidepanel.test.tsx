import { testAccount } from "@/test/msw/handlers";
import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, delay, http } from "msw";
import { beforeAll, describe, expect, it, vi } from "vitest";

import { AccountsAnalyticsSidepanel } from "./AccountsAnalyticsSidepanel";

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

beforeAll(() => {
  Object.defineProperty(Element.prototype, "scrollTo", {
    writable: true,
    value: vi.fn(),
  });
});

const zerodha = {
  id: 2,
  name: "Zerodha",
  bank_type: "investment",
  currency: "inr",
  created_by: 1,
};

function analyticsHandler(
  accountAnalytics: unknown[],
  ms = 0,
  accounts: unknown[] = [testAccount, zerodha]
) {
  return [
    http.get("*/api/v1/analytics/monthly", () =>
      HttpResponse.json({
        data: { total_income: 0, total_expenses: 0, total_amount: 0 },
      })
    ),
    http.get("*/api/v1/analytics/account", async () => {
      if (ms) await delay(ms);
      return HttpResponse.json({
        data: { account_analytics: accountAnalytics },
      });
    }),
    http.get("*/api/v1/account", () =>
      HttpResponse.json({ message: "ok", data: accounts })
    ),
  ];
}

describe("AccountsAnalyticsSidepanel", () => {
  it("shows skeletons while the data loads", () => {
    server.use(
      ...analyticsHandler(
        [{ account_id: 1, current_balance: 1000, balance_one_month_ago: 900 }],
        100
      )
    );
    renderWithProviders(<AccountsAnalyticsSidepanel />);

    expect(
      document.querySelectorAll('[data-slot="skeleton"]').length
    ).toBeGreaterThan(0);
  });

  it("invites the user to add a first account", async () => {
    const user = userEvent.setup();
    server.use(...analyticsHandler([], 0, []));
    renderWithProviders(<AccountsAnalyticsSidepanel />);

    expect(await screen.findByText("No accounts yet")).toBeInTheDocument();
    expect(
      screen.getByText("Add an account to start tracking")
    ).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Add Account" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });

  it("separates investments from bank accounts", async () => {
    server.use(
      ...analyticsHandler([
        { account_id: 1, current_balance: 1000, balance_one_month_ago: 900 },
        {
          account_id: 2,
          current_balance: 0,
          balance_one_month_ago: 0,
          current_value: 5000,
          xirr: 12.34,
        },
      ])
    );
    renderWithProviders(<AccountsAnalyticsSidepanel />);

    expect(await screen.findByText("Investments")).toBeInTheDocument();
    expect(screen.getByText("Banks")).toBeInTheDocument();
    expect(screen.getByText("Zerodha")).toBeInTheDocument();
    expect(screen.getByText("₹5K")).toBeInTheDocument();
    expect(screen.getByText("12.3%")).toBeInTheDocument();
    expect(screen.getByText("HDFC Savings")).toBeInTheDocument();
    expect(screen.getByText("₹2K")).toBeInTheDocument();
    expect(screen.getByText("11.1%")).toBeInTheDocument();
  });

  it("reports no change when the previous balance is zero", async () => {
    server.use(
      ...analyticsHandler([
        { account_id: 1, current_balance: 1000, balance_one_month_ago: 0 },
      ])
    );
    renderWithProviders(<AccountsAnalyticsSidepanel />);
    await screen.findByText("HDFC Savings");

    const card = screen.getByText("HDFC Savings").closest(".group")!;
    expect(within(card as HTMLElement).getByText("0.0%")).toBeInTheDocument();
  });
});
