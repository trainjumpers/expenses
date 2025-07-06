import { useUpdateCategory } from "@/components/hooks/useCategories";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { IconName } from "@/components/ui/icon-picker";
import { Category } from "@/lib/models/category";
import { Tag } from "lucide-react";
import { useState } from "react";

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
  const updateCategoryMutation = useUpdateCategory();

  const handleSubmit = async (formData: { name: string; icon: IconName }) => {
    setIsSubmitting(true);
    const categoryData = {
      name: formData.name,
      icon: formData.icon,
    };
    updateCategoryMutation.mutate(
      { id: category.id, data: categoryData },
      {
        onSuccess: () => {
          if (onCategoryUpdated) onCategoryUpdated();
          onOpenChange(false);
        },
        onError: (error) => {
          console.error("Failed to update category:", error);
        },
        onSettled: () => setIsSubmitting(false),
      }
    );
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
