import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it, vi } from "vitest";

import AnalyticsPage from "./page";

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
    refresh: vi.fn(),
    back: vi.fn(),
    forward: vi.fn(),
    prefetch: vi.fn(),
  }),
  usePathname: () => "/analytics",
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
  uncategorized_count: 0,
  uncategorized_amount: 0,
  realized_interest: 0,
};

function handlers() {
  return [
    http.get("*/api/v1/analytics/insights", () =>
      HttpResponse.json({
        message: "ok",
        data: {
          summary,
          monthly: [
            { month: "2026-09", income: 60000, expenses: 45000, net: 15000 },
          ],
          categories: [],
          top_expenses: [],
          investments: [],
          spending_summary: {
            expense_count: 1,
            average_transaction: 100,
            median_transaction: 100,
            largest_expense: 100,
            active_spending_days: 1,
            no_spend_days: 0,
          },
          category_movement: [],
          weekday_behavior: { days: [], weekend_share: 0 },
          trend: {
            recent_month: "2026-09",
            prior_month: "2026-08",
            recent_expenses: 45000,
            prior_expenses: 40000,
            change: 5000,
            trailing_three_month_average: 43000,
          },
          data_confidence: {
            uncategorized_share: 0,
            multi_category_count: 0,
            multi_category_share: 0,
            latest_transaction_date: "2026-09-10",
            stale_days: 4,
            multiple_currencies: false,
            currencies: ["INR"],
          },
        },
      })
    ),
    http.get("*/api/v1/analytics/cash-balance", () =>
      HttpResponse.json({
        message: "ok",
        data: {
          initial_balance: 0,
          total_income: 0,
          total_expenses: 0,
          time_series: [],
        },
      })
    ),
    http.get("*/api/v1/analytics/monthly", () =>
      HttpResponse.json({
        data: { total_income: 0, total_expenses: 0, total_amount: 0 },
      })
    ),
    http.get("*/api/v1/analytics/account", () =>
      HttpResponse.json({ data: { account_analytics: [] } })
    ),
    http.get("*/api/v1/transaction", () =>
      HttpResponse.json({
        data: { transactions: [], total: 0, page: 1, page_size: 1 },
      })
    ),
    http.get("*/api/v1/statement", () =>
      HttpResponse.json({
        data: { statements: [], total: 0, page: 1, page_size: 5 },
      })
    ),
  ];
}

describe("analytics page", () => {
  it("renders the analytics view inside the dashboard", async () => {
    server.use(...handlers());
    renderWithProviders(<AnalyticsPage />);

    expect(
      await screen.findByRole("heading", { name: "Analytics" })
    ).toBeInTheDocument();
    expect(
      await screen.findByRole("heading", { name: "Cash flow" })
    ).toBeInTheDocument();
  });
});
