"use client";

import { useCreateAccount } from "@/components/hooks/useAccounts";
import { LoadingButton } from "@/components/ui/LoadingButton";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Account } from "@/lib/models/account";
import { useState } from "react";
import { toast } from "sonner";

interface AddAccountModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  onAccountAdded?: (account: Account) => void;
}

export function AddAccountModal({
  isOpen,
  onOpenChange,
  onAccountAdded,
}: AddAccountModalProps) {
  const createAccountMutation = useCreateAccount();
  const [formData, setFormData] = useState({
    name: "",
    bank_type: "",
    currency: "inr",
    balance: "",
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!formData.name || !formData.bank_type || !formData.currency) {
      toast.error("Please fill all required fields.");
      return;
    }

    const input = {
      name: formData.name,
      bank_type:
        formData.bank_type.toLowerCase() as import("@/lib/models/account").BankType,
      currency:
        formData.currency.toLowerCase() as import("@/lib/models/account").Currency,
      balance: formData.balance ? Number(formData.balance) : undefined,
    };

    createAccountMutation.mutate(input, {
      onSuccess: (newAccount) => {
        setFormData({ name: "", bank_type: "", currency: "inr", balance: "" });
        onOpenChange(false);
        if (onAccountAdded) onAccountAdded(newAccount);
      },
    });
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Add Account</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="grid gap-4 py-4">
            <div className="grid grid-cols-3 items-center gap-4">
              <Label htmlFor="name" className="text-right">
                Account Name
              </Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) =>
                  setFormData({ ...formData, name: e.target.value })
                }
                placeholder="Enter account name"
                className="col-span-2 w-[220px]"
              />
            </div>

            <div className="grid grid-cols-3 items-center gap-4">
              <Label htmlFor="bank_type" className="text-right">
                Bank
              </Label>
              <Select
                value={formData.bank_type}
                onValueChange={(value) =>
                  setFormData({ ...formData, bank_type: value })
                }
              >
                <SelectTrigger className="col-span-2 w-[220px]">
                  <SelectValue placeholder="Select bank" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="axis">Axis Bank</SelectItem>
                  <SelectItem value="sbi">State Bank of India</SelectItem>
                  <SelectItem value="hdfc">HDFC Bank</SelectItem>
                  <SelectItem value="icici">ICICI Bank</SelectItem>
                  <SelectItem value="icici_credit">
                    ICICI Bank (Credit Card)
                  </SelectItem>
                  <SelectItem value="investment">Investment Account</SelectItem>
                  <SelectItem value="others">Others</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="grid grid-cols-3 items-center gap-4">
              <Label htmlFor="currency" className="text-right">
                Currency
              </Label>
              <Select
                value={formData.currency}
                onValueChange={(value) =>
                  setFormData({ ...formData, currency: value })
                }
              >
                <SelectTrigger className="col-span-2 w-[220px]">
                  <SelectValue placeholder="Select currency" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="inr">Indian Rupee (INR)</SelectItem>
                  <SelectItem value="usd">US Dollar (USD)</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="grid grid-cols-3 items-center gap-4">
              <Label htmlFor="balance" className="text-right">
                Initial Balance
              </Label>
              <Input
                id="balance"
                type="number"
                value={formData.balance}
                onChange={(e) =>
                  setFormData({ ...formData, balance: e.target.value })
                }
                placeholder="Enter initial balance"
                className="col-span-2 w-[220px]"
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <LoadingButton
              type="submit"
              loading={createAccountMutation.isPending}
              fixedWidth="140px"
              disabled={createAccountMutation.isPending}
            >
              Add Account
            </LoadingButton>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
