import { useAccounts } from "@/components/custom/Provider/AccountProvider";
import { useCategories } from "@/components/custom/Provider/CategoryProvider";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { updateTransaction } from "@/lib/api/transaction";
import { Transaction } from "@/lib/models/transaction";
import { toast } from "sonner";

import { TransactionForm } from "./TransactionForm";

interface UpdateTransactionModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  transaction: Transaction | null;
  onTransactionUpdated?: () => Promise<void> | void;
  isRefreshing?: boolean;
}

export function UpdateTransactionModal({
  isOpen,
  onOpenChange,
  transaction,
  onTransactionUpdated,
  isRefreshing = false,
}: UpdateTransactionModalProps) {
  const { read: readAccounts } = useAccounts();
  const { read: readCategories } = useCategories();
  const accounts = readAccounts();
  const categories = readCategories();

  const handleSubmit = async (formData: {
    name: string;
    description: string;
    amount: string;
    date: Date;
    category_ids: number[];
    account_id: number;
  }) => {
    if (!transaction) return;
    try {
      const updateData = {
        name: formData.name,
        description: formData.description || undefined,
        amount: parseFloat(formData.amount),
        date: formData.date.toISOString(),
        category_ids: formData.category_ids,
        account_id: formData.account_id,
      };
      await updateTransaction(transaction.id, updateData);
      toast.success("Transaction updated successfully!");
      if (onTransactionUpdated) {
        await onTransactionUpdated();
      }
      onOpenChange(false);
    } catch (error) {
      console.error("Failed to update transaction:", error);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Update Transaction</DialogTitle>
        </DialogHeader>
        {transaction && (
          <TransactionForm
            initialValues={{
              name: transaction.name,
              description: transaction.description || "",
              amount: transaction.amount.toString(),
              date: new Date(transaction.date),
              category_ids: transaction.category_ids,
              account_id: transaction.account_id,
            }}
            onSubmit={handleSubmit}
            loading={false}
            isRefreshing={isRefreshing}
            accounts={accounts}
            categories={categories}
            onOpenChange={onOpenChange}
            submitText="Update"
          />
        )}
      </DialogContent>
    </Dialog>
  );
}

export default UpdateTransactionModal;
