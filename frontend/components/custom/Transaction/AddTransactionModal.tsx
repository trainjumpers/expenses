import { useAccounts } from "@/components/custom/Provider/AccountProvider";
import { useCategories } from "@/components/custom/Provider/CategoryProvider";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { createTransaction } from "@/lib/api/transaction";
import { CreateTransaction } from "@/lib/models/transaction";
import { useState } from "react";
import { toast } from "sonner";

import { TransactionForm } from "./TransactionModal";

interface AddTransactionModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  onTransactionAdded?: () => Promise<void> | void;
  isRefreshing?: boolean;
}

export function AddTransactionModal({
  isOpen,
  onOpenChange,
  onTransactionAdded,
  isRefreshing = false,
}: AddTransactionModalProps) {
  const { read: readAccounts } = useAccounts();
  const { read: readCategories } = useCategories();
  const accounts = readAccounts();
  const categories = readCategories();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (formData: {
    name: string;
    description: string;
    amount: string;
    date: Date;
    category_ids: number[];
    account_id: number;
  }) => {
    setIsSubmitting(true);
    try {
      const transactionData: CreateTransaction = {
        name: formData.name,
        description: formData.description || undefined,
        amount: parseFloat(formData.amount),
        date: formData.date.toISOString().split("T")[0],
        category_ids: formData.category_ids,
        account_id: formData.account_id,
      };
      await createTransaction(transactionData);
      toast.success("Transaction added successfully!");
      if (onTransactionAdded) {
        await onTransactionAdded();
      }
      onOpenChange(false);
    } catch (error) {
      console.error("Failed to create transaction:", error);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Add Transaction</DialogTitle>
        </DialogHeader>
        <TransactionForm
          initialValues={{
            name: "",
            description: "",
            amount: "",
            date: new Date(),
            category_ids: [],
            account_id: 0,
          }}
          onSubmit={handleSubmit}
          loading={isSubmitting}
          isRefreshing={isRefreshing}
          accounts={accounts}
          categories={categories}
          onOpenChange={onOpenChange}
          submitText="Add"
        />
      </DialogContent>
    </Dialog>
  );
}
