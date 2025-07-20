import { FileText, Upload } from "lucide-react";

interface ImportTypeSelectionProps {
  onImportTypeSelect: (type: 'bank' | 'custom') => void;
}

export function ImportTypeSelection({ onImportTypeSelect }: ImportTypeSelectionProps) {
  return (
    <div className="space-y-4">
      <div className="text-center">
        <h3 className="text-lg font-medium mb-2">Choose Import Method</h3>
        <p className="text-sm text-muted-foreground mb-6">
          Select how you&apos;d like to import your statement
        </p>
      </div>

      <div className="grid grid-cols-1 gap-4">
        <button
          type="button"
          onClick={() => onImportTypeSelect('bank')}
          className="p-4 border-2 border-dashed rounded-lg hover:border-primary hover:bg-primary/5 transition-colors text-left"
        >
          <div className="flex items-start space-x-3">
            <FileText className="h-6 w-6 text-blue-500 mt-1" />
            <div>
              <h4 className="font-medium">Bank Statement</h4>
              <p className="text-sm text-muted-foreground">
                Import using your bank&apos;s specific format (SBI, HDFC, etc.)
              </p>
            </div>
          </div>
        </button>

        <button
          type="button"
          onClick={() => onImportTypeSelect('custom')}
          className="p-4 border-2 border-dashed rounded-lg hover:border-primary hover:bg-primary/5 transition-colors text-left"
        >
          <div className="flex items-start space-x-3">
            <Upload className="h-6 w-6 text-green-500 mt-1" />
            <div>
              <h4 className="font-medium">Custom CSV/Excel</h4>
              <p className="text-sm text-muted-foreground">
                Import any CSV or Excel file with custom column mapping
              </p>
            </div>
          </div>
        </button>
      </div>
    </div>
  );
}