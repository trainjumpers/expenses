import { ConfirmDialog } from "@/components/custom/Modal/ConfirmDialog";
import { AddRuleModal } from "@/components/custom/Modal/Rule/AddRuleModal";
import { EditRuleModal } from "@/components/custom/Modal/Rule/EditRuleModal";
import { RuleListSkeleton } from "@/components/custom/Skeletons/RuleSkeletons";
import { useDeleteRule, useRules } from "@/components/hooks/useRules";
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
import type { Rule } from "@/lib/models/rule";
import { BookOpen, Search, Trash2 } from "lucide-react";
import { useState, useMemo, useCallback, useEffect } from "react";

interface ViewRulesModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
}

export const ViewRulesModal = ({
  isOpen,
  onOpenChange,
}: ViewRulesModalProps) => {
  const [currentPage, setCurrentPage] = useState(1);
  const [searchTerm, setSearchTerm] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [editRuleId, setEditRuleId] = useState<number | null>(null);
  const [isAddRuleModalOpen, setIsAddRuleModalOpen] = useState(false);
  const [loadingId, setLoadingId] = useState<number | null>(null);
  const [confirmDeleteRule, setConfirmDeleteRule] = useState<Rule | null>(null);
  const [confirmLoading, setConfirmLoading] = useState(false);
  const deleteRuleMutation = useDeleteRule();

  const pageSize = 5;

  // Debounce search term
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchTerm);
      setCurrentPage(1); // Reset to first page on search
    }, 300);

    return () => clearTimeout(timer);
  }, [searchTerm]);

  const queryParams = useMemo(() => ({
    page: currentPage,
    page_size: pageSize,
    search: debouncedSearch || undefined,
  }), [currentPage, pageSize, debouncedSearch]);

  const { data: response, isLoading, refetch } = useRules(queryParams);

  const rules = response?.rules || [];
  const totalItems = response?.total || 0;
  const totalPages = Math.ceil(totalItems / pageSize);

  const handlePageChange = useCallback((page: number) => {
    setCurrentPage(page);
  }, []);

  const openDeleteDialog = (rule: Rule) => {
    setConfirmDeleteRule(rule);
    setConfirmLoading(false);
  };

  const handleConfirmDelete = async () => {
    if (!confirmDeleteRule) return;
    setConfirmLoading(true);
    setLoadingId(confirmDeleteRule.id);
    await deleteRuleMutation.mutateAsync(confirmDeleteRule.id);
    setConfirmDeleteRule(null);
    setConfirmLoading(false);
    setLoadingId(null);
    refetch();
  };

  return (
    <>
      <Dialog open={isOpen} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-[600px] max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <BookOpen className="h-5 w-5" />
              View Rules
            </DialogTitle>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            {/* Search Input */}
            <div className="relative">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="Search rules by name or description..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10"
              />
            </div>

            {/* Rules List */}
            {isLoading ? (
              <RuleListSkeleton count={3} />
            ) : rules.length === 0 ? (
              <div className="text-muted-foreground text-center py-8">
                {debouncedSearch ? "No rules match your search." : "No rules to display yet."}
              </div>
            ) : (
              <>
                <div className="grid gap-4">
                  {rules.map((rule) => (
                    <div
                      key={rule.id}
                      className="flex items-center justify-between p-4 rounded-lg border border-border"
                    >
                      <div className="flex-1 min-w-0">
                        <div className="font-medium truncate">{rule.name}</div>
                        {rule.description && (
                          <div className="text-sm text-muted-foreground truncate">
                            {rule.description}
                          </div>
                        )}
                      </div>
                      <div className="flex gap-2 ml-4">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => setEditRuleId(rule.id)}
                        >
                          Edit
                        </Button>
                        <Button
                          variant="destructive"
                          size="sm"
                          disabled={loadingId === rule.id}
                          onClick={() => openDeleteDialog(rule)}
                        >
                          <Trash2 className="h-4 w-4" />
                          <span className="sr-only">Delete</span>
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>

                {/* Pagination */}
                {totalPages > 1 && (
                  <div className="flex items-center justify-between text-sm text-muted-foreground">
                    <Pagination>
                      <PaginationContent>
                        <PaginationItem>
                          <PaginationPrevious
                            onClick={() => handlePageChange(currentPage - 1)}
                            className={currentPage <= 1 ? "pointer-events-none opacity-50" : "cursor-pointer"}
                          />
                        </PaginationItem>
                        
                        {Array.from({ length: totalPages }, (_, i) => i + 1).map((page) => (
                          <PaginationItem key={page}>
                            <PaginationLink
                              onClick={() => handlePageChange(page)}
                              isActive={currentPage === page}
                              className="cursor-pointer"
                            >
                              {page}
                            </PaginationLink>
                          </PaginationItem>
                        ))}
                        
                        <PaginationItem>
                          <PaginationNext
                            onClick={() => handlePageChange(currentPage + 1)}
                            className={currentPage >= totalPages ? "pointer-events-none opacity-50" : "cursor-pointer"}
                          />
                        </PaginationItem>
                      </PaginationContent>
                    </Pagination>
                  </div>
                )}
              </>
            )}

            {/* Add Rule Button */}
            <Button
              onClick={() => setIsAddRuleModalOpen(true)}
              className="w-full"
            >
              Add Rule
            </Button>
          </div>
        </DialogContent>
      </Dialog>
      <AddRuleModal
        isOpen={isAddRuleModalOpen}
        onOpenChange={setIsAddRuleModalOpen}
      />
      {editRuleId !== null && (
        <EditRuleModal
          isOpen={true}
          onOpenChange={(open) => {
            if (!open) {
              setEditRuleId(null);
              refetch();
            }
          }}
          ruleId={editRuleId}
        />
      )}
      <ConfirmDialog
        isOpen={!!confirmDeleteRule}
        onOpenChange={(open) => {
          if (!open) setConfirmDeleteRule(null);
        }}
        title="Delete Rule"
        description={
          confirmDeleteRule
            ? `Are you sure you want to delete the rule "${confirmDeleteRule.name}"? This action cannot be undone.`
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
};
