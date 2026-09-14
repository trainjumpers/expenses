import { renderWithProviders } from "@/test/render";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { RuleActions } from "./RuleActions";

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

describe("RuleActions", () => {
  it("adds an empty category action", async () => {
    const user = userEvent.setup();
    const onActionsChange = vi.fn();
    renderWithProviders(
      <RuleActions
        actions={[{ action_type: "name", action_value: "coffee" }]}
        onActionsChange={onActionsChange}
      />
    );

    await user.click(screen.getByRole("button", { name: /add action/i }));

    expect(onActionsChange).toHaveBeenCalledWith([
      { action_type: "name", action_value: "coffee" },
      { action_type: "category", action_value: "" },
    ]);
  });

  it("changes the action type", async () => {
    const user = userEvent.setup();
    const onActionsChange = vi.fn();
    renderWithProviders(
      <RuleActions
        actions={[{ action_type: "category", action_value: "1" }]}
        onActionsChange={onActionsChange}
      />
    );

    await user.selectOptions(screen.getAllByRole("combobox")[0], "set_name");

    expect(onActionsChange).toHaveBeenCalledWith([
      { action_type: "name", action_value: "1" },
    ]);
  });

  it("removes an action beyond the first", async () => {
    const user = userEvent.setup();
    const onActionsChange = vi.fn();
    renderWithProviders(
      <RuleActions
        actions={[
          { action_type: "category", action_value: "1" },
          { action_type: "name", action_value: "coffee" },
        ]}
        onActionsChange={onActionsChange}
      />
    );

    await user.click(
      screen.getAllByRole("button", { name: "Remove action" })[1]
    );

    expect(onActionsChange).toHaveBeenCalledWith([
      { action_type: "category", action_value: "1" },
    ]);
  });
});
