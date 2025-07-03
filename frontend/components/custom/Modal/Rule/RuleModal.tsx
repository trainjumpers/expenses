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
import { BaseRuleAction, BaseRuleCondition } from "@/lib/models/rule";
import { useEffect, useState } from "react";

type RuleModalMode = "add" | "edit";

export interface RuleModalInitialData {
  ruleName?: string;
  ruleDescription?: string;
  conditions?: BaseRuleCondition[];
  actions?: BaseRuleAction[];
  effectiveScope?: "all" | "from";
  effectiveFromDate?: Date;
}

interface RuleModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  mode: RuleModalMode;
  initialData?: RuleModalInitialData;
  onSubmit: (data: {
    ruleName: string;
    ruleDescription: string;
    conditions: BaseRuleCondition[];
    actions: BaseRuleAction[];
    effectiveScope: "all" | "from";
    effectiveFromDate?: Date;
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
  const [ruleName, setRuleName] = useState(initialData?.ruleName || "");
  const [ruleDescription, setRuleDescription] = useState(
    initialData?.ruleDescription || ""
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
  const [effectiveScope, setEffectiveScope] = useState<"all" | "from">(
    initialData?.effectiveScope || "all"
  );
  const [effectiveFromDate, setEffectiveFromDate] = useState<Date | undefined>(
    initialData?.effectiveFromDate
  );
  const [localError, setLocalError] = useState<string | null>(null);

  // Update form state with initialData when modal is opened or initialData changes (for edit mode)
  useEffect(() => {
    if (isOpen && initialData) {
      setRuleName(initialData.ruleName || "");
      setRuleDescription(initialData.ruleDescription || "");
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
      setEffectiveScope(initialData.effectiveScope || "all");
      setEffectiveFromDate(initialData.effectiveFromDate);
      setLocalError(null);
    }
  }, [isOpen, initialData]);

  // Validation and submit handler
  const handleSubmit = async () => {
    setLocalError(null);

    // Validation
    if (!ruleName.trim()) {
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

    // Pass form data to parent handler
    await onSubmit({
      ruleName,
      ruleDescription,
      conditions,
      actions,
      effectiveScope,
      effectiveFromDate,
    });
  };

  // Reset form state
  const resetForm = () => {
    setRuleName(initialData?.ruleName || "");
    setRuleDescription(initialData?.ruleDescription || "");
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
    setEffectiveScope(initialData?.effectiveScope || "all");
    setEffectiveFromDate(initialData?.effectiveFromDate);
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
