import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import RootLayout from "./layout";

vi.mock("./globals.css", () => ({}));

vi.mock("next/font/google", () => ({
  Geist: () => ({ variable: "--font-geist-sans" }),
  Geist_Mono: () => ({ variable: "--font-geist-mono" }),
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

describe("RootLayout", () => {
  it("wraps pages in the theme, query and auth providers", async () => {
    render(
      <RootLayout>
        <p>Page content</p>
      </RootLayout>
    );

    expect(await screen.findByText("Page content")).toBeInTheDocument();
  });
});
