import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

export function TransactionsTableSkeleton() {
  return (
    <div className="w-full">
      <div className="border border-border bg-card rounded-t-md">
        <div className="h-[700px] flex flex-col m-4">
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-muted/50">
                <TableHead className="w-[50px] py-4">
                  <Skeleton className="h-4 w-4" />
                </TableHead>
                <TableHead className="py-4">
                  <Skeleton className="h-4 w-20" />
                </TableHead>
                <TableHead className="py-4">
                  <Skeleton className="h-4 w-24" />
                </TableHead>
                <TableHead className="py-4">
                  <Skeleton className="h-4 w-32" />
                </TableHead>
                <TableHead className="py-4">
                  <Skeleton className="h-4 w-24" />
                </TableHead>
                <TableHead className="py-4">
                  <Skeleton className="h-4 w-24" />
                </TableHead>
                <TableHead className="text-right py-4">
                  <Skeleton className="h-4 w-16 ml-auto" />
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {Array.from({ length: 10 }).map((_, index) => (
                <TableRow key={index} className="hover:bg-muted/50">
                  <TableCell className="py-4">
                    <Skeleton className="h-4 w-4" />
                  </TableCell>
                  <TableCell className="py-4">
                    <Skeleton className="h-4 w-24" />
                  </TableCell>
                  <TableCell className="py-4">
                    <Skeleton className="h-4 w-32" />
                  </TableCell>
                  <TableCell className="py-4">
                    <Skeleton className="h-4 w-40" />
                  </TableCell>
                  <TableCell className="py-4">
                    <Skeleton className="h-4 w-24" />
                  </TableCell>
                  <TableCell className="py-4">
                    <Skeleton className="h-4 w-24" />
                  </TableCell>
                  <TableCell className="text-right py-4">
                    <Skeleton className="h-4 w-20 ml-auto" />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </div>
      <div className="flex items-center justify-center py-3 bg-card rounded-b-md border-t border-border">
        <div className="flex items-center space-x-2">
          <Skeleton className="h-8 w-24" />
          <div className="flex items-center gap-1">
            {Array.from({ length: 3 }).map((_, index) => (
              <Skeleton key={index} className="h-8 w-8" />
            ))}
          </div>
          <Skeleton className="h-8 w-24" />
        </div>
      </div>
    </div>
  );
}
