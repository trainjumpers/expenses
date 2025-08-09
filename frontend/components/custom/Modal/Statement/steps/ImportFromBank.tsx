import { LoadingButton } from "@/components/ui/LoadingButton";
import { Button } from "@/components/ui/button";
import { DialogFooter } from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Account } from "@/lib/models/account";
import { StatementUploadResponse } from "@/lib/models/statement";
import { CreateStatementRequest } from "@/lib/models/statement";
import { UseMutationResult } from "@tanstack/react-query";
import {
  AlertCircle,
  ChevronDownIcon,
  FileText,
  Upload,
  X,
} from "lucide-react";

interface ImportFromBankProps {
  accounts: Account[];
  selectedAccountId: number;
  onSelectedAccountIdChange: (id: number) => void;
  selectedFiles: File[];
  onFileInputChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onAdditionalFilesChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onFileRemove: (index?: number) => void;
  error: string;
  dragActive: boolean;
  handleDrag: (e: React.DragEvent) => void;
  handleDrop: (e: React.DragEvent) => void;
  handleSubmit: (e: React.FormEvent) => void;
  onStepChange: (step: number) => void;
  uploadStatementMutation: UseMutationResult<
    StatementUploadResponse,
    Error,
    CreateStatementRequest,
    unknown
  >;
  uploadProgress: {
    current: number;
    total: number;
    processing: boolean;
  };
  fileInputRef: React.RefObject<HTMLInputElement | null>;
  additionalFileInputRef: React.RefObject<HTMLInputElement | null>;
}

export function ImportFromBank({
  accounts,
  selectedAccountId,
  onSelectedAccountIdChange,
  selectedFiles,
  onFileInputChange,
  onAdditionalFilesChange,
  onFileRemove,
  error,
  dragActive,
  handleDrag,
  handleDrop,
  handleSubmit,
  onStepChange,
  uploadStatementMutation,
  uploadProgress,
  fileInputRef,
  additionalFileInputRef,
}: ImportFromBankProps) {
  return (
    <form onSubmit={handleSubmit}>
      <div className="space-y-6 py-4">
        {/* Account Selection */}
        <div className="space-y-2">
          <Label>Account</Label>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="outline"
                className="w-full justify-start text-left font-normal flex items-center"
                type="button"
              >
                {(() => {
                  const selected = accounts.find(
                    (acc) => acc.id === selectedAccountId
                  );
                  return selected
                    ? `${selected.name} (${selected.bank_type.toUpperCase()})`
                    : "Select account";
                })()}
                <ChevronDownIcon className="ml-auto w-4 h-4 opacity-60" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent className="w-56 max-h-64 overflow-y-auto">
              {accounts.map((account) => (
                <DropdownMenuItem
                  key={account.id}
                  onClick={() => onSelectedAccountIdChange(account.id)}
                  className={`py-1 px-2 text-sm min-h-0 h-8 cursor-pointer flex items-center ${
                    selectedAccountId === account.id
                      ? "bg-accent/40 font-semibold"
                      : ""
                  }`}
                >
                  {account.name} ({account.bank_type.toUpperCase()})
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        {/* File Upload */}
        <div className="space-y-2">
          <Label>Statement Files</Label>
          <div className="space-y-4">
            {selectedFiles.length === 0 ? (
              <div
                onDragEnter={handleDrag}
                onDragLeave={handleDrag}
                onDragOver={handleDrag}
                onDrop={handleDrop}
                className={`border-2 border-dashed rounded-lg p-6 text-center cursor-pointer transition-colors ${
                  dragActive
                    ? "border-primary bg-primary/5"
                    : "border-border hover:border-primary dark:border-border dark:hover:border-primary"
                }`}
                onClick={() => document.getElementById("file-input")?.click()}
              >
                <Upload className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
                <p className="text-sm text-foreground mb-2">
                  {dragActive
                    ? "Drop the files here..."
                    : "Drag & drop your bank statements here, or click to select"}
                </p>
                <p className="text-xs text-muted-foreground">
                  Supports CSV, XLS, XLSX, TXT files (max 256KB each, up to 10
                  files)
                </p>
                <Input
                  id="file-input"
                  ref={fileInputRef}
                  type="file"
                  accept=".csv,.xls,.xlsx,.txt"
                  multiple
                  onChange={onFileInputChange}
                  className="hidden"
                />
              </div>
            ) : (
              <div className="space-y-2">
                {selectedFiles.map((file, index) => (
                  <div
                    key={index}
                    className="border rounded-lg p-4 bg-muted/50 dark:bg-muted/20"
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-3">
                        <FileText className="h-8 w-8 text-blue-500 dark:text-blue-400" />
                        <div>
                          <p className="text-sm font-medium text-foreground">
                            {file.name}
                          </p>
                          <p className="text-xs text-muted-foreground">
                            {(file.size / 1024).toFixed(1)} KB
                          </p>
                        </div>
                      </div>
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => onFileRemove(index)}
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                ))}

                {/* Add more files button */}
                {selectedFiles.length < 10 && (
                  <div
                    className="border-2 border-dashed rounded-lg p-4 text-center cursor-pointer transition-colors hover:border-primary"
                    onClick={() =>
                      document.getElementById("file-input-additional")?.click()
                    }
                  >
                    <p className="text-sm text-muted-foreground">
                      Click to add more files ({selectedFiles.length}/10)
                    </p>
                    <Input
                      id="file-input-additional"
                      ref={additionalFileInputRef}
                      type="file"
                      accept=".csv,.xls,.xlsx,.txt"
                      multiple
                      onChange={onAdditionalFilesChange}
                      className="hidden"
                    />
                  </div>
                )}
              </div>
            )}

            {/* Upload Progress */}
            {uploadProgress.processing && (
              <div className="bg-blue-50 dark:bg-blue-950/50 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
                <div className="flex items-center space-x-3">
                  <div className="text-sm text-blue-700 dark:text-blue-300">
                    <p className="font-medium">
                      Processing files... ({uploadProgress.current}/
                      {uploadProgress.total})
                    </p>
                    <div className="w-full bg-blue-200 dark:bg-blue-800 rounded-full h-2 mt-2">
                      <div
                        className="bg-blue-500 h-2 rounded-full transition-all duration-300"
                        style={{
                          width: `${(uploadProgress.current / uploadProgress.total) * 100}%`,
                        }}
                      />
                    </div>
                  </div>
                </div>
              </div>
            )}

            {/* Error Display */}
            {error && (
              <div className="text-sm text-destructive flex items-center space-x-2">
                <AlertCircle className="h-4 w-4" />
                <span>{error}</span>
              </div>
            )}
          </div>
        </div>

        <div className="bg-blue-50 dark:bg-blue-950/50 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
          <div className="flex items-start space-x-3">
            <AlertCircle className="h-5 w-5 text-blue-500 dark:text-blue-400 mt-0.5" />
            <div className="text-sm text-blue-700 dark:text-blue-300">
              <p className="font-medium mb-1">Processing Information:</p>
              <ul className="text-xs space-y-1 text-blue-600 dark:text-blue-400">
                <li>• Your statement will be processed in the background</li>
                <li>
                  • You can check the processing status in the statements
                  history
                </li>
              </ul>
            </div>
          </div>
        </div>
      </div>

      <DialogFooter>
        <Button type="button" variant="outline" onClick={() => onStepChange(1)}>
          Back
        </Button>
        <LoadingButton
          type="submit"
          loading={uploadStatementMutation.isPending}
          disabled={
            uploadStatementMutation.isPending || selectedFiles.length === 0
          }
          fixedWidth="140px"
        >
          Import Statement
        </LoadingButton>
      </DialogFooter>
    </form>
  );
}
