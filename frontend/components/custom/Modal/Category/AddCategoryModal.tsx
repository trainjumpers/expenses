import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { IconName } from "@/components/ui/icon-picker";
import { createCategory } from "@/lib/api/category";
import { Category } from "@/lib/models/category";
import { Tag } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { CategoryForm } from "./CategoryForm";

interface AddCategoryModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  onCategoryAdded?: (category: Category) => void;
}

export function AddCategoryModal({
  isOpen,
  onOpenChange,
  onCategoryAdded,
}: AddCategoryModalProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (formData: { name: string; icon: IconName }) => {
    setIsSubmitting(true);
    try {
      const categoryData = {
        name: formData.name,
        icon: formData.icon,
      };
      const category = await createCategory(categoryData);
      toast.success("Category added successfully!");
      if (onCategoryAdded) {
        onCategoryAdded(category);
      }
      onOpenChange(false);
    } catch (error) {
      console.error("Failed to create category:", error);
      toast.error("Failed to create category");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Tag className="h-5 w-5" />
            Add Category
          </DialogTitle>
        </DialogHeader>
        <CategoryForm
          initialValues={{
            name: "",
            icon: "circle-dashed",
          }}
          onSubmit={handleSubmit}
          loading={isSubmitting}
          submitText="Add"
          onOpenChange={onOpenChange}
        />
      </DialogContent>
    </Dialog>
  );
}
