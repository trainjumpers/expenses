import { Button } from "@/components/ui/button";
import { DialogFooter } from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { AlertCircle } from "lucide-react";
import { useState } from "react";

interface MapColumnsProps {
  headers: string[];
  onStepChange: (step: number) => void;
  onCancel: () => void;
  onSubmit: (mappings: Record<string, string>) => void;
}

const requiredFields = [
  { value: "txn_date", label: "Transaction Date" },
  { value: "name", label: "Name" },
  { value: "description", label: "Description (Optional)" },
  { value: "amount", label: "Amount" },
  { value: "credit", label: "Credit" },
  { value: "debit", label: "Debit" },
];

export function MapColumns({
  headers,
  onStepChange,
  onCancel,
  onSubmit,
}: MapColumnsProps) {
  const [mappings, setMappings] = useState<Record<string, string>>({});
  const [error, setError] = useState<string>("");

  const handleMappingChange = (field: string, header: string) => {
    setMappings((prev) => {
      const updated = { ...prev };
      if (header === "" || header === "none") {
        delete updated[field];
      } else {
        updated[field] = header;
      }
      return updated;
    });
  };

  const validateMappings = () => {
    if (!mappings.txn_date) {
      return "Transaction Date must be mapped.";
    }
    if (!mappings.name) {
      return "Name must be mapped.";
    }
    const hasAmount = !!mappings.amount;
    const hasCreditDebit = !!mappings.credit && !!mappings.debit;
    if (!hasAmount && !hasCreditDebit) {
      return "You must map either 'Amount' or both 'Credit' and 'Debit'.";
    }
    if (hasAmount && (mappings.credit || mappings.debit)) {
      return "You cannot map 'Amount' with 'Credit' or 'Debit'. Please choose one method.";
    }
    return "";
  };

  const handleNext = () => {
    const validationError = validateMappings();
    if (validationError) {
      setError(validationError);
      return;
    }
    setError("");
    onSubmit(mappings);
    onCancel();
  };

  return (
    <div className="space-y-4 py-4">
      <div className="bg-blue-50 dark:bg-blue-950/50 border border-blue-200 dark:border-blue-800 rounded-lg p-3">
        <div className="flex items-start space-x-3">
          <AlertCircle className="h-5 w-5 text-blue-500 dark:text-blue-400 mt-0.5 flex-shrink-0" />
          <div className="text-sm text-blue-700 dark:text-blue-300">
            <p className="font-medium">Map Your Columns</p>
            <p className="text-xs mt-1 text-blue-600 dark:text-blue-400">
              Match the columns from your file to the required transaction
              fields. For amounts, you can use a single &apos;Amount&apos;
              column or separate &apos;Credit&apos; and &apos;Debit&apos;
              columns.
            </p>
          </div>
        </div>
      </div>

      <div className="space-y-4">
        {requiredFields.map((field) => (
          <div key={field.value} className="flex items-center justify-between">
            <Label>{field.label}</Label>
            <div className="w-1/2">
              <Select
                value={mappings[field.value] || ""}
                onValueChange={(value) =>
                  handleMappingChange(field.value, value)
                }
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Select a column" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">None</SelectItem>
                  {headers
                    .filter((h) => h)
                    .map((header, i) => (
                      <SelectItem key={i} value={header}>
                        {header}
                      </SelectItem>
                    ))}
                </SelectContent>
              </Select>
            </div>
          </div>
        ))}
      </div>

      {error && (
        <div className="text-sm text-destructive flex items-center space-x-2 pt-2">
          <AlertCircle className="h-4 w-4" />
          <span>{error}</span>
        </div>
      )}

      <DialogFooter>
        <Button type="button" variant="outline" onClick={() => onStepChange(3)}>
          Back
        </Button>
        <Button type="button" onClick={handleNext}>
          Process Transactions
        </Button>
      </DialogFooter>
    </div>
  );
}
