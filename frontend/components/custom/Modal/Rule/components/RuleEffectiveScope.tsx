import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { cn } from "@/lib/utils";
import { format } from "date-fns";
import { CalendarIcon, ChevronDownIcon } from "lucide-react";

interface RuleEffectiveScopeProps {
  effectiveScope: "all" | "from";
  effectiveFromDate: Date | undefined;
  onEffectiveScopeChange: (scope: "all" | "from") => void;
  onEffectiveFromDateChange: (date: Date | undefined) => void;
  disabled?: boolean;
}

export function RuleEffectiveScope({
  effectiveScope,
  effectiveFromDate,
  onEffectiveScopeChange,
  onEffectiveFromDateChange,
  disabled = false,
}: RuleEffectiveScopeProps) {
  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold">FOR</h3>
      <div className="flex flex-col w-full gap-2">
        <div className="flex w-full gap-2">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="outline"
                className="flex-1 justify-between pr-4"
                disabled={disabled}
              >
                {effectiveScope === "all"
                  ? "All past and future transactions"
                  : effectiveFromDate
                    ? `Starting from ${format(effectiveFromDate, "dd/MM/yyyy")}`
                    : "Starting from (choose date)"}
                <ChevronDownIcon className="ml-2 h-4 w-4 opacity-50" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              align="start"
              className="w-full min-w-[260px]"
            >
              <DropdownMenuItem
                onSelect={() => onEffectiveScopeChange("all")}
                className={cn(
                  "cursor-pointer",
                  effectiveScope === "all" && "font-semibold"
                )}
                disabled={disabled}
              >
                All past and future transactions
              </DropdownMenuItem>
              <DropdownMenuItem
                onSelect={() => onEffectiveScopeChange("from")}
                className={cn(
                  "cursor-pointer",
                  effectiveScope === "from" && "font-semibold"
                )}
                disabled={disabled}
              >
                Starting from (choose date)
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
          {effectiveScope === "from" && (
            <Popover>
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  className={cn(
                    "flex-1 justify-start text-left font-normal pr-4",
                    !effectiveFromDate && "text-muted-foreground"
                  )}
                  disabled={disabled}
                >
                  <CalendarIcon className="mr-2 h-4 w-4" />
                  {effectiveFromDate ? (
                    format(effectiveFromDate, "dd/MM/yyyy")
                  ) : (
                    <span>Select date</span>
                  )}
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-auto p-0" align="start">
                <Calendar
                  mode="single"
                  selected={effectiveFromDate}
                  onSelect={onEffectiveFromDateChange}
                  initialFocus
                />
              </PopoverContent>
            </Popover>
          )}
        </div>
      </div>
    </div>
  );
}
