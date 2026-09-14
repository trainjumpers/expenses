import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { describe, expect, it, vi } from "vitest";

import { CommandCenterModal } from "./CommandCenterModal";

vi.mock("@/components/ui/icon-picker", () => ({
  Icon: () => null,
  IconPicker: () => null,
}));

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

function render() {
  return renderWithProviders(
    <CommandCenterModal isOpen onOpenChange={vi.fn()} />
  );
}

describe("CommandCenterModal", () => {
  it("lists the quick actions", () => {
    render();

    expect(
      screen.getByRole("dialog", { name: "Command Center" })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /add account/i })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /add category/i })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /add transaction/i })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /add rule/i })
    ).toBeInTheDocument();
  });

  it("opens the account form", async () => {
    const user = userEvent.setup();
    render();

    await user.click(screen.getByRole("button", { name: /add account/i }));

    expect(
      await screen.findByRole("dialog", { name: "Add Account" })
    ).toBeInTheDocument();
  });

  it("opens the category form", async () => {
    const user = userEvent.setup();
    render();

    await user.click(screen.getByRole("button", { name: /add category/i }));

    expect(
      await screen.findByRole("dialog", { name: /add category/i })
    ).toBeInTheDocument();
  });

  it("opens the transaction form", async () => {
    const user = userEvent.setup();
    server.use(
      http.get("*/api/v1/account", () =>
        HttpResponse.json({ message: "ok", data: [] })
      ),
      http.get("*/api/v1/category", () =>
        HttpResponse.json({ message: "ok", data: [] })
      )
    );
    render();

    await user.click(screen.getByRole("button", { name: /add transaction/i }));

    expect(
      await screen.findByRole("dialog", { name: "Add New Transaction" })
    ).toBeInTheDocument();
  });

  it("opens the rule form", async () => {
    const user = userEvent.setup();
    render();

    await user.click(screen.getByRole("button", { name: /add rule/i }));

    expect(
      await screen.findByRole("dialog", { name: /rule/i })
    ).toBeInTheDocument();
  });
});
