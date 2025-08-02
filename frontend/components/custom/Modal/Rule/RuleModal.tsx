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
import {
  BaseRule,
  BaseRuleAction,
  BaseRuleCondition,
  ConditionLogic,
  Rule,
} from "@/lib/models/rule";
import { getEffectiveScopeAndDate } from "@/lib/utils/rule";
import { useEffect, useState } from "react";

type RuleModalMode = "add" | "edit";

export interface RuleModalInitialData {
  rule?: Rule;
  conditions?: BaseRuleCondition[];
  actions?: BaseRuleAction[];
}

interface RuleModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  mode: RuleModalMode;
  initialData?: RuleModalInitialData;
  onSubmit: (data: {
    rule: BaseRule;
    conditions: BaseRuleCondition[];
    actions: BaseRuleAction[];
  }) => Promise<void>;
  loading?: boolean;
  fetching?: boolean;
  error?: string | null;
  dialogTitle?: string;
  submitButtonText?: string;
}

export function RuleModal({
  isOpen,
  onOpenChange,
  mode,
  initialData,
  onSubmit,
  loading = false,
  fetching = false,
  error = null,
  dialogTitle,
  submitButtonText,
}: RuleModalProps) {
  // Form state
  const [rule, setRule] = useState<BaseRule>(
    initialData?.rule || {
      name: "",
      description: "",
      condition_logic: ConditionLogic.AND,
      effective_from: new Date(0).toISOString(),
    }
  );
  const [conditions, setConditions] = useState<BaseRuleCondition[]>(
    initialData?.conditions && initialData.conditions.length > 0
      ? initialData.conditions
      : [
          {
            condition_type: "name",
            condition_operator: "contains",
            condition_value: "",
          },
        ]
  );
  const [actions, setActions] = useState<BaseRuleAction[]>(
    initialData?.actions && initialData.actions.length > 0
      ? initialData.actions
      : [{ action_type: "category", action_value: "" }]
  );
  const [conditionLogic, setConditionLogic] = useState<ConditionLogic>(
    initialData?.rule?.condition_logic || ConditionLogic.AND
  );
  const [effectiveScope, setEffectiveScope] = useState<"all" | "from">("all");
  const [effectiveFromDate, setEffectiveFromDate] = useState<Date | undefined>(
    undefined
  );
  const [localError, setLocalError] = useState<string | null>(null);

  // Update form state with initialData when modal is opened or initialData changes (for edit mode)
  useEffect(() => {
    if (isOpen && initialData) {
      setRule(
        initialData.rule || {
          name: "",
          description: "",
          condition_logic: ConditionLogic.AND,
          effective_from: new Date(0).toISOString(),
        }
      );
      setConditions(
        initialData.conditions && initialData.conditions.length > 0
          ? initialData.conditions
          : [
              {
                condition_type: "name",
                condition_operator: "contains",
                condition_value: "",
              },
            ]
      );
      setActions(
        initialData.actions && initialData.actions.length > 0
          ? initialData.actions
          : [{ action_type: "category", action_value: "" }]
      );
      setConditionLogic(
        initialData.rule?.condition_logic || ConditionLogic.AND
      );
      const { effectiveScope, effectiveFromDate } = getEffectiveScopeAndDate(
        initialData.rule?.effective_from
      );
      setEffectiveScope(effectiveScope);
      setEffectiveFromDate(effectiveFromDate);
      setLocalError(null);
    }
  }, [isOpen, initialData]);

  // Validation and submit handler
  const handleSubmit = async () => {
    setLocalError(null);

    // Validation
    if (!rule.name.trim()) {
      setLocalError("Rule name is required.");
      return;
    }
    if (conditions.some((c) => !String(c.condition_value).trim())) {
      setLocalError("All condition values must be filled.");
      return;
    }
    if (actions.some((a) => !String(a.action_value).trim())) {
      setLocalError("All action values must be filled.");
      return;
    }
    if (effectiveScope === "from" && !effectiveFromDate) {
      setLocalError("Please select an effective date.");
      return;
    }

    let effective_from: string;
    if (effectiveScope === "from" && effectiveFromDate) {
      effective_from = effectiveFromDate.toISOString();
    } else {
      effective_from = new Date(0).toISOString();
    }

    await onSubmit({
      rule: {
        ...rule,
        effective_from,
        condition_logic: conditionLogic,
      },
      conditions,
      actions,
    });
  };

  // Reset form state
  const resetForm = () => {
    const defaultRule = {
      name: "",
      description: "",
      condition_logic: ConditionLogic.AND,
      effective_from: new Date(0).toISOString(),
    };
    setRule(initialData?.rule || defaultRule);
    setConditions(
      initialData?.conditions && initialData.conditions.length > 0
        ? initialData.conditions
        : [
            {
              condition_type: "name",
              condition_operator: "contains",
              condition_value: "",
            },
          ]
    );
    setActions(
      initialData?.actions && initialData.actions.length > 0
        ? initialData.actions
        : [{ action_type: "category", action_value: "" }]
    );
    setConditionLogic(initialData?.rule?.condition_logic || ConditionLogic.AND);
    const { effectiveScope, effectiveFromDate } = getEffectiveScopeAndDate(
      initialData?.rule?.effective_from
    );
    setEffectiveScope(effectiveScope);
    setEffectiveFromDate(effectiveFromDate);
    setLocalError(null);
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
            {dialogTitle ||
              (mode === "add"
                ? "New transaction rule"
                : "Edit transaction rule")}
          </DialogTitle>
        </DialogHeader>

        {fetching ? (
          <EditRuleModalSkeleton />
        ) : (
          <>
            <RuleBasicInfo
              rule={rule}
              onRuleChange={setRule}
              disabled={loading}
            />

            {mode === "edit" && (
              <div className="text-xs text-muted-foreground mb-2">
                <em>
                  Note: Deleting an action or condition is not supported. If you
                  want to, please delete the rule and recreate it.
                </em>
              </div>
            )}
            <RuleConditions
              conditions={conditions}
              onConditionsChange={setConditions}
              conditionLogic={conditionLogic}
              onConditionLogicChange={setConditionLogic}
              disabled={mode === "edit"}
              loading={loading}
            />

            <RuleActions
              actions={actions}
              onActionsChange={setActions}
              disabled={loading || mode === "edit"}
            />

            <RuleEffectiveScope
              effectiveScope={effectiveScope}
              effectiveFromDate={effectiveFromDate}
              onEffectiveScopeChange={setEffectiveScope}
              onEffectiveFromDateChange={setEffectiveFromDate}
              disabled={loading}
            />

            {(localError || error) && (
              <div className="text-destructive text-sm bg-destructive/10 border border-destructive/20 rounded-lg p-3">
                {localError || error}
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
                {submitButtonText ||
                  (mode === "add" ? "Create Rule" : "Save Changes")}
              </LoadingButton>
            </div>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}
