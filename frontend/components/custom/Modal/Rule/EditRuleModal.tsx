import {
  RuleActions,
  RuleBasicInfo,
  RuleConditions,
  RuleEffectiveScope,
} from "@/components/custom/Modal/Rule/components";
import { EditRuleModalSkeleton } from "@/components/custom/Skeletons/RuleSkeletons";
import { LoadingButton } from "@/components/ui/LoadingButton";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { getRule, updateRule } from "@/lib/api/rule";
import {
  DescribeRuleResponse,
  RuleFieldType,
  RuleOperator,
  UpdateRuleInput,
} from "@/lib/models/rule";
import { useEffect, useState } from "react";
import { toast } from "sonner";

interface EditRuleModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  ruleId: number;
}

interface RuleCondition {
  condition_type: RuleFieldType;
  condition_operator: RuleOperator;
  condition_value: string;
}

interface RuleAction {
  action_type: RuleFieldType;
  action_value: string;
}

export function EditRuleModal({
  isOpen,
  onOpenChange,
  ruleId,
}: EditRuleModalProps) {
  const [ruleName, setRuleName] = useState("");
  const [ruleDescription, setRuleDescription] = useState("");
  const [conditions, setConditions] = useState<RuleCondition[]>([]);
  const [actions, setActions] = useState<RuleAction[]>([]);
  const [effectiveScope, setEffectiveScope] = useState<"all" | "from">("all");
  const [effectiveFromDate, setEffectiveFromDate] = useState<Date | undefined>(
    undefined
  );
  const [loading, setLoading] = useState(false);
  const [fetching, setFetching] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Fetch rule details on open
  useEffect(() => {
    if (!isOpen) return;
    setFetching(true);
    setError(null);
    getRule(ruleId)
      .then((data: DescribeRuleResponse) => {
        setRuleName(data.rule.name || "");
        setRuleDescription(data.rule.description || "");
        setConditions(
          data.conditions.map((c) => ({
            condition_type: c.condition_type,
            condition_operator: c.condition_operator,
            condition_value:
              c.condition_type === "category"
                ? String(c.condition_value)
                : c.condition_value,
          }))
        );
        setActions(
          data.actions.map((a) => ({
            action_type: a.action_type,
            action_value:
              a.action_type === "category"
                ? String(a.action_value)
                : a.action_value,
          }))
        );
        // Effective date logic
        if (data.rule.effective_from) {
          const effDate = new Date(data.rule.effective_from);
          const isToday = effDate.toDateString() === new Date().toDateString();
          if (isToday) {
            setEffectiveScope("all");
            setEffectiveFromDate(undefined);
          } else {
            setEffectiveScope("from");
            setEffectiveFromDate(effDate);
          }
        } else {
          setEffectiveScope("all");
          setEffectiveFromDate(undefined);
        }
      })
      .catch((e) => {
        setError(e?.message || "Failed to fetch rule details");
      })
      .finally(() => setFetching(false));
  }, [isOpen, ruleId]);

  const resetForm = () => {
    setRuleName("");
    setRuleDescription("");
    setConditions([]);
    setActions([]);
    setEffectiveScope("all");
    setEffectiveFromDate(undefined);
    setError(null);
  };

  const handleSubmit = async () => {
    setLoading(true);
    setError(null);

    // Validation
    if (!ruleName.trim()) {
      setError("Rule name is required.");
      setLoading(false);
      return;
    }
    if (conditions.some((c) => !c.condition_value.trim())) {
      setError("All condition values must be filled.");
      setLoading(false);
      return;
    }
    if (actions.some((a) => !a.action_value.trim())) {
      setError("All action values must be filled.");
      setLoading(false);
      return;
    }
    if (effectiveScope === "from" && !effectiveFromDate) {
      setError("Please select an effective date.");
      setLoading(false);
      return;
    }

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
    <Dialog
      open={isOpen}
      onOpenChange={(open) => {
        if (!open) resetForm();
        onOpenChange(open);
      }}
    >
      <DialogContent className="sm:max-w-[640px] max-h-[90vh] overflow-y-auto">
        <DialogHeader className="flex flex-row items-center justify-between space-y-0 pb-6">
          <DialogTitle className="text-xl font-semibold">
            Edit transaction rule
          </DialogTitle>
        </DialogHeader>

        {fetching ? (
          <EditRuleModalSkeleton />
        ) : (
          <>
            <RuleBasicInfo
              ruleName={ruleName}
              ruleDescription={ruleDescription}
              onRuleNameChange={setRuleName}
              onRuleDescriptionChange={setRuleDescription}
              disabled={loading}
            />

            <RuleConditions
              conditions={conditions}
              onConditionsChange={setConditions}
              disabled={loading}
            />

            <RuleActions
              actions={actions}
              onActionsChange={setActions}
              disabled={loading}
            />

            <RuleEffectiveScope
              effectiveScope={effectiveScope}
              effectiveFromDate={effectiveFromDate}
              onEffectiveScopeChange={setEffectiveScope}
              onEffectiveFromDateChange={setEffectiveFromDate}
              disabled={loading}
            />

            {error && (
              <div className="text-destructive text-sm bg-destructive/10 border border-destructive/20 rounded-lg p-3">
                {error}
              </div>
            )}

            <div className="pt-6 mt-8 border-t flex gap-3 justify-end">
              <Button
                variant="outline"
                onClick={() => {
                  resetForm();
                  onOpenChange(false);
                }}
                disabled={loading}
                type="button"
              >
                Cancel
              </Button>
              <LoadingButton
                onClick={handleSubmit}
                loading={loading}
                type="button"
              >
                Save Changes
              </LoadingButton>
            </div>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}
