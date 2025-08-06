import { useCategories } from "@/components/hooks/useCategories";
import { useAccounts } from "@/components/hooks/useAccounts";
import { RuleFieldType } from "@/lib/models/rule";

interface RuleInputProps {
  fieldType: RuleFieldType;
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  disabled?: boolean;
  className?: string;
}

export function RuleInput({
  fieldType,
  value,
  onChange,
  placeholder = "Enter value",
  disabled = false,
  className = "flex-1 px-3 py-2 bg-background border border-border rounded-lg text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring",
}: RuleInputProps) {
  const { data: categories = [] } = useCategories();
  const { data: accounts = [] } = useAccounts();

  if (fieldType === "amount") {
    return (
      <input
        type="number"
        className={className}
        placeholder={placeholder}
        value={value}
        onChange={(e) => onChange(e.target.value.replace(/[^0-9.]/g, ""))}
        disabled={disabled}
      />
    );
  }

  if (fieldType === "category") {
    return (
      <select
        className={className}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        disabled={disabled}
      >
        <option value="">Select category</option>
        {categories.map((cat) => (
          <option key={cat.id} value={String(cat.id)}>
            {cat.name}
          </option>
        ))}
      </select>
    );
  }

  if (fieldType === "transfer") {
    return (
      <select
        className={className}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        disabled={disabled}
      >
        <option value="">Select account</option>
        {accounts.map((account) => (
          <option key={account.id} value={String(account.id)}>
            {account.name}
          </option>
        ))}
      </select>
    );
  }

  return (
    <input
      type="text"
      className={className}
      placeholder={placeholder}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      disabled={disabled}
    />
  );
}
