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
import { AlertCircle, ArrowLeft } from "lucide-react";
import { useState } from "react";
import {
  ColumnMapping
} from "@/lib/models/statement";
import {
  ImportTypeSelection,
  FileUploadStep,
  CSVPreview as CSVPreviewComponent,
  ColumnMapping as ColumnMappingComponent
} from "./components";

type Step = 'import-type' | 'file-upload' | 'preview' | 'mapping';

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

  const [currentStep, setCurrentStep] = useState<Step>('import-type');
  const [importType, setImportType] = useState<'bank' | 'custom'>('bank');
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [selectedAccountId, setSelectedAccountId] = useState<number>(
    accounts[0]?.id || 0
  );
  const [dragActive, setDragActive] = useState(false);
  const [error, setError] = useState<string>("");
  const [csvPreview, setCsvPreview] = useState<{
    columns: string[];
    rows: string[][];
    total: number;
  } | null>(null);
  const [skipRows, setSkipRows] = useState<number>(0);
  const [columnMappings, setColumnMappings] = useState<ColumnMapping[]>([]);
  const [isLoadingPreview, setIsLoadingPreview] = useState(false);
  const [isRefreshingPreview, setIsRefreshingPreview] = useState(false);

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





  const handleCancel = () => {
    // Reset all state
    setCurrentStep('import-type');
    setImportType('bank');
    setSelectedFile(null);
    setSelectedAccountId(accounts[0]?.id || 0);
    setCsvPreview(null);
    setSkipRows(0);
    setColumnMappings([]);
    setError("");
    onOpenChange(false);
  };

  const removeFile = () => {
    setSelectedFile(null);
    setError("");
  };

  const handleImportTypeSelect = (type: 'bank' | 'custom') => {
    setImportType(type);
    setCurrentStep('file-upload');
  };

  const handleFileNext = async () => {
    if (!selectedFile || !selectedAccountId) {
      setError("Please select a file and account");
      return;
    }

    if (importType === 'bank') {
      // For bank import, use unified endpoint with empty metadata
      uploadStatementMutation.mutate(
        {
          account_id: selectedAccountId,
          file: selectedFile,
          // No metadata for bank imports - will use defaults (skip_rows=0, mappings=[])
        },
        {
          onSuccess: () => {
            handleCancel();
          },
        }
      );
    } else {
      // For custom import, generate a simple preview from the file
      setIsLoadingPreview(true);
      try {
        const text = await selectedFile.text();
        const lines = text.split('\n').filter(line => line.trim());
        
        if (lines.length === 0) {
          throw new Error('File appears to be empty');
        }

        // Parse CSV headers (first line)
        const headers = lines[0].split(',').map(h => h.trim().replace(/"/g, ''));
        
        // Get preview rows (next few lines)
        const previewRows = lines.slice(1, Math.min(6, lines.length))
          .map(line => line.split(',').map(cell => cell.trim().replace(/"/g, '')));

        setCsvPreview({
          columns: headers,
          rows: previewRows,
          total: lines.length - 1, // Exclude header
        });
        setCurrentStep('preview');
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to preview CSV');
      } finally {
        setIsLoadingPreview(false);
      }
    }
  };

  const handleSkipRowsChange = async (newSkipRows: number) => {
    if (!selectedFile) return;

    setSkipRows(newSkipRows);
    setIsRefreshingPreview(true);

    try {
      const text = await selectedFile.text();
      const lines = text.split('\n').filter(line => line.trim());
      
      if (lines.length <= newSkipRows) {
        throw new Error('Skip rows exceeds file length');
      }

      // Skip the specified number of rows, then parse headers
      const remainingLines = lines.slice(newSkipRows);
      const headers = remainingLines[0].split(',').map(h => h.trim().replace(/"/g, ''));
      
      // Get preview rows (next few lines after headers)
      const previewRows = remainingLines.slice(1, Math.min(6, remainingLines.length))
        .map(line => line.split(',').map(cell => cell.trim().replace(/"/g, '')));

      setCsvPreview({
        columns: headers,
        rows: previewRows,
        total: remainingLines.length - 1, // Exclude header
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to refresh CSV preview');
    } finally {
      setIsRefreshingPreview(false);
    }
  };

  const handlePreviewNext = () => {
    setCurrentStep('mapping');
  };

  const handleCustomImportSubmit = async () => {
    if (!selectedFile || !selectedAccountId || columnMappings.length === 0) {
      setError("Please complete all required fields");
      return;
    }

    // Validate required mappings
    const hasName = columnMappings.some(m => m.target_field === 'name');
    const hasDate = columnMappings.some(m => m.target_field === 'date');
    const hasAmount = columnMappings.some(m => m.target_field === 'amount');
    const hasCredit = columnMappings.some(m => m.target_field === 'credit');
    const hasDebit = columnMappings.some(m => m.target_field === 'debit');

    if (!hasName) {
      setError("Name field is required");
      return;
    }

    if (!hasDate) {
      setError("Date field is required");
      return;
    }

    if (!hasAmount && !(hasCredit && hasDebit)) {
      setError("Either Amount OR both Credit and Debit fields are required");
      return;
    }

    try {
      // Use the unified uploadStatementMutation with metadata
      uploadStatementMutation.mutate(
        {
          account_id: selectedAccountId,
          file: selectedFile,
          metadata: {
            skip_rows: skipRows,
            mappings: columnMappings,
          },
        },
        {
          onSuccess: () => {
            handleCancel();
          },
        }
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to import CSV");
    }
  };

  const handleBack = () => {
    switch (currentStep) {
      case 'file-upload':
        setCurrentStep('import-type');
        break;
      case 'preview':
        setCurrentStep('file-upload');
        break;
      case 'mapping':
        setCurrentStep('preview');
        break;
    }
  };

  return (
    <>
      <Dialog open={isOpen} onOpenChange={onOpenChange}>
        <DialogContent className={`${currentStep === 'preview' || currentStep === 'mapping' ? 'sm:max-w-[800px]' : 'sm:max-w-[500px]'}`}>
          <DialogHeader>
            <DialogTitle>Import Bank Statement</DialogTitle>
          </DialogHeader>

          <div className="space-y-6 py-4">
            {/* Step 1: Import Type Selection */}
            {currentStep === 'import-type' && (
              <ImportTypeSelection onImportTypeSelect={handleImportTypeSelect} />
            )}

            {/* Step 2: File Upload */}
            {currentStep === 'file-upload' && (
              <FileUploadStep
                accounts={accounts}
                selectedAccountId={selectedAccountId}
                selectedFile={selectedFile}
                dragActive={dragActive}
                importType={importType}
                onAccountSelect={setSelectedAccountId}
                onFileSelect={handleFileSelect}
                onRemoveFile={removeFile}
                onDragStateChange={setDragActive}
              />
            )}

            {/* Step 3: CSV Preview */}
            {currentStep === 'preview' && csvPreview && (
              <CSVPreviewComponent
                csvPreview={csvPreview}
                skipRows={skipRows}
                isRefreshingPreview={isRefreshingPreview}
                onSkipRowsChange={handleSkipRowsChange}
              />
            )}

            {/* Step 4: Column Mapping */}
            {currentStep === 'mapping' && csvPreview && (
              <ColumnMappingComponent
                csvPreview={csvPreview}
                columnMappings={columnMappings}
                onColumnMappingChange={setColumnMappings}
              />
            )}



            {/* Error Display */}
            {error && (
              <div className="text-sm text-destructive flex items-center space-x-2">
                <AlertCircle className="h-4 w-4" />
                <span>{error}</span>
              </div>
            )}
          </div>

          <DialogFooter>
            {currentStep !== 'import-type' && (
              <Button
                type="button"
                variant="outline"
                onClick={handleBack}
                disabled={isLoadingPreview || uploadStatementMutation.isPending}
              >
                <ArrowLeft className="h-4 w-4 mr-2" />
                Back
              </Button>
            )}

            <Button
              type="button"
              variant="outline"
              onClick={handleCancel}
              disabled={isLoadingPreview || uploadStatementMutation.isPending}
            >
              Cancel
            </Button>

            {currentStep === 'file-upload' && (
              <LoadingButton
                type="button"
                onClick={handleFileNext}
                loading={isLoadingPreview || uploadStatementMutation.isPending}
                disabled={!selectedFile || !selectedAccountId}
                fixedWidth="140px"
              >
                {importType === 'bank' ? 'Import Statement' : 'Preview'}
              </LoadingButton>
            )}

            {currentStep === 'preview' && (
              <Button
                type="button"
                onClick={handlePreviewNext}
                disabled={!csvPreview}
              >
                Next: Map Columns
              </Button>
            )}

            {currentStep === 'mapping' && (
              <LoadingButton
                type="button"
                onClick={handleCustomImportSubmit}
                loading={uploadStatementMutation.isPending}
                disabled={columnMappings.length === 0}
                fixedWidth="140px"
              >
                Import CSV
              </LoadingButton>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
