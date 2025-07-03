import { useCategories } from "@/components/custom/Provider/CategoryProvider";
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
  const { read: readCategories } = useCategories();
  const categories = readCategories();

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
