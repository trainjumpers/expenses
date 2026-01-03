import type { TransactionFiltersState } from "@/app/transaction/page";
import { Button } from "@/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import type { Account } from "@/lib/models/account";
import type { Category } from "@/lib/models/category";
import {
  Calendar,
  Filter as FilterIcon,
  Hash,
  Layers,
  List,
  Search,
} from "lucide-react";
import React, { useState } from "react";

import { AccountFilter } from "./Filter/AccountFilter";
import { AmountFilter } from "./Filter/AmountFilter";
import { CategoryFilter } from "./Filter/CategoryFilter";
import { DateFilter } from "./Filter/DateFilter";

interface TransactionFiltersProps {
  accounts: Account[];
  categories: Category[];
  filters: TransactionFiltersState;
  onFilterChange: (newFilters: Partial<TransactionFiltersState>) => void;
  onClear: () => void;
}

type FilterType = "account" | "date" | "amount" | "category";

const FILTERS: { key: FilterType; label: string; icon: React.ReactNode }[] = [
  { key: "account", label: "Account", icon: <Layers className="w-4 h-4" /> },
  { key: "date", label: "Date", icon: <Calendar className="w-4 h-4" /> },
  { key: "amount", label: "Amount", icon: <Hash className="w-4 h-4" /> },
  { key: "category", label: "Category", icon: <List className="w-3 h-3" /> },
];

const TransactionFilters: React.FC<TransactionFiltersProps> = ({
  accounts,
  categories,
  filters,
  onFilterChange,
  onClear,
}) => {
  const [popoverOpen, setPopoverOpen] = useState(false);
  const [selectedFilter, setSelectedFilter] = useState<FilterType>("account");
  const [tempFilters, setTempFilters] =
    useState<TransactionFiltersState>(filters);

  React.useEffect(() => {
    if (popoverOpen) {
      setTempFilters(filters);
    }
  }, [popoverOpen, filters]);

  const handleSearchSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    onFilterChange({ search: tempFilters.search });
  };

  const handleApply = () => {
    onFilterChange({
      accountId: tempFilters.accountId,
      categoryId: tempFilters.categoryId,
      uncategorized: tempFilters.uncategorized,
      minAmount: tempFilters.minAmount,
      maxAmount: tempFilters.maxAmount,
      dateFrom: tempFilters.dateFrom,
      dateTo: tempFilters.dateTo,
    });
    setPopoverOpen(false);
  };

  const handleClear = () => {
    setPopoverOpen(false);
    onClear();
  };

  const renderFilterControl = () => {
    switch (selectedFilter) {
      case "account":
        return (
          <AccountFilter
            filters={tempFilters}
            setFilters={setTempFilters}
            accounts={accounts}
          />
        );
      case "category":
        return (
          <CategoryFilter
            filters={tempFilters}
            setFilters={setTempFilters}
            categories={categories}
          />
        );
      case "amount":
        return (
          <AmountFilter filters={tempFilters} setFilters={setTempFilters} />
        );
      case "date":
        return <DateFilter filters={tempFilters} setFilters={setTempFilters} />;
      default:
        return <div className="text-muted-foreground">Coming soon...</div>;
    }
  };

  return (
    <div className="flex items-center gap-2 rounded-lg p-4">
      <form
        onSubmit={handleSearchSubmit}
        className="flex-1 flex items-center bg-transparent border border-border rounded-md px-3 py-2 focus-within:ring-2 focus-within:ring-primary"
      >
        <Search className="w-3 h-3 text-muted-foreground mr-1" />
        <input
          type="text"
          placeholder="Search transactions ..."
          className="flex-1 bg-transparent outline-none text-foreground placeholder:text-muted-foreground"
          value={tempFilters.search}
          onChange={(e) =>
            setTempFilters({ ...tempFilters, search: e.target.value })
          }
          onBlur={() => {
            if (tempFilters.search !== filters.search) {
              onFilterChange({ search: tempFilters.search });
            }
          }}
          style={{ height: 28 }}
        />
      </form>
      <Popover open={popoverOpen} onOpenChange={setPopoverOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="ghost"
            className="flex items-center gap-2"
            type="button"
          >
            <FilterIcon className="w-3 h-3" />
            Filter
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[500px] p-0" align="end">
          <div className="flex">
            <div className="w-1/3 border-r">
              <div className="p-2">
                <h3 className="text-lg font-semibold">Filters</h3>
                <div className="mt-4 flex flex-col items-start">
                  {FILTERS.map((f) => (
                    <Button
                      key={f.key}
                      variant={selectedFilter === f.key ? "secondary" : "ghost"}
                      onClick={() => setSelectedFilter(f.key)}
                      className="w-full justify-start gap-2"
                    >
                      {f.icon} {f.label}
                    </Button>
                  ))}
                </div>
              </div>
            </div>
            <div className="w-2/3 p-4 flex flex-col justify-between">
              {renderFilterControl()}
              <div className="flex justify-end gap-2 mt-4">
                <Button variant="ghost" onClick={handleClear}>
                  Clear
                </Button>
                <Button onClick={handleApply}>Apply</Button>
              </div>
            </div>
          </div>
        </PopoverContent>
      </Popover>
    </div>
  );
};

export default TransactionFilters;
