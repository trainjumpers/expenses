import { RuleInput } from "@/components/custom/Modal/Rule/components/RuleInput";
import { Button } from "@/components/ui/button";
import {
  BaseRuleCondition,
  ConditionLogic,
  RULE_FIELD_TYPES,
  RULE_OPERATORS,
  RuleFieldType,
  RuleOperator,
} from "@/lib/models/rule";
import { Plus, Trash2 } from "lucide-react";

interface RuleConditionsProps {
  conditions: BaseRuleCondition[];
  onConditionsChange: (conditions: BaseRuleCondition[]) => void;
  conditionLogic: ConditionLogic;
  onConditionLogicChange: (logic: ConditionLogic) => void;
  disabled?: boolean;
}

export function RuleConditions({
  conditions,
  onConditionsChange,
  conditionLogic,
  onConditionLogicChange,
  disabled = false,
}: RuleConditionsProps) {
  const handleAddCondition = () => {
    const newConditions = [
      ...conditions,
      {
        condition_type: "name" as RuleFieldType,
        condition_operator: "contains" as RuleOperator,
        condition_value: "",
      },
    ];
    onConditionsChange(newConditions);
  };

  const handleRemoveCondition = (idx: number) => {
    if (conditions.length > 1) {
      const newConditions = conditions.filter((_, i) => i !== idx);
      onConditionsChange(newConditions);
    }
  };

  const handleConditionChange = (idx: number, field: string, value: string) => {
    const newConditions = conditions.map((c, i) =>
      i === idx ? { ...c, [field]: value } : c
    );
    onConditionsChange(newConditions);
  };

  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold">IF</h3>
      <div className="flex items-center gap-2 text-sm">
        <span>Match</span>
        <select
          value={conditionLogic}
          onChange={(e) =>
            onConditionLogicChange(e.target.value as ConditionLogic)
          }
          disabled={disabled || conditions.length < 2}
          className="px-2 py-1 bg-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
        >
          <option value={ConditionLogic.AND}>All</option>
          <option value={ConditionLogic.OR}>Any</option>
        </select>
        <span>of the following conditions:</span>
      </div>
      <div className="space-y-3 pl-4 border-l-2">
        {conditions.map((condition, idx) => (
          <div key={idx} className="space-y-3">
            {idx > 0 && (
              <div className="flex items-center">
                <span className="text-sm font-medium text-muted-foreground bg-muted px-2 py-1 rounded">
                  {conditionLogic}
                </span>
              </div>
            )}
            <div className="flex items-center gap-2">
              <select
                className="px-3 py-2 bg-background border border-border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-ring min-w-[140px]"
                value={condition.condition_type}
                onChange={(e) =>
                  handleConditionChange(idx, "condition_type", e.target.value)
                }
                disabled={disabled}
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
                disabled={disabled}
              >
                {RULE_OPERATORS.map((op) => (
                  <option key={op.value} value={op.value}>
                    {op.label}
                  </option>
                ))}
              </select>
              <RuleInput
                fieldType={condition.condition_type}
                value={condition.condition_value}
                onChange={(value) =>
                  handleConditionChange(idx, "condition_value", value)
                }
                placeholder={
                  condition.condition_type === "amount"
                    ? "Enter amount"
                    : "Enter a value"
                }
                disabled={disabled}
              />
              <Button
                variant="ghost"
                size="icon"
                onClick={() => handleRemoveCondition(idx)}
                disabled={disabled || conditions.length === 1}
                className="h-8 w-8 text-muted-foreground hover:text-destructive"
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          </div>
        ))}
        <div className="flex items-center gap-4">
          <Button
            variant="ghost"
            size="sm"
            onClick={handleAddCondition}
            disabled={disabled}
            className="text-muted-foreground hover:text-foreground"
          >
            <Plus className="h-4 w-4 mr-2" />
            Add condition
          </Button>
        </div>
      </div>
    </div>
  );
}
