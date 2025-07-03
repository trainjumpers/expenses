import { RuleInput } from "@/components/custom/Modal/Rule/components/RuleInput";
import { Button } from "@/components/ui/button";
import { BaseRuleAction, RuleFieldType } from "@/lib/models/rule";
import { Plus, Trash2 } from "lucide-react";

interface RuleActionsProps {
  actions: BaseRuleAction[];
  onActionsChange: (actions: BaseRuleAction[]) => void;
  disabled?: boolean;
}

export function RuleActions({
  actions,
  onActionsChange,
  disabled = false,
}: RuleActionsProps) {
  const handleAddAction = () => {
    const newActions = [
      ...actions,
      { action_type: "category" as RuleFieldType, action_value: "" },
    ];
    onActionsChange(newActions);
  };

  const handleRemoveAction = (idx: number) => {
    if (actions.length > 1) {
      const newActions = actions.filter((_, i) => i !== idx);
      onActionsChange(newActions);
    }
  };

  const handleActionChange = (idx: number, field: string, value: string) => {
    const newActions = actions.map((a, i) =>
      i === idx ? { ...a, [field]: value } : a
    );
    onActionsChange(newActions);
  };

  return (
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
              disabled={disabled}
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
            <RuleInput
              fieldType={action.action_type}
              value={action.action_value}
              onChange={(value) =>
                handleActionChange(idx, "action_value", value)
              }
              placeholder={
                action.action_type === "amount" ? "Enter amount" : "Enter value"
              }
              disabled={disabled}
            />
            <Button
              variant="ghost"
              size="icon"
              onClick={() => handleRemoveAction(idx)}
              disabled={disabled || actions.length === 1}
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
        disabled={disabled}
        className="text-muted-foreground hover:text-foreground"
      >
        <Plus className="h-4 w-4 mr-2" />
        Add action
      </Button>
    </div>
  );
}
