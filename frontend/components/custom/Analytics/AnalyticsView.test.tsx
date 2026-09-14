import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, delay, http } from "msw";
import { describe, expect, it, vi } from "vitest";

import { AnalyticsView } from "./AnalyticsView";

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

class ResizeObserverStub {
  observe() {}
  unobserve() {}
  disconnect() {}
}
vi.stubGlobal("ResizeObserver", ResizeObserverStub);

const summary = {
  net_worth: 150000,
  investment_value: 50000,
  bank_value: 100000,
  period_income: 120000,
  period_expenses: 90000,
  period_net: 30000,
  savings_rate: 0.25,
  uncategorized_count: 2,
  uncategorized_amount: 5000,
  realized_interest: 1200,
};

const insights = {
  summary,
  monthly: [
    { month: "2026-08", income: 60000, expenses: 45000, net: 15000 },
    { month: "2026-09", income: 60000, expenses: 45000, net: 15000 },
  ],
  categories: [{ category_id: 1, category_name: "Food", total_amount: -20000 }],
  top_expenses: [
    { name: "Big Bazaar", amount: 12000, count: 5, share: 0.2, average: 2400 },
  ],
  investments: [
    {
      account_id: 2,
      name: "Zerodha",
      current_value: 50000,
      contributed: 40000,
      distributed: 0,
      realized_interest: 1200,
      xirr: 14.2,
      percentage_increase: 25,
    },
  ],
  spending_summary: {
    expense_count: 42,
    average_transaction: 1200,
    median_transaction: 900,
    largest_expense: 8000,
    active_spending_days: 20,
    no_spend_days: 10,
  },
  category_movement: [
    {
      category_id: 1,
      category_name: "Food",
      recent_total: 12000,
      prior_total: 10000,
      recent_share: 0.25,
      prior_share: 0.22,
      change: 2000,
    },
  ],
  weekday_behavior: {
    days: [{ weekday: 1, total: 5000, count: 10, average: 500, share: 0.2 }],
    weekend_share: 0.4,
  },
  trend: {
    recent_month: "2026-08",
    prior_month: "2026-07",
    recent_expenses: 45000,
    prior_expenses: 40000,
    change: 5000,
    trailing_three_month_average: 43000,
  },
  data_confidence: {
    uncategorized_share: 0.05,
    multi_category_count: 1,
    multi_category_share: 0.02,
    latest_transaction_date: "2026-09-10",
    stale_days: 4,
    multiple_currencies: false,
    currencies: ["INR"],
  },
};

const cashHistory = {
  initial_balance: 10000,
  total_income: 120000,
  total_expenses: 90000,
  time_series: [
    { date: "2026-09-01", cash_balance: 40000 },
    { date: "2026-09-02", cash_balance: 41000 },
  ],
};

function insightsHandler(data: unknown = insights, ms = 0) {
  return http.get("*/api/v1/analytics/insights", async () => {
    if (ms) await delay(ms);
    return HttpResponse.json({ message: "ok", data });
  });
}

function cashHandler() {
  return http.get("*/api/v1/analytics/cash-balance", () =>
    HttpResponse.json({ message: "ok", data: cashHistory })
  );
}

function firstTransactionHandler() {
  return http.get("*/api/v1/transaction", () =>
    HttpResponse.json({
      message: "ok",
      data: {
        transactions: [
          {
            id: 1,
            date: "2025-01-05T00:00:00.000Z",
            name: "Opening",
            description: null,
            amount: 0,
            category_ids: [],
            account_id: 1,
          },
        ],
        total: 1,
        page: 1,
        page_size: 1,
      },
    })
  );
}

describe("AnalyticsView", () => {
  it("shows a skeleton while the insights load", () => {
    server.use(insightsHandler(insights, 100), cashHandler());
    renderWithProviders(<AnalyticsView />);

    expect(
      document.querySelectorAll('[data-slot="skeleton"]').length
    ).toBeGreaterThan(0);
  });

  it("offers a retry when insights fail", async () => {
    const user = userEvent.setup();
    let attempts = 0;
    server.use(
      http.get("*/api/v1/analytics/insights", () => {
        attempts += 1;
        return HttpResponse.json({ error: "boom" }, { status: 500 });
      }),
      cashHandler()
    );
    renderWithProviders(<AnalyticsView />);

    expect(
      await screen.findByText("Analytics unavailable")
    ).toBeInTheDocument();
    expect(
      screen.getByText("Could not load analytics insights.")
    ).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Try again" }));
    await waitFor(() => expect(attempts).toBe(2));
  });

  it("explains an inactive range", async () => {
    server.use(
      insightsHandler({
        ...insights,
        monthly: [{ month: "2026-09", income: 0, expenses: 0, net: 0 }],
        investments: [],
        summary: { ...summary, net_worth: 0 },
      }),
      cashHandler()
    );
    renderWithProviders(<AnalyticsView />);

    expect(
      await screen.findByText("No activity in this range")
    ).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: "import a statement" })
    ).toHaveAttribute("href", "/transaction");
  });

  it("renders the full household report", async () => {
    server.use(insightsHandler(), cashHandler(), firstTransactionHandler());
    renderWithProviders(<AnalyticsView />);

    expect(await screen.findByText("Surplus")).toBeInTheDocument();
    expect(screen.getByText("Savings rate")).toBeInTheDocument();
    expect(screen.getByText("25.0%")).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: "Cash flow" })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: "Category movement" })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: "Spending habits" })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: "Top payees" })
    ).toBeInTheDocument();
    expect(screen.getByText("Big Bazaar")).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: "Investments" })
    ).toBeInTheDocument();
    expect(screen.getByText("Zerodha")).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: "Current net worth" })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: "Cash balance history" })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: "Data health" })
    ).toBeInTheDocument();
  });

  it("labels a negative period as a deficit", async () => {
    server.use(
      insightsHandler({
        ...insights,
        summary: { ...summary, period_net: -5000 },
      }),
      cashHandler()
    );
    renderWithProviders(<AnalyticsView />);

    expect(await screen.findByText("Deficit")).toBeInTheDocument();
  });
});
