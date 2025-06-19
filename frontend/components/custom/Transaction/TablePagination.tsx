import { Button } from "@/components/ui/button";
import { ChevronLeft, ChevronRight, MoreHorizontal } from "lucide-react";
import React from "react";

interface TablePaginationProps {
  currentPage: number;
  totalPages: number;
  setCurrentPage: (page: number) => void;
}

const TablePagination: React.FC<TablePaginationProps> = ({
  currentPage,
  totalPages,
  setCurrentPage,
}) => {
  const renderPaginationItems = () => {
    const items = [];
    const maxVisiblePages = 3;

    if (totalPages <= maxVisiblePages) {
      for (let i = 1; i <= totalPages; i++) {
        items.push(
          <Button
            key={i}
            variant={currentPage === i ? "default" : "outline"}
            size="sm"
            onClick={() => setCurrentPage(i)}
            className={
              currentPage === i
                ? "bg-primary text-primary-foreground"
                : "hover:bg-muted"
            }
          >
            {i}
          </Button>
        );
      }
    } else {
      items.push(
        <Button
          key={1}
          variant={currentPage === 1 ? "default" : "outline"}
          size="sm"
          onClick={() => setCurrentPage(1)}
          className={
            currentPage === 1
              ? "bg-primary text-primary-foreground"
              : "hover:bg-muted"
          }
        >
          1
        </Button>
      );
      if (currentPage > 2) {
        items.push(
          <span key="start-ellipsis" className="px-2">
            <MoreHorizontal className="h-4 w-4" />
          </span>
        );
      }
      for (
        let i = Math.max(2, currentPage - 1);
        i <= Math.min(totalPages - 1, currentPage + 1);
        i++
      ) {
        items.push(
          <Button
            key={i}
            variant={currentPage === i ? "default" : "outline"}
            size="sm"
            onClick={() => setCurrentPage(i)}
            className={
              currentPage === i
                ? "bg-primary text-primary-foreground"
                : "hover:bg-muted"
            }
          >
            {i}
          </Button>
        );
      }
      if (currentPage < totalPages - 1) {
        items.push(
          <span key="end-ellipsis" className="px-2">
            <MoreHorizontal className="h-4 w-4" />
          </span>
        );
      }
      items.push(
        <Button
          key={totalPages}
          variant={currentPage === totalPages ? "default" : "outline"}
          size="sm"
          onClick={() => setCurrentPage(totalPages)}
          className={
            currentPage === totalPages
              ? "bg-primary text-primary-foreground"
              : "hover:bg-muted"
          }
        >
          {totalPages}
        </Button>
      );
    }
    return items;
  };

  return (
    <div className="flex items-center justify-center py-3 bg-card rounded-b-md border-t border-border">
      <div className="flex items-center space-x-2">
        <Button
          variant="outline"
          size="sm"
          onClick={() => setCurrentPage(Math.max(currentPage - 1, 1))}
          disabled={currentPage === 1}
          className="hover:bg-muted"
        >
          <ChevronLeft className="h-4 w-4 mr-2" />
          Previous
        </Button>
        <div className="flex items-center gap-1">{renderPaginationItems()}</div>
        <Button
          variant="outline"
          size="sm"
          onClick={() => setCurrentPage(Math.min(currentPage + 1, totalPages))}
          disabled={currentPage === totalPages}
          className="hover:bg-muted"
        >
          Next
          <ChevronRight className="h-4 w-4 ml-2" />
        </Button>
      </div>
    </div>
  );
};

export default TablePagination;
