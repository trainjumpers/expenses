import { AddAccountModal } from "@/components/custom/Modal/Accounts/AddAccountModal";
import { UpdateAccountModal } from "@/components/custom/Modal/Accounts/UpdateAccountModal";
import { ConfirmDialog } from "@/components/custom/Modal/ConfirmDialog";
import { useAccounts } from "@/components/custom/Provider/AccountProvider";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Account } from "@/lib/models/account";
import { Trash2, Wallet } from "lucide-react";
import { useState } from "react";

interface ViewAccountsModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ViewAccountsModal({
  isOpen,
  onOpenChange,
}: ViewAccountsModalProps) {
  const { read: readAccounts, delete: deleteAccount, refresh } = useAccounts();
  const [isAddAccountModalOpen, setIsAddAccountModalOpen] = useState(false);
  const [selectedAccount, setSelectedAccount] = useState<Account | null>(null);
  const [loadingId, setLoadingId] = useState<number | null>(null);
  const [confirmDeleteAccount, setConfirmDeleteAccount] =
    useState<Account | null>(null);
  const [confirmLoading, setConfirmLoading] = useState(false);
  const accounts = readAccounts();

  const handleAccountUpdated = () => {
    setSelectedAccount(null);
  };

  const openDeleteDialog = (account: Account) => {
    setConfirmDeleteAccount(account);
    setConfirmLoading(false);
  };

  const handleConfirmDelete = async () => {
    if (!confirmDeleteAccount) return;
    setConfirmLoading(true);
    setLoadingId(confirmDeleteAccount.id);
    await deleteAccount(confirmDeleteAccount.id);
    refresh();
    setConfirmDeleteAccount(null);
    setConfirmLoading(false);
    setLoadingId(null);
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
                    <div className="flex gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setSelectedAccount(account)}
                      >
                        Edit
                      </Button>
                      <Button
                        variant="destructive"
                        size="sm"
                        disabled={loadingId === account.id}
                        onClick={() => openDeleteDialog(account)}
                      >
                        <Trash2 className="h-4 w-4" />
                        <span className="sr-only">Delete</span>
                      </Button>
                    </div>
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
      <ConfirmDialog
        isOpen={!!confirmDeleteAccount}
        onOpenChange={(open) => {
          if (!open) setConfirmDeleteAccount(null);
        }}
        title="Delete Account"
        description={
          confirmDeleteAccount
            ? `Are you sure you want to delete the account "${confirmDeleteAccount.name}"? This action cannot be undone.`
            : ""
        }
        confirmLabel="Delete"
        cancelLabel="Cancel"
        destructive
        loading={confirmLoading}
        onConfirm={handleConfirmDelete}
      />
    </>
  );
}
