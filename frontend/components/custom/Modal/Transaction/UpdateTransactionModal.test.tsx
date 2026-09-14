import type { Transaction } from "@/lib/models/transaction";
import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { afterAll, describe, expect, it, vi } from "vitest";

import { UpdateTransactionModal } from "./UpdateTransactionModal";

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

const transaction: Transaction = {
  id: 1,
  date: "2026-08-15T00:00:00.000Z",
  name: "Coffee",
  description: "Morning coffee",
  amount: 250,
  category_ids: [1],
  account_id: 1,
};

function setup(initial: Transaction | null = transaction) {
  const onOpenChange = vi.fn();
  renderWithProviders(
    <UpdateTransactionModal
      isOpen
      onOpenChange={onOpenChange}
      transaction={initial}
    />
  );
  return { onOpenChange };
}

describe("UpdateTransactionModal", () => {
  const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});

  afterAll(() => {
    consoleError.mockRestore();
  });

  it("renders nothing without a transaction", () => {
    setup(null);

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("saves the edited transaction", async () => {
    const user = userEvent.setup();
    let body: unknown;
    server.use(
      http.patch("*/api/v1/transaction/1", async ({ request }) => {
        body = await request.json();
        return HttpResponse.json({ data: { ...transaction, name: "Latte" } });
      })
    );
    const { onOpenChange } = setup();

    const name = screen.getByDisplayValue("Coffee");
    await user.clear(name);
    await user.type(name, "Latte");
    await user.click(screen.getByRole("button", { name: "Update" }));

    await waitFor(() => expect(onOpenChange).toHaveBeenCalledWith(false));
    expect(body).toMatchObject({
      name: "Latte",
      description: "Morning coffee",
      amount: 250,
      category_ids: [1],
      account_id: 1,
    });
  });

  it("keeps the dialog open when the update fails", async () => {
    const user = userEvent.setup();
    server.use(
      http.patch("*/api/v1/transaction/1", () =>
        HttpResponse.json({ error: "boom" }, { status: 500 })
      )
    );
    const { onOpenChange } = setup();

    await user.click(screen.getByRole("button", { name: "Update" }));

    await waitFor(() => expect(consoleError).toHaveBeenCalledWith("boom"));
    expect(onOpenChange).not.toHaveBeenCalledWith(false);
  });
});
