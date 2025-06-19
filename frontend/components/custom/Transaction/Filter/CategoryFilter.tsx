import { TransactionFiltersState } from "@/app/transaction/page";
import { Category } from "@/lib/models/category";
import React from "react";

interface CategoryFilterProps {
  filters: TransactionFiltersState;
  setFilters: React.Dispatch<React.SetStateAction<TransactionFiltersState>>;
  categories: Category[];
}

export const CategoryFilter: React.FC<CategoryFilterProps> = ({
  filters,
  setFilters,
  categories,
}) => {
  return (
    <div className="w-full">
      <label className="block text-[15px] mb-1">Category</label>
      <select
        className="border rounded px-2 py-1 w-full h-8"
        value={filters.categoryId ?? ""}
        onChange={(e) =>
          setFilters({
            ...filters,
            categoryId: e.target.value ? Number(e.target.value) : undefined,
          })
        }
      >
        <option value="">All</option>
        {categories.map((cat) => (
          <option key={cat.id} value={cat.id}>
            {cat.name}
          </option>
        ))}
      </select>
    </div>
  );
};
