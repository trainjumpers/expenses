import { useAccounts } from "@/components/hooks/useAccounts";
import { useUploadStatement } from "@/components/hooks/useStatements";
import { LoadingButton } from "@/components/ui/LoadingButton";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  AlertCircle,
  ChevronDownIcon,
  FileText,
  Upload,
  X,
} from "lucide-react";
import { useCallback, useState } from "react";

interface ImportStatementModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ImportStatementModal({
  isOpen,
  onOpenChange,
}: ImportStatementModalProps) {
  const { data: accounts = [] } = useAccounts();
  const uploadStatementMutation = useUploadStatement();

  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [selectedAccountId, setSelectedAccountId] = useState<number>(
    accounts[0]?.id || 0
  );
  const [dragActive, setDragActive] = useState(false);
  const [error, setError] = useState<string>("");

  const validateFile = (file: File): string | null => {
    // Check file size (256KB)
    if (file.size > 256 * 1024) {
      return "File size must be less than 256KB";
    }

    // Check file type
    const validTypes = [
      "text/csv",
      "application/vnd.ms-excel",
      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
    ];
    const validExtensions = [".csv", ".xls", ".xlsx"];

    const isValidType =
      validTypes.includes(file.type) ||
      validExtensions.some((ext) => file.name.toLowerCase().endsWith(ext));

    if (!isValidType) {
      return "File must be CSV or Excel format (.csv, .xls, .xlsx)";
    }

    return null;
  };

  const handleFileSelect = (file: File) => {
    const validationError = validateFile(file);
    if (validationError) {
      setError(validationError);
      setSelectedFile(null);
      return;
    }

    setError("");
    setSelectedFile(file);
  };

  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      handleFileSelect(file);
    }
  };

  const handleDrag = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === "dragenter" || e.type === "dragover") {
      setDragActive(true);
    } else if (e.type === "dragleave") {
      setDragActive(false);
    }
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);

    const file = e.dataTransfer.files?.[0];
    if (file) {
      const validationError = validateFile(file);
      if (validationError) {
        setError(validationError);
        setSelectedFile(null);
        return;
      }

      setError("");
      setSelectedFile(file);
    }
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!selectedFile) {
      setError("Please select a file");
      return;
    }

    if (!selectedAccountId) {
      setError("Please select an account");
      return;
    }

    uploadStatementMutation.mutate(
      {
        account_id: selectedAccountId,
        file: selectedFile,
      },
      {
        onSuccess: () => {
          handleCancel();
        },
      }
    );
  };

  const handleCancel = () => {
    setSelectedFile(null);
    setSelectedAccountId(accounts[0]?.id || 0);
    setError("");
    onOpenChange(false);
  };

  const removeFile = () => {
    setSelectedFile(null);
    setError("");
  };

  return (
    <>
      <Dialog open={isOpen} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>Import Bank Statement</DialogTitle>
          </DialogHeader>

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
                        onClick={() => setSelectedAccountId(account.id)}
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
                <Label>Statement File</Label>
                <div className="space-y-4">
                  {!selectedFile ? (
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
                      onClick={() =>
                        document.getElementById("file-input")?.click()
                      }
                    >
                      <Upload className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
                      <p className="text-sm text-foreground mb-2">
                        {dragActive
                          ? "Drop the file here..."
                          : "Drag & drop your bank statement here, or click to select"}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        Supports CSV, XLS, XLSX files (max 256KB)
                      </p>
                      <Input
                        id="file-input"
                        type="file"
                        accept=".csv,.xls,.xlsx"
                        onChange={handleFileInputChange}
                        className="hidden"
                      />
                    </div>
                  ) : (
                    <div className="border rounded-lg p-4 bg-muted/50 dark:bg-muted/20">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-3">
                          <FileText className="h-8 w-8 text-blue-500 dark:text-blue-400" />
                          <div>
                            <p className="text-sm font-medium text-foreground">
                              {selectedFile.name}
                            </p>
                            <p className="text-xs text-muted-foreground">
                              {(selectedFile.size / 1024).toFixed(1)} KB
                            </p>
                          </div>
                        </div>
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={removeFile}
                        >
                          <X className="h-4 w-4" />
                        </Button>
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
                      <li>
                        • Your statement will be processed in the background
                      </li>
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
              <Button
                type="button"
                variant="outline"
                onClick={handleCancel}
                disabled={uploadStatementMutation.isPending}
              >
                Cancel
              </Button>
              <LoadingButton
                type="submit"
                loading={uploadStatementMutation.isPending}
                disabled={uploadStatementMutation.isPending || !selectedFile}
                fixedWidth="140px"
              >
                Import Statement
              </LoadingButton>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </>
  );
}
