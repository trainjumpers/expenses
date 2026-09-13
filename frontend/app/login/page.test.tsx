import { loginRequests, testUser } from "@/test/msw/handlers";
import { renderWithProviders } from "@/test/render";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { toast } from "sonner";
import { beforeEach, describe, expect, it, vi } from "vitest";

import LoginPage from "./page";

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
  usePathname: () => "/login",
  useSearchParams: () => new URLSearchParams(),
}));

describe("LoginPage", () => {
  beforeEach(() => {
    loginRequests.length = 0;
  });

  it("submits the credentials and redirects home", async () => {
    const user = userEvent.setup();
    renderWithProviders(<LoginPage />);

    await user.type(screen.getByPlaceholderText("Email"), testUser.email);
    await user.type(screen.getByPlaceholderText("Password"), "password123");
    await user.click(screen.getByRole("button", { name: /sign in/i }));

    await waitFor(() => expect(mockPush).toHaveBeenCalledWith("/"));
    expect(loginRequests).toEqual([
      { email: testUser.email, password: "password123" },
    ]);
    expect(toast.success).toHaveBeenCalledWith("Welcome back!");
  });

  it("stays on the page and shows an error for rejected credentials", async () => {
    const user = userEvent.setup();
    renderWithProviders(<LoginPage />);

    await user.type(screen.getByPlaceholderText("Email"), testUser.email);
    await user.type(screen.getByPlaceholderText("Password"), "wrong-password");
    await user.click(screen.getByRole("button", { name: /sign in/i }));

    await waitFor(() =>
      expect(toast.error).toHaveBeenCalledWith(
        "The email or password is incorrect",
        { id: "login-error" }
      )
    );
    expect(mockPush).not.toHaveBeenCalled();
  });
});
