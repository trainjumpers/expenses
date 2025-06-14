import { useCategories } from "@/components/custom/Provider/CategoryProvider";
import { LoadingButton } from "@/components/ui/LoadingButton";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { type IconName, IconPicker } from "@/components/ui/icon-picker";
import { Input } from "@/components/ui/input";
import { useState } from "react";
import { toast } from "sonner";

interface AddCategoryModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AddCategoryModal({
  isOpen,
  onOpenChange,
}: AddCategoryModalProps) {
  const { create } = useCategories();
  const [formData, setFormData] = useState({
    name: "",
    icon: "" as IconName | undefined,
  });
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    try {
      if (!formData.name) {
        toast.error("Please enter a category name.");
        return;
      }
      await create({ name: formData.name, icon: formData.icon });
      toast.success("Category created successfully!");
      setFormData({ name: "", icon: undefined });
      onOpenChange(false);
    } catch (error: unknown) {
      console.error(error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Create Category</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="grid gap-4 py-4">
            <div className="flex items-center gap-4">
              <Input
                id="name"
                value={formData.name}
                onChange={(e) =>
                  setFormData({ ...formData, name: e.target.value })
                }
                placeholder="Enter category name"
                className="flex-1"
              />
              <div className="flex items-center gap-2">
                <div>
                  <IconPicker
                    value={formData.icon || undefined}
                    onValueChange={(value) =>
                      setFormData({ ...formData, icon: value })
                    }
                    defaultValue="circle-dashed"
                    modal={true}
                  />
                </div>
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <LoadingButton
              type="submit"
              loading={isLoading}
              fixedWidth="140px"
              disabled={isLoading}
            >
              Create Category
            </LoadingButton>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
