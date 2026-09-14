import type { TransactionFiltersState } from "@/app/transaction/page";
import type { Account } from "@/lib/models/account";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { AccountFilter } from "./AccountFilter";

const filters: TransactionFiltersState = {
  accountId: undefined,
  categoryId: undefined,
  uncategorized: undefined,
  minAmount: undefined,
  maxAmount: undefined,
  dateFrom: undefined,
  dateTo: undefined,
  search: "",
};

const accounts: Account[] = [
  {
    id: 1,
    name: "HDFC Savings",
    bank_type: "hdfc",
    currency: "inr",
    created_by: 1,
  },
  {
    id: 2,
    name: "ICICI Salary",
    bank_type: "icici",
    currency: "inr",
    created_by: 1,
  },
];

describe("AccountFilter", () => {
  it("lists the accounts and picks one", async () => {
    const user = userEvent.setup();
    const setFilters = vi.fn();
    render(
      <AccountFilter
        filters={filters}
        setFilters={setFilters}
        accounts={accounts}
      />
    );

    expect(
      screen.getByRole("option", { name: "HDFC Savings" })
    ).toBeInTheDocument();

    await user.selectOptions(screen.getByRole("combobox"), "2");

    expect(setFilters).toHaveBeenCalledWith({ ...filters, accountId: 2 });
  });

  it("goes back to all accounts", async () => {
    const user = userEvent.setup();
    const setFilters = vi.fn();
    render(
      <AccountFilter
        filters={{ ...filters, accountId: 2 }}
        setFilters={setFilters}
        accounts={accounts}
      />
    );

    await user.selectOptions(screen.getByRole("combobox"), "");

    expect(setFilters).toHaveBeenCalledWith({
      ...filters,
      accountId: undefined,
    });
  });
});
