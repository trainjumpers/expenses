import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { IconName } from "@/components/ui/icon-picker";
import { updateCategory } from "@/lib/api/category";
import { Category } from "@/lib/models/category";
import { Tag } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { CategoryForm } from "./CategoryForm";

interface UpdateCategoryModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  category: Category;
  onCategoryUpdated?: () => void;
}

export function UpdateCategoryModal({
  isOpen,
  onOpenChange,
  category,
  onCategoryUpdated,
}: UpdateCategoryModalProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (formData: { name: string; icon: IconName }) => {
    setIsSubmitting(true);
    try {
      const categoryData = {
        name: formData.name,
        icon: formData.icon,
      };
      await updateCategory(category.id, categoryData);
      toast.success("Category updated successfully!");
      if (onCategoryUpdated) {
        onCategoryUpdated();
      }
      onOpenChange(false);
    } catch (error) {
      console.error("Failed to update category:", error);
      toast.error("Failed to update category");
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
            Update Category
          </DialogTitle>
        </DialogHeader>
        <CategoryForm
          initialValues={{
            name: category.name,
            icon: (category.icon || "circle-dashed") as IconName,
          }}
          onSubmit={handleSubmit}
          loading={isSubmitting}
          submitText="Update"
          onOpenChange={onOpenChange}
        />
      </DialogContent>
    </Dialog>
  );
}
