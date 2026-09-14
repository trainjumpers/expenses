import { ViewRulesModal } from "@/components/custom/Modal/Rule/ViewRulesModal";
import type { Rule } from "@/lib/models/rule";
import { ConditionLogic } from "@/lib/models/rule";
import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
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

const teaRule: Rule = {
  ...rule,
  id: 2,
  name: "Tea rule",
  description: "Matches tea",
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

function render() {
  return renderWithProviders(<ViewRulesModal isOpen onOpenChange={vi.fn()} />);
}

function listHandler(rules: Rule[]) {
  return http.get("*/api/v1/rule", () =>
    HttpResponse.json({
      data: { rules, total: rules.length, page: 1, page_size: 5 },
    })
  );
}

describe("ViewRulesModal", () => {
  it("lists rules with their descriptions", async () => {
    server.use(listHandler([rule]));
    render();

    expect(await screen.findByText("Coffee rule")).toBeInTheDocument();
    expect(screen.getByText("Matches coffee")).toBeInTheDocument();
  });

  it("shows an empty state without rules", async () => {
    server.use(listHandler([]));
    render();

    expect(
      await screen.findByText("No rules to display yet.")
    ).toBeInTheDocument();
  });

  it("searches rules after the debounce", async () => {
    const user = userEvent.setup();
    let lastSearch: string | null = null;
    server.use(
      http.get("*/api/v1/rule", ({ request }) => {
        lastSearch = new URL(request.url).searchParams.get("search");
        const rules = lastSearch ? [rule] : [rule, teaRule];
        return HttpResponse.json({
          data: { rules, total: rules.length, page: 1, page_size: 5 },
        });
      })
    );
    render();
    await screen.findByText("Tea rule");

    await user.type(
      screen.getByPlaceholderText("Search rules by name or description..."),
      "coffee"
    );

    await waitFor(() => expect(lastSearch).toBe("coffee"));
    await waitFor(() =>
      expect(screen.queryByText("Tea rule")).not.toBeInTheDocument()
    );
    expect(screen.getByText("Coffee rule")).toBeInTheDocument();
  });

  it("pages through long rule lists", async () => {
    const user = userEvent.setup();
    const allRules = Array.from({ length: 6 }, (_, i) => ({
      ...rule,
      id: i + 1,
      name: `Rule ${i + 1}`,
    }));
    server.use(
      http.get("*/api/v1/rule", ({ request }) => {
        const page = Number(
          new URL(request.url).searchParams.get("page") ?? "1"
        );
        return HttpResponse.json({
          data: {
            rules: allRules.slice((page - 1) * 5, page * 5),
            total: 6,
            page,
            page_size: 5,
          },
        });
      })
    );
    render();
    await screen.findByText("Rule 1");
    expect(screen.queryByText("Rule 6")).not.toBeInTheDocument();

    await user.click(screen.getByText("2"));

    expect(await screen.findByText("Rule 6")).toBeInTheDocument();
  });

  it("deletes a rule after confirming", async () => {
    const user = userEvent.setup();
    let deleted = false;
    server.use(
      listHandler([rule]),
      http.delete("*/api/v1/rule/1", () => {
        deleted = true;
        return new HttpResponse(null, { status: 204 });
      })
    );
    render();
    await screen.findByText("Coffee rule");

    await user.click(screen.getByRole("button", { name: "Delete" }));
    const confirm = await screen.findByRole("dialog", { name: "Delete Rule" });
    await user.click(within(confirm).getByRole("button", { name: "Delete" }));

    await waitFor(() => expect(deleted).toBe(true));
  });

  it("opens the edit dialog with the rule loaded", async () => {
    const user = userEvent.setup();
    server.use(
      listHandler([rule]),
      http.get("*/api/v1/rule/1", () =>
        HttpResponse.json({ data: describeResponse })
      )
    );
    render();
    await screen.findByText("Coffee rule");

    await user.click(screen.getByRole("button", { name: "Edit" }));

    expect(
      await screen.findByRole("dialog", { name: "Edit transaction rule" })
    ).toBeInTheDocument();
    expect(await screen.findByDisplayValue("Coffee rule")).toBeInTheDocument();
  });
});
