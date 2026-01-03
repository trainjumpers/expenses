import type { BaseRule } from "@/lib/models/rule";

interface RuleBasicInfoProps {
  rule: BaseRule;
  onRuleChange: (rule: BaseRule) => void;
  disabled?: boolean;
}

export function RuleBasicInfo({
  rule,
  onRuleChange,
  disabled = false,
}: RuleBasicInfoProps) {
  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center gap-2">
        <label className="text-sm font-medium text-muted-foreground min-w-[90px]">
          Rule name
        </label>
        <input
          type="text"
          className="flex-1 px-4 py-2 bg-background border border-border rounded-lg text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
          placeholder="Enter a name for this rule"
          value={rule.name}
          onChange={(e) => onRuleChange({ ...rule, name: e.target.value })}
          disabled={disabled}
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
          value={rule.description || ""}
          onChange={(e) =>
            onRuleChange({ ...rule, description: e.target.value })
          }
          disabled={disabled}
        />
      </div>
    </div>
  );
}
