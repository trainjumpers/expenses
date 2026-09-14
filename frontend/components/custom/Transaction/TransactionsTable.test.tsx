import type { Transaction } from "@/lib/models/transaction";
import { testAccount, testCategory } from "@/test/msw/handlers";
import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import type { ComponentProps } from "react";
import { describe, expect, it, vi } from "vitest";

import { TransactionsTable } from "./TransactionsTable";

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
    refresh: vi.fn(),
    back: vi.fn(),
    forward: vi.fn(),
    prefetch: vi.fn(),
  }),
  usePathname: () => "/transaction",
  useSearchParams: () => new URLSearchParams(),
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

function setup(
  overrides: Partial<ComponentProps<typeof TransactionsTable>> = {}
) {
  const props: ComponentProps<typeof TransactionsTable> = {
    selectedRows: new Set<number>(),
    setSelectedRows: vi.fn(),
    transactions,
    loading: false,
    error: null,
    currentPage: 1,
    setCurrentPage: vi.fn(),
    total: transactions.length,
    pageSize: 15,
    sortBy: "date",
    sortOrder: "desc",
    setSortBy: vi.fn(),
    setSortOrder: vi.fn(),
    onFilterChange: vi.fn(),
    ...overrides,
  };

  return { ...props, ...renderWithProviders(<TransactionsTable {...props} />) };
}

describe("TransactionsTable", () => {
  it("shows a skeleton while loading", () => {
    const { container } = setup({ loading: true, transactions: [] });

    expect(
      container.querySelectorAll('[data-slot="skeleton"]').length
    ).toBeGreaterThan(0);
    expect(screen.queryByText("No transactions found")).not.toBeInTheDocument();
  });

  it("shows the error message", () => {
    setup({ error: "Failed to get transactions" });

    expect(screen.getByText("Failed to get transactions")).toBeInTheDocument();
  });

  it("disables pagination when there are no transactions", () => {
    setup({ transactions: [], total: 0 });

    expect(screen.getByText("No transactions found")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Previous" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Next" })).toBeDisabled();
  });

  it("renders amounts, categories, accounts and dates", async () => {
    setup();

    expect(await screen.findByText("Coffee")).toBeInTheDocument();
    expect(screen.getByText("Refund for shoes")).toBeInTheDocument();
    expect(await screen.findByText("₹250.00")).toBeInTheDocument();
    expect(screen.getByText("₹500.00")).toBeInTheDocument();
    expect(await screen.findByText("Food")).toBeInTheDocument();
    expect(await screen.findAllByText("HDFC Savings")).toHaveLength(2);
    expect(screen.getByText("2026-08-15")).toBeInTheDocument();
  });

  it("sorts by name without clearing the amount filters", async () => {
    const user = userEvent.setup();
    const { setSortBy, setSortOrder, onFilterChange } = setup();

    await user.click(screen.getByRole("button", { name: "Name" }));

    expect(setSortBy).toHaveBeenCalledWith("name");
    expect(setSortOrder).toHaveBeenCalledWith("asc");
    expect(onFilterChange).not.toHaveBeenCalled();
  });

  it("filters credits from the credit column", async () => {
    const user = userEvent.setup();
    const { setSortBy, setSortOrder, onFilterChange } = setup();

    await user.click(screen.getByRole("button", { name: "Credit" }));

    expect(setSortBy).toHaveBeenCalledWith("amount");
    expect(setSortOrder).toHaveBeenCalledWith("asc");
    expect(onFilterChange).toHaveBeenCalledWith({
      maxAmount: -0.01,
      minAmount: undefined,
    });
  });

  it("filters debits from the debit column", async () => {
    const user = userEvent.setup();
    const { onFilterChange } = setup();

    await user.click(screen.getByRole("button", { name: "Debit" }));

    expect(onFilterChange).toHaveBeenCalledWith({
      minAmount: 0.01,
      maxAmount: undefined,
    });
  });

  it("selects every row", async () => {
    const user = userEvent.setup();
    const { setSelectedRows } = setup();

    await user.click(screen.getByRole("checkbox", { name: "Select all" }));

    expect(setSelectedRows).toHaveBeenCalledWith(new Set([1, 2]));
  });

  it("toggles a single row", async () => {
    const user = userEvent.setup();
    const setSelectedRows = vi.fn();
    setup({ setSelectedRows });

    await user.click(screen.getByRole("checkbox", { name: "Select Coffee" }));

    const updater = setSelectedRows.mock.calls[0][0] as (
      previous: Set<number>
    ) => Set<number>;
    expect(updater(new Set())).toEqual(new Set([1]));
    expect(updater(new Set([1, 2]))).toEqual(new Set([2]));
  });

  it("toggles the order when sorting the active column", async () => {
    const user = userEvent.setup();
    const { setSortBy, setSortOrder } = setup({
      sortBy: "name",
      sortOrder: "asc",
    });

    await user.click(screen.getByRole("button", { name: "Name" }));

    expect(setSortOrder).toHaveBeenCalledWith("desc");
    expect(setSortBy).not.toHaveBeenCalled();
  });

  it("sorts by description and date", async () => {
    const user = userEvent.setup();
    const { setSortBy } = setup({ sortBy: "name" });

    await user.click(screen.getByRole("button", { name: "Description" }));
    expect(setSortBy).toHaveBeenCalledWith("description");

    await user.click(screen.getByRole("button", { name: "Date" }));
    expect(setSortBy).toHaveBeenCalledWith("date");
  });

  it("selects every row from the header checkbox", async () => {
    const user = userEvent.setup();
    const setSelectedRows = vi.fn();
    setup({ setSelectedRows });

    await screen.findByText("Coffee");
    await user.click(screen.getByRole("checkbox", { name: "Select all" }));

    expect(setSelectedRows).toHaveBeenCalledWith(new Set([1, 2]));
  });

  it("clears the selection when every row is selected", async () => {
    const user = userEvent.setup();
    const setSelectedRows = vi.fn();
    setup({ selectedRows: new Set([1, 2]), setSelectedRows });

    await screen.findByText("Coffee");
    await user.click(screen.getByRole("checkbox", { name: "Select all" }));

    expect(setSelectedRows).toHaveBeenCalledWith(new Set());
  });

  it("adds and removes categories inline", async () => {
    const user = userEvent.setup();
    const patchBodies: unknown[] = [];
    server.use(
      http.get("*/api/v1/category", () =>
        HttpResponse.json({
          message: "ok",
          data: [testCategory, { id: 2, name: "Travel", created_by: 1 }],
        })
      ),
      http.patch("*/api/v1/transaction/1", async ({ request }) => {
        patchBodies.push(await request.json());
        return HttpResponse.json({ data: transactions[0] });
      })
    );
    setup();
    await screen.findByText("Coffee");
    await screen.findByText("Food");

    await user.click(screen.getByText("Food"));
    await user.click(await screen.findByRole("button", { name: "Travel" }));
    await waitFor(() =>
      expect(patchBodies[0]).toEqual({ category_ids: [1, 2] })
    );

    await user.click(screen.getByText("Food"));
    await user.click(await screen.findByRole("button", { name: "Food" }));
    await waitFor(() => expect(patchBodies[1]).toEqual({ category_ids: [] }));
  });

  it("changes the account inline and ignores a same-account pick", async () => {
    const user = userEvent.setup();
    const patchBodies: unknown[] = [];
    server.use(
      http.get("*/api/v1/account", () =>
        HttpResponse.json({
          message: "ok",
          data: [
            testAccount,
            {
              id: 2,
              name: "ICICI Salary",
              bank_type: "icici",
              currency: "inr",
              created_by: 1,
            },
          ],
        })
      ),
      http.patch("*/api/v1/transaction/1", async ({ request }) => {
        patchBodies.push(await request.json());
        return HttpResponse.json({ data: transactions[0] });
      })
    );
    setup();
    await screen.findByText("Coffee");
    await screen.findAllByText("HDFC Savings");

    await user.click(screen.getAllByText("HDFC Savings")[0]);
    await user.click(
      await screen.findByRole("button", { name: "ICICI Salary" })
    );
    await waitFor(() => expect(patchBodies[0]).toEqual({ account_id: 2 }));

    await user.click(screen.getAllByText("HDFC Savings")[0]);
    await user.click(
      await screen.findByRole("button", { name: "HDFC Savings" })
    );
    await waitFor(() => expect(patchBodies).toHaveLength(1));
  });
});
