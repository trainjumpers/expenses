import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { screen } from "@testing-library/react";
import { HttpResponse, delay, http } from "msw";
import { describe, expect, it, vi } from "vitest";

import { NetWorth } from "./NetWorth";

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

const dateRange = {
  from: new Date(2026, 8, 1),
  to: new Date(2026, 8, 30),
};

function historyHandler(
  timeSeries: { date: string; cash_balance: number }[],
  initialBalance = 10000,
  ms = 0
) {
  return [
    http.get("*/api/v1/analytics/cash-balance", async () => {
      if (ms) await delay(ms);
      return HttpResponse.json({
        message: "ok",
        data: {
          initial_balance: initialBalance,
          total_income: 0,
          total_expenses: 0,
          time_series: timeSeries,
        },
      });
    }),
    http.get("*/api/v1/transaction", () =>
      HttpResponse.json({
        message: "ok",
        data: { transactions: [], total: 0, page: 1, page_size: 1 },
      })
    ),
  ];
}

describe("NetWorth", () => {
  it("shows a skeleton while the history loads", () => {
    server.use(
      ...historyHandler(
        [{ date: "2026-09-01", cash_balance: 40000 }],
        10000,
        100
      )
    );
    renderWithProviders(<NetWorth dateRange={dateRange} />);

    expect(
      document.querySelectorAll('[data-slot="skeleton"]').length
    ).toBeGreaterThan(0);
  });

  it("shows the closing balance and the range dates", async () => {
    server.use(
      ...historyHandler([
        { date: "2026-09-01", cash_balance: 40000 },
        { date: "2026-09-02", cash_balance: 41000 },
      ])
    );
    renderWithProviders(<NetWorth dateRange={dateRange} />);

    expect(await screen.findByText("Cash balance")).toBeInTheDocument();
    expect(screen.getByText("₹41,000.00")).toBeInTheDocument();
    expect(screen.getByText(/Up ₹31K/)).toBeInTheDocument();
    expect(screen.getByText("Sep 01, 2026")).toBeInTheDocument();
    expect(screen.getByText("Sep 02, 2026")).toBeInTheDocument();
  });

  it("reports a falling balance against the starting point", async () => {
    server.use(
      ...historyHandler([{ date: "2026-09-02", cash_balance: 8000 }], 10000)
    );
    renderWithProviders(<NetWorth dateRange={dateRange} />);

    expect(await screen.findByText("₹8,000.00")).toBeInTheDocument();
    expect(screen.getByText(/Down ₹2K/)).toBeInTheDocument();
  });

  it("renders an empty history without dates", async () => {
    server.use(...historyHandler([]));
    renderWithProviders(<NetWorth dateRange={dateRange} />);

    expect(await screen.findByText("₹0.00")).toBeInTheDocument();
    expect(screen.getByText(/Down ₹10K/)).toBeInTheDocument();
  });
});
