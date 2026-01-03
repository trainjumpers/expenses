import { useUpdateAccount } from "@/components/hooks/useAccounts";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import type { Account, BankType, Currency } from "@/lib/models/account";
import { Wallet } from "lucide-react";
import { useState } from "react";

import { AccountForm } from "./AccountForm";

interface UpdateAccountModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  account: Account;
  onAccountUpdated?: () => void;
}

export function UpdateAccountModal({
  isOpen,
  onOpenChange,
  account,
  onAccountUpdated,
}: UpdateAccountModalProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const updateAccountMutation = useUpdateAccount();

  const handleSubmit = async (formData: {
    name: string;
    bank_type: BankType;
    currency: Currency;
    balance: string;
  }) => {
    setIsSubmitting(true);
    const accountData = {
      name: formData.name,
      bank_type: formData.bank_type,
      currency: formData.currency,
      balance: formData.balance ? Number(formData.balance) : undefined,
    };
    updateAccountMutation.mutate(
      { id: account.id, data: accountData },
      {
        onSuccess: () => {
          if (onAccountUpdated) onAccountUpdated();
          onOpenChange(false);
        },
        onError: (error) => {
          console.error("Failed to update account:", error);
        },
        onSettled: () => setIsSubmitting(false),
      }
    );
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-106.25">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Wallet className="h-5 w-5" />
            Update Account
          </DialogTitle>
        </DialogHeader>
        <AccountForm
          initialValues={{
            name: account.name,
            bank_type: account.bank_type as BankType,
            currency: account.currency as Currency,
            balance: account.balance?.toString() || "0",
          }}
          onSubmit={handleSubmit}
          loading={isSubmitting}
          submitText="Update"
          onOpenChange={onOpenChange}
        />
      </DialogContent>
    </Dialog>
  );
}
