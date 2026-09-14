import type { Account } from "@/lib/models/account";
import { testAccount } from "@/test/msw/handlers";
import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { describe, expect, it, vi } from "vitest";

import { TransactionForm } from "./TransactionForm";

const initialValues = {
  name: "",
  description: "",
  amount: "",
  date: new Date("2026-09-01"),
  category_ids: [],
  account_id: 0,
};

function formProps() {
  return {
    accounts: [],
    categories: [],
    onSubmit: vi.fn(),
    loading: false,
    submitText: "Add",
    onOpenChange: vi.fn(),
    initialValues,
  };
}

describe("TransactionForm", () => {
  it("keeps typed values when the parent re-renders", async () => {
    const user = userEvent.setup();
    const { rerender } = renderWithProviders(
      <TransactionForm {...formProps()} />
    );

    const name = screen.getByPlaceholderText("Enter transaction name");
    await user.type(name, "Coffee");
    await user.type(screen.getByPlaceholderText("Enter amount"), "12");

    // The modals build a new initialValues object on every render.
    rerender(<TransactionForm {...formProps()} />);

    expect(name).toHaveValue("Coffee");
    expect(screen.getByPlaceholderText("Enter amount")).toHaveValue(12);
  });

  it("submits the collected values", async () => {
    const user = userEvent.setup();
    const props = formProps();
    renderWithProviders(<TransactionForm {...props} />);

    await user.type(
      screen.getByPlaceholderText("Enter transaction name"),
      "Coffee"
    );
    await user.type(screen.getByPlaceholderText("Enter amount"), "12");
    await user.click(screen.getByRole("button", { name: "Add" }));

    expect(props.onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({ name: "Coffee", amount: "12" })
    );
  });

  const accounts: Account[] = [
    testAccount,
    {
      id: 2,
      name: "ICICI Salary",
      bank_type: "icici",
      currency: "inr",
      created_by: 1,
    },
  ];
  const categories = [
    { id: 1, name: "Food", created_by: 1 },
    { id: 2, name: "Travel", created_by: 1 },
  ];

  function setup(overrides: Record<string, unknown> = {}) {
    const props = { ...formProps(), accounts, categories, ...overrides };
    renderWithProviders(<TransactionForm {...props} />);
    return props;
  }

  it("selects an account and closes from cancel", async () => {
    const user = userEvent.setup();
    const props = setup();

    await user.click(
      await screen.findByRole("button", { name: /HDFC Savings/ })
    );
    await user.click(
      await screen.findByRole("menuitem", { name: "ICICI Salary" })
    );
    expect(
      screen.getByRole("button", { name: /ICICI Salary/ })
    ).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Cancel" }));
    expect(props.onOpenChange).toHaveBeenCalledWith(false);
  });

  it("toggles a category on and off", async () => {
    const user = userEvent.setup();
    setup();

    await user.click(
      screen.getByRole("button", { name: /select categories/i })
    );
    await user.click(await screen.findByRole("menuitem", { name: "Food" }));
    expect(screen.getByRole("button", { name: "Food" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Food" }));
    await user.click(await screen.findByRole("menuitem", { name: "Food" }));
    expect(
      screen.getByRole("button", { name: /select categories/i })
    ).toBeInTheDocument();
  });

  it("truncates very long category selections", async () => {
    const user = userEvent.setup();
    const longName = "Groceries ".repeat(12).trim();
    const props = setup({
      categories: [{ id: 3, name: longName, created_by: 1 }],
    });

    await user.click(
      screen.getByRole("button", { name: /select categories/i })
    );
    await user.click(await screen.findByRole("menuitem", { name: longName }));

    expect(
      screen.getByRole("button", { name: `${longName.slice(0, 100)}...` })
    ).toBeInTheDocument();
    expect(props.onSubmit).not.toHaveBeenCalled();
  });

  it("adds a new account and picks it for the transaction", async () => {
    const user = userEvent.setup();
    let created: unknown;
    server.use(
      http.post("*/api/v1/account", async ({ request }) => {
        created = await request.json();
        return HttpResponse.json(
          { data: { id: 2, name: "ICICI Salary" } },
          { status: 201 }
        );
      })
    );
    const props = setup();

    await user.click(
      await screen.findByRole("button", { name: /HDFC Savings/ })
    );
    await user.click(
      await screen.findByRole("menuitem", { name: "+ Add new account" })
    );
    expect(
      await screen.findByRole("dialog", { name: "Add Account" })
    ).toBeInTheDocument();

    await user.type(
      screen.getByPlaceholderText("Enter account name"),
      "New Bank"
    );
    await user.click(screen.getAllByRole("combobox")[0]);
    await user.click(await screen.findByRole("option", { name: "HDFC Bank" }));
    await user.click(screen.getByRole("button", { name: "Add Account" }));

    await waitFor(() =>
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument()
    );
    expect(created).toEqual({
      name: "New Bank",
      bank_type: "hdfc",
      currency: "inr",
      balance: undefined,
      current_value: undefined,
    });

    expect(
      screen.getByRole("button", { name: /ICICI Salary/ })
    ).toBeInTheDocument();
    expect(props.onSubmit).not.toHaveBeenCalled();

    await user.type(
      screen.getByPlaceholderText("Enter transaction name"),
      "Coffee"
    );
    await user.type(screen.getByPlaceholderText("Enter amount"), "12");
    await user.click(screen.getByRole("button", { name: "Add" }));
    expect(props.onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({ account_id: 2, name: "Coffee" })
    );
  });

  it("opens the add category dialog", async () => {
    const user = userEvent.setup();
    setup();

    await user.click(
      screen.getByRole("button", { name: /select categories/i })
    );
    await user.click(
      await screen.findByRole("menuitem", { name: "+ Add new category" })
    );

    expect(
      await screen.findByRole("dialog", { name: "Add Category" })
    ).toBeInTheDocument();
  });

  it("picks a date from the calendar", async () => {
    const user = userEvent.setup();
    setup();

    await user.click(
      screen.getByRole("button", {
        name: new Date(2026, 8, 1).toLocaleDateString(),
      })
    );
    await user.click(
      await screen.findByRole("button", { name: /September 15th, 2026/ })
    );

    expect(
      screen.getByRole("button", {
        name: new Date(2026, 8, 15).toLocaleDateString(),
      })
    ).toBeInTheDocument();
  });
});
