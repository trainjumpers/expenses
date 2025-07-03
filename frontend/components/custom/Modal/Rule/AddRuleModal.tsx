import { RuleModal } from "@/components/custom/Modal/Rule/RuleModal";
import { useCategories } from "@/components/custom/Provider/CategoryProvider";
import { createRule } from "@/lib/api/rule";
import {
  BaseRule,
  BaseRuleAction,
  BaseRuleCondition,
  CreateRuleInput,
} from "@/lib/models/rule";
import {
  normalizeRuleActions,
  normalizeRuleConditions,
} from "@/lib/utils/rule";
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
    rule,
    conditions,
    actions,
  }: {
    rule: BaseRule;
    conditions: BaseRuleCondition[];
    actions: BaseRuleAction[];
  }) => {
    setLoading(true);
    setError(null);

    try {
      const payload: CreateRuleInput = {
        rule,
        actions: normalizeRuleActions(actions, categories),
        conditions: normalizeRuleConditions(conditions, categories),
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
