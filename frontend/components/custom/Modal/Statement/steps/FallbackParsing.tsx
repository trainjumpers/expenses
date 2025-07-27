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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { Account } from "@/lib/models/account";
import { StatementPreviewResponse } from "@/lib/models/statement";
import { UseMutationResult } from "@tanstack/react-query";
import {
  AlertCircle,
  ChevronDownIcon,
  FileText,
  Upload,
  X,
} from "lucide-react";

interface FallbackParsingProps {
  accounts: Account[];
  selectedAccountId: number;
  onSelectedAccountIdChange: (id: number) => void;
  selectedFile: File | null;
  onFileInputChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onFileRemove: () => void;
  error: string;
  dragActive: boolean;
  handleDrag: (e: React.DragEvent) => void;
  handleDrop: (e: React.DragEvent) => void;
  skipRows: number;
  onSkipRowsChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  rowSize: number;
  onRowSizeChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  previewData: StatementPreviewResponse | null;
  previewStatementMutation: UseMutationResult<
    StatementPreviewResponse,
    Error,
    { file: File; skipRows: number; rowSize: number }
  >;
  onStepChange: (step: number) => void;
}

export function FallbackParsing({
  accounts,
  selectedAccountId,
  onSelectedAccountIdChange,
  selectedFile,
  onFileInputChange,
  onFileRemove,
  error,
  dragActive,
  handleDrag,
  handleDrop,
  skipRows,
  onSkipRowsChange,
  rowSize,
  onRowSizeChange,
  previewData,
  previewStatementMutation,
  onStepChange,
}: FallbackParsingProps) {
  return (
    <div className="space-y-4 py-4">
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

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <div className="flex items-center space-x-2">
            <Label htmlFor="skip-rows">Skip Rows</Label>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <AlertCircle className="h-4 w-4 text-muted-foreground" />
                </TooltipTrigger>
                <TooltipContent>
                  <p>Number of rows to skip from the top of the file.</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
          <Input
            id="skip-rows"
            type="number"
            value={skipRows}
            onChange={onSkipRowsChange}
            min="0"
          />
        </div>
        <div className="space-y-2">
          <div className="flex items-center space-x-2">
            <Label htmlFor="row-size">Row Size</Label>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <AlertCircle className="h-4 w-4 text-muted-foreground" />
                </TooltipTrigger>
                <TooltipContent>
                  <p>Number of rows to display in the preview.</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
          <Input
            id="row-size"
            type="number"
            value={rowSize}
            onChange={onRowSizeChange}
            min="1"
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label>Statement File</Label>
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
              document.getElementById("file-input-fallback")?.click()
            }
          >
            <Upload className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-sm text-foreground mb-2">
              {dragActive
                ? "Drop the file here..."
                : "Drag & drop your bank statement here, or click to select"}
            </p>
            <p className="text-xs text-muted-foreground">
              Supports CSV, XLS files (max 256KB)
            </p>
            <Input
              id="file-input-fallback"
              type="file"
              accept=".csv,.xls"
              onChange={onFileInputChange}
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
                onClick={onFileRemove}
              >
                <X className="h-4 w-4" />
              </Button>
            </div>
          </div>
        )}
      </div>

      {(previewData || previewStatementMutation.isPending) && (
        <div className="w-110">
          <div className="relative max-h-64 overflow-auto border rounded-md">
            <Table>
              <TableHeader className="sticky top-0 bg-background">
                <TableRow className="bg-muted/50">
                  {previewData?.headers.map((header, i) => (
                    <TableHead key={i}>{header}</TableHead>
                  ))}
                </TableRow>
              </TableHeader>
              <TableBody>
                {previewStatementMutation.isPending ? (
                  <TableRow>
                    <TableCell
                      colSpan={previewData?.headers.length || 5}
                      className="text-center"
                    >
                      Loading preview...
                    </TableCell>
                  </TableRow>
                ) : (
                  previewData?.rows.map((row, i) => (
                    <TableRow key={i}>
                      {row.map((cell, j) => (
                        <TableCell key={j}>{cell}</TableCell>
                      ))}
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </div>
      )}

      {error && (
        <div className="text-sm text-destructive flex items-center space-x-2">
          <AlertCircle className="h-4 w-4" />
          <span>{error}</span>
        </div>
      )}

      <DialogFooter>
        <Button type="button" variant="outline" onClick={() => onStepChange(1)}>
          Back
        </Button>
        <Button
          type="button"
          disabled={!previewData || !selectedAccountId}
          onClick={() => onStepChange(4)}
        >
          Next
        </Button>
      </DialogFooter>
    </div>
  );
}
