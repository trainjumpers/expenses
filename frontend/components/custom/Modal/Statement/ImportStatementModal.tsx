import { useAccounts } from "@/components/hooks/useAccounts";
import {
  usePreviewStatement,
  useUploadStatement,
} from "@/components/hooks/useStatements";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { StatementPreviewResponse } from "@/lib/models/statement";
import { useCallback, useState } from "react";

import { FallbackParsing } from "./steps/FallbackParsing";
import { ImportFromBank } from "./steps/ImportFromBank";
import { MapColumns } from "./steps/MapColumns";
import { SelectBank } from "./steps/SelectBank";

// Constants
const PREVIEW_SIZE = 5;
enum ImportStep {
  SelectBank = 1,
  ImportFromBank = 2,
  Preview = 3,
  MapColumns = 4,
}

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
  const previewStatementMutation = usePreviewStatement();

  const [step, setStep] = useState<ImportStep>(ImportStep.SelectBank);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [skipRows, setSkipRows] = useState(0);
  const [rowSize, setRowSize] = useState(PREVIEW_SIZE);
  const [previewData, setPreviewData] =
    useState<StatementPreviewResponse | null>(null);
  const [selectedAccountId, setSelectedAccountId] = useState<number>(
    accounts[0]?.id || 0
  );
  const [dragActive, setDragActive] = useState(false);
  const [error, setError] = useState<string>("");

  const validateFile = (file: File, forBank: boolean): string | null => {
    if (file.size > 256 * 1024) {
      return "File size must be less than 256KB";
    }
    const validExtensions = forBank
      ? [".csv", ".xls", ".xlsx"]
      : [".csv", ".xls"];
    if (!validExtensions.some((ext) => file.name.toLowerCase().endsWith(ext))) {
      return `File must be ${validExtensions.join(", ")} format`;
    }
    return null;
  };

  const handleFileSelect = useCallback(
    (file: File) => {
      const validationError = validateFile(
        file,
        step === ImportStep.ImportFromBank
      );
      if (validationError) {
        setError(validationError);
        setSelectedFile(null);
        setPreviewData(null);
        return;
      }

      setError("");
      setSelectedFile(file);
      if (step === ImportStep.Preview) {
        previewStatementMutation.mutate(
          { file, skipRows, rowSize },
          {
            onSuccess: (data) => setPreviewData(data),
            onError: () => setPreviewData(null),
          }
        );
      }
    },
    [step, skipRows, rowSize, previewStatementMutation]
  );

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

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      e.stopPropagation();
      setDragActive(false);
      const file = e.dataTransfer.files?.[0];
      if (file) {
        handleFileSelect(file);
      }
    },
    [handleFileSelect]
  );

  const handleSkipRowsChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = parseInt(e.target.value, 10);
    setSkipRows(value);
    if (selectedFile) {
      previewStatementMutation.mutate(
        { file: selectedFile, skipRows: value, rowSize },
        {
          onSuccess: (data) => setPreviewData(data),
          onError: () => setPreviewData(null),
        }
      );
    }
  };

  const handleRowSizeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = parseInt(e.target.value, PREVIEW_SIZE);
    setRowSize(isNaN(value) ? PREVIEW_SIZE : value);
    if (selectedFile) {
      previewStatementMutation.mutate(
        { file: selectedFile, skipRows, rowSize: value },
        {
          onSuccess: (data) => setPreviewData(data),
          onError: () => setPreviewData(null),
        }
      );
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedFile || !selectedAccountId) {
      setError("Please select a file and an account");
      return;
    }
    uploadStatementMutation.mutate(
      { account_id: selectedAccountId, file: selectedFile },
      { onSuccess: () => handleCancel() }
    );
  };

  const handleCancel = () => {
    setSelectedFile(null);
    setSelectedAccountId(accounts[0]?.id || 0);
    setError("");
    setStep(ImportStep.SelectBank);
    setPreviewData(null);
    setSkipRows(0);
    setRowSize(PREVIEW_SIZE);
    onOpenChange(false);
  };

  const removeFile = () => {
    setSelectedFile(null);
    setError("");
    setPreviewData(null);
  };

  const handleProcessStatement = (mappings: Record<string, string>) => {
    if (!selectedFile) {
      setError("Something went wrong, no file selected.");
      return;
    }

    const metadata = {
      skipRows: skipRows,
      columnMapping: mappings,
    };

    uploadStatementMutation.mutate(
      {
        account_id: selectedAccountId,
        file: selectedFile,
        bank_type: "others",
        metadata: JSON.stringify(metadata),
      },
      {
        onSuccess: () => {
          handleCancel();
        },
      }
    );
  };

  const renderStep = () => {
    switch (step) {
      case ImportStep.ImportFromBank:
        return (
          <ImportFromBank
            accounts={accounts}
            selectedAccountId={selectedAccountId}
            onSelectedAccountIdChange={setSelectedAccountId}
            selectedFile={selectedFile}
            onFileInputChange={handleFileInputChange}
            onFileRemove={removeFile}
            error={error}
            dragActive={dragActive}
            handleDrag={handleDrag}
            handleDrop={handleDrop}
            handleSubmit={handleSubmit}
            onStepChange={setStep}
            uploadStatementMutation={uploadStatementMutation}
          />
        );
      case ImportStep.Preview:
        return (
          <FallbackParsing
            accounts={accounts}
            selectedAccountId={selectedAccountId}
            onSelectedAccountIdChange={setSelectedAccountId}
            selectedFile={selectedFile}
            onFileInputChange={handleFileInputChange}
            onFileRemove={removeFile}
            error={error}
            dragActive={dragActive}
            handleDrag={handleDrag}
            handleDrop={handleDrop}
            skipRows={skipRows}
            onSkipRowsChange={handleSkipRowsChange}
            rowSize={rowSize}
            onRowSizeChange={handleRowSizeChange}
            previewData={previewData}
            previewStatementMutation={previewStatementMutation}
            onStepChange={setStep}
          />
        );
      case ImportStep.MapColumns:
        return (
          <MapColumns
            headers={previewData?.headers || []}
            onStepChange={setStep}
            onCancel={handleCancel}
            onSubmit={handleProcessStatement}
          />
        );
      default:
        return <SelectBank onStepChange={setStep} />;
    }
  };

  const getTitle = () => {
    switch (step) {
      case ImportStep.ImportFromBank:
        return "Import from Bank";
      case ImportStep.Preview:
        return "Preview";
      case ImportStep.MapColumns:
        return "Map Columns";
      default:
        return "Select Import Method";
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleCancel}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{getTitle()}</DialogTitle>
        </DialogHeader>
        {renderStep()}
      </DialogContent>
    </Dialog>
  );
}
