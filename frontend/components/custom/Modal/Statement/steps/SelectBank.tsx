import { Button } from "@/components/ui/button";
import { FileCog, Landmark } from "lucide-react";

interface SelectBankProps {
  onStepChange: (step: number) => void;
}

export function SelectBank({ onStepChange }: SelectBankProps) {
  return (
    <div className="py-4 space-y-4">
      <Button
        variant="outline"
        className="w-full justify-start p-4 h-auto flex items-center space-x-4"
        onClick={() => onStepChange(2)}
      >
        <div className="bg-blue-100 dark:bg-blue-900/50 p-3 rounded-lg">
          <Landmark className="h-6 w-6 text-blue-600 dark:text-blue-400" />
        </div>
        <div className="text-left">
          <p className="font-semibold">Import from Bank</p>
          <p className="text-xs text-muted-foreground">
            Recommended for supported banks.
          </p>
        </div>
      </Button>
      <Button
        variant="outline"
        className="w-full justify-start p-4 h-auto flex items-center space-x-4"
        onClick={() => onStepChange(3)}
      >
        <div className="bg-green-100 dark:bg-green-900/50 p-3 rounded-lg">
          <FileCog className="h-6 w-6 text-green-600 dark:text-green-300" />
        </div>
        <div className="text-left">
          <p className="font-semibold">Custom Parsing</p>
          <p className="text-xs text-muted-foreground">
            For unsupported banks or parsing failures.
          </p>
        </div>
      </Button>
    </div>
  );
}
