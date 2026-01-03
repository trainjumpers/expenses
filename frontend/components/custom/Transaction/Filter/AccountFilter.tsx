import type { TransactionFiltersState } from "@/app/transaction/page";
import type { Account } from "@/lib/models/account";
import React from "react";

interface AccountFilterProps {
  filters: TransactionFiltersState;
  setFilters: React.Dispatch<React.SetStateAction<TransactionFiltersState>>;
  accounts: Account[];
}

export const AccountFilter: React.FC<AccountFilterProps> = ({
  filters,
  setFilters,
  accounts,
}) => {
  return (
    <div className="w-full">
      <label className="block text-[15px] mb-1">Account</label>
      <select
        className="border rounded px-2 py-1 w-full h-8"
        value={filters.accountId ?? ""}
        onChange={(e) =>
          setFilters({
            ...filters,
            accountId: e.target.value ? Number(e.target.value) : undefined,
          })
        }
      >
        <option value="">All</option>
        {accounts.map((acc) => (
          <option key={acc.id} value={acc.id}>
            {acc.name}
          </option>
        ))}
      </select>
    </div>
  );
};
