interface RuleBasicInfoProps {
  ruleName: string;
  ruleDescription: string;
  onRuleNameChange: (name: string) => void;
  onRuleDescriptionChange: (description: string) => void;
  disabled?: boolean;
}

export function RuleBasicInfo({
  ruleName,
  ruleDescription,
  onRuleNameChange,
  onRuleDescriptionChange,
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
          value={ruleName}
          onChange={(e) => onRuleNameChange(e.target.value)}
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
          value={ruleDescription}
          onChange={(e) => onRuleDescriptionChange(e.target.value)}
          disabled={disabled}
        />
      </div>
    </div>
  );
}
