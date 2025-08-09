import { LoadingButton } from "@/components/ui/LoadingButton";
import { Button } from "@/components/ui/button";
import { DialogFooter } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { BankType, Currency } from "@/lib/models/account";
import { ChangeEvent, useState } from "react";

interface AccountFormProps {
  initialValues: {
    name: string;
    bank_type: BankType;
    currency: Currency;
    balance: string;
  };
  onSubmit: (formData: {
    name: string;
    bank_type: BankType;
    currency: Currency;
    balance: string;
  }) => Promise<void>;
  loading: boolean;
  submitText: string;
  onOpenChange: (open: boolean) => void;
}

export function AccountForm({
  initialValues,
  onSubmit,
  loading,
  submitText,
  onOpenChange,
}: AccountFormProps) {
  const [formData, setFormData] = useState(initialValues);

  const handleFormSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    await onSubmit(formData);
  };

  return (
    <form onSubmit={handleFormSubmit}>
      <div className="grid gap-4 py-4">
        <div className="grid grid-cols-3 items-center gap-4">
          <Label htmlFor="name" className="text-right">
            Account Name
          </Label>
          <Input
            id="name"
            value={formData.name}
            onChange={(e: ChangeEvent<HTMLInputElement>) =>
              setFormData({ ...formData, name: e.target.value })
            }
            placeholder="Enter account name"
            className="col-span-2 w-[220px]"
            required
          />
        </div>

        <div className="grid grid-cols-3 items-center gap-4">
          <Label htmlFor="bank_type" className="text-right">
            Bank
          </Label>
          <Select
            value={formData.bank_type}
            onValueChange={(value: BankType) =>
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
            onValueChange={(value: Currency) =>
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
            Balance
          </Label>
          <Input
            id="balance"
            type="number"
            value={formData.balance}
            onChange={(e: ChangeEvent<HTMLInputElement>) =>
              setFormData({ ...formData, balance: e.target.value })
            }
            placeholder="Enter balance"
            className="col-span-2 w-[220px]"
          />
        </div>
      </div>
      <DialogFooter>
        <Button
          type="button"
          variant="outline"
          onClick={() => onOpenChange(false)}
          disabled={loading}
        >
          Cancel
        </Button>
        <LoadingButton
          type="submit"
          loading={loading}
          fixedWidth="100px"
          disabled={loading}
        >
          {submitText}
        </LoadingButton>
      </DialogFooter>
    </form>
  );
}
