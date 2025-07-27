import { Checkbox } from "@/components/ui/checkbox";
import { Icon, IconName } from "@/components/ui/icon-picker";
import { Account, Currency } from "@/lib/models/account";
import { Category } from "@/lib/models/category";
import { Transaction } from "@/lib/models/transaction";
import React from "react";

import DropdownCell from "./DropdownCell";

const currencyMap: Record<Currency, string> = {
  usd: "$",
  inr: "₹",
};

interface TransactionTableRowProps {
  transaction: Transaction;
  selected: boolean;
  onSelect: () => void;
  accounts: Account[];
  categories: Category[];
  editing: { id: number; field: "category" | "account" | null };
  setEditing: React.Dispatch<
    React.SetStateAction<{ id: number; field: "category" | "account" | null }>
  >;
  handleUpdate: (original: Transaction, updated: Transaction) => Promise<void>;
  renderCategoryPills: (ids: number[]) => React.ReactNode;
  getAccountName: (id: number) => string;
}

const TransactionTableRow: React.FC<TransactionTableRowProps> = ({
  transaction,
  selected,
  onSelect,
  accounts,
  categories,
  editing,
  setEditing,
  handleUpdate,
  renderCategoryPills,
  getAccountName,
}) => {
  return (
    <tr className="hover:bg-muted/50">
      <td className="py-4 px-6">
        <Checkbox
          checked={selected}
          onCheckedChange={onSelect}
          aria-label={`Select ${transaction.name}`}
          className="translate-y-[2px]"
        />
      </td>
      <td className="text-foreground py-4 px-6 text-center">
        {transaction.name}
      </td>
      <td className="text-foreground py-4 px-6 text-center">
        {transaction.description || "-"}
      </td>
      <td className="text-foreground py-4 px-6 text-center">
        <DropdownCell<Category>
          isOpen={editing.id === transaction.id && editing.field === "category"}
          onOpen={() => setEditing({ id: transaction.id, field: "category" })}
          onClose={() => setEditing({ id: -1, field: null })}
          options={categories}
          renderOption={(cat: Category) => (
            <>
              {cat.icon && (
                <Icon name={cat.icon as IconName} className="w-4 h-4" />
              )}{" "}
              {cat.name}
            </>
          )}
          onSelect={async (catId: number) => {
            const originalTransaction = transaction;
            const prev = transaction.category_ids;
            let newIds: number[];
            if (prev.includes(catId)) {
              newIds = prev.filter((id) => id !== catId);
            } else {
              newIds = [...prev, catId];
            }

            const updatedTransaction = {
              ...transaction,
              category_ids: newIds,
            };
            await handleUpdate(originalTransaction, updatedTransaction);
            setEditing({ id: -1, field: null });
          }}
          selectedIds={transaction.category_ids}
        >
          {transaction.category_ids.length > 0 ? (
            renderCategoryPills(transaction.category_ids)
          ) : (
            <span className="text-muted-foreground">-</span>
          )}
        </DropdownCell>
      </td>
      <td className="text-right text-foreground font-medium py-4 px-6">
        {(() => {
          const account = accounts.find(
            (acc) => acc.id === transaction.account_id
          );
          const currency = account?.currency || "usd";
          const symbol = currencyMap[currency as Currency] || "$";
          // Credit: show amount if negative (< 0)
          return transaction.amount < 0
            ? `${symbol}${Math.abs(transaction.amount).toFixed(2)}`
            : "-";
        })()}
      </td>
      <td className="text-right text-foreground font-medium py-4 px-6">
        {(() => {
          const account = accounts.find(
            (acc) => acc.id === transaction.account_id
          );
          const currency = account?.currency || "usd";
          const symbol = currencyMap[currency as Currency] || "$";
          // Debit: show amount if positive (> 0)
          return transaction.amount > 0
            ? `${symbol}${transaction.amount.toFixed(2)}`
            : "-";
        })()}
      </td>
      <td className="text-foreground py-4 px-6 text-center">
        <DropdownCell<Account>
          isOpen={editing.id === transaction.id && editing.field === "account"}
          onOpen={() => setEditing({ id: transaction.id, field: "account" })}
          onClose={() => setEditing({ id: -1, field: null })}
          options={accounts}
          renderOption={(acc: Account) => acc.name}
          onSelect={async (accId: number) => {
            const originalTransaction = transaction;
            const updatedTransaction = { ...transaction, account_id: accId };
            await handleUpdate(originalTransaction, updatedTransaction);
            setEditing({ id: -1, field: null });
          }}
          selectedIds={[transaction.account_id]}
        >
          {getAccountName(transaction.account_id)}
        </DropdownCell>
      </td>
      <td className="text-foreground py-4 px-6 text-center">
        {transaction.date.split("T")[0]}
      </td>
    </tr>
  );
};

export default TransactionTableRow;
