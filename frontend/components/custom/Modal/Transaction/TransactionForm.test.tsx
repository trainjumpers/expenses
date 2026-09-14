import { renderWithProviders } from "@/test/render";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
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
});
