import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { updateAccount } from "@/lib/api/account";
import { Account, BankType, Currency } from "@/lib/models/account";
import { Wallet } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
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

  const handleSubmit = async (formData: {
    name: string;
    bank_type: BankType;
    currency: Currency;
    balance: string;
  }) => {
    setIsSubmitting(true);
    try {
      const accountData = {
        name: formData.name,
        bank_type: formData.bank_type,
        currency: formData.currency,
        balance: formData.balance ? Number(formData.balance) : undefined,
      };
      await updateAccount(account.id, accountData);
      toast.success("Account updated successfully!");
      if (onAccountUpdated) {
        onAccountUpdated();
      }
      onOpenChange(false);
    } catch (error) {
      console.error("Failed to update account:", error);
      toast.error("Failed to update account");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
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