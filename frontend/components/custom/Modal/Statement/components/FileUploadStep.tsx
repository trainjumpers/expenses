import { AlertCircle } from "lucide-react";
import { ImportType } from "@/lib/models/statement";
import { AccountSelector } from "./AccountSelector";
import { FileUpload } from "./FileUpload";

interface Account {
  id: number;
  name: string;
  bank_type: string;
}

interface FileUploadStepProps {
  accounts: Account[];
  selectedAccountId: number;
  selectedFile: File | null;
  dragActive: boolean;
  importType: ImportType;
  onAccountSelect: (accountId: number) => void;
  onFileSelect: (file: File) => void;
  onRemoveFile: () => void;
  onDragStateChange: (active: boolean) => void;
}

export function FileUploadStep({
  accounts,
  selectedAccountId,
  selectedFile,
  dragActive,
  importType,
  onAccountSelect,
  onFileSelect,
  onRemoveFile,
  onDragStateChange,
}: FileUploadStepProps) {
  return (
    <div className="space-y-4">
      <AccountSelector
        accounts={accounts}
        selectedAccountId={selectedAccountId}
        onAccountSelect={onAccountSelect}
      />

      <FileUpload
        selectedFile={selectedFile}
        dragActive={dragActive}
        onFileSelect={onFileSelect}
        onRemoveFile={onRemoveFile}
        onDragStateChange={onDragStateChange}
      />

      {importType === 'bank' && (
        <div className="bg-blue-50 dark:bg-blue-950/50 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
          <div className="flex items-start space-x-3">
            <AlertCircle className="h-5 w-5 text-blue-500 dark:text-blue-400 mt-0.5" />
            <div className="text-sm text-blue-700 dark:text-blue-300">
              <p className="font-medium mb-1">Processing Information:</p>
              <ul className="text-xs space-y-1 text-blue-600 dark:text-blue-400">
                <li>• Your statement will be processed in the background</li>
                <li>• You can check the processing status in the statements history</li>
              </ul>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}