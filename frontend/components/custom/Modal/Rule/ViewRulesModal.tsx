import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { BookOpen } from "lucide-react";
import { type FC } from "react";

interface ViewRulesModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
}

export const ViewRulesModal: FC<ViewRulesModalProps> = ({
  isOpen,
  onOpenChange,
}) => {
  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <BookOpen className="h-5 w-5" />
            View Rules
          </DialogTitle>
        </DialogHeader>
        <div className="py-4">
          {/* Placeholder for rules list */}
          <div className="text-muted-foreground text-center">
            No rules to display yet.
          </div>
          <div className="flex justify-end mt-6">
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              Close
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
};
