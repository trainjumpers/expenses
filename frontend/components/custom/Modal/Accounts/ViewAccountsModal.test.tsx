import { testAccount } from "@/test/msw/handlers";
import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { toast } from "sonner";
import { describe, expect, it, vi } from "vitest";

import { ViewAccountsModal } from "./ViewAccountsModal";

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

const iciciAccount = {
  id: 2,
  name: "ICICI Salary",
  bank_type: "icici",
  currency: "inr",
  created_by: 1,
} as const;

function render() {
  return renderWithProviders(
    <ViewAccountsModal isOpen onOpenChange={vi.fn()} />
  );
}

describe("ViewAccountsModal", () => {
  it("lists accounts with their bank and currency", async () => {
    render();

    expect(await screen.findByText("HDFC Savings")).toBeInTheDocument();
    expect(screen.getByText("HDFC - INR")).toBeInTheDocument();
  });

  it("filters accounts by the search term", async () => {
    const user = userEvent.setup();
    server.use(
      http.get("*/api/v1/account", () =>
        HttpResponse.json({ message: "ok", data: [testAccount, iciciAccount] })
      )
    );
    render();
    await screen.findByText("HDFC Savings");

    await user.type(screen.getByLabelText("Search accounts"), "icici");

    await waitFor(() =>
      expect(screen.queryByText("HDFC Savings")).not.toBeInTheDocument()
    );
    expect(screen.getByText("ICICI Salary")).toBeInTheDocument();
  });

  it("shows an empty state without accounts", async () => {
    server.use(
      http.get("*/api/v1/account", () =>
        HttpResponse.json({ message: "ok", data: [] })
      )
    );
    render();

    expect(
      await screen.findByText("No accounts found. Add one to get started!")
    ).toBeInTheDocument();
  });

  it("deletes an account after confirming", async () => {
    const user = userEvent.setup();
    server.use(
      http.delete(
        "*/api/v1/account/1",
        () => new HttpResponse(null, { status: 204 })
      )
    );
    render();
    await screen.findByText("HDFC Savings");

    const view = screen.getByRole("dialog", { name: "View Accounts" });
    await user.click(within(view).getByRole("button", { name: "Delete" }));

    const confirm = await screen.findByRole("dialog", {
      name: "Delete Account",
    });
    await user.click(within(confirm).getByRole("button", { name: "Delete" }));

    await waitFor(() =>
      expect(toast.success).toHaveBeenCalledWith("Account deleted successfully")
    );
    await waitFor(() =>
      expect(screen.queryByText("HDFC Savings")).not.toBeInTheDocument()
    );
  });

  it("opens the edit dialog for an account", async () => {
    const user = userEvent.setup();
    render();
    await screen.findByText("HDFC Savings");

    const view = screen.getByRole("dialog", { name: "View Accounts" });
    await user.click(within(view).getByRole("button", { name: "Edit" }));

    expect(
      await screen.findByRole("dialog", { name: "Update Account" })
    ).toBeInTheDocument();
  });

  it("pages through long account lists", async () => {
    const user = userEvent.setup();
    const many = Array.from({ length: 6 }, (_, i) => ({
      ...testAccount,
      id: i + 1,
      name: `Account ${i + 1}`,
    }));
    server.use(
      http.get("*/api/v1/account", () =>
        HttpResponse.json({ message: "ok", data: many })
      )
    );
    render();
    await screen.findByText("Account 1");
    expect(screen.queryByText("Account 6")).not.toBeInTheDocument();

    await user.click(screen.getByText("2"));

    expect(await screen.findByText("Account 6")).toBeInTheDocument();
  });
});
