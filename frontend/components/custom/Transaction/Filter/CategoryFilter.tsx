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
        value={
          filters.uncategorized
            ? "uncategorized"
            : (filters.categoryId?.toString() ?? "")
        }
        onChange={(e) => {
          if (e.target.value === "uncategorized") {
            setFilters({
              ...filters,
              categoryId: undefined,
              uncategorized: true,
            });
          } else if (e.target.value === "") {
            setFilters({
              ...filters,
              categoryId: undefined,
              uncategorized: undefined,
            });
          } else {
            setFilters({
              ...filters,
              categoryId: Number(e.target.value),
              uncategorized: undefined,
            });
          }
        }}
      >
        <option value="">All</option>
        <option value="uncategorized">Uncategorized</option>
        {categories.map((cat) => (
          <option key={cat.id} value={cat.id}>
            {cat.name}
          </option>
        ))}
      </select>
    </div>
  );
};
