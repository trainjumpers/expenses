import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Label } from "@/components/ui/label";
import { ChevronDownIcon } from "lucide-react";

interface Account {
  id: number;
  name: string;
  bank_type: string;
}

interface AccountSelectorProps {
  accounts: Account[];
  selectedAccountId: number;
  onAccountSelect: (accountId: number) => void;
}

export function AccountSelector({ 
  accounts, 
  selectedAccountId, 
  onAccountSelect 
}: AccountSelectorProps) {
  return (
    <div className="space-y-2">
      <Label>Account</Label>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="outline"
            className="w-full justify-start text-left font-normal flex items-center"
            type="button"
          >
            {(() => {
              const selected = accounts.find(
                (acc) => acc.id === selectedAccountId
              );
              return selected
                ? `${selected.name} (${selected.bank_type.toUpperCase()})`
                : "Select account";
            })()}
            <ChevronDownIcon className="ml-auto w-4 h-4 opacity-60" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent className="w-56 max-h-64 overflow-y-auto">
          {accounts.map((account) => (
            <DropdownMenuItem
              key={account.id}
              onClick={() => onAccountSelect(account.id)}
              className={`py-1 px-2 text-sm min-h-0 h-8 cursor-pointer flex items-center ${
                selectedAccountId === account.id
                  ? "bg-accent/40 font-semibold"
                  : ""
              }`}
            >
              {account.name} ({account.bank_type.toUpperCase()})
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}