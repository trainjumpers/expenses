import type { TransactionFiltersState } from "@/app/transaction/page";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { CategoryFilter } from "./CategoryFilter";

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

const categories = [{ id: 1, name: "Food", created_by: 1 }];

describe("CategoryFilter", () => {
  it("picks a category", async () => {
    const user = userEvent.setup();
    const setFilters = vi.fn();
    render(
      <CategoryFilter
        filters={filters}
        setFilters={setFilters}
        categories={categories}
      />
    );

    await user.selectOptions(screen.getByRole("combobox"), "1");

    expect(setFilters).toHaveBeenCalledWith({
      ...filters,
      categoryId: 1,
      uncategorized: undefined,
    });
  });

  it("picks the uncategorized bucket", async () => {
    const user = userEvent.setup();
    const setFilters = vi.fn();
    render(
      <CategoryFilter
        filters={filters}
        setFilters={setFilters}
        categories={categories}
      />
    );

    await user.selectOptions(screen.getByRole("combobox"), "uncategorized");

    expect(setFilters).toHaveBeenCalledWith({
      ...filters,
      categoryId: undefined,
      uncategorized: true,
    });
  });

  it("clears the category", async () => {
    const user = userEvent.setup();
    const setFilters = vi.fn();
    render(
      <CategoryFilter
        filters={{ ...filters, categoryId: 1 }}
        setFilters={setFilters}
        categories={categories}
      />
    );

    await user.selectOptions(screen.getByRole("combobox"), "");

    expect(setFilters).toHaveBeenCalledWith({
      ...filters,
      categoryId: undefined,
      uncategorized: undefined,
    });
  });

  it("reflects the uncategorized state", () => {
    render(
      <CategoryFilter
        filters={{ ...filters, uncategorized: true }}
        setFilters={vi.fn()}
        categories={categories}
      />
    );

    expect(screen.getByRole("combobox")).toHaveValue("uncategorized");
  });
});
