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
import { useCallback, useRef, useState } from "react";

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
  const [selectedFiles, setSelectedFiles] = useState<File[]>([]);
  const [currentFileIndex, setCurrentFileIndex] = useState(0);
  const [skipRows, setSkipRows] = useState(0);
  const [rowSize, setRowSize] = useState(PREVIEW_SIZE);
  const [previewData, setPreviewData] =
    useState<StatementPreviewResponse | null>(null);
  const [selectedAccountId, setSelectedAccountId] = useState<number>(
    accounts[0]?.id || 0
  );
  const [dragActive, setDragActive] = useState(false);
  const [error, setError] = useState<string>("");
  const [uploadProgress, setUploadProgress] = useState<{
    current: number;
    total: number;
    processing: boolean;
  }>({ current: 0, total: 0, processing: false });

  // Refs for file inputs
  const fileInputRef = useRef<HTMLInputElement>(null);
  const additionalFileInputRef = useRef<HTMLInputElement>(null);
  const fallbackFileInputRef = useRef<HTMLInputElement>(null);

  // Helper function to reset file inputs
  const resetFileInputs = useCallback(() => {
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
    if (additionalFileInputRef.current) {
      additionalFileInputRef.current.value = "";
    }
    if (fallbackFileInputRef.current) {
      fallbackFileInputRef.current.value = "";
    }
  }, []);

  const validateFile = (file: File, forBank: boolean): string | null => {
    if (file.size > 256 * 1024) {
      return "File size must be less than 256KB";
    }
    const validExtensions = forBank
      ? [".csv", ".xls", ".xlsx", ".txt"]
      : [".csv", ".xls", ".txt"];
    if (!validExtensions.some((ext) => file.name.toLowerCase().endsWith(ext))) {
      return `File must be ${validExtensions.join(", ")} format`;
    }
    return null;
  };

  const validateFiles = useCallback(
    (files: File[], forBank: boolean): string | null => {
      if (files.length === 0) return "Please select at least one file";

      // For custom parser (non-bank), only allow single file
      if (!forBank && files.length > 1) {
        return "Custom parser only supports single file upload";
      }

      // For bank parsing, allow up to 10 files
      if (forBank && files.length > 10) {
        return "Maximum 10 files allowed";
      }

      for (const file of files) {
        const error = validateFile(file, forBank);
        if (error) return `${file.name}: ${error}`;
      }
      return null;
    },
    []
  );

  const handleFilesSelect = useCallback(
    (files: File[]) => {
      const validationError = validateFiles(
        files,
        step === ImportStep.ImportFromBank
      );
      if (validationError) {
        setError(validationError);
        setSelectedFiles([]);
        setPreviewData(null);
        return;
      }

      setError("");
      setSelectedFiles(files);
      setCurrentFileIndex(0);

      if (step === ImportStep.Preview && files.length > 0) {
        previewStatementMutation.mutate(
          { file: files[0], skipRows, rowSize },
          {
            onSuccess: (data) => setPreviewData(data),
            onError: () => setPreviewData(null),
          }
        );
      }
    },
    [step, skipRows, rowSize, previewStatementMutation, validateFiles]
  );

  const handleFileInputChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const files = Array.from(e.target.files || []);
      if (files.length > 0) {
        // For custom parser (Preview step), only take the first file
        if (step === ImportStep.Preview) {
          handleFilesSelect([files[0]]);
        } else {
          handleFilesSelect(files);
        }
      }
      resetFileInputs();
    },
    [step, handleFilesSelect, resetFileInputs]
  );

  const handleAdditionalFilesChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const newFiles = Array.from(e.target.files || []);
      if (newFiles.length > 0) {
        // Filter out duplicates based on file name and size
        const existingFileKeys = selectedFiles.map(
          (f) => `${f.name}-${f.size}`
        );
        const uniqueNewFiles = newFiles.filter(
          (f) => !existingFileKeys.includes(`${f.name}-${f.size}`)
        );

        if (uniqueNewFiles.length === 0) {
          setError("All selected files are already added");
          resetFileInputs();
          return;
        }

        const combinedFiles = [...selectedFiles, ...uniqueNewFiles];

        // Validate the combined files
        const validationError = validateFiles(
          combinedFiles,
          step === ImportStep.ImportFromBank
        );
        if (validationError) {
          setError(validationError);
          resetFileInputs();
          return;
        }

        setError("");
        setSelectedFiles(combinedFiles);
      }
      resetFileInputs();
    },
    [selectedFiles, validateFiles, step, resetFileInputs]
  );

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
      const files = Array.from(e.dataTransfer.files || []);
      if (files.length > 0) {
        // For custom parser (Preview step), only take the first file
        if (step === ImportStep.Preview) {
          handleFilesSelect([files[0]]);
        } else {
          handleFilesSelect(files);
        }
      }
    },
    [handleFilesSelect, step]
  );

  const handleSkipRowsChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = parseInt(e.target.value, 10);
    setSkipRows(value);
    const currentFile = selectedFiles[currentFileIndex];
    if (currentFile) {
      previewStatementMutation.mutate(
        { file: currentFile, skipRows: value, rowSize },
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
    const currentFile = selectedFiles[currentFileIndex];
    if (currentFile) {
      previewStatementMutation.mutate(
        { file: currentFile, skipRows, rowSize: value },
        {
          onSuccess: (data) => setPreviewData(data),
          onError: () => setPreviewData(null),
        }
      );
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (selectedFiles.length === 0 || !selectedAccountId) {
      setError("Please select at least one file and an account");
      return;
    }

    setUploadProgress({
      current: 0,
      total: selectedFiles.length,
      processing: true,
    });

    for (let i = 0; i < selectedFiles.length; i++) {
      try {
        await new Promise<void>((resolve, reject) => {
          uploadStatementMutation.mutate(
            { account_id: selectedAccountId, file: selectedFiles[i] },
            {
              onSuccess: () => {
                setUploadProgress((prev) => ({ ...prev, current: i + 1 }));
                resolve();
              },
              onError: (err) => {
                setError(
                  `Failed to upload ${selectedFiles[i].name}: ${err.message}`
                );
                reject(err);
              },
            }
          );
        });
      } catch {
        setUploadProgress({ current: 0, total: 0, processing: false });
        return;
      }
    }

    setUploadProgress({ current: 0, total: 0, processing: false });
    handleCancel();
  };

  const handleCancel = () => {
    setSelectedFiles([]);
    setCurrentFileIndex(0);
    setSelectedAccountId(accounts[0]?.id || 0);
    setError("");
    setStep(ImportStep.SelectBank);
    setPreviewData(null);
    setSkipRows(0);
    setRowSize(PREVIEW_SIZE);
    setUploadProgress({ current: 0, total: 0, processing: false });
    onOpenChange(false);
  };

  const removeFile = (index?: number) => {
    if (index !== undefined) {
      const newFiles = selectedFiles.filter((_, i) => i !== index);
      setSelectedFiles(newFiles);
      if (currentFileIndex >= newFiles.length) {
        setCurrentFileIndex(Math.max(0, newFiles.length - 1));
      }
      if (newFiles.length === 0) {
        setPreviewData(null);
      } else if (index === currentFileIndex && step === ImportStep.Preview) {
        const newCurrentFile =
          newFiles[Math.min(currentFileIndex, newFiles.length - 1)];
        previewStatementMutation.mutate(
          { file: newCurrentFile, skipRows, rowSize },
          {
            onSuccess: (data) => setPreviewData(data),
            onError: () => setPreviewData(null),
          }
        );
      }
    } else {
      setSelectedFiles([]);
      setCurrentFileIndex(0);
      setPreviewData(null);
    }
    setError("");
  };

  const handleProcessStatement = async (mappings: Record<string, string>) => {
    if (selectedFiles.length === 0) {
      setError("Something went wrong, no files selected.");
      return;
    }

    const metadata = {
      skipRows: skipRows,
      columnMapping: mappings,
    };

    setUploadProgress({
      current: 0,
      total: selectedFiles.length,
      processing: true,
    });

    for (let i = 0; i < selectedFiles.length; i++) {
      try {
        await new Promise<void>((resolve, reject) => {
          uploadStatementMutation.mutate(
            {
              account_id: selectedAccountId,
              file: selectedFiles[i],
              bank_type: "others",
              metadata: JSON.stringify(metadata),
            },
            {
              onSuccess: () => {
                setUploadProgress((prev) => ({ ...prev, current: i + 1 }));
                resolve();
              },
              onError: (err) => {
                setError(
                  `Failed to process ${selectedFiles[i].name}: ${err.message}`
                );
                reject(err);
              },
            }
          );
        });
      } catch {
        setUploadProgress({ current: 0, total: 0, processing: false });
        return;
      }
    }

    setUploadProgress({ current: 0, total: 0, processing: false });
    handleCancel();
  };

  const renderStep = () => {
    switch (step) {
      case ImportStep.ImportFromBank:
        return (
          <ImportFromBank
            accounts={accounts}
            selectedAccountId={selectedAccountId}
            onSelectedAccountIdChange={setSelectedAccountId}
            selectedFiles={selectedFiles}
            onFileInputChange={handleFileInputChange}
            onAdditionalFilesChange={handleAdditionalFilesChange}
            onFileRemove={removeFile}
            error={error}
            dragActive={dragActive}
            handleDrag={handleDrag}
            handleDrop={handleDrop}
            handleSubmit={handleSubmit}
            onStepChange={setStep}
            uploadStatementMutation={uploadStatementMutation}
            uploadProgress={uploadProgress}
            fileInputRef={fileInputRef}
            additionalFileInputRef={additionalFileInputRef}
          />
        );
      case ImportStep.Preview:
        return (
          <FallbackParsing
            accounts={accounts}
            selectedAccountId={selectedAccountId}
            onSelectedAccountIdChange={setSelectedAccountId}
            selectedFile={selectedFiles[0] || null}
            onFileInputChange={handleFileInputChange}
            onFileRemove={() => removeFile(0)}
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
            fileInputRef={fallbackFileInputRef}
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
