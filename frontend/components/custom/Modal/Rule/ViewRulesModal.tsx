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
import type { Rule } from "@/lib/models/rule";
import { BookOpen, Trash2 } from "lucide-react";
import { useState } from "react";

interface ViewRulesModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
}

export const ViewRulesModal = ({
  isOpen,
  onOpenChange,
}: ViewRulesModalProps) => {
  const { data: rules = [], isLoading, refetch } = useRules();
  const [editRuleId, setEditRuleId] = useState<number | null>(null);
  const [isAddRuleModalOpen, setIsAddRuleModalOpen] = useState(false);
  const [loadingId, setLoadingId] = useState<number | null>(null);
  const [confirmDeleteRule, setConfirmDeleteRule] = useState<Rule | null>(null);
  const [confirmLoading, setConfirmLoading] = useState(false);
  const deleteRuleMutation = useDeleteRule();

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
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <BookOpen className="h-5 w-5" />
              View Rules
            </DialogTitle>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            {isLoading ? (
              <RuleListSkeleton count={3} />
            ) : rules.length === 0 ? (
              <div className="text-muted-foreground text-center">
                No rules to display yet.
              </div>
            ) : (
              <div className="grid gap-4">
                {rules.map((rule) => (
                  <div
                    key={rule.id}
                    className="flex items-center justify-between p-4 rounded-lg border border-border"
                  >
                    <div>
                      <div className="font-medium">{rule.name}</div>
                      {rule.description && (
                        <div className="text-sm text-muted-foreground">
                          {rule.description}
                        </div>
                      )}
                    </div>
                    <div className="flex gap-2">
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
            )}
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
