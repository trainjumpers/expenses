import type { Transaction } from "@/lib/models/transaction";
import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { afterAll, beforeEach, describe, expect, it, vi } from "vitest";

import TransactionPage from "./page";

const nav = vi.hoisted(() => ({
  replace: vi.fn(),
  searchParams: new URLSearchParams(),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: vi.fn(),
    replace: nav.replace,
    refresh: vi.fn(),
    back: vi.fn(),
    forward: vi.fn(),
    prefetch: vi.fn(),
  }),
  usePathname: () => "/transaction",
  useSearchParams: () => nav.searchParams,
}));

const transactions: Transaction[] = [
  {
    id: 1,
    date: "2026-08-15T00:00:00.000Z",
    name: "Coffee",
    description: null,
    amount: 250,
    category_ids: [1],
    account_id: 1,
  },
  {
    id: 2,
    date: "2026-08-16T00:00:00.000Z",
    name: "Shoe Refund",
    description: "Refund for shoes",
    amount: -500,
    category_ids: [],
    account_id: 1,
  },
];

let transactionUrl = "";
let deletedIds: number[] = [];
let executedBody: unknown;

function baseHandlers() {
  return [
    http.get("*/api/v1/analytics/monthly", () =>
      HttpResponse.json({
        data: { total_income: 0, total_expenses: 0, total_amount: 0 },
      })
    ),
    http.get("*/api/v1/analytics/account", () =>
      HttpResponse.json({ data: { account_analytics: [] } })
    ),
    http.get("*/api/v1/statement", () =>
      HttpResponse.json({
        data: { statements: [], total: 0, page: 1, page_size: 5 },
      })
    ),
    http.get("*/api/v1/transaction", ({ request }) => {
      transactionUrl = request.url;
      return HttpResponse.json({
        message: "ok",
        data: {
          transactions,
          total: transactions.length,
          page: 1,
          page_size: 15,
        },
      });
    }),
    http.delete("*/api/v1/transaction/:id", ({ params }) => {
      deletedIds.push(Number(params.id));
      return new HttpResponse(null, { status: 204 });
    }),
    http.post("*/api/v1/rule/execute", async ({ request }) => {
      executedBody = await request.json();
      return HttpResponse.json({
        data: { modified: [], processed_transactions: [] },
      });
    }),
  ];
}

async function openActions(
  user: ReturnType<typeof userEvent.setup>,
  item: string
) {
  await user.click(screen.getByRole("button", { name: "More actions" }));
  await user.click(await screen.findByRole("menuitem", { name: item }));
}

beforeEach(() => {
  nav.searchParams = new URLSearchParams();
  nav.replace.mockClear();
  transactionUrl = "";
  deletedIds = [];
  executedBody = undefined;
});

describe("TransactionPage", () => {
  const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});

  afterAll(() => {
    consoleError.mockRestore();
  });

  it("restores the query state from the url", async () => {
    nav.searchParams = new URLSearchParams(
      "page=2&sort_by=amount&sort_order=asc&account_id=1&category_id=2&uncategorized=true&min_amount=10&max_amount=500&date_from=2026-08-01&date_to=2026-08-31&search=coffee"
    );
    server.use(...baseHandlers());
    renderWithProviders(<TransactionPage />);
    await screen.findByText("Coffee");

    const params = new URL(transactionUrl).searchParams;
    expect(params.get("page")).toBe("2");
    expect(params.get("sort_by")).toBe("amount");
    expect(params.get("sort_order")).toBe("asc");
    expect(params.get("account_id")).toBe("1");
    expect(params.get("category_id")).toBe("2");
    expect(params.get("uncategorized")).toBe("true");
    expect(params.get("min_amount")).toBe("10");
    expect(params.get("max_amount")).toBe("500");
    expect(params.get("date_from")).toBe("2026-08-01");
    expect(params.get("date_to")).toBe("2026-08-31");
    expect(params.get("search")).toBe("coffee");
  });

  it("writes a page change back to the url", async () => {
    const user = userEvent.setup();
    const many = Array.from({ length: 30 }, (_, i) => ({
      ...transactions[0],
      id: i + 1,
      name: `Transaction ${i + 1}`,
    }));
    server.use(
      http.get("*/api/v1/analytics/monthly", () =>
        HttpResponse.json({
          data: { total_income: 0, total_expenses: 0, total_amount: 0 },
        })
      ),
      http.get("*/api/v1/analytics/account", () =>
        HttpResponse.json({ data: { account_analytics: [] } })
      ),
      http.get("*/api/v1/transaction", ({ request }) => {
        transactionUrl = request.url;
        return HttpResponse.json({
          message: "ok",
          data: { transactions: many, total: 30, page: 1, page_size: 15 },
        });
      })
    );
    renderWithProviders(<TransactionPage />);
    await screen.findByText("Transaction 1");

    await user.click(screen.getByRole("button", { name: "2" }));

    await waitFor(() =>
      expect(nav.replace).toHaveBeenCalledWith(
        expect.stringContaining("page=2"),
        { scroll: false }
      )
    );
  });

  it("opens the update dialog for a single selected row", async () => {
    const user = userEvent.setup();
    server.use(...baseHandlers());
    renderWithProviders(<TransactionPage />);
    await screen.findByText("Coffee");

    await user.click(screen.getByRole("checkbox", { name: "Select Coffee" }));
    await openActions(user, "Update");

    expect(
      await screen.findByRole("dialog", { name: "Update Transaction" })
    ).toBeInTheDocument();
  });

  it("deletes a single selected transaction", async () => {
    const user = userEvent.setup();
    server.use(...baseHandlers());
    renderWithProviders(<TransactionPage />);
    await screen.findByText("Coffee");

    await user.click(screen.getByRole("checkbox", { name: "Select Coffee" }));
    await openActions(user, "Delete");

    await waitFor(() => expect(deletedIds).toEqual([1]));
    await waitFor(() =>
      expect(screen.queryByText("Coffee")).not.toBeInTheDocument()
    );
  });

  it("deletes several transactions at once", async () => {
    const user = userEvent.setup();
    server.use(...baseHandlers());
    renderWithProviders(<TransactionPage />);
    await screen.findByText("Coffee");

    await user.click(screen.getByRole("checkbox", { name: "Select all" }));
    await openActions(user, "Delete");

    await waitFor(() => expect(deletedIds).toHaveLength(2));
    expect(deletedIds.sort()).toEqual([1, 2]);
  });

  it("runs the rules against the selection", async () => {
    const user = userEvent.setup();
    server.use(...baseHandlers());
    renderWithProviders(<TransactionPage />);
    await screen.findByText("Coffee");

    await user.click(screen.getByRole("checkbox", { name: "Select Coffee" }));
    await openActions(user, "Execute Rules");

    await waitFor(() =>
      expect(executedBody).toEqual({ transaction_ids: [1] })
    );
  });

  it("opens the import statement dialog", async () => {
    const user = userEvent.setup();
    server.use(...baseHandlers());
    renderWithProviders(<TransactionPage />);
    await screen.findByText("Coffee");

    await openActions(user, "Import Statement");

    expect(
      await screen.findByRole("dialog", { name: "Select Import Method" })
    ).toBeInTheDocument();
  });

  it("opens the add transaction dialog", async () => {
    const user = userEvent.setup();
    server.use(...baseHandlers());
    renderWithProviders(<TransactionPage />);
    await screen.findByText("Coffee");

    await user.click(screen.getByRole("button", { name: /add transaction/i }));

    expect(
      await screen.findByRole("dialog", { name: "Add New Transaction" })
    ).toBeInTheDocument();
  });

  it("logs deletion failures without announcing success", async () => {
    const user = userEvent.setup();
    server.use(...baseHandlers());
    server.use(
      http.delete("*/api/v1/transaction/:id", () =>
        HttpResponse.json({ error: "boom" }, { status: 500 })
      )
    );
    renderWithProviders(<TransactionPage />);
    await screen.findByText("Coffee");

    await user.click(screen.getByRole("checkbox", { name: "Select Coffee" }));
    await openActions(user, "Delete");

    await waitFor(() =>
      expect(consoleError).toHaveBeenCalledWith(
        "Failed to delete transaction"
      )
    );
  });
});
