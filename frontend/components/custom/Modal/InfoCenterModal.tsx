import { ViewAccountsModal } from "@/components/custom/Modal/Accounts/ViewAccountsModal";
import { ViewCategoriesModal } from "@/components/custom/Modal/Category/ViewCategoriesModal";
import { ViewRulesModal } from "@/components/custom/Modal/Rule/ViewRulesModal";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Banknote, BookOpen, Eye, LucideIcon, Tag, Wallet } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";

interface InfoCenterModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
}

interface ViewOption {
  title: string;
  description: string;
  icon: LucideIcon;
  onClick: () => void;
}

export function InfoCenterModal({
  isOpen,
  onOpenChange,
}: InfoCenterModalProps) {
  const router = useRouter();
  const [isViewAccountsModalOpen, setIsViewAccountsModalOpen] = useState(false);
  const [isViewCategoriesModalOpen, setIsViewCategoriesModalOpen] =
    useState(false);
  const [isViewRulesModalOpen, setIsViewRulesModalOpen] = useState(false);

  const options: ViewOption[] = [
    {
      title: "Categories",
      description: "View and manage your spending categories",
      icon: Tag,
      onClick: () => {
        onOpenChange(false);
        setIsViewCategoriesModalOpen(true);
      },
    },
    {
      title: "Accounts",
      description: "View and manage your bank accounts",
      icon: Wallet,
      onClick: () => {
        onOpenChange(false);
        setIsViewAccountsModalOpen(true);
      },
    },
    {
      title: "Transactions",
      description: "View and manage your transactions",
      icon: Banknote,
      onClick: () => {
        onOpenChange(false);
        router.push("/transaction");
      },
    },
    {
      title: "Rules",
      description: "View and manage your rules",
      icon: BookOpen,
      onClick: () => {
        onOpenChange(false);
        setIsViewRulesModalOpen(true);
      },
    },
  ];

  return (
    <>
      <Dialog open={isOpen} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Eye className="h-5 w-5" />
              View Center
            </DialogTitle>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            {options.map((option) => (
              <Button
                key={option.title}
                variant="outline"
                className="h-auto p-4 justify-start gap-4"
                onClick={option.onClick}
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

      <ViewAccountsModal
        isOpen={isViewAccountsModalOpen}
        onOpenChange={setIsViewAccountsModalOpen}
      />
      <ViewCategoriesModal
        isOpen={isViewCategoriesModalOpen}
        onOpenChange={setIsViewCategoriesModalOpen}
      />
      <ViewRulesModal
        isOpen={isViewRulesModalOpen}
        onOpenChange={setIsViewRulesModalOpen}
      />
    </>
  );
}
