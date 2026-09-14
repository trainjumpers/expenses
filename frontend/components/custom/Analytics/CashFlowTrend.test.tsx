import { renderWithProviders } from "@/test/render";
import { screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { CashFlowTrend } from "./CashFlowTrend";

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

const monthly = [
  { month: "2026-08", income: 60000, expenses: 45000, net: 15000 },
];

const trend = {
  recent_month: "2026-08",
  prior_month: "2026-07",
  recent_expenses: 45000,
  prior_expenses: 40000,
  change: 5000,
  trailing_three_month_average: 43000,
};

describe("CashFlowTrend", () => {
  it("reports a spending decrease", () => {
    renderWithProviders(
      <CashFlowTrend
        monthly={monthly}
        trend={{ ...trend, change: -5000 }}
        movement={[]}
      />
    );

    expect(screen.getByText(/down/)).toBeInTheDocument();
    expect(screen.getByText(/12\.5%/)).toBeInTheDocument();
  });

  it("reports flat spending without a percentage", () => {
    renderWithProviders(
      <CashFlowTrend
        monthly={monthly}
        trend={{ ...trend, change: 0, prior_expenses: 0 }}
        movement={[]}
      />
    );

    expect(screen.getByText(/flat/)).toBeInTheDocument();
    expect(screen.queryByText(/%/)).not.toBeInTheDocument();
  });

  it("notes a range without an earlier month", () => {
    renderWithProviders(
      <CashFlowTrend
        monthly={monthly}
        trend={{ ...trend, prior_month: "" }}
        movement={[]}
      />
    );

    expect(
      screen.getByText(/does not include an earlier complete month/)
    ).toBeInTheDocument();
  });

  it("notes an empty cash flow", () => {
    renderWithProviders(
      <CashFlowTrend monthly={[]} trend={trend} movement={[]} />
    );

    expect(
      screen.getByText("No household cash flow in this range.")
    ).toBeInTheDocument();
  });
});
