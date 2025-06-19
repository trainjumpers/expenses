import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import * as React from "react";

export type DropdownCellProps<T> = {
  isOpen: boolean;
  onOpen: () => void;
  onClose: () => void;
  options: T[];
  renderOption: (opt: T) => React.ReactNode;
  onSelect: (id: number) => void | Promise<void>;
  children: React.ReactNode;
  selectedIds?: number[];
};

function DropdownCell<T extends { id: number }>({
  isOpen,
  onOpen,
  onClose,
  options,
  renderOption,
  onSelect,
  children,
  selectedIds,
}: DropdownCellProps<T>) {
  return (
    <Popover
      open={isOpen}
      onOpenChange={(open) => (open ? onOpen() : onClose())}
    >
      <PopoverTrigger asChild>
        <div className="cursor-pointer inline-block">{children}</div>
      </PopoverTrigger>
      <PopoverContent align="start" className="p-2 min-w-[160px] w-auto">
        <div className="flex flex-col gap-1">
          {options.map((opt) => (
            <button
              key={opt.id}
              className={`flex items-center gap-2 px-2 py-1 rounded hover:bg-muted text-xs ${Array.isArray(selectedIds) && selectedIds.includes(opt.id) ? "bg-primary/10 font-semibold" : ""}`}
              onClick={async (e) => {
                e.stopPropagation();
                await onSelect(opt.id);
              }}
            >
              {renderOption(opt)}
            </button>
          ))}
        </div>
      </PopoverContent>
    </Popover>
  );
}

export default DropdownCell;
