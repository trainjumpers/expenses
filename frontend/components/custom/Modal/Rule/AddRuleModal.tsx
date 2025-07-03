import {
  RuleActions,
  RuleBasicInfo,
  RuleConditions,
  RuleEffectiveScope,
} from "@/components/custom/Modal/Rule/components";
import { useCategories } from "@/components/custom/Provider/CategoryProvider";
import { LoadingButton } from "@/components/ui/LoadingButton";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { createRule } from "@/lib/api/rule";
import {
  CreateRuleInput,
  RuleFieldType,
  RuleOperator,
} from "@/lib/models/rule";
import { useState } from "react";
import { toast } from "sonner";

interface AddRuleModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
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

export function AddRuleModal({ isOpen, onOpenChange }: AddRuleModalProps) {
  const [ruleName, setRuleName] = useState("");
  const [ruleDescription, setRuleDescription] = useState("");
  const [conditions, setConditions] = useState<RuleCondition[]>([
    {
      condition_type: "name",
      condition_operator: "contains",
      condition_value: "",
    },
  ]);
  const [actions, setActions] = useState<RuleAction[]>([
    { action_type: "category", action_value: "" },
  ]);
  const { read: readCategories } = useCategories();
  const categories = readCategories();
  const [effectiveScope, setEffectiveScope] = useState<"all" | "from">("all");
  const [effectiveFromDate, setEffectiveFromDate] = useState<Date | undefined>(
    undefined
  );
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const resetForm = () => {
    setRuleName("");
    setRuleDescription("");
    setConditions([
      {
        condition_type: "name",
        condition_operator: "contains",
        condition_value: "",
      },
    ]);
    setActions([{ action_type: "category", action_value: "" }]);
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
      const payload: CreateRuleInput = {
        rule: {
          name: ruleName,
          description: ruleDescription || undefined,
          effective_from:
            effectiveScope === "from" && effectiveFromDate
              ? effectiveFromDate.toISOString()
              : new Date().toISOString(),
          created_by: 1, // TODO: Replace with actual user ID
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
      resetForm();
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
            New transaction rule
          </DialogTitle>
        </DialogHeader>

        {/* Rule Name and Description */}
        <RuleBasicInfo
          ruleName={ruleName}
          ruleDescription={ruleDescription}
          onRuleNameChange={setRuleName}
          onRuleDescriptionChange={setRuleDescription}
          disabled={loading}
        />

        {/* IF Section */}
        <RuleConditions
          conditions={conditions}
          onConditionsChange={setConditions}
          disabled={loading}
        />

        {/* THEN Section */}
        <RuleActions
          actions={actions}
          onActionsChange={setActions}
          disabled={loading}
        />

        {/* FOR Section */}
        <RuleEffectiveScope
          effectiveScope={effectiveScope}
          effectiveFromDate={effectiveFromDate}
          onEffectiveScopeChange={setEffectiveScope}
          onEffectiveFromDateChange={setEffectiveFromDate}
          disabled={loading}
        />

        {/* Error Display */}
        {error && (
          <div className="text-destructive text-sm bg-destructive/10 border border-destructive/20 rounded-lg p-3">
            {error}
          </div>
        )}

        {/* Footer */}
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
          <LoadingButton onClick={handleSubmit} loading={loading} type="button">
            Create Rule
          </LoadingButton>
        </div>
      </DialogContent>
    </Dialog>
  );
}
