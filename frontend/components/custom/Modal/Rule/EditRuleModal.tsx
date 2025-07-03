import {
  RuleModal,
  RuleModalInitialData,
} from "@/components/custom/Modal/Rule/RuleModal";
import { getRule, updateRule } from "@/lib/api/rule";
import {
  BaseRuleAction,
  BaseRuleCondition,
  DescribeRuleResponse,
  UpdateRuleInput,
} from "@/lib/models/rule";
import { useEffect, useState } from "react";
import { toast } from "sonner";

interface EditRuleModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  ruleId: number;
}

export function EditRuleModal({
  isOpen,
  onOpenChange,
  ruleId,
}: EditRuleModalProps) {
  const [initialData, setInitialData] = useState<
    RuleModalInitialData | undefined
  >(undefined);
  const [fetching, setFetching] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Fetch rule details on open
  useEffect(() => {
    if (!isOpen) return;
    setFetching(true);
    setError(null);
    getRule(ruleId)
      .then((data: DescribeRuleResponse) => {
        // Prepare initial data for RuleModal
        let effectiveScope: "all" | "from" = "all";
        let effectiveFromDate: Date | undefined = undefined;
        if (data.rule.effective_from) {
          const effDate = new Date(data.rule.effective_from);
          const isToday = effDate.toDateString() === new Date().toDateString();
          if (isToday) {
            effectiveScope = "all";
            effectiveFromDate = undefined;
          } else {
            effectiveScope = "from";
            effectiveFromDate = effDate;
          }
        }
        setInitialData({
          ruleName: data.rule.name || "",
          ruleDescription: data.rule.description || "",
          conditions: data.conditions.map((c) => ({
            condition_type: c.condition_type,
            condition_operator: c.condition_operator,
            condition_value:
              c.condition_type === "category"
                ? String(c.condition_value)
                : c.condition_value,
          })),
          actions: data.actions.map((a) => ({
            action_type: a.action_type,
            action_value:
              a.action_type === "category"
                ? String(a.action_value)
                : a.action_value,
          })),
          effectiveScope,
          effectiveFromDate,
        });
      })
      .catch((e) => {
        setError(e?.message || "Failed to fetch rule details");
      })
      .finally(() => setFetching(false));
  }, [isOpen, ruleId]);

  // Handler for RuleModal submit
  const handleSubmit = async ({
    ruleName,
    ruleDescription,
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
      // Only send changed fields for the rule itself
      const ruleUpdate: UpdateRuleInput = {
        name: ruleName,
        description: ruleDescription || undefined,
        effective_from:
          effectiveScope === "from" && effectiveFromDate
            ? effectiveFromDate.toISOString()
            : new Date().toISOString(),
      };

      await updateRule(ruleId, ruleUpdate);
      // Note: You may want to update actions/conditions as well if your API supports it

      toast.success("Rule updated successfully!");
      onOpenChange(false);
    } catch (e) {
      const err = e as Error;
      setError(err.message || "Failed to update rule");
    } finally {
      setLoading(false);
    }
  };

  return (
    <RuleModal
      isOpen={isOpen}
      onOpenChange={onOpenChange}
      mode="edit"
      initialData={initialData}
      onSubmit={handleSubmit}
      loading={loading}
      fetching={fetching}
      error={error}
      dialogTitle="Edit transaction rule"
      submitButtonText="Save Changes"
    />
  );
}
