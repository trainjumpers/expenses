import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { describe, expect, it, vi } from "vitest";

import { AddTransactionModal } from "./AddTransactionModal";

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

describe("AddTransactionModal", () => {
  it("creates a transaction for the default account", async () => {
    const user = userEvent.setup();
    let body: unknown;
    server.use(
      http.post("*/api/v1/transaction", async ({ request }) => {
        body = await request.json();
        return HttpResponse.json({ data: { id: 5 } }, { status: 201 });
      })
    );
    const onOpenChange = vi.fn();
    renderWithProviders(
      <AddTransactionModal isOpen onOpenChange={onOpenChange} />
    );

    await user.type(
      await screen.findByPlaceholderText("Enter transaction name"),
      "Coffee"
    );
    await user.type(screen.getByPlaceholderText("Enter amount"), "12");
    await user.click(screen.getByRole("button", { name: "Add" }));

    await waitFor(() => expect(onOpenChange).toHaveBeenCalledWith(false));
    expect(body).toMatchObject({
      name: "Coffee",
      amount: 12,
      account_id: 1,
    });
  });
});
