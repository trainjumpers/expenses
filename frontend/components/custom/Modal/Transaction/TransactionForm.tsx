import { AddAccountModal } from "@/components/custom/Modal/Accounts/AddAccountModal";
import { AddCategoryModal } from "@/components/custom/Modal/Category/AddCategoryModal";
import { LoadingButton } from "@/components/ui/LoadingButton";
import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import { DialogFooter } from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Icon } from "@/components/ui/icon-picker";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Textarea } from "@/components/ui/textarea";
import { Account } from "@/lib/models/account";
import { Category } from "@/lib/models/category";
import { ChevronDownIcon } from "lucide-react";
import { ChangeEvent, useEffect, useState } from "react";

interface TransactionFormProps {
  initialValues: {
    name: string;
    description: string;
    amount: string;
    date: Date;
    category_ids: number[];
    account_id: number;
  };
  onSubmit: (formData: {
    name: string;
    description: string;
    amount: string;
    date: Date;
    category_ids: number[];
    account_id: number;
  }) => Promise<void>;
  loading: boolean;
  isRefreshing?: boolean;
  accounts: Account[];
  categories: Category[];
  submitText: string;
  onOpenChange: (open: boolean) => void;
}

export function TransactionForm({
  initialValues,
  onSubmit,
  loading,
  isRefreshing = false,
  accounts,
  categories,
  submitText,
  onOpenChange,
}: TransactionFormProps) {
  const [formData, setFormData] = useState(initialValues);
  const [showAddAccount, setShowAddAccount] = useState(false);
  const [showAddCategory, setShowAddCategory] = useState(false);
  const [openCalendar, setOpenCalendar] = useState(false);

  useEffect(() => {
    setFormData(initialValues);
  }, [initialValues]);

  const handleAccountAdded = (account: Account) => {
    setShowAddAccount(false);
    setFormData((prev) => ({ ...prev, account_id: account.id }));
  };

  const handleInputChange = (
    e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) => {
    const { id, value } = e.target;
    setFormData((prev) => ({ ...prev, [id]: value }));
  };

  const handleFormSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    await onSubmit(formData);
  };

  return (
    <form onSubmit={handleFormSubmit}>
      <div className="grid gap-4 py-4" aria-disabled={loading || isRefreshing}>
        <div className="grid grid-cols-3 items-center gap-4">
          <Label htmlFor="name" className="text-right">
            Name
          </Label>
          <Input
            id="name"
            value={formData.name}
            onChange={handleInputChange}
            placeholder="Enter transaction name"
            className="col-span-2 w-[220px]"
            required
          />
        </div>

        <div className="grid grid-cols-3 items-center gap-4">
          <Label htmlFor="description" className="text-right">
            Description
          </Label>
          <Textarea
            id="description"
            value={formData.description}
            onChange={handleInputChange}
            placeholder="Enter transaction description"
            className="col-span-2 w-[220px]"
          />
        </div>

        <div className="grid grid-cols-3 items-center gap-4">
          <Label htmlFor="amount" className="text-right">
            Amount
          </Label>
          <Input
            id="amount"
            type="number"
            step="0.01"
            value={formData.amount}
            onChange={handleInputChange}
            placeholder="Enter amount"
            className="col-span-2 w-[220px]"
            required
          />
        </div>

        <div className="grid grid-cols-3 items-center gap-4">
          <Label htmlFor="date" className="text-right">
            Date
          </Label>
          <Popover open={openCalendar} onOpenChange={setOpenCalendar}>
            <PopoverTrigger asChild>
              <Button
                variant="outline"
                className={
                  "col-span-2 w-[220px] justify-between font-normal" +
                  (!formData.date ? " text-muted-foreground" : "")
                }
              >
                {formData.date
                  ? formData.date.toLocaleDateString()
                  : "Pick a date"}
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-auto p-0" align="start">
              <Calendar
                mode="single"
                selected={formData.date}
                onSelect={(date) => {
                  if (date) setFormData((prev) => ({ ...prev, date }));
                  setOpenCalendar(false);
                }}
              />
            </PopoverContent>
          </Popover>
        </div>

        <div className="grid grid-cols-3 items-center gap-4">
          <Label className="text-right">Account</Label>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="outline"
                className="col-span-2 w-[220px] justify-start text-left font-normal flex items-center"
                type="button"
              >
                {(() => {
                  const selected = accounts.find(
                    (acc) => acc.id === formData.account_id
                  );
                  return selected ? selected.name : "Select account";
                })()}
                <ChevronDownIcon className="ml-auto w-4 h-4 opacity-60" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent className="w-56 max-h-64 overflow-y-auto">
              {accounts.map((account) => (
                <DropdownMenuItem
                  key={account.id}
                  onClick={() =>
                    setFormData((prev) => ({
                      ...prev,
                      account_id: account.id,
                    }))
                  }
                  className={`py-1 px-2 text-sm min-h-0 h-8 cursor-pointer flex items-center ${formData.account_id === account.id ? "bg-accent/40 font-semibold" : ""}`}
                >
                  {account.name}
                </DropdownMenuItem>
              ))}
              <DropdownMenuItem
                onClick={() => setShowAddAccount(true)}
                className="py-1 px-2 text-sm min-h-0 h-8 cursor-pointer flex items-center text-primary font-semibold border-t border-border"
              >
                + Add new account
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
          <AddAccountModal
            isOpen={showAddAccount}
            onOpenChange={setShowAddAccount}
            onAccountAdded={handleAccountAdded}
          />
        </div>
        <div className="grid grid-cols-3 items-center gap-4">
          <Label className="text-right">Categories</Label>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="outline"
                className="col-span-2 w-[220px] justify-start text-left font-normal flex items-center"
                type="button"
              >
                {(() => {
                  if (formData.category_ids.length === 0)
                    return "Select categories";
                  const selectedNames = categories
                    .filter((cat) => formData.category_ids.includes(cat.id))
                    .map((cat) => cat.name);
                  const joined = selectedNames.join(", ");
                  if (joined.length > 100) {
                    return joined.slice(0, 100) + "...";
                  }
                  return joined;
                })()}
                <ChevronDownIcon className="ml-auto w-4 h-4 opacity-60" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent className="w-56 max-h-64 overflow-y-auto">
              {categories.map((category) => {
                const selected = formData.category_ids.includes(category.id);
                return (
                  <DropdownMenuItem
                    key={category.id}
                    onClick={() => {
                      setFormData((prev) => {
                        const ids = prev.category_ids;
                        if (selected) {
                          return {
                            ...prev,
                            category_ids: ids.filter(
                              (id) => id !== category.id
                            ),
                          };
                        } else {
                          return {
                            ...prev,
                            category_ids: [...ids, category.id],
                          };
                        }
                      });
                    }}
                    className={`py-1 px-2 text-sm min-h-0 h-8 cursor-pointer flex items-center ${selected ? "bg-accent/40 font-semibold" : ""}`}
                  >
                    <Icon
                      name={
                        (category.icon
                          ? category.icon
                          : "circle-dashed") as import("@/components/ui/icon-picker").IconName
                      }
                      className="mr-2 w-4 h-4 inline-block align-middle"
                    />
                    <span className="align-middle">{category.name}</span>
                  </DropdownMenuItem>
                );
              })}
              <DropdownMenuItem
                onClick={() => setShowAddCategory(true)}
                className="py-1 px-2 text-sm min-h-0 h-8 cursor-pointer flex items-center text-primary font-semibold border-t border-border"
              >
                + Add new category
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
          <AddCategoryModal
            isOpen={showAddCategory}
            onOpenChange={setShowAddCategory}
          />
        </div>
      </div>
      <DialogFooter>
        <Button
          type="button"
          variant="outline"
          onClick={() => onOpenChange(false)}
          disabled={loading || isRefreshing}
        >
          Cancel
        </Button>
        <LoadingButton
          type="submit"
          loading={loading || isRefreshing}
          fixedWidth="100px"
          disabled={loading || isRefreshing}
        >
          {submitText}
        </LoadingButton>
      </DialogFooter>
    </form>
  );
}
