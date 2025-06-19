import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Wallet } from "lucide-react";
import { useState } from "react";
import { useAccounts } from "@/components/custom/Provider/AccountProvider";
import { AddAccountModal } from "@/components/custom/Modal/Category/AddAccountModal";
import { Account } from "@/lib/models/account";
import { UpdateAccountModal } from "@/components/custom/Modal/Accounts/UpdateAccountModal";

interface ViewAccountsModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ViewAccountsModal({
  isOpen,
  onOpenChange,
}: ViewAccountsModalProps) {
  const { read: readAccounts } = useAccounts();
  const [isAddAccountModalOpen, setIsAddAccountModalOpen] = useState(false);
  const [selectedAccount, setSelectedAccount] = useState<Account | null>(null);
  const accounts = readAccounts();

  const handleAccountUpdated = () => {
    setSelectedAccount(null);
  };

  return (
    <>
      <Dialog open={isOpen} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Wallet className="h-5 w-5" />
              View Accounts
            </DialogTitle>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            {accounts.length === 0 ? (
              <p className="text-center text-muted-foreground">
                No accounts found. Add one to get started!
              </p>
            ) : (
              <div className="grid gap-4">
                {accounts.map((account) => (
                  <div
                    key={account.id}
                    className="flex items-center justify-between p-4 rounded-lg border border-border"
                  >
                    <div>
                      <h3 className="font-medium">{account.name}</h3>
                      <p className="text-sm text-muted-foreground">
                        {account.bank_type.toUpperCase()} -{" "}
                        {account.currency.toUpperCase()}
                      </p>
                    </div>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setSelectedAccount(account)}
                    >
                      Edit
                    </Button>
                  </div>
                ))}
              </div>
            )}
            <Button
              onClick={() => setIsAddAccountModalOpen(true)}
              className="w-full"
            >
              Add Account
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      <AddAccountModal
        isOpen={isAddAccountModalOpen}
        onOpenChange={setIsAddAccountModalOpen}
      />

      {selectedAccount && (
        <UpdateAccountModal
          isOpen={selectedAccount !== null}
          onOpenChange={() => setSelectedAccount(null)}
          account={selectedAccount}
          onAccountUpdated={handleAccountUpdated}
        />
      )}
    </>
  );
} 