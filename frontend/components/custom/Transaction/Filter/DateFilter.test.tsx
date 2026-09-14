import type { TransactionFiltersState } from "@/app/transaction/page";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { format } from "date-fns";
import { describe, expect, it, vi } from "vitest";

import { DateFilter } from "./DateFilter";

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

function pickFifteenth() {
  const day = new Date();
  day.setDate(15);
  return { day, label: format(day, "EEEE, MMMM do, yyyy") };
}

describe("DateFilter", () => {
  it("shows the chosen dates", () => {
    render(
      <DateFilter
        filters={{ ...filters, dateFrom: "2026-09-01", dateTo: "2026-09-30" }}
        setFilters={vi.fn()}
      />
    );

    expect(
      screen.getByRole("button", {
        name: new Date("2026-09-01").toLocaleDateString(),
      })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", {
        name: new Date("2026-09-30").toLocaleDateString(),
      })
    ).toBeInTheDocument();
  });

  it("sets the from date", async () => {
    const user = userEvent.setup();
    const setFilters = vi.fn();
    const { day, label } = pickFifteenth();
    render(<DateFilter filters={filters} setFilters={setFilters} />);

    await user.click(
      screen.getAllByRole("button", { name: /select date/i })[0]
    );
    await user.click(await screen.findByRole("button", { name: label }));

    expect(setFilters).toHaveBeenCalledWith({
      ...filters,
      dateFrom: format(day, "yyyy-MM-dd"),
    });
  });

  it("sets the to date", async () => {
    const user = userEvent.setup();
    const setFilters = vi.fn();
    const { day, label } = pickFifteenth();
    render(<DateFilter filters={filters} setFilters={setFilters} />);

    await user.click(
      screen.getAllByRole("button", { name: /select date/i })[1]
    );
    await user.click(await screen.findByRole("button", { name: label }));

    expect(setFilters).toHaveBeenCalledWith({
      ...filters,
      dateTo: format(day, "yyyy-MM-dd"),
    });
  });

  it("clears the from date", async () => {
    const user = userEvent.setup();
    const setFilters = vi.fn();
    const { day, label } = pickFifteenth();
    const withDate = { ...filters, dateFrom: format(day, "yyyy-MM-dd") };
    render(<DateFilter filters={withDate} setFilters={setFilters} />);

    await user.click(
      screen.getByRole("button", {
        name: new Date(withDate.dateFrom).toLocaleDateString(),
      })
    );
    await user.click(
      await screen.findByRole("button", { name: new RegExp(label) })
    );

    expect(setFilters).toHaveBeenCalledWith({
      ...withDate,
      dateFrom: undefined,
    });
  });

  it("clears the to date", async () => {
    const user = userEvent.setup();
    const setFilters = vi.fn();
    const { day, label } = pickFifteenth();
    const withDate = { ...filters, dateTo: format(day, "yyyy-MM-dd") };
    render(<DateFilter filters={withDate} setFilters={setFilters} />);

    await user.click(
      screen.getByRole("button", {
        name: new Date(withDate.dateTo).toLocaleDateString(),
      })
    );
    await user.click(
      await screen.findByRole("button", { name: new RegExp(label) })
    );

    expect(setFilters).toHaveBeenCalledWith({
      ...withDate,
      dateTo: undefined,
    });
  });
});
