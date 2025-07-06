import { useAccounts } from "@/components/hooks/useAccounts";
import { useCategories } from "@/components/hooks/useCategories";
import { useUpdateTransaction } from "@/components/hooks/useTransactions";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Transaction } from "@/lib/models/transaction";

import { TransactionForm } from "./TransactionForm";

interface UpdateTransactionModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  transaction: Transaction | null;
}

export function UpdateTransactionModal({
  isOpen,
  onOpenChange,
  transaction,
}: UpdateTransactionModalProps) {
  const { data: accounts = [] } = useAccounts();
  const { data: categories = [] } = useCategories();
  const updateTransactionMutation = useUpdateTransaction();

  const handleSubmit = async (formData: {
    name: string;
    description: string;
    amount: string;
    date: Date;
    category_ids: number[];
    account_id: number;
  }) => {
    if (!transaction) return;

    const updateData = {
      name: formData.name,
      description: formData.description || undefined,
      amount: parseFloat(formData.amount),
      date: formData.date.toISOString(),
      category_ids: formData.category_ids,
      account_id: formData.account_id,
    };

    updateTransactionMutation.mutate(
      { id: transaction.id, data: updateData },
      {
        onSuccess: () => {
          onOpenChange(false);
        },
      }
    );
  };
  if (!transaction) return null;

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[600px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Update Transaction</DialogTitle>
        </DialogHeader>
        <TransactionForm
          accounts={accounts}
          categories={categories}
          onSubmit={handleSubmit}
          loading={updateTransactionMutation.isPending}
          submitText="Update"
          onOpenChange={onOpenChange}
          initialValues={{
            name: transaction.name,
            description: transaction.description || "",
            amount: transaction.amount.toString(),
            date: new Date(transaction.date),
            category_ids: transaction.category_ids,
            account_id: transaction.account_id,
          }}
        />
      </DialogContent>
    </Dialog>
  );
}

export default UpdateTransactionModal;
