import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { AlertCircle } from "lucide-react";
import { CSVPreview as CSVPreviewType } from "@/lib/models/statement";

interface CSVPreviewProps {
  csvPreview: CSVPreviewType;
  skipRows: number;
  isRefreshingPreview: boolean;
  onSkipRowsChange: (skipRows: number) => void;
}

export function CSVPreview({
  csvPreview,
  skipRows,
  isRefreshingPreview,
  onSkipRowsChange,
}: CSVPreviewProps) {
  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <Label>CSV Preview</Label>
        <p className="text-sm text-muted-foreground">
          Preview of your CSV file ({csvPreview.total} total rows)
        </p>
      </div>

      {/* Skip Rows Control */}
      <div className="space-y-2">
        <Label>Skip Rows</Label>
        <div className="flex items-center space-x-2">
          <Input
            type="number"
            min="0"
            max={Math.max(0, csvPreview.total - 2)}
            value={skipRows}
            onChange={(e) => onSkipRowsChange(parseInt(e.target.value) || 0)}
            className="w-20"
            disabled={isRefreshingPreview}
          />
          <span className="text-sm text-muted-foreground">
            rows from the top
          </span>
          {isRefreshingPreview && (
            <span className="text-xs text-muted-foreground">
              Refreshing preview...
            </span>
          )}
        </div>
      </div>

      {/* Preview Table */}
      <div className="border rounded-lg overflow-hidden" style={{width: "750px"}}>
        <div className="overflow-auto max-h-64">
          <table className="text-sm min-w-max whitespace-nowrap">
            <thead className="bg-muted">
              <tr>
                {csvPreview.columns.map((column, index) => (
                  <th
                    key={index}
                    className="px-3 py-2 text-left font-medium whitespace-nowrap"
                    style={{ width: '100px' }}
                  >
                    <div className="truncate" title={column}>
                      {column}
                    </div>
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {csvPreview.rows.map((row, rowIndex) => (
                <tr key={rowIndex} className="border-t border-border">
                  {row.map((cell, cellIndex) => (
                    <td
                      key={cellIndex}
                      className="px-3 py-2 border-r border-border last:border-r-0 whitespace-nowrap"
                      style={{ width: '100px' }}
                    >
                      <div className="truncate" title={cell || ''}>
                        {cell || ''}
                      </div>
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {csvPreview.columns.length > 4 && (
          <div className="px-3 py-2 bg-muted/50 text-xs text-muted-foreground text-center border-t border-border">
            ← Scroll horizontally to see all {csvPreview.columns.length} columns →
          </div>
        )}
      </div>

      <div className="bg-amber-50 dark:bg-amber-950/50 border border-amber-200 dark:border-amber-800 rounded-lg p-4">
        <div className="flex items-start space-x-3">
          <AlertCircle className="h-5 w-5 text-amber-500 dark:text-amber-400 mt-0.5" />
          <div className="text-sm text-amber-700 dark:text-amber-300">
            <p className="font-medium mb-1">Next Step:</p>
            <p className="text-xs text-amber-600 dark:text-amber-400">
              You&apos;ll be able to map these columns to transaction fields (Name, Amount, Date, etc.)
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}