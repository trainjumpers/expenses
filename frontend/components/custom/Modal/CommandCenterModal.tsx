import { AddAccountModal } from "@/components/custom/Modal/Accounts/AddAccountModal";
import { AddCategoryModal } from "@/components/custom/Modal/Category/AddCategoryModal";
import { AddTransactionModal } from "@/components/custom/Modal/Transaction/AddTransactionModal";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { LucideIcon, Receipt, Tag, Wallet } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";

interface CommandCenterModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
}

interface CommandOption {
  title: string;
  description: string;
  icon: LucideIcon;
  onClick?: () => void;
  href?: string;
}

export function CommandCenterModal({
  isOpen,
  onOpenChange,
}: CommandCenterModalProps) {
  const router = useRouter();
  const [isAddAccountModalOpen, setIsAddAccountModalOpen] = useState(false);
  const [isAddCategoryModalOpen, setIsAddCategoryModalOpen] = useState(false);
  const [isAddTransactionModalOpen, setIsAddTransactionModalOpen] =
    useState(false);

  const options: CommandOption[] = [
    {
      title: "Add Account",
      description: "Add a new bank account or credit card",
      icon: Wallet,
      onClick: () => setIsAddAccountModalOpen(true),
    },
    {
      title: "Add Category",
      description: "Create a new spending category",
      icon: Tag,
      onClick: () => setIsAddCategoryModalOpen(true),
    },
    {
      title: "Add Transaction",
      description: "Record a new income or expense",
      icon: Receipt,
      onClick: () => setIsAddTransactionModalOpen(true),
    },
  ];

  const handleOptionClick = (option: CommandOption) => {
    if (option.href) {
      onOpenChange(false);
      router.push(option.href);
    } else if (option.onClick) {
      onOpenChange(false);
      option.onClick();
    }
  };

  return (
    <>
      <Dialog open={isOpen} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Command Center</DialogTitle>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            {options.map((option) => (
              <Button
                key={option.title}
                variant="outline"
                className="h-auto p-4 justify-start gap-4"
                onClick={() => handleOptionClick(option)}
              >
                <option.icon className="h-5 w-5" />
                <div className="flex flex-col items-start">
                  <span className="font-medium">{option.title}</span>
                  <span className="text-sm text-muted-foreground">
                    {option.description}
                  </span>
                </div>
              </Button>
            ))}
          </div>
        </DialogContent>
      </Dialog>
      <AddAccountModal
        isOpen={isAddAccountModalOpen}
        onOpenChange={setIsAddAccountModalOpen}
      />
      <AddCategoryModal
        isOpen={isAddCategoryModalOpen}
        onOpenChange={setIsAddCategoryModalOpen}
      />
      <AddTransactionModal
        isOpen={isAddTransactionModalOpen}
        onOpenChange={setIsAddTransactionModalOpen}
      />
    </>
  );
}
