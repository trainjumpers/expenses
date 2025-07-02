import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { createRule } from "@/lib/api/rule";
import {
  CreateRuleInput,
  RULE_FIELD_TYPES,
  RULE_OPERATORS,
} from "@/lib/models/rule";
import { cn } from "@/lib/utils";
import { format } from "date-fns";
import { CalendarIcon, ChevronDownIcon, Plus, Trash2, X } from "lucide-react";
import { useState } from "react";

interface AddRuleModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AddRuleModal({ isOpen, onOpenChange }: AddRuleModalProps) {
  const [ruleName, setRuleName] = useState("");
  const [ruleDescription, setRuleDescription] = useState("");
  const [conditions, setConditions] = useState<
    {
      condition_type: (typeof RULE_FIELD_TYPES)[number]["value"];
      condition_operator: (typeof RULE_OPERATORS)[number]["value"];
      condition_value: string;
    }[]
  >([
    {
      condition_type: "name",
      condition_operator: "contains",
      condition_value: "",
    },
  ]);
  const [actions, setActions] = useState<
    {
      action_type: (typeof RULE_FIELD_TYPES)[number]["value"];
      action_value: string;
    }[]
  >([{ action_type: "category", action_value: "" }]);
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

  const handleAddCondition = () => {
    setConditions([
      ...conditions,
      {
        condition_type: "name",
        condition_operator: "contains",
        condition_value: "",
      },
    ]);
  };

  const handleRemoveCondition = (idx: number) => {
    if (conditions.length > 1) {
      setConditions(conditions.filter((_, i) => i !== idx));
    }
  };

  const handleConditionChange = (idx: number, field: string, value: string) => {
    setConditions(
      conditions.map((c, i) => (i === idx ? { ...c, [field]: value } : c))
    );
  };

  const handleAddAction = () => {
    setActions([...actions, { action_type: "category", action_value: "" }]);
  };

  const handleRemoveAction = (idx: number) => {
    if (actions.length > 1) {
      setActions(actions.filter((_, i) => i !== idx));
    }
  };

  const handleActionChange = (idx: number, field: string, value: string) => {
    setActions(
      actions.map((a, i) => (i === idx ? { ...a, [field]: value } : a))
    );
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
          action_type:
            a.action_type as (typeof RULE_FIELD_TYPES)[number]["value"],
          action_value: a.action_value,
        })),
        conditions: conditions.map((c) => ({
          condition_type:
            c.condition_type as (typeof RULE_FIELD_TYPES)[number]["value"],
          condition_operator:
            c.condition_operator as (typeof RULE_OPERATORS)[number]["value"],
          condition_value: c.condition_value,
        })),
      };

      await createRule(payload);

      resetForm();
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
          <Button
            variant="ghost"
            size="icon"
            onClick={() => onOpenChange(false)}
            className="h-6 w-6"
          >
            <X className="h-4 w-4" />
          </Button>
        </DialogHeader>

        {/* Rule Name and Description */}
        <div className="flex flex-col gap-3">
          <div className="flex items-center gap-2">
            <label className="text-sm font-medium text-muted-foreground min-w-[90px]">
              Rule name
            </label>
            <input
              type="text"
              className="flex-1 px-4 py-2 bg-background border border-border rounded-lg text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
              placeholder="Enter a name for this rule"
              value={ruleName}
              onChange={(e) => setRuleName(e.target.value)}
              disabled={loading}
              required
            />
          </div>
          <div className="flex items-center gap-2">
            <label className="text-sm font-medium text-muted-foreground min-w-[90px]">
              Description
            </label>
            <input
              type="text"
              className="flex-1 px-4 py-2 bg-background border border-border rounded-lg text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
              placeholder="Enter a description for this rule"
              value={ruleDescription}
              onChange={(e) => setRuleDescription(e.target.value)}
              disabled={loading}
            />
          </div>
        </div>

        {/* IF Section */}
        <div className="space-y-4">
          <h3 className="text-lg font-semibold">IF</h3>
          <div className="space-y-3">
            {conditions.map((condition, idx) => (
              <div key={idx} className="space-y-3">
                {idx > 0 && (
                  <div className="flex items-center">
                    <span className="text-sm font-medium text-muted-foreground bg-muted px-2 py-1 rounded">
                      AND
                    </span>
                  </div>
                )}
                <div className="flex items-center gap-2">
                  <select
                    className="px-3 py-2 bg-background border border-border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-ring min-w-[140px]"
                    value={condition.condition_type}
                    onChange={(e) =>
                      handleConditionChange(
                        idx,
                        "condition_type",
                        e.target.value
                      )
                    }
                    disabled={loading}
                  >
                    {RULE_FIELD_TYPES.map((ft) => (
                      <option key={ft.value} value={ft.value}>
                        {ft.label}
                      </option>
                    ))}
                  </select>
                  <select
                    className="px-3 py-2 bg-background border border-border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-ring min-w-[100px]"
                    value={condition.condition_operator}
                    onChange={(e) =>
                      handleConditionChange(
                        idx,
                        "condition_operator",
                        e.target.value
                      )
                    }
                    disabled={loading}
                  >
                    {RULE_OPERATORS.map((op) => (
                      <option key={op.value} value={op.value}>
                        {op.label}
                      </option>
                    ))}
                  </select>
                  <input
                    type="text"
                    className="flex-1 px-3 py-2 bg-background border border-border rounded-lg text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                    placeholder="Enter a value"
                    value={condition.condition_value}
                    onChange={(e) =>
                      handleConditionChange(
                        idx,
                        "condition_value",
                        e.target.value
                      )
                    }
                    disabled={loading}
                  />
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => handleRemoveCondition(idx)}
                    disabled={loading || conditions.length === 1}
                    className="h-8 w-8 text-muted-foreground hover:text-destructive"
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
          <div className="flex items-center gap-4">
            <Button
              variant="ghost"
              size="sm"
              onClick={handleAddCondition}
              disabled={loading}
              className="text-muted-foreground hover:text-foreground"
            >
              <Plus className="h-4 w-4 mr-2" />
              Add condition
            </Button>
            <Button
              variant="ghost"
              size="sm"
              disabled={loading}
              className="text-muted-foreground hover:text-foreground"
            >
              <Plus className="h-4 w-4 mr-2" />
              Add condition group
            </Button>
          </div>
        </div>

        {/* THEN Section */}
        <div className="space-y-4">
          <h3 className="text-lg font-semibold">THEN</h3>
          <div className="space-y-3">
            {actions.map((action, idx) => (
              <div key={idx} className="flex items-center gap-2">
                <select
                  className="px-3 py-2 bg-background border border-border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-ring min-w-[180px]"
                  value={`set_${action.action_type}`}
                  onChange={(e) =>
                    handleActionChange(
                      idx,
                      "action_type",
                      e.target.value.replace("set_", "")
                    )
                  }
                  disabled={loading}
                >
                  <option value="set_category">Set transaction category</option>
                  <option value="set_name">Set transaction name</option>
                  <option value="set_description">
                    Set transaction description
                  </option>
                  <option value="set_amount">Set transaction amount</option>
                </select>
                <span className="text-sm font-medium text-muted-foreground px-2">
                  TO
                </span>
                <input
                  type="text"
                  className="flex-1 px-3 py-2 bg-background border border-border rounded-lg text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                  placeholder="Enter value"
                  value={action.action_value}
                  onChange={(e) =>
                    handleActionChange(idx, "action_value", e.target.value)
                  }
                  disabled={loading}
                />
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => handleRemoveAction(idx)}
                  disabled={loading || actions.length === 1}
                  className="h-8 w-8 text-muted-foreground hover:text-destructive"
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            ))}
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={handleAddAction}
            disabled={loading}
            className="text-muted-foreground hover:text-foreground"
          >
            <Plus className="h-4 w-4 mr-2" />
            Add action
          </Button>
        </div>

        {/* FOR Section */}
        <div className="space-y-4">
          <h3 className="text-lg font-semibold">FOR</h3>
          <div className="flex items-center gap-2">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="outline"
                  className="justify-between"
                  disabled={loading}
                >
                  {effectiveScope === "all"
                    ? "All past and future transactions"
                    : effectiveFromDate
                      ? `Starting from ${format(effectiveFromDate, "dd/MM/yyyy")}`
                      : "Starting from (choose date)"}
                  <ChevronDownIcon className="ml-2 h-4 w-4 opacity-50" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent
                align="start"
                className="w-full min-w-[260px]"
              >
                <DropdownMenuItem
                  onSelect={() => setEffectiveScope("all")}
                  className={cn(
                    "cursor-pointer",
                    effectiveScope === "all" && "font-semibold"
                  )}
                  disabled={loading}
                >
                  All past and future transactions
                </DropdownMenuItem>
                <DropdownMenuItem
                  onSelect={() => setEffectiveScope("from")}
                  className={cn(
                    "cursor-pointer",
                    effectiveScope === "from" && "font-semibold"
                  )}
                  disabled={loading}
                >
                  Starting from (choose date)
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
            {effectiveScope === "from" && (
              <Popover>
                <PopoverTrigger asChild>
                  <Button
                    variant="outline"
                    className={cn(
                      "justify-start text-left font-normal",
                      !effectiveFromDate && "text-muted-foreground"
                    )}
                    disabled={loading}
                  >
                    <CalendarIcon className="mr-2 h-4 w-4" />
                    {effectiveFromDate ? (
                      format(effectiveFromDate, "dd/MM/yyyy")
                    ) : (
                      <span>Select date</span>
                    )}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0" align="start">
                  <Calendar
                    mode="single"
                    selected={effectiveFromDate}
                    onSelect={setEffectiveFromDate}
                    initialFocus
                  />
                </PopoverContent>
              </Popover>
            )}
          </div>

          {/* Error Display */}
          {error && (
            <div className="text-destructive text-sm bg-destructive/10 border border-destructive/20 rounded-lg p-3">
              {error}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="pt-6 mt-8 border-t">
          <Button
            onClick={handleSubmit}
            disabled={loading}
            className="w-full py-3 text-base font-medium"
          >
            {loading ? "Creating..." : "Create Rule"}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
