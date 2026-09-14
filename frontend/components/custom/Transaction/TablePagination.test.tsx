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

describe("TablePagination long lists", () => {
  it("shows the edges, middle pages and both ellipses", async () => {
    const setCurrentPage = vi.fn();
    render(
      <TablePagination
        currentPage={5}
        totalPages={10}
        setCurrentPage={setCurrentPage}
      />
    );

    expect(screen.getByRole("button", { name: "1" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "4" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "5" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "6" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "10" })).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "1" }));
    expect(setCurrentPage).toHaveBeenCalledWith(1);

    await userEvent.click(screen.getByRole("button", { name: "10" }));
    expect(setCurrentPage).toHaveBeenCalledWith(10);

    await userEvent.click(screen.getByRole("button", { name: "Previous" }));
    expect(setCurrentPage).toHaveBeenCalledWith(4);
  });
});
