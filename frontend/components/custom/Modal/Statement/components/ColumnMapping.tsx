import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Label } from "@/components/ui/label";
import { AlertCircle, ChevronDownIcon } from "lucide-react";
import { ColumnMapping as ColumnMappingType, CSVPreview } from "@/lib/models/statement";

interface ColumnMappingProps {
  csvPreview: CSVPreview;
  columnMappings: ColumnMappingType[];
  onColumnMappingChange: (mappings: ColumnMappingType[]) => void;
}

export function ColumnMapping({
  csvPreview,
  columnMappings,
  onColumnMappingChange,
}: ColumnMappingProps) {
  const handleMappingChange = (column: string, targetField: string | null) => {
    const filtered = columnMappings.filter(m => m.source_column !== column);
    if (targetField) {
      const newMappings = [...filtered, { 
        source_column: column, 
        target_field: targetField as 'name' | 'amount' | 'description' | 'date' | 'credit' | 'debit' 
      }];
      onColumnMappingChange(newMappings);
    } else {
      onColumnMappingChange(filtered);
    }
  };

  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <Label>Column Mapping</Label>
        <p className="text-sm text-muted-foreground">
          Map your CSV columns to transaction fields
        </p>
      </div>

      <div className="space-y-3 max-h-64 overflow-y-auto">
        {csvPreview.columns.map((column, index) => (
          <div key={index} className="flex items-center space-x-3">
            <div className="flex-1">
              <Label className="text-sm font-medium">{column}</Label>
            </div>
            <div className="flex-1">
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="outline" className="w-full justify-between">
                    {columnMappings.find(m => m.source_column === column)?.target_field || 'Select field'}
                    <ChevronDownIcon className="h-4 w-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent>
                  <DropdownMenuItem onClick={() => handleMappingChange(column, null)}>
                    None
                  </DropdownMenuItem>
                  {['name', 'amount', 'description', 'date', 'credit', 'debit'].map(field => (
                    <DropdownMenuItem
                      key={field}
                      onClick={() => handleMappingChange(column, field)}
                    >
                      {field.charAt(0).toUpperCase() + field.slice(1)}
                      {['name', 'date'].includes(field) && <span className="text-red-500 ml-1">*</span>}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </div>
        ))}
      </div>

      <div className="bg-blue-50 dark:bg-blue-950/50 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
        <div className="flex items-start space-x-3">
          <AlertCircle className="h-5 w-5 text-blue-500 dark:text-blue-400 mt-0.5" />
          <div className="text-sm text-blue-700 dark:text-blue-300">
            <p className="font-medium mb-1">Required Fields:</p>
            <ul className="text-xs space-y-1 text-blue-600 dark:text-blue-400">
              <li>• Name and Date are required</li>
              <li>• Either Amount OR both Credit and Debit must be mapped</li>
              <li>• Description is optional</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
}