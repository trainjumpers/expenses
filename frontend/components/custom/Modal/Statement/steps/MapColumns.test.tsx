import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { MapColumns } from "./MapColumns";

const headers = ["Date", "Narration", "Amount", "Credit", "Debit"];

function setup() {
  const onStepChange = vi.fn();
  const onCancel = vi.fn();
  const onSubmit = vi.fn();

  render(
    <MapColumns
      headers={headers}
      onStepChange={onStepChange}
      onCancel={onCancel}
      onSubmit={onSubmit}
    />
  );

  return { onStepChange, onCancel, onSubmit };
}

async function mapField(
  user: ReturnType<typeof userEvent.setup>,
  index: number,
  option: string
) {
  await user.click(screen.getAllByRole("combobox")[index]);
  await user.click(await screen.findByRole("option", { name: option }));
}

describe("MapColumns", () => {
  it("requires the transaction date and name", async () => {
    const user = userEvent.setup();
    setup();

    await user.click(
      screen.getByRole("button", { name: "Process Transactions" })
    );
    expect(
      screen.getByText("Transaction Date must be mapped.")
    ).toBeInTheDocument();

    await mapField(user, 0, "Date");
    await user.click(
      screen.getByRole("button", { name: "Process Transactions" })
    );
    expect(screen.getByText("Name must be mapped.")).toBeInTheDocument();
  });

  it("requires an amount or a credit and debit pair", async () => {
    const user = userEvent.setup();
    setup();

    await mapField(user, 0, "Date");
    await mapField(user, 1, "Narration");
    await user.click(
      screen.getByRole("button", { name: "Process Transactions" })
    );

    expect(
      screen.getByText(
        "You must map either 'Amount' or both 'Credit' and 'Debit'."
      )
    ).toBeInTheDocument();
  });

  it("rejects mixing amount with credit or debit", async () => {
    const user = userEvent.setup();
    setup();

    await mapField(user, 0, "Date");
    await mapField(user, 1, "Narration");
    await mapField(user, 3, "Amount");
    await mapField(user, 4, "Credit");
    await user.click(
      screen.getByRole("button", { name: "Process Transactions" })
    );

    expect(
      screen.getByText(
        "You cannot map 'Amount' with 'Credit' or 'Debit'. Please choose one method."
      )
    ).toBeInTheDocument();
  });

  it("submits a single amount mapping", async () => {
    const user = userEvent.setup();
    const { onSubmit, onCancel } = setup();

    await mapField(user, 0, "Date");
    await mapField(user, 1, "Narration");
    await mapField(user, 3, "Amount");
    await user.click(
      screen.getByRole("button", { name: "Process Transactions" })
    );

    expect(onSubmit).toHaveBeenCalledWith({
      txn_date: "Date",
      name: "Narration",
      amount: "Amount",
    });
    expect(onCancel).toHaveBeenCalled();
  });

  it("submits a credit and debit mapping", async () => {
    const user = userEvent.setup();
    const { onSubmit } = setup();

    await mapField(user, 0, "Date");
    await mapField(user, 1, "Narration");
    await mapField(user, 4, "Credit");
    await mapField(user, 5, "Debit");
    await user.click(
      screen.getByRole("button", { name: "Process Transactions" })
    );

    expect(onSubmit).toHaveBeenCalledWith({
      txn_date: "Date",
      name: "Narration",
      credit: "Credit",
      debit: "Debit",
    });
  });

  it("clears a mapping set back to None", async () => {
    const user = userEvent.setup();
    setup();

    await mapField(user, 0, "Date");
    await mapField(user, 1, "Narration");
    await mapField(user, 3, "Amount");
    await mapField(user, 3, "None");
    await user.click(
      screen.getByRole("button", { name: "Process Transactions" })
    );

    expect(
      screen.getByText(
        "You must map either 'Amount' or both 'Credit' and 'Debit'."
      )
    ).toBeInTheDocument();
  });

  it("goes back to the preview step", async () => {
    const user = userEvent.setup();
    const { onStepChange } = setup();

    await user.click(screen.getByRole("button", { name: "Back" }));

    expect(onStepChange).toHaveBeenCalledWith(3);
  });
});
