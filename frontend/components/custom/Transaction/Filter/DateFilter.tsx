import { TransactionFiltersState } from "@/app/transaction/page";
import { Button } from "@/components/ui/button";
import { Calendar as DatePicker } from "@/components/ui/calendar";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Calendar } from "lucide-react";
import React from "react";

interface DateFilterProps {
  filters: TransactionFiltersState;
  setFilters: React.Dispatch<React.SetStateAction<TransactionFiltersState>>;
}

export const DateFilter: React.FC<DateFilterProps> = ({
  filters,
  setFilters,
}) => {
  return (
    <div className="flex flex-col gap-y-2 w-full">
      <div>
        <label className="block text-[15px] mb-1">Date From</label>
        <Popover>
          <PopoverTrigger asChild>
            <Button
              variant="outline"
              className="w-40 justify-between font-normal"
            >
              {filters.dateFrom
                ? new Date(filters.dateFrom).toLocaleDateString()
                : "Select date"}
              <Calendar className="w-4 h-4 ml-2 opacity-50" />
            </Button>
          </PopoverTrigger>
          <PopoverContent align="start" className="p-0">
            <DatePicker
              mode="single"
              selected={
                filters.dateFrom ? new Date(filters.dateFrom) : undefined
              }
              onSelect={(date) => {
                setFilters({
                  ...filters,
                  dateFrom: date ? date.toISOString().slice(0, 10) : undefined,
                });
              }}
            />
          </PopoverContent>
        </Popover>
      </div>
      <div>
        <label className="block text-[15px] mb-1">Date To</label>
        <Popover>
          <PopoverTrigger asChild>
            <Button
              variant="outline"
              className="w-40 justify-between font-normal"
            >
              {filters.dateTo
                ? new Date(filters.dateTo).toLocaleDateString()
                : "Select date"}
              <Calendar className="w-4 h-4 ml-2 opacity-50" />
            </Button>
          </PopoverTrigger>
          <PopoverContent align="start" className="p-0">
            <DatePicker
              mode="single"
              selected={filters.dateTo ? new Date(filters.dateTo) : undefined}
              onSelect={(date) => {
                setFilters({
                  ...filters,
                  dateTo: date ? date.toISOString().slice(0, 10) : undefined,
                });
              }}
            />
          </PopoverContent>
        </Popover>
      </div>
    </div>
  );
};
