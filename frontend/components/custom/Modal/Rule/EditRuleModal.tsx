import {
  RuleModal,
  RuleModalInitialData,
} from "@/components/custom/Modal/Rule/RuleModal";
import {
  getRule,
  updateRule,
  updateRuleActions,
  updateRuleConditions,
} from "@/lib/api/rule";
import {
  BaseRuleAction,
  BaseRuleCondition,
  DescribeRuleResponse,
  UpdateRuleInput,
} from "@/lib/models/rule";
import {
  normalizeRuleActions,
  normalizeRuleConditions,
} from "@/lib/utils/rule";
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
        setInitialData({
          rule: {
            name: data.rule.name || "",
            description: data.rule.description || "",
            effective_from:
              data.rule.effective_from || new Date().toISOString().slice(0, 10),
          },
          conditions: normalizeRuleConditions(
            data.conditions as BaseRuleCondition[],
            []
          ),
          actions: normalizeRuleActions(data.actions as BaseRuleAction[], []),
        });
      })
      .catch((e) => {
        setError(e?.message || "Failed to fetch rule details");
      })
      .finally(() => setFetching(false));
  }, [isOpen, ruleId]);

  // Handler for RuleModal submit
  const handleSubmit = async ({
    rule,
    actions,
    conditions,
  }: {
    rule: UpdateRuleInput;
    actions: BaseRuleAction[];
    conditions: BaseRuleCondition[];
  }) => {
    setLoading(true);
    setError(null);

    try {
      const ruleUpdate: UpdateRuleInput = {
        ...rule,
      };
      await updateRule(ruleId, ruleUpdate);
      await updateRuleActions(ruleId, actions);
      await updateRuleConditions(ruleId, conditions);

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
