import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { FileText, Upload, X } from "lucide-react";
import { useCallback } from "react";

interface FileUploadProps {
  selectedFile: File | null;
  dragActive: boolean;
  onFileSelect: (file: File) => void;
  onRemoveFile: () => void;
  onDragStateChange: (active: boolean) => void;
}

export function FileUpload({
  selectedFile,
  dragActive,
  onFileSelect,
  onRemoveFile,
  onDragStateChange,
}: FileUploadProps) {
  const handleDrag = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === "dragenter" || e.type === "dragover") {
      onDragStateChange(true);
    } else if (e.type === "dragleave") {
      onDragStateChange(false);
    }
  }, [onDragStateChange]);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    onDragStateChange(false);

    const file = e.dataTransfer.files?.[0];
    if (file) {
      onFileSelect(file);
    }
  }, [onFileSelect, onDragStateChange]);

  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      onFileSelect(file);
    }
  };

  return (
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
                : "Drag & drop your statement file here, or click to select"}
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
                onClick={onRemoveFile}
              >
                <X className="h-4 w-4" />
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}