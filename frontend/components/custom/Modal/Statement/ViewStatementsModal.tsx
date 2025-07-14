import TablePagination from "@/components/custom/Transaction/TablePagination";
import { useAccounts } from "@/components/hooks/useAccounts";
import { useStatements } from "@/components/hooks/useStatements";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Building2, Calendar, FileText } from "lucide-react";
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
  const pageSize = 5;
  const { data, isLoading, error } = useStatements(page, pageSize);
  const statements = data?.statements || [];
  const total = data?.total || 0;
  const totalPages = Math.ceil(total / pageSize);
  const { data: accounts = [] } = useAccounts();

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const getFileTypeIcon = () => {
    return <FileText className="h-4 w-4 text-blue-500" />;
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[800px] max-h-[80vh]">
        <DialogHeader>
          <DialogTitle>Statement History</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          {isLoading ? (
            <div className="flex items-center justify-center py-8">
              <div className="text-sm text-muted-foreground">
                Loading statements...
              </div>
            </div>
          ) : error ? (
            <div className="flex items-center justify-center py-8">
              <div className="text-sm text-red-600">
                Failed to load statements
              </div>
            </div>
          ) : statements.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 space-y-3">
              <FileText className="h-12 w-12 text-muted-foreground" />
              <div className="text-sm text-muted-foreground text-center">
                <p className="font-medium">No statements uploaded yet</p>
                <p>Upload your first bank statement to get started</p>
              </div>
            </div>
          ) : (
            <div className="border rounded-lg overflow-hidden">
              <div className="max-h-[400px] overflow-y-auto">
                <Table>
                  <TableHeader className="sticky top-0 bg-background">
                    <TableRow>
                      <TableHead className="w-[50px]">File</TableHead>
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
                              className="text-sm font-medium truncate max-w-[200px]"
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
                            <Calendar className="h-4 w-4 text-muted-foreground" />
                            <span className="text-sm">
                              {formatDate(statement.created_at)}
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
