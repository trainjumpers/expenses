import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { ToggleTheme } from "./ToggleTheme";

const setTheme = vi.hoisted(() => vi.fn());

vi.mock("next-themes", () => ({
  useTheme: () => ({ setTheme }),
}));

describe("ToggleTheme", () => {
  it("offers the three theme choices", async () => {
    const user = userEvent.setup();
    render(<ToggleTheme />);

    await user.click(screen.getByRole("button", { name: "Toggle theme" }));
    await user.click(await screen.findByRole("menuitem", { name: "Dark" }));

    expect(setTheme).toHaveBeenCalledWith("dark");
  });

  it("switches back to light and system", async () => {
    const user = userEvent.setup();
    render(<ToggleTheme />);

    await user.click(screen.getByRole("button", { name: "Toggle theme" }));
    await user.click(await screen.findByRole("menuitem", { name: "Light" }));
    expect(setTheme).toHaveBeenCalledWith("light");

    await user.click(screen.getByRole("button", { name: "Toggle theme" }));
    await user.click(await screen.findByRole("menuitem", { name: "System" }));
    expect(setTheme).toHaveBeenCalledWith("system");
  });
});
