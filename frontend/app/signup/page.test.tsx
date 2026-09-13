import { signupRequests } from "@/test/msw/handlers";
import { renderWithProviders } from "@/test/render";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { toast } from "sonner";
import { beforeEach, describe, expect, it, vi } from "vitest";

import SignupPage from "./page";

const mockPush = vi.fn();
const mockReplace = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: mockPush,
    replace: mockReplace,
    refresh: vi.fn(),
    back: vi.fn(),
    forward: vi.fn(),
    prefetch: vi.fn(),
  }),
  usePathname: () => "/signup",
  useSearchParams: () => new URLSearchParams(),
}));

describe("SignupPage", () => {
  beforeEach(() => {
    signupRequests.length = 0;
  });

  it("submits the account details and redirects home", async () => {
    const user = userEvent.setup();
    renderWithProviders(<SignupPage />);

    await user.type(screen.getByPlaceholderText("Name"), "New User");
    await user.type(screen.getByPlaceholderText("Email"), "new@example.com");
    await user.type(screen.getByPlaceholderText("Password"), "password123");
    await user.click(screen.getByRole("button", { name: /sign up/i }));

    await waitFor(() => expect(mockPush).toHaveBeenCalledWith("/"));
    expect(signupRequests).toEqual([
      { name: "New User", email: "new@example.com", password: "password123" },
    ]);
    expect(toast.success).toHaveBeenCalledWith("Account created successfully!");
  });

  it("shows a conflict toast when the account already exists", async () => {
    const user = userEvent.setup();
    renderWithProviders(<SignupPage />);

    await user.type(screen.getByPlaceholderText("Name"), "Existing User");
    await user.type(
      screen.getByPlaceholderText("Email"),
      "existing@example.com"
    );
    await user.type(screen.getByPlaceholderText("Password"), "password123");
    await user.click(screen.getByRole("button", { name: /sign up/i }));

    await waitFor(() =>
      expect(toast.info).toHaveBeenCalledWith(
        "Account already exists.",
        expect.objectContaining({
          action: expect.objectContaining({ label: "Login" }),
        })
      )
    );
    expect(mockPush).not.toHaveBeenCalled();
  });
});
