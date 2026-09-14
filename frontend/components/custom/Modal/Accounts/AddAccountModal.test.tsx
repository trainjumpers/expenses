import { AddAccountModal } from "@/components/custom/Modal/Accounts/AddAccountModal";
import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { toast } from "sonner";
import { describe, expect, it, vi } from "vitest";

function setup() {
  const onOpenChange = vi.fn();
  const onAccountAdded = vi.fn();
  renderWithProviders(
    <AddAccountModal
      isOpen
      onOpenChange={onOpenChange}
      onAccountAdded={onAccountAdded}
    />
  );
  return { onOpenChange, onAccountAdded };
}

async function selectOption(
  user: ReturnType<typeof userEvent.setup>,
  index: number,
  option: string
) {
  await user.click(screen.getAllByRole("combobox")[index]);
  await user.click(await screen.findByRole("option", { name: option }));
}

describe("AddAccountModal", () => {
  it("requires the name and bank", async () => {
    const user = userEvent.setup();
    setup();

    await user.click(screen.getByRole("button", { name: "Add Account" }));

    expect(toast.error).toHaveBeenCalledWith(
      "Please fill all required fields."
    );
  });

  it("creates a bank account", async () => {
    const user = userEvent.setup();
    let body: unknown;
    server.use(
      http.post("*/api/v1/account", async ({ request }) => {
        body = await request.json();
        return HttpResponse.json(
          {
            message: "Account created successfully",
            data: { id: 9, name: "ICICI", bank_type: "icici", currency: "inr" },
          },
          { status: 201 }
        );
      })
    );
    const { onOpenChange, onAccountAdded } = setup();

    await user.type(screen.getByPlaceholderText("Enter account name"), "ICICI");
    await selectOption(user, 0, "ICICI Bank");
    await user.type(
      screen.getByPlaceholderText("Enter initial balance"),
      "5000"
    );
    await user.click(screen.getByRole("button", { name: "Add Account" }));

    await waitFor(() =>
      expect(toast.success).toHaveBeenCalledWith("Account created successfully")
    );
    expect(body).toEqual({
      name: "ICICI",
      bank_type: "icici",
      currency: "inr",
      balance: 5000,
      current_value: undefined,
    });
    expect(onOpenChange).toHaveBeenCalledWith(false);
    expect(onAccountAdded).toHaveBeenCalledWith(
      expect.objectContaining({ id: 9 })
    );
  });

  it("collects the current value for an investment account", async () => {
    const user = userEvent.setup();
    let body: unknown;
    server.use(
      http.post("*/api/v1/account", async ({ request }) => {
        body = await request.json();
        return HttpResponse.json(
          { data: { id: 10, name: "Zerodha", bank_type: "investment" } },
          { status: 201 }
        );
      })
    );
    setup();

    await user.type(
      screen.getByPlaceholderText("Enter account name"),
      "Zerodha"
    );
    await selectOption(user, 0, "Investment Account");
    expect(
      await screen.findByPlaceholderText("Enter current value")
    ).toBeInTheDocument();
    await user.type(
      screen.getByPlaceholderText("Enter current value"),
      "75000"
    );
    await user.click(screen.getByRole("button", { name: "Add Account" }));

    await waitFor(() => expect(body).toMatchObject({ current_value: 75000 }));
  });

  it("closes without saving", async () => {
    const user = userEvent.setup();
    const { onOpenChange } = setup();

    await user.click(screen.getByRole("button", { name: "Cancel" }));

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
