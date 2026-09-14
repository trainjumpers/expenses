import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { toast } from "sonner";
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

  it("rejects an inverted amount range without applying", async () => {
    const user = userEvent.setup();
    const { onFilterChange } = setup();

    await user.click(screen.getByRole("button", { name: "Filter" }));
    await user.click(screen.getByRole("button", { name: "Amount" }));

    const [minAmount, maxAmount] = screen.getAllByRole("spinbutton");
    await user.type(minAmount, "5000");
    await user.type(maxAmount, "100");
    await user.click(screen.getByRole("button", { name: "Apply" }));

    expect(toast.error).toHaveBeenCalledWith(
      "Min amount cannot exceed max amount"
    );
    expect(onFilterChange).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: "Apply" })).toBeInTheDocument();
  });

  it("switches between the filter panels", async () => {
    const user = userEvent.setup();
    setup();

    await user.click(screen.getByRole("button", { name: "Filter" }));
    await user.click(screen.getByRole("button", { name: /^amount$/i }));
    expect(screen.getByText("Min Amount")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /^date$/i }));
    expect(screen.getByText("Date From")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /^category$/i }));
    expect(
      screen.getByRole("option", { name: "Uncategorized" })
    ).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /^account$/i }));
    expect(screen.getByRole("option", { name: "All" })).toBeInTheDocument();
  });

  it("commits the search on blur", async () => {
    const user = userEvent.setup();
    const { onFilterChange } = setup();

    const input = screen.getByPlaceholderText("Search transactions ...");
    await user.type(input, "coffee");
    await user.tab();

    expect(onFilterChange).toHaveBeenCalledWith({ search: "coffee" });
  });
});
