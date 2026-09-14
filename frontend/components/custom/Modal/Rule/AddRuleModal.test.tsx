import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { toast } from "sonner";
import { describe, expect, it, vi } from "vitest";

import { AddRuleModal } from "./AddRuleModal";

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

function setup() {
  const onOpenChange = vi.fn();
  renderWithProviders(<AddRuleModal isOpen onOpenChange={onOpenChange} />);
  return { onOpenChange };
}

async function fillRequiredFields(user: ReturnType<typeof userEvent.setup>) {
  await user.type(
    screen.getByPlaceholderText("Enter a name for this rule"),
    "Coffee rule"
  );
  await user.type(screen.getByPlaceholderText("Enter a value"), "coffee");
  const selects = screen.getAllByRole("combobox");
  const categorySelect = selects.find((select) =>
    within(select).queryByRole("option", { name: "Select category" })
  )!;
  await user.selectOptions(categorySelect, "1");
}

describe("AddRuleModal", () => {
  it("creates a rule", async () => {
    const user = userEvent.setup();
    let body: unknown;
    server.use(
      http.post("*/api/v1/rule", async ({ request }) => {
        body = await request.json();
        return HttpResponse.json(
          { data: { id: 9, name: "Coffee rule" } },
          { status: 201 }
        );
      })
    );
    const { onOpenChange } = setup();

    await fillRequiredFields(user);
    await user.click(screen.getByRole("button", { name: "Create Rule" }));

    await waitFor(() =>
      expect(toast.success).toHaveBeenCalledWith("Rule created successfully!")
    );
    expect(body).toMatchObject({
      rule: { name: "Coffee rule", condition_logic: "AND" },
      actions: [{ action_type: "category", action_value: "1" }],
      conditions: [
        {
          condition_type: "name",
          condition_operator: "contains",
          condition_value: "coffee",
        },
      ],
    });
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("shows the failure reason and stays open", async () => {
    const user = userEvent.setup();
    server.use(
      http.post("*/api/v1/rule", () =>
        HttpResponse.json(
          { error: "Rule name already exists" },
          { status: 409 }
        )
      )
    );
    const { onOpenChange } = setup();

    await fillRequiredFields(user);
    await user.click(screen.getByRole("button", { name: "Create Rule" }));

    expect(
      await screen.findByText("Rule name already exists")
    ).toBeInTheDocument();
    expect(onOpenChange).not.toHaveBeenCalledWith(false);
  });
});
