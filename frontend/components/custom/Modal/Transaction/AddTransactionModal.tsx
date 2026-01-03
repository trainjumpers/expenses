import { useAccounts } from "@/components/hooks/useAccounts";
import { useCategories } from "@/components/hooks/useCategories";
import { useCreateTransaction } from "@/components/hooks/useTransactions";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import type { CreateTransaction } from "@/lib/models/transaction";

import { TransactionForm } from "./TransactionForm";

interface AddTransactionModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AddTransactionModal({
  isOpen,
  onOpenChange,
}: AddTransactionModalProps) {
  const { data: accounts = [] } = useAccounts();
  const { data: categories = [] } = useCategories();
  const createTransactionMutation = useCreateTransaction();

  const handleSubmit = async (formData: {
    name: string;
    description: string;
    amount: string;
    date: Date;
    category_ids: number[];
    account_id: number;
  }) => {
    const transactionData: CreateTransaction = {
      name: formData.name,
      description: formData.description || undefined,
      amount: parseFloat(formData.amount),
      date: formData.date.toISOString().split("T")[0],
      category_ids: formData.category_ids,
      account_id: formData.account_id,
    };

    createTransactionMutation.mutate(transactionData, {
      onSuccess: () => {
        onOpenChange(false);
      },
    });
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[600px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Add New Transaction</DialogTitle>
        </DialogHeader>
        <TransactionForm
          accounts={accounts}
          categories={categories}
          onSubmit={handleSubmit}
          loading={createTransactionMutation.isPending}
          submitText="Add"
          onOpenChange={onOpenChange}
          initialValues={{
            name: "",
            description: "",
            amount: "",
            date: new Date(),
            category_ids: [],
            account_id: accounts[0]?.id || 0,
          }}
        />
      </DialogContent>
    </Dialog>
  );
}
