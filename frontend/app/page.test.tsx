import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, delay, http } from "msw";
import { describe, expect, it, vi } from "vitest";

import Page from "./page";

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

function dashboardHandlers(categoryDelay = 0) {
  return [
    http.get("*/api/v1/analytics/monthly", () =>
      HttpResponse.json({
        data: { total_income: 0, total_expenses: 0, total_amount: 0 },
      })
    ),
    http.get("*/api/v1/analytics/account", () =>
      HttpResponse.json({ data: { account_analytics: [] } })
    ),
    http.get("*/api/v1/analytics/category", async () => {
      if (categoryDelay) await delay(categoryDelay);
      return HttpResponse.json({ data: { category_transactions: [] } });
    }),
    http.get("*/api/v1/analytics/cash-balance", () =>
      HttpResponse.json({
        data: {
          initial_balance: 0,
          total_income: 0,
          total_expenses: 0,
          time_series: [],
        },
      })
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
    http.get("*/api/v1/rule", () =>
      HttpResponse.json({
        data: { rules: [], total: 0, page: 1, page_size: 5 },
      })
    ),
  ];
}

describe("dashboard page", () => {
  it("greets the user and shows the empty analytics panels", async () => {
    server.use(...dashboardHandlers());
    renderWithProviders(<Page />);

    expect(await screen.findByText("Welcome back, Test")).toBeInTheDocument();
    expect(
      screen.getByText(/what's happening with your finances/)
    ).toBeInTheDocument();
    expect(
      await screen.findByText("No category activity for the selected filter.")
    ).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: "No accounts yet" })
    ).toBeInTheDocument();
    expect(screen.getByText("Cash balance")).toBeInTheDocument();
  });

  it("shows skeletons while the analytics load", () => {
    server.use(...dashboardHandlers(100));
    renderWithProviders(<Page />);

    expect(
      document.querySelectorAll('[data-slot="skeleton"]').length
    ).toBeGreaterThan(0);
  });

  it("reports analytics failures", async () => {
    server.use(...dashboardHandlers());
    server.use(
      http.get("*/api/v1/analytics/category", () =>
        HttpResponse.json({ error: "boom" }, { status: 500 })
      )
    );
    renderWithProviders(<Page />);

    expect(
      await screen.findByText("Error loading analytics data.")
    ).toBeInTheDocument();
  });

  it("opens the command center and the view center", async () => {
    const user = userEvent.setup();
    server.use(...dashboardHandlers());
    renderWithProviders(<Page />);
    await screen.findByText("Welcome back, Test");

    await user.click(screen.getByRole("button", { name: /^new$/i }));
    expect(
      await screen.findByRole("dialog", { name: "Command Center" })
    ).toBeInTheDocument();

    await user.keyboard("{Escape}");
    await waitFor(() =>
      expect(
        screen.queryByRole("dialog", { name: "Command Center" })
      ).not.toBeInTheDocument()
    );

    await user.click(screen.getByRole("button", { name: /^view$/i }));
    expect(
      await screen.findByRole("dialog", { name: "View Center" })
    ).toBeInTheDocument();
  });
});
