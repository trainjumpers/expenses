import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { AuthGuard } from "./AuthGuard";

const mockUseSession = vi.fn();
const mockPathname = vi.fn();

vi.mock("@/components/hooks/useSession", () => ({
  PUBLIC_ROUTES: ["/login", "/signup"],
  useSession: () => mockUseSession(),
}));

vi.mock("next/navigation", () => ({
  usePathname: () => mockPathname(),
}));

function renderGuard() {
  return render(
    <AuthGuard>
      <div>dashboard content</div>
    </AuthGuard>
  );
}

describe("AuthGuard", () => {
  it("renders children on public routes regardless of session state", () => {
    mockPathname.mockReturnValue("/login");
    mockUseSession.mockReturnValue({ isAuthenticated: false, isLoading: true });

    renderGuard();

    expect(screen.getByText("dashboard content")).toBeInTheDocument();
  });

  it("renders the skeleton while the session is loading", () => {
    mockPathname.mockReturnValue("/");
    mockUseSession.mockReturnValue({ isAuthenticated: false, isLoading: true });

    const { container } = renderGuard();

    expect(screen.queryByText("dashboard content")).not.toBeInTheDocument();
    expect(
      container.querySelectorAll('[data-slot="skeleton"]').length
    ).toBeGreaterThan(0);
  });

  it("hides children for an unauthenticated session", () => {
    mockPathname.mockReturnValue("/");
    mockUseSession.mockReturnValue({
      isAuthenticated: false,
      isLoading: false,
    });

    renderGuard();

    expect(screen.queryByText("dashboard content")).not.toBeInTheDocument();
  });

  it("renders children for an authenticated session", () => {
    mockPathname.mockReturnValue("/");
    mockUseSession.mockReturnValue({ isAuthenticated: true, isLoading: false });

    renderGuard();

    expect(screen.getByText("dashboard content")).toBeInTheDocument();
  });
});
