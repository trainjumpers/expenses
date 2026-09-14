import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { fireEvent, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, delay, http } from "msw";
import { beforeAll, describe, expect, it, vi } from "vitest";

import { MonthlyAnalyticsCard } from "./MonthlyAnalyticsCard";

const monthlyData = {
  total_income: 50000,
  total_expenses: 20000,
  total_amount: 30000,
};

function monthlyHandler(data: unknown = monthlyData, ms = 0) {
  return http.get("*/api/v1/analytics/monthly", async () => {
    if (ms) await delay(ms);
    return HttpResponse.json({ message: "ok", data });
  });
}

beforeAll(() => {
  Object.defineProperty(Element.prototype, "scrollTo", {
    writable: true,
    value: vi.fn(),
  });
});

describe("MonthlyAnalyticsCard", () => {
  it("renders the metrics for every period", async () => {
    server.use(monthlyHandler());
    renderWithProviders(<MonthlyAnalyticsCard />);

    expect(await screen.findAllByText("₹50,000.00")).toHaveLength(6);
    expect(screen.getAllByText("₹20,000.00")).toHaveLength(6);
    expect(screen.getAllByText("₹30,000.00")).toHaveLength(6);
    expect(screen.getAllByText("Income")).toHaveLength(6);
    expect(screen.getAllByText("Expenses")).toHaveLength(6);
    expect(screen.getAllByText("Net Flow")).toHaveLength(6);
    expect(screen.getByText("This Month")).toBeInTheDocument();
    expect(screen.getByText("All Time")).toBeInTheDocument();
  });

  it("shows a placeholder when a period has no data", async () => {
    server.use(monthlyHandler(null));
    renderWithProviders(<MonthlyAnalyticsCard />);

    expect(await screen.findAllByText("No data available")).toHaveLength(6);
  });

  it("shows a skeleton while loading", async () => {
    server.use(monthlyHandler(monthlyData, 100));
    renderWithProviders(<MonthlyAnalyticsCard />);

    expect(
      document.querySelectorAll('[data-slot="skeleton"]').length
    ).toBeGreaterThan(0);
    expect(await screen.findAllByText("Income")).toHaveLength(6);
  });

  it("steps through the ranges with the arrows", async () => {
    const user = userEvent.setup();
    server.use(monthlyHandler());
    renderWithProviders(<MonthlyAnalyticsCard />);
    await screen.findAllByText("Income");

    const previous = screen.getByRole("button", {
      name: "View previous range",
    });
    const next = screen.getByRole("button", { name: "View next range" });
    expect(previous).toBeDisabled();

    await user.click(next);
    expect(previous).toBeEnabled();

    await user.click(
      screen.getByRole("button", { name: "View All Time analytics" })
    );
    expect(next).toBeDisabled();
  });

  it("follows the scroll position", async () => {
    server.use(monthlyHandler());
    const { container } = renderWithProviders(<MonthlyAnalyticsCard />);
    await screen.findAllByText("Income");

    const scroller = container.querySelector(".overflow-x-auto") as HTMLElement;
    Object.defineProperty(scroller, "scrollLeft", {
      value: 200,
      writable: true,
    });
    Object.defineProperty(scroller, "clientWidth", { value: 100 });

    fireEvent.scroll(scroller);

    await waitFor(() =>
      expect(
        screen.getByRole("button", { name: "View previous range" })
      ).toBeEnabled()
    );
  });

  it("steps back with the previous arrow", async () => {
    const user = userEvent.setup();
    server.use(monthlyHandler());
    renderWithProviders(<MonthlyAnalyticsCard />);
    await screen.findAllByText("Income");

    await user.click(screen.getByRole("button", { name: "View next range" }));
    await user.click(
      screen.getByRole("button", { name: "View previous range" })
    );

    expect(
      screen.getByRole("button", { name: "View previous range" })
    ).toBeDisabled();
  });
});
