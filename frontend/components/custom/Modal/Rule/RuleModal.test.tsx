import { renderWithProviders } from "@/test/render";
import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { RuleModal } from "./RuleModal";

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

function setup(overrides: Partial<Parameters<typeof RuleModal>[0]> = {}) {
  const onSubmit = vi.fn().mockResolvedValue(undefined);
  const onOpenChange = vi.fn();

  renderWithProviders(
    <RuleModal
      isOpen
      onOpenChange={onOpenChange}
      mode="add"
      onSubmit={onSubmit}
      {...overrides}
    />
  );

  return { onSubmit, onOpenChange };
}

async function fillRequiredFields(
  user: ReturnType<typeof userEvent.setup>,
  { name = "Coffee rule", value = "coffee", category = "1" } = {}
) {
  await user.type(
    screen.getByPlaceholderText("Enter a name for this rule"),
    name
  );
  await user.type(screen.getByPlaceholderText("Enter a value"), value);

  const selects = screen.getAllByRole("combobox");
  const categorySelect = selects.find((select) =>
    within(select).queryByRole("option", { name: "Select category" })
  )!;
  await user.selectOptions(categorySelect, category);
}

describe("RuleModal", () => {
  it("requires a rule name", async () => {
    const user = userEvent.setup();
    setup();

    await user.click(screen.getByRole("button", { name: "Create Rule" }));

    expect(screen.getByText("Rule name is required.")).toBeInTheDocument();
  });

  it("requires every condition value", async () => {
    const user = userEvent.setup();
    setup();

    await user.type(
      screen.getByPlaceholderText("Enter a name for this rule"),
      "Coffee rule"
    );
    await user.click(screen.getByRole("button", { name: "Create Rule" }));

    expect(
      screen.getByText("All condition values must be filled.")
    ).toBeInTheDocument();
  });

  it("requires every action value", async () => {
    const user = userEvent.setup();
    setup();

    await user.type(
      screen.getByPlaceholderText("Enter a name for this rule"),
      "Coffee rule"
    );
    await user.type(screen.getByPlaceholderText("Enter a value"), "coffee");
    await user.click(screen.getByRole("button", { name: "Create Rule" }));

    expect(
      screen.getByText("All action values must be filled.")
    ).toBeInTheDocument();
  });

  it("submits the rule, conditions and actions", async () => {
    const user = userEvent.setup();
    const { onSubmit } = setup();

    await fillRequiredFields(user);
    await user.click(screen.getByRole("button", { name: "Create Rule" }));

    await waitFor(() =>
      expect(onSubmit).toHaveBeenCalledWith({
        rule: expect.objectContaining({
          name: "Coffee rule",
          condition_logic: "AND",
          effective_from: new Date(0).toISOString(),
        }),
        conditions: [
          {
            condition_type: "name",
            condition_operator: "contains",
            condition_value: "coffee",
          },
        ],
        actions: [{ action_type: "category", action_value: "1" }],
      })
    );
  });

  it("adds and removes conditions and switches the match to Any", async () => {
    const user = userEvent.setup();
    setup();

    await user.click(screen.getByRole("button", { name: /add condition/i }));
    expect(screen.getAllByPlaceholderText("Enter a value")).toHaveLength(2);

    const match = screen.getByRole("button", { name: "All" });
    expect(match).toBeEnabled();
    await user.click(match);
    await user.click(await screen.findByRole("menuitem", { name: "Any" }));
    expect(screen.getByRole("button", { name: "Any" })).toBeInTheDocument();

    const removeButtons = screen.getAllByRole("button", {
      name: "Remove condition",
    });
    await user.click(removeButtons[1]);
    expect(screen.getAllByPlaceholderText("Enter a value")).toHaveLength(1);
  });

  it("requires a date when the rule starts from a date", async () => {
    const user = userEvent.setup();
    setup();

    await fillRequiredFields(user);
    await user.click(
      screen.getByRole("button", {
        name: /all past and future transactions/i,
      })
    );
    await user.click(
      await screen.findByRole("menuitem", {
        name: "Starting from (choose date)",
      })
    );
    await user.click(screen.getByRole("button", { name: "Create Rule" }));

    expect(
      screen.getByText("Please select an effective date.")
    ).toBeInTheDocument();
  });

  it("shows a server error passed by the caller", () => {
    setup({ error: "Rule overlaps an existing rule" });

    expect(
      screen.getByText("Rule overlaps an existing rule")
    ).toBeInTheDocument();
  });

  it("shows a skeleton while fetching", () => {
    setup({ fetching: true });

    expect(
      document.querySelectorAll('[data-slot="skeleton"]').length
    ).toBeGreaterThan(0);
    expect(screen.queryByText("IF")).not.toBeInTheDocument();
  });
});
