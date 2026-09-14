import { renderWithProviders } from "@/test/render";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { CashBalanceHistory } from "./NetWorthAndCash";

class ResizeObserverStub {
  observe() {}
  unobserve() {}
  disconnect() {}
}
vi.stubGlobal("ResizeObserver", ResizeObserverStub);

function history(values: number[]) {
  return {
    initial_balance: 0,
    total_income: 0,
    total_expenses: 0,
    time_series: values.map((value, i) => ({
      date: `2026-09-0${i + 1}`,
      cash_balance: value,
    })),
  };
}

describe("CashBalanceHistory", () => {
  it("shows the latest balance", () => {
    renderWithProviders(
      <CashBalanceHistory
        history={history([40000, 41000])}
        isLoading={false}
        isError={false}
        onRetry={vi.fn()}
      />
    );

    expect(screen.getAllByText("₹41,000.00").length).toBeGreaterThan(0);
  });

  it("renders a flat series with a padded domain", () => {
    renderWithProviders(
      <CashBalanceHistory
        history={history([40000, 40000])}
        isLoading={false}
        isError={false}
        onRetry={vi.fn()}
      />
    );

    expect(
      screen.getByRole("heading", { name: "Cash balance history" })
    ).toBeInTheDocument();
  });

  it("notes an empty history", () => {
    renderWithProviders(
      <CashBalanceHistory
        history={history([])}
        isLoading={false}
        isError={false}
        onRetry={vi.fn()}
      />
    );

    expect(
      screen.getByText("No cash activity in this range.")
    ).toBeInTheDocument();
  });

  it("shows a loading skeleton", () => {
    renderWithProviders(
      <CashBalanceHistory
        history={undefined}
        isLoading
        isError={false}
        onRetry={vi.fn()}
      />
    );

    expect(
      document.querySelectorAll('[data-slot="skeleton"]').length
    ).toBeGreaterThan(0);
  });

  it("offers a retry after a failure", async () => {
    const user = userEvent.setup();
    const onRetry = vi.fn();
    renderWithProviders(
      <CashBalanceHistory
        history={undefined}
        isLoading={false}
        isError
        onRetry={onRetry}
      />
    );

    expect(
      screen.getByText("Could not load cash balance history.")
    ).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Try again" }));

    expect(onRetry).toHaveBeenCalled();
  });
});
