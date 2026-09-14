import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { describe, expect, it, vi } from "vitest";

import { InfoCenterModal } from "./InfoCenterModal";

const nav = vi.hoisted(() => ({ push: vi.fn() }));

vi.mock("@/components/ui/icon-picker", () => ({
  Icon: () => null,
  IconPicker: () => null,
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: nav.push,
    replace: vi.fn(),
    refresh: vi.fn(),
    back: vi.fn(),
    forward: vi.fn(),
    prefetch: vi.fn(),
  }),
  usePathname: () => "/",
  useSearchParams: () => new URLSearchParams(),
}));

function render() {
  return renderWithProviders(<InfoCenterModal isOpen onOpenChange={vi.fn()} />);
}

describe("InfoCenterModal", () => {
  it("lists the places to view", () => {
    render();

    expect(
      screen.getByRole("dialog", { name: "View Center" })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /accounts/i })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /categories/i })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /transactions/i })
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /rules/i })).toBeInTheDocument();
  });

  it("opens the accounts list", async () => {
    const user = userEvent.setup();
    render();

    await user.click(screen.getByRole("button", { name: /accounts/i }));

    expect(
      await screen.findByRole("dialog", { name: "View Accounts" })
    ).toBeInTheDocument();
  });

  it("opens the categories list", async () => {
    const user = userEvent.setup();
    render();

    await user.click(screen.getByRole("button", { name: /categories/i }));

    expect(
      await screen.findByRole("dialog", { name: "Categories" })
    ).toBeInTheDocument();
  });

  it("opens the rules list", async () => {
    const user = userEvent.setup();
    server.use(
      http.get("*/api/v1/rule", () =>
        HttpResponse.json({
          data: { rules: [], total: 0, page: 1, page_size: 5 },
        })
      )
    );
    render();

    await user.click(screen.getByRole("button", { name: /rules/i }));

    expect(
      await screen.findByRole("dialog", { name: "View Rules" })
    ).toBeInTheDocument();
  });

  it("navigates to the transaction page", async () => {
    const user = userEvent.setup();
    render();

    await user.click(screen.getByRole("button", { name: /transactions/i }));

    expect(nav.push).toHaveBeenCalledWith("/transaction");
  });
});
