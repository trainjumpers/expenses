import TablePagination from "@/components/custom/Transaction/TablePagination";
import { useAccounts } from "@/components/hooks/useAccounts";
import { useStatements } from "@/components/hooks/useStatements";
import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { format } from "date-fns";
import { Calendar as CalendarIcon } from "lucide-react";
import { Building2, FileText, Search, X } from "lucide-react";
import { useState } from "react";

interface ViewStatementsModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
}

interface StatementStatusProps {
  status: string;
}

function StatementStatus({ status }: StatementStatusProps) {
  const statusClass =
    `capitalize text-xs px-2 py-0.5 rounded font-medium ` +
    (status === "pending"
      ? "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/40 dark:text-yellow-200"
      : status === "processing"
        ? "bg-blue-100 text-blue-800 dark:bg-blue-900/40 dark:text-blue-200"
        : status === "done"
          ? "bg-green-100 text-green-800 dark:bg-green-900/40 dark:text-green-200"
          : status === "error"
            ? "bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-200"
            : "bg-gray-100 text-gray-800 dark:bg-gray-900/40 dark:text-gray-200");

  return (
    <div className="flex items-center space-x-2">
      <span
        className={statusClass}
        style={{ minWidth: 60, textAlign: "center", display: "inline-block" }}
      >
        {status}
      </span>
    </div>
  );
}

export function ViewStatementsModal({
  isOpen,
  onOpenChange,
}: ViewStatementsModalProps) {
  const [page, setPage] = useState(1);
  const [accountId, setAccountId] = useState<number | undefined>();
  const [dateFrom, setDateFrom] = useState<Date | undefined>();
  const [dateTo, setDateTo] = useState<Date | undefined>();
  const [search, setSearch] = useState("");

  const pageSize = 5;

  const { data, isLoading, error } = useStatements({
    page,
    page_size: pageSize,
    account_id: typeof accountId === "number" ? accountId : undefined,
    date_from: dateFrom ? format(new Date(dateFrom), "yyyy-MM-dd") : undefined,
    date_to: dateTo ? format(new Date(dateTo), "yyyy-MM-dd") : undefined,
    search: search || undefined,
  });
  const statements = data?.statements || [];
  const total = data?.total || 0;
  const totalPages = Math.ceil(total / pageSize);
  const { data: accounts = [] } = useAccounts();

  const getFileTypeIcon = () => {
    return <FileText className="h-4 w-4 text-blue-500" />;
  };

  const handleClearFilters = () => {
    setAccountId(undefined);
    setDateFrom(undefined);
    setDateTo(undefined);
    setSearch("");
    setPage(1);
  };

  const hasActiveFilters =
    typeof accountId === "number" || dateFrom || dateTo || search;

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-200 max-h-[80vh]">
        <DialogHeader>
          <DialogTitle>Statement History</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <div className="relative flex-1">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search by filename..."
                  value={search}
                  onChange={(e) => {
                    setSearch(e.target.value);
                    setPage(1);
                  }}
                  className="pl-9"
                />
              </div>
              <Select
                value={accountId?.toString() || "all"}
                onValueChange={(value) => {
                  setAccountId(value === "all" ? undefined : Number(value));
                  setPage(1);
                }}
              >
                <SelectTrigger className="w-50">
                  <SelectValue placeholder="All Accounts" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Accounts</SelectItem>
                  {accounts.map((account) => (
                    <SelectItem key={account.id} value={account.id.toString()}>
                      {account.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="flex items-center gap-2">
              <Popover>
                <PopoverTrigger asChild>
                  <Button variant="ghost" className="justify-start text-sm">
                    <CalendarIcon className="h-4 w-4 mr-2" />
                    {dateFrom
                      ? format(new Date(dateFrom), "MMM d, yyyy")
                      : "From Date"}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0" align="start">
                  <Calendar
                    mode="single"
                    selected={dateFrom ? new Date(dateFrom) : undefined}
                    onSelect={(date) => {
                      setDateFrom(date);
                      setPage(1);
                    }}
                    initialFocus
                  />
                </PopoverContent>
              </Popover>

              <Popover>
                <PopoverTrigger asChild>
                  <Button variant="ghost" className="justify-start text-sm">
                    <CalendarIcon className="h-4 w-4 mr-2" />
                    {dateTo
                      ? format(new Date(dateTo), "MMM d, yyyy")
                      : "To Date"}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0" align="start">
                  <Calendar
                    mode="single"
                    selected={dateTo ? new Date(dateTo) : undefined}
                    onSelect={(date) => {
                      setDateTo(date);
                      setPage(1);
                    }}
                    initialFocus
                  />
                </PopoverContent>
              </Popover>
            </div>

            {hasActiveFilters && (
              <Button
                variant="ghost"
                size="sm"
                onClick={handleClearFilters}
                className="ml-auto"
              >
                <X className="h-4 w-4 mr-1" />
                Clear Filters
              </Button>
            )}
          </div>

          {isLoading ? (
            <div className="flex items-center justify-center py-8">
              <div className="text-sm text-muted-foreground">
                Loading statements...
              </div>
            </div>
          ) : error ? (
            <div className="flex items-center justify-center py-8">
              <div className="text-sm text-red-600 dark:text-red-300">
                Failed to load statements
              </div>
            </div>
          ) : statements.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 space-y-3">
              <FileText className="h-12 w-12 text-muted-foreground" />
              <div className="text-sm text-muted-foreground text-center">
                <p className="font-medium">
                  {hasActiveFilters
                    ? "No statements found"
                    : "No statements uploaded yet"}
                </p>
                <p>
                  {hasActiveFilters
                    ? "Try adjusting your filters"
                    : "Upload your first bank statement to get started"}
                </p>
              </div>
            </div>
          ) : (
            <div className="border rounded-lg overflow-hidden">
              <div className="max-h-100 overflow-y-auto">
                <Table>
                  <TableHeader className="sticky top-0 bg-background">
                    <TableRow>
                      <TableHead className="w-12">File</TableHead>
                      <TableHead>Filename</TableHead>
                      <TableHead>Account</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Uploaded</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {statements.map((statement) => (
                      <TableRow key={statement.id}>
                        <TableCell>
                          <div className="flex items-center justify-center">
                            {getFileTypeIcon()}
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="space-y-1">
                            <p
                              className="text-sm font-medium truncate max-w-50"
                              title={statement.original_filename}
                            >
                              {statement.original_filename}
                            </p>
                            <p className="text-xs text-muted-foreground uppercase">
                              {statement.file_type}
                            </p>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center space-x-2">
                            <Building2 className="h-4 w-4 text-muted-foreground" />
                            <span className="text-sm">
                              {accounts.find(
                                (acc) => acc.id === statement.account_id
                              )?.name || "Unknown Account"}
                            </span>
                          </div>
                        </TableCell>
                        <TableCell>
                          <StatementStatus status={statement.status} />
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center space-x-2">
                            <CalendarIcon className="h-4 w-4 text-muted-foreground" />
                            <span className="text-sm">
                              {format(
                                new Date(statement.created_at),
                                "MMM d, yyyy, h:mm aa"
                              )}
                            </span>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
              <TablePagination
                currentPage={page}
                totalPages={totalPages}
                setCurrentPage={setPage}
              />
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
