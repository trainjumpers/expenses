import { testAccount } from "@/test/msw/handlers";
import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { toast } from "sonner";
import { afterAll, describe, expect, it, vi } from "vitest";

import { UpdateAccountModal } from "./UpdateAccountModal";

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
    refresh: vi.fn(),
    back: vi.fn(),
    forward: vi.fn(),
    prefetch: vi.fn(),
  }),
  usePathname: () => "/",
  useSearchParams: () => new URLSearchParams(),
}));

function setup(
  account: Parameters<typeof UpdateAccountModal>[0]["account"] = testAccount
) {
  const onOpenChange = vi.fn();
  const onAccountUpdated = vi.fn();
  renderWithProviders(
    <UpdateAccountModal
      isOpen
      onOpenChange={onOpenChange}
      account={account}
      onAccountUpdated={onAccountUpdated}
    />
  );
  return { onOpenChange, onAccountUpdated };
}

describe("UpdateAccountModal", () => {
  const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});

  afterAll(() => {
    consoleError.mockRestore();
  });

  it("saves the edited account", async () => {
    const user = userEvent.setup();
    let body: unknown;
    server.use(
      http.patch("*/api/v1/account/1", async ({ request }) => {
        body = await request.json();
        return HttpResponse.json({
          data: { ...testAccount, name: "HDFC Salary" },
        });
      })
    );
    const { onOpenChange, onAccountUpdated } = setup();

    const name = screen.getByDisplayValue("HDFC Savings");
    await user.clear(name);
    await user.type(name, "HDFC Salary");
    await user.click(screen.getByRole("button", { name: "Update" }));

    await waitFor(() =>
      expect(toast.success).toHaveBeenCalledWith("Account updated successfully")
    );
    expect(body).toEqual({
      name: "HDFC Salary",
      bank_type: "hdfc",
      currency: "inr",
      balance: 1000,
      current_value: undefined,
    });
    expect(onAccountUpdated).toHaveBeenCalled();
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("collects the current value for an investment account", async () => {
    const user = userEvent.setup();
    let body: unknown;
    server.use(
      http.patch("*/api/v1/account/2", async ({ request }) => {
        body = await request.json();
        return HttpResponse.json({ data: { id: 2 } });
      })
    );
    setup({
      id: 2,
      name: "Zerodha",
      bank_type: "investment",
      currency: "inr",
      balance: 0,
      current_value: 5000,
      created_by: 1,
    });

    const currentValue = await screen.findByDisplayValue("5000");
    await user.clear(currentValue);
    await user.type(currentValue, "7500");
    await user.click(screen.getByRole("button", { name: "Update" }));

    await waitFor(() =>
      expect(body).toMatchObject({
        bank_type: "investment",
        current_value: 7500,
      })
    );
  });

  it("keeps the dialog open when the update fails", async () => {
    const user = userEvent.setup();
    server.use(
      http.patch("*/api/v1/account/1", () =>
        HttpResponse.json({ error: "boom" }, { status: 500 })
      )
    );
    const { onOpenChange } = setup();

    await user.click(screen.getByRole("button", { name: "Update" }));

    await waitFor(() =>
      expect(consoleError).toHaveBeenCalledWith(
        "Failed to update account:",
        expect.any(Error)
      )
    );
    expect(onOpenChange).not.toHaveBeenCalledWith(false);
  });

  it("closes without saving", async () => {
    const user = userEvent.setup();
    const { onOpenChange } = setup();

    await user.click(screen.getByRole("button", { name: "Cancel" }));

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
