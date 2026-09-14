import { EditRuleModal } from "@/components/custom/Modal/Rule/EditRuleModal";
import type { Rule } from "@/lib/models/rule";
import { ConditionLogic } from "@/lib/models/rule";
import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, delay, http } from "msw";
import { toast } from "sonner";
import { describe, expect, it, vi } from "vitest";

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

const rule: Rule = {
  id: 1,
  name: "Coffee rule",
  description: "Matches coffee",
  condition_logic: ConditionLogic.AND,
  effective_from: new Date(0).toISOString(),
  created_by: 1,
};

const describeResponse = {
  rule,
  actions: [{ id: 21, rule_id: 1, action_type: "category", action_value: "1" }],
  conditions: [
    {
      id: 11,
      rule_id: 1,
      condition_type: "name",
      condition_operator: "contains",
      condition_value: "coffee",
    },
  ],
};

function describeHandler() {
  return http.get("*/api/v1/rule/1", () =>
    HttpResponse.json({ data: describeResponse })
  );
}

function setup() {
  const onOpenChange = vi.fn();
  renderWithProviders(
    <EditRuleModal isOpen onOpenChange={onOpenChange} ruleId={1} />
  );
  return { onOpenChange };
}

describe("EditRuleModal", () => {
  it("loads the rule into the form", async () => {
    server.use(describeHandler());
    setup();

    expect(
      await screen.findByRole("dialog", { name: "Edit transaction rule" })
    ).toBeInTheDocument();
    expect(await screen.findByDisplayValue("Coffee rule")).toBeInTheDocument();
    expect(screen.getByDisplayValue("Matches coffee")).toBeInTheDocument();
  });

  it("shows a skeleton while the rule is loading", async () => {
    server.use(
      http.get("*/api/v1/rule/1", async () => {
        await delay(100);
        return HttpResponse.json({ data: describeResponse });
      })
    );
    setup();

    expect(
      document.querySelectorAll('[data-slot="skeleton"]').length
    ).toBeGreaterThan(0);
    expect(await screen.findByDisplayValue("Coffee rule")).toBeInTheDocument();
  });

  it("saves the rule, conditions and actions in one submit", async () => {
    const user = userEvent.setup();
    let ruleBody: unknown;
    let actionsBody: unknown;
    let conditionsBody: unknown;
    server.use(
      describeHandler(),
      http.patch("*/api/v1/rule/1", async ({ request }) => {
        ruleBody = await request.json();
        return HttpResponse.json({ data: rule });
      }),
      http.put("*/api/v1/rule/1/actions", async ({ request }) => {
        actionsBody = await request.json();
        return HttpResponse.json({ data: [] });
      }),
      http.put("*/api/v1/rule/1/conditions", async ({ request }) => {
        conditionsBody = await request.json();
        return HttpResponse.json({ data: [] });
      })
    );
    const { onOpenChange } = setup();
    await screen.findByDisplayValue("Coffee rule");

    await user.click(screen.getByRole("button", { name: "Save Changes" }));

    await waitFor(() =>
      expect(toast.success).toHaveBeenCalledWith("Rule updated successfully!")
    );
    expect(ruleBody).toMatchObject({ name: "Coffee rule" });
    expect(actionsBody).toEqual({
      actions: [{ action_type: "category", action_value: "1" }],
    });
    expect(conditionsBody).toEqual({
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

  it("shows the fetch failure reason", async () => {
    server.use(
      http.get("*/api/v1/rule/1", () =>
        HttpResponse.json({ error: "Rule not found" }, { status: 404 })
      )
    );
    setup();

    expect(await screen.findByText("Rule not found")).toBeInTheDocument();
  });

  it("shows the save failure reason and keeps the dialog open", async () => {
    const user = userEvent.setup();
    server.use(
      describeHandler(),
      http.patch("*/api/v1/rule/1", () =>
        HttpResponse.json(
          { error: "Rule name already exists" },
          { status: 409 }
        )
      ),
      http.put("*/api/v1/rule/1/actions", () =>
        HttpResponse.json({ data: [] })
      ),
      http.put("*/api/v1/rule/1/conditions", () =>
        HttpResponse.json({ data: [] })
      )
    );
    const { onOpenChange } = setup();
    await screen.findByDisplayValue("Coffee rule");

    await user.click(screen.getByRole("button", { name: "Save Changes" }));

    expect(
      await screen.findByText("Rule name already exists")
    ).toBeInTheDocument();
    expect(onOpenChange).not.toHaveBeenCalledWith(false);
  });
});
