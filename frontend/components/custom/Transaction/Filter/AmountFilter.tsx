import { TransactionFiltersState } from "@/app/transaction/page";
import { Input } from "@/components/ui/input";
import React from "react";

interface AmountFilterProps {
  filters: TransactionFiltersState;
  setFilters: React.Dispatch<React.SetStateAction<TransactionFiltersState>>;
}

export const AmountFilter: React.FC<AmountFilterProps> = ({
  filters,
  setFilters,
}) => {
  return (
    <div className="flex flex-col gap-y-2 w-full">
      <div>
        <label className="block text-[15px] mb-1">Min Amount</label>
        <Input
          type="number"
          value={filters.minAmount ?? ""}
          onChange={(e) =>
            setFilters({
              ...filters,
              minAmount: e.target.value ? Number(e.target.value) : undefined,
            })
          }
          className="w-20 h-8 px-2"
        />
      </div>
      <div>
        <label className="block text-[15px] mb-1">Max Amount</label>
        <Input
          type="number"
          value={filters.maxAmount ?? ""}
          onChange={(e) =>
            setFilters({
              ...filters,
              maxAmount: e.target.value ? Number(e.target.value) : undefined,
            })
          }
          className="w-20 h-8 px-2"
        />
      </div>
    </div>
  );
};
