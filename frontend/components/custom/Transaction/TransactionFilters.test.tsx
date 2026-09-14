import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import TransactionFilters from "./TransactionFilters";

const filters = {
  accountId: undefined,
  categoryId: undefined,
  uncategorized: undefined,
  minAmount: undefined,
  maxAmount: undefined,
  dateFrom: undefined,
  dateTo: undefined,
  search: "",
};

function setup() {
  const onFilterChange = vi.fn();
  const onClear = vi.fn();

  render(
    <TransactionFilters
      accounts={[]}
      categories={[]}
      filters={filters}
      onFilterChange={onFilterChange}
      onClear={onClear}
    />
  );

  return { onFilterChange, onClear };
}

describe("TransactionFilters", () => {
  it("submits the search term", async () => {
    const user = userEvent.setup();
    const { onFilterChange } = setup();

    await user.type(
      screen.getByPlaceholderText("Search transactions ..."),
      "coffee"
    );
    await user.keyboard("{Enter}");

    expect(onFilterChange).toHaveBeenCalledWith({ search: "coffee" });
  });

  it("applies amount filters", async () => {
    const user = userEvent.setup();
    const { onFilterChange } = setup();

    await user.click(screen.getByRole("button", { name: "Filter" }));
    await user.click(screen.getByRole("button", { name: "Amount" }));

    const [minAmount, maxAmount] = screen.getAllByRole("spinbutton");
    await user.type(minAmount, "100");
    await user.type(maxAmount, "5000");
    await user.click(screen.getByRole("button", { name: "Apply" }));

    expect(onFilterChange).toHaveBeenCalledWith(
      expect.objectContaining({ minAmount: 100, maxAmount: 5000 })
    );
  });

  it("hands the clear action to the page", async () => {
    const user = userEvent.setup();
    const { onClear } = setup();

    await user.click(screen.getByRole("button", { name: "Filter" }));
    await user.click(screen.getByRole("button", { name: "Clear" }));

    expect(onClear).toHaveBeenCalled();
  });
});
