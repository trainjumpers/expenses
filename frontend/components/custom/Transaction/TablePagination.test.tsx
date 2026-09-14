import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import TablePagination from "./TablePagination";

describe("TablePagination", () => {
  it("disables both controls when there are no results", () => {
    render(
      <TablePagination
        currentPage={1}
        totalPages={0}
        setCurrentPage={vi.fn()}
      />
    );

    expect(screen.getByRole("button", { name: "Previous" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Next" })).toBeDisabled();
  });

  it("disables previous on the first page and next on the last page", () => {
    const { rerender } = render(
      <TablePagination
        currentPage={1}
        totalPages={3}
        setCurrentPage={vi.fn()}
      />
    );

    expect(screen.getByRole("button", { name: "Previous" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Next" })).toBeEnabled();

    rerender(
      <TablePagination
        currentPage={3}
        totalPages={3}
        setCurrentPage={vi.fn()}
      />
    );

    expect(screen.getByRole("button", { name: "Previous" })).toBeEnabled();
    expect(screen.getByRole("button", { name: "Next" })).toBeDisabled();
  });

  it("moves to the next page", async () => {
    const setCurrentPage = vi.fn();
    render(
      <TablePagination
        currentPage={1}
        totalPages={3}
        setCurrentPage={setCurrentPage}
      />
    );

    await userEvent.click(screen.getByRole("button", { name: "Next" }));

    expect(setCurrentPage).toHaveBeenCalledWith(2);
  });

  it("jumps to a numbered page", async () => {
    const setCurrentPage = vi.fn();
    render(
      <TablePagination
        currentPage={1}
        totalPages={3}
        setCurrentPage={setCurrentPage}
      />
    );

    await userEvent.click(screen.getByRole("button", { name: "3" }));

    expect(setCurrentPage).toHaveBeenCalledWith(3);
  });

  it("collapses long ranges around the current page", () => {
    render(
      <TablePagination
        currentPage={10}
        totalPages={20}
        setCurrentPage={vi.fn()}
      />
    );

    expect(screen.getByRole("button", { name: "1" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "9" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "10" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "11" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "20" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "5" })).not.toBeInTheDocument();
  });
});
