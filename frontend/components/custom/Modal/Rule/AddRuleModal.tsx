import { RuleModal } from "@/components/custom/Modal/Rule/RuleModal";
import { useCategories } from "@/components/hooks/useCategories";
import { useCreateRule } from "@/components/hooks/useRules";
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
  const { data: categories = [] } = useCategories();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const createRuleMutation = useCreateRule();

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
    const payload: CreateRuleInput = {
      rule,
      actions: normalizeRuleActions(actions, categories),
      conditions: normalizeRuleConditions(conditions, categories),
    };
    createRuleMutation.mutate(payload, {
      onSuccess: () => {
        toast.success("Rule created successfully!");
        onOpenChange(false);
      },
      onError: (e: unknown) => {
        const err = e as Error;
        setError(err.message || "Failed to create rule");
      },
      onSettled: () => {
        setLoading(false);
      },
    });
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
