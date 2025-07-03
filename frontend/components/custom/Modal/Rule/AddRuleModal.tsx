import { RuleModal } from "@/components/custom/Modal/Rule/RuleModal";
import { useCategories } from "@/components/custom/Provider/CategoryProvider";
import { createRule } from "@/lib/api/rule";
import {
  BaseRuleAction,
  BaseRuleCondition,
  CreateRuleInput,
} from "@/lib/models/rule";
import { useState } from "react";
import { toast } from "sonner";

interface AddRuleModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AddRuleModal({ isOpen, onOpenChange }: AddRuleModalProps) {
  const { read: readCategories } = useCategories();
  const categories = readCategories();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Handler for RuleModal submit
  const handleSubmit = async ({
    ruleName,
    ruleDescription,
    conditions,
    actions,
    effectiveScope,
    effectiveFromDate,
  }: {
    ruleName: string;
    ruleDescription: string;
    conditions: BaseRuleCondition[];
    actions: BaseRuleAction[];
    effectiveScope: "all" | "from";
    effectiveFromDate?: Date;
  }) => {
    setLoading(true);
    setError(null);

    try {
      const payload: CreateRuleInput = {
        rule: {
          name: ruleName,
          description: ruleDescription || undefined,
          effective_from:
            effectiveScope === "from" && effectiveFromDate
              ? effectiveFromDate.toISOString()
              : new Date().toISOString(),
        },
        actions: actions.map((a) => ({
          action_type: a.action_type,
          action_value:
            a.action_type === "category"
              ? (() => {
                  const cat = categories.find(
                    (cat) =>
                      String(cat.id) === a.action_value ||
                      cat.name === a.action_value
                  );
                  return cat ? String(cat.id) : a.action_value;
                })()
              : a.action_value,
        })),
        conditions: conditions.map((c) => ({
          condition_type: c.condition_type,
          condition_operator: c.condition_operator,
          condition_value:
            c.condition_type === "category"
              ? (() => {
                  const cat = categories.find(
                    (cat) =>
                      String(cat.id) === c.condition_value ||
                      cat.name === c.condition_value
                  );
                  return cat ? String(cat.id) : c.condition_value;
                })()
              : c.condition_value,
        })),
      };

      await createRule(payload);
      toast.success("Rule created successfully!");
      onOpenChange(false);
    } catch (e) {
      const err = e as Error;
      setError(err.message || "Failed to create rule");
    } finally {
      setLoading(false);
    }
  };

  return (
    <RuleModal
      isOpen={isOpen}
      onOpenChange={onOpenChange}
      mode="add"
      onSubmit={handleSubmit}
      loading={loading}
      error={error}
      dialogTitle="New transaction rule"
      submitButtonText="Create Rule"
    />
  );
}
