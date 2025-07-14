import { useAccounts } from "@/components/hooks/useAccounts";
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
import { Checkbox } from "@/components/ui/checkbox";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { 
  Upload, 
  FileText, 
  AlertCircle, 
  ChevronDownIcon, 
  X, 
  ArrowRight,
  CheckCircle
} from "lucide-react";
import { useCallback, useState } from "react";
import { toast } from "sonner";
import { 
  previewStatement, 
  parseStatementDirect,
  type ColumnMapping,
  type PreviewStatementResponse,
  type ParsedStatementResult
} from "@/lib/api/custom-parser";

interface CustomParserModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
}

type Step = "upload" | "preview" | "mapping" | "result";

export function CustomParserModal({
  isOpen,
  onOpenChange,
}: CustomParserModalProps) {
  const { data: accounts = [] } = useAccounts();
  
  const [currentStep, setCurrentStep] = useState<Step>("upload");
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [selectedAccountId, setSelectedAccountId] = useState<number>(accounts[0]?.id || 0);
  const [dragActive, setDragActive] = useState(false);
  const [error, setError] = useState<string>("");
  const [isLoading, setIsLoading] = useState(false);
  
  // Preview data
  const [previewData, setPreviewData] = useState<PreviewStatementResponse | null>(null);
  const [hasHeaders, setHasHeaders] = useState(true);
  
  // Column mapping
  const [columnMapping, setColumnMapping] = useState<ColumnMapping>({
    date_column: -1,
    description_column: -1,
    amount_column: -1,
    reference_column: -1,
  });
  
  // Parse result
  const [parseResult, setParseResult] = useState<ParsedStatementResult | null>(null);

  const validateFile = (file: File): string | null => {
    if (file.size > 256 * 1024) {
      return "File size must be less than 256KB";
    }

    const validTypes = [
      'text/csv',
      'application/vnd.ms-excel',
      'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
    ];
    const validExtensions = ['.csv', '.xls', '.xlsx'];
    
    const isValidType = validTypes.includes(file.type) || 
                       validExtensions.some(ext => file.name.toLowerCase().endsWith(ext));
    
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

  const handlePreview = async () => {
    if (!selectedFile || !selectedAccountId) {
      setError("Please select a file and account");
      return;
    }

    setIsLoading(true);
    setError("");

    try {
      const response = await previewStatement({
        account_id: selectedAccountId,
        file: selectedFile,
      });

      setPreviewData(response);
      setCurrentStep("preview");
    } catch (error) {
      setError(error instanceof Error ? error.message : "Failed to preview statement");
    } finally {
      setIsLoading(false);
    }
  };

  const handleSetupMapping = () => {
    if (!previewData) return;
    
    // Auto-detect common column mappings
    const headers = previewData.preview.headers.map(h => h.toLowerCase());
    
    const dateIndex = headers.findIndex(h => 
      h.includes('date') || h.includes('time') || h.includes('transaction')
    );
    const descIndex = headers.findIndex(h => 
      h.includes('description') || h.includes('narration') || h.includes('details')
    );
    const amountIndex = headers.findIndex(h => 
      h.includes('amount') || h.includes('debit') || h.includes('credit')
    );
    const refIndex = headers.findIndex(h => 
      h.includes('reference') || h.includes('ref') || h.includes('id')
    );

    setColumnMapping({
      date_column: dateIndex >= 0 ? dateIndex : -1,
      description_column: descIndex >= 0 ? descIndex : -1,
      amount_column: amountIndex >= 0 ? amountIndex : -1,
      reference_column: refIndex >= 0 ? refIndex : -1,
    });

    setCurrentStep("mapping");
  };

  const handleParse = async () => {
    if (!selectedFile || !selectedAccountId || !previewData) {
      setError("Missing required data");
      return;
    }

    // Validate mapping
    if (columnMapping.date_column === -1 || columnMapping.amount_column === -1) {
      setError("Date and Amount columns are required");
      return;
    }

    setIsLoading(true);
    setError("");

    try {
      const response = await parseStatementDirect({
        account_id: selectedAccountId,
        file: selectedFile,
        has_headers: hasHeaders,
        column_mapping: columnMapping,
      });

      setParseResult(response.parse_result);
      setCurrentStep("result");
      
      if (response.parse_result.successful_rows > 0) {
        toast.success(`Successfully parsed ${response.parse_result.successful_rows} transactions!`);
      }
    } catch (error) {
      setError(error instanceof Error ? error.message : "Failed to parse statement");
    } finally {
      setIsLoading(false);
    }
  };

  const handleCancel = () => {
    setCurrentStep("upload");
    setSelectedFile(null);
    setPreviewData(null);
    setParseResult(null);
    setColumnMapping({
      date_column: -1,
      description_column: -1,
      amount_column: -1,
      reference_column: -1,
    });
    setError("");
    onOpenChange(false);
  };

  const handleComplete = () => {
    handleCancel();
    // Refresh statements list
    window.location.reload(); // Simple refresh for now
  };

  const renderUploadStep = () => (
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
                const selected = accounts.find(acc => acc.id === selectedAccountId);
                return selected ? `${selected.name} (${selected.bank_type.toUpperCase()})` : "Select account";
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
                  selectedAccountId === account.id ? "bg-accent/40 font-semibold" : ""
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
              onClick={() => document.getElementById('custom-file-input')?.click()}
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
                id="custom-file-input"
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
                    <p className="text-sm font-medium text-foreground">{selectedFile.name}</p>
                    <p className="text-xs text-muted-foreground">
                      {(selectedFile.size / 1024).toFixed(1)} KB
                    </p>
                  </div>
                </div>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setSelectedFile(null)}
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
    </div>
  );

  const renderPreviewStep = () => {
    if (!previewData) return null;

    return (
      <div className="space-y-4 py-4">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-medium">File Preview</h3>
            <p className="text-sm text-muted-foreground">
              {previewData.filename} • {previewData.preview.total_rows} rows
            </p>
          </div>
          <div className="flex items-center space-x-2">
            <Checkbox
              id="has-headers"
              checked={hasHeaders}
              onCheckedChange={(checked) => setHasHeaders(checked as boolean)}
            />
            <Label htmlFor="has-headers" className="text-sm">
              First row contains headers
            </Label>
          </div>
        </div>

        <div className="border rounded-lg overflow-hidden">
          <div className="max-h-64 overflow-auto">
            <Table>
              <TableHeader className="sticky top-0 bg-background">
                <TableRow>
                  {previewData.preview.headers.map((header, index) => (
                    <TableHead key={index} className="min-w-[120px]">
                      {hasHeaders ? header : `Column ${index + 1}`}
                    </TableHead>
                  ))}
                </TableRow>
              </TableHeader>
              <TableBody>
                {previewData.preview.sample_data.map((row, rowIndex) => (
                  <TableRow key={rowIndex}>
                    {row.map((cell, cellIndex) => (
                      <TableCell key={cellIndex} className="max-w-[200px] truncate">
                        {cell}
                      </TableCell>
                    ))}
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </div>
      </div>
    );
  };

  const renderMappingStep = () => {
    if (!previewData) return null;

    const headers = previewData.preview.headers;

    return (
      <div className="space-y-4 py-4">
        <div>
          <h3 className="text-lg font-medium">Column Mapping</h3>
          <p className="text-sm text-muted-foreground">
            Map your file columns to our transaction fields
          </p>
        </div>

        <div className="grid grid-cols-2 gap-4">
          {/* Date Column */}
          <div className="space-y-2">
            <Label className="text-sm font-medium">
              Date Column <span className="text-red-500">*</span>
            </Label>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" className="w-full justify-start">
                  {columnMapping.date_column >= 0 
                    ? headers[columnMapping.date_column] 
                    : "Select column"}
                  <ChevronDownIcon className="ml-auto w-4 h-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-56">
                {headers.map((header, index) => (
                  <DropdownMenuItem
                    key={index}
                    onClick={() => setColumnMapping(prev => ({ ...prev, date_column: index }))}
                  >
                    {header}
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>

          {/* Amount Column */}
          <div className="space-y-2">
            <Label className="text-sm font-medium">
              Amount Column <span className="text-red-500">*</span>
            </Label>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" className="w-full justify-start">
                  {columnMapping.amount_column >= 0 
                    ? headers[columnMapping.amount_column] 
                    : "Select column"}
                  <ChevronDownIcon className="ml-auto w-4 h-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-56">
                {headers.map((header, index) => (
                  <DropdownMenuItem
                    key={index}
                    onClick={() => setColumnMapping(prev => ({ ...prev, amount_column: index }))}
                  >
                    {header}
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>

          {/* Description Column */}
          <div className="space-y-2">
            <Label className="text-sm font-medium">Description Column</Label>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" className="w-full justify-start">
                  {columnMapping.description_column >= 0 
                    ? headers[columnMapping.description_column] 
                    : "Select column (optional)"}
                  <ChevronDownIcon className="ml-auto w-4 h-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-56">
                <DropdownMenuItem onClick={() => setColumnMapping(prev => ({ ...prev, description_column: -1 }))}>
                  None
                </DropdownMenuItem>
                {headers.map((header, index) => (
                  <DropdownMenuItem
                    key={index}
                    onClick={() => setColumnMapping(prev => ({ ...prev, description_column: index }))}
                  >
                    {header}
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>

          {/* Reference Column */}
          <div className="space-y-2">
            <Label className="text-sm font-medium">Reference Column</Label>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" className="w-full justify-start">
                  {columnMapping.reference_column >= 0 
                    ? headers[columnMapping.reference_column] 
                    : "Select column (optional)"}
                  <ChevronDownIcon className="ml-auto w-4 h-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-56">
                <DropdownMenuItem onClick={() => setColumnMapping(prev => ({ ...prev, reference_column: -1 }))}>
                  None
                </DropdownMenuItem>
                {headers.map((header, index) => (
                  <DropdownMenuItem
                    key={index}
                    onClick={() => setColumnMapping(prev => ({ ...prev, reference_column: index }))}
                  >
                    {header}
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>

        {error && (
          <div className="text-sm text-destructive flex items-center space-x-2">
            <AlertCircle className="h-4 w-4" />
            <span>{error}</span>
          </div>
        )}
      </div>
    );
  };

  const renderResultStep = () => {
    if (!parseResult) return null;

    return (
      <div className="space-y-4 py-4">
        <div>
          <h3 className="text-lg font-medium">Parse Results</h3>
          <p className="text-sm text-muted-foreground">
            Statement processing completed
          </p>
        </div>

        {/* Summary */}
        <div className="grid grid-cols-3 gap-4">
          <div className="text-center p-4 border rounded-lg">
            <div className="text-2xl font-bold text-green-600">{parseResult.successful_rows}</div>
            <div className="text-sm text-muted-foreground">Successful</div>
          </div>
          <div className="text-center p-4 border rounded-lg">
            <div className="text-2xl font-bold text-red-600">{parseResult.failed_rows}</div>
            <div className="text-sm text-muted-foreground">Failed</div>
          </div>
          <div className="text-center p-4 border rounded-lg">
            <div className="text-2xl font-bold">{parseResult.total_rows}</div>
            <div className="text-sm text-muted-foreground">Total</div>
          </div>
        </div>

        {/* Errors */}
        {parseResult.errors && parseResult.errors.length > 0 && (
          <div className="space-y-2">
            <Label className="text-sm font-medium text-red-600">Errors:</Label>
            <div className="max-h-32 overflow-y-auto space-y-1">
              {parseResult.errors.map((error, index) => (
                <div key={index} className="text-xs text-red-600 bg-red-50 dark:bg-red-950/20 p-2 rounded">
                  {error}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Success message */}
        {parseResult.successful_rows > 0 && (
          <div className="bg-green-50 dark:bg-green-950/20 border border-green-200 dark:border-green-800 rounded-lg p-4">
            <div className="flex items-center space-x-2">
              <CheckCircle className="h-5 w-5 text-green-500" />
              <div className="text-sm text-green-700 dark:text-green-300">
                <p className="font-medium">Statement processed successfully!</p>
                <p>{parseResult.successful_rows} transactions have been created and are now available in your transaction list.</p>
              </div>
            </div>
          </div>
        )}
      </div>
    );
  };

  const getStepTitle = () => {
    switch (currentStep) {
      case "upload": return "Upload Statement";
      case "preview": return "Preview Data";
      case "mapping": return "Map Columns";
      case "result": return "Parse Results";
      default: return "Custom Parser";
    }
  };

  const getNextButtonText = () => {
    switch (currentStep) {
      case "upload": return "Preview";
      case "preview": return "Setup Mapping";
      case "mapping": return "Parse Statement";
      case "result": return "Complete";
      default: return "Next";
    }
  };

  const handleNext = () => {
    switch (currentStep) {
      case "upload": handlePreview(); break;
      case "preview": handleSetupMapping(); break;
      case "mapping": handleParse(); break;
      case "result": handleComplete(); break;
    }
  };

  const canProceed = () => {
    switch (currentStep) {
      case "upload": return selectedFile && selectedAccountId;
      case "preview": return previewData;
      case "mapping": return columnMapping.date_column >= 0 && columnMapping.amount_column >= 0;
      case "result": return true;
      default: return false;
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[800px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{getStepTitle()}</DialogTitle>
        </DialogHeader>

        {/* Step indicator */}
        <div className="flex items-center space-x-2 mb-4">
          {["upload", "preview", "mapping", "result"].map((step, index) => (
            <div key={step} className="flex items-center">
              <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium ${
                currentStep === step 
                  ? "bg-primary text-primary-foreground" 
                  : index < ["upload", "preview", "mapping", "result"].indexOf(currentStep)
                    ? "bg-green-500 text-white"
                    : "bg-muted text-muted-foreground"
              }`}>
                {index < ["upload", "preview", "mapping", "result"].indexOf(currentStep) ? (
                  <CheckCircle className="w-4 h-4" />
                ) : (
                  index + 1
                )}
              </div>
              {index < 3 && (
                <ArrowRight className="w-4 h-4 mx-2 text-muted-foreground" />
              )}
            </div>
          ))}
        </div>

        {/* Step content */}
        <div className="min-h-[300px]">
          {currentStep === "upload" && renderUploadStep()}
          {currentStep === "preview" && renderPreviewStep()}
          {currentStep === "mapping" && renderMappingStep()}
          {currentStep === "result" && renderResultStep()}
        </div>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={handleCancel}
            disabled={isLoading}
          >
            Cancel
          </Button>
          <Button
            type="button"
            onClick={handleNext}
            disabled={!canProceed() || isLoading}
          >
            {isLoading ? "Processing..." : getNextButtonText()}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
