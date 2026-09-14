import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { toast } from "sonner";
import { describe, expect, it, vi } from "vitest";

import { ProfileDropdown } from "./ProfileDropdown";

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

describe("ProfileDropdown", () => {
  it("shows the user initials once the profile loads", async () => {
    renderWithProviders(<ProfileDropdown />);

    expect(
      await screen.findByRole("button", { name: "TU" })
    ).toBeInTheDocument();
  });

  it("opens the menu with the profile actions", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ProfileDropdown />);

    await user.click(await screen.findByRole("button", { name: "TU" }));

    expect(await screen.findByText("Test User")).toBeInTheDocument();
    expect(screen.getByText("test1@example.com")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /profile/i })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /change password/i })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /view statements/i })
    ).toBeInTheDocument();
  });

  it("logs out from the menu", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ProfileDropdown />);
    await user.click(await screen.findByRole("button", { name: "TU" }));

    await user.click(screen.getByRole("button", { name: /log out/i }));

    await waitFor(() =>
      expect(toast.success).toHaveBeenCalledWith("Logged out successfully")
    );
  });

  it("opens the profile editor", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ProfileDropdown />);
    await user.click(await screen.findByRole("button", { name: "TU" }));

    await user.click(screen.getByRole("button", { name: /profile/i }));

    expect(
      await screen.findByRole("dialog", { name: "Edit Profile" })
    ).toBeInTheDocument();
  });

  it("opens the password dialog", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ProfileDropdown />);
    await user.click(await screen.findByRole("button", { name: "TU" }));

    await user.click(screen.getByRole("button", { name: /change password/i }));

    expect(
      await screen.findByRole("dialog", { name: "Change Password" })
    ).toBeInTheDocument();
  });

  it("opens the statements dialog", async () => {
    const user = userEvent.setup();
    server.use(
      http.get("*/api/v1/statement", () =>
        HttpResponse.json({
          data: { statements: [], total: 0, page: 1, page_size: 10 },
        })
      )
    );
    renderWithProviders(<ProfileDropdown />);
    await user.click(await screen.findByRole("button", { name: "TU" }));

    await user.click(screen.getByRole("button", { name: /view statements/i }));

    expect(
      await screen.findByRole("dialog", { name: "Statement History" })
    ).toBeInTheDocument();
  });

  it("falls back to a neutral avatar without a profile", async () => {
    server.use(
      http.get("*/api/v1/user", () => HttpResponse.json({ data: null }))
    );
    renderWithProviders(<ProfileDropdown />);

    await waitFor(() =>
      expect(screen.getByRole("button", { name: "" })).toBeInTheDocument()
    );
  });
});
