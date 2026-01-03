import { AddAccountModal } from "@/components/custom/Modal/Accounts/AddAccountModal";
import { UpdateAccountModal } from "@/components/custom/Modal/Accounts/UpdateAccountModal";
import { ConfirmDialog } from "@/components/custom/Modal/ConfirmDialog";
import { useAccounts, useDeleteAccount } from "@/components/hooks/useAccounts";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import {
  Pagination,
  PaginationContent,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination";
import type { Account } from "@/lib/models/account";
import { Search, Trash2, Wallet } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

interface ViewAccountsModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ViewAccountsModal({
  isOpen,
  onOpenChange,
}: ViewAccountsModalProps) {
  const { data: accounts = [] } = useAccounts();
  const deleteAccountMutation = useDeleteAccount();

  const [isAddAccountModalOpen, setIsAddAccountModalOpen] = useState(false);
  const [selectedAccount, setSelectedAccount] = useState<Account | null>(null);
  const [confirmDeleteAccount, setConfirmDeleteAccount] =
    useState<Account | null>(null);

  // Frontend-only search + pagination
  const [searchTerm, setSearchTerm] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [currentPage, setCurrentPage] = useState(1);
  const pageSize = 5;

  useEffect(() => {
    const t = setTimeout(() => setDebouncedSearch(searchTerm.trim()), 300);
    return () => clearTimeout(t);
  }, [searchTerm]);

  const filtered = useMemo(() => {
    if (!debouncedSearch) return accounts;
    const s = debouncedSearch.toLowerCase();
    return accounts.filter(
      (a) =>
        a.name.toLowerCase().includes(s) ||
        a.bank_type.toLowerCase().includes(s) ||
        a.currency.toLowerCase().includes(s)
    );
  }, [accounts, debouncedSearch]);

  const totalPages = Math.ceil(filtered.length / pageSize) || 1;
  const pagedAccounts = useMemo(() => {
    const start = (currentPage - 1) * pageSize;
    return filtered.slice(start, start + pageSize);
  }, [filtered, currentPage]);

  const handleAccountUpdated = () => {
    setSelectedAccount(null);
  };

  const openDeleteDialog = (account: Account) => {
    setConfirmDeleteAccount(account);
  };

  const handleConfirmDelete = async () => {
    if (!confirmDeleteAccount) return;

    deleteAccountMutation.mutate(confirmDeleteAccount.id, {
      onSuccess: () => {
        setConfirmDeleteAccount(null);
      },
    });
  };

  return (
    <>
      <Dialog open={isOpen} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-[600px] max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Wallet className="h-5 w-5" />
              View Accounts
            </DialogTitle>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                aria-label="Search accounts"
                placeholder="Search by name, bank, or currency..."
                value={searchTerm}
                onChange={(e) => {
                  setSearchTerm(e.target.value);
                  setCurrentPage(1);
                }}
                className="pl-10"
              />
            </div>
            {pagedAccounts.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                {debouncedSearch
                  ? "No accounts match your search."
                  : "No accounts found. Add one to get started!"}
              </div>
            ) : (
              <div className="grid gap-4">
                {pagedAccounts.map((account) => (
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
                        disabled={deleteAccountMutation.isPending}
                        onClick={() => openDeleteDialog(account)}
                      >
                        <Trash2 className="h-4 w-4" />
                        <span className="sr-only">Delete</span>
                      </Button>
                    </div>
                  </div>
                ))}

                {totalPages > 1 && (
                  <div className="flex items-center justify-between text-sm text-muted-foreground">
                    <Pagination>
                      <PaginationContent>
                        <PaginationItem>
                          <PaginationPrevious
                            onClick={() =>
                              setCurrentPage((p) => Math.max(1, p - 1))
                            }
                            className={
                              currentPage <= 1
                                ? "pointer-events-none opacity-50"
                                : "cursor-pointer"
                            }
                          />
                        </PaginationItem>
                        {Array.from(
                          { length: totalPages },
                          (_, i) => i + 1
                        ).map((page) => (
                          <PaginationItem key={page}>
                            <PaginationLink
                              onClick={() => setCurrentPage(page)}
                              isActive={currentPage === page}
                              className="cursor-pointer"
                            >
                              {page}
                            </PaginationLink>
                          </PaginationItem>
                        ))}
                        <PaginationItem>
                          <PaginationNext
                            onClick={() =>
                              setCurrentPage((p) => Math.min(totalPages, p + 1))
                            }
                            className={
                              currentPage >= totalPages
                                ? "pointer-events-none opacity-50"
                                : "cursor-pointer"
                            }
                          />
                        </PaginationItem>
                      </PaginationContent>
                    </Pagination>
                  </div>
                )}
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
        loading={deleteAccountMutation.isPending}
        onConfirm={handleConfirmDelete}
      />
    </>
  );
}
