import { BaseRuleAction, BaseRuleCondition } from "@/lib/models/rule";

type Category = { id: number; name: string };

function getCategoryId(value: string, categories: Category[]): string {
  const cat = categories.find(
    (cat) => String(cat.id) === value || cat.name === value
  );
  return cat ? String(cat.id) : value;
}

export function normalizeRuleActions(
  actions: BaseRuleAction[],
  categories: Category[]
): BaseRuleAction[] {
  return actions.map((a) => ({
    action_type: a.action_type,
    action_value:
      a.action_type === "category"
        ? getCategoryId(a.action_value, categories)
        : a.action_value,
  }));
}

// Helper to determine effectiveScope and effectiveFromDate from effective_from string
export function getEffectiveScopeAndDate(effective_from?: string): {
  effectiveScope: "all" | "from";
  effectiveFromDate: Date | undefined;
} {
  if (effective_from) {
    const effDate = new Date(effective_from);
    // Check for invalid date
    if (isNaN(effDate.getTime())) {
      return { effectiveScope: "all", effectiveFromDate: undefined };
    }

    // A rule with effective_from set to the earliest possible date is considered to be for "all" time.
    const isEarliest = effDate.getTime() === new Date(0).getTime();

    if (isEarliest) {
      return { effectiveScope: "all", effectiveFromDate: undefined };
    } else {
      return { effectiveScope: "from", effectiveFromDate: effDate };
    }
  }
  return { effectiveScope: "all", effectiveFromDate: undefined };
}

export function normalizeRuleConditions(
  conditions: BaseRuleCondition[],
  categories: Category[]
): BaseRuleCondition[] {
  return conditions.map((c) => ({
    condition_type: c.condition_type,
    condition_operator: c.condition_operator,
    condition_value:
      c.condition_type === "category"
        ? getCategoryId(c.condition_value, categories)
        : c.condition_value,
  }));
}
