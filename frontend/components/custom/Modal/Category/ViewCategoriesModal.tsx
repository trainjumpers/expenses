import { AddCategoryModal } from "@/components/custom/Modal/Category/AddCategoryModal";
import { UpdateCategoryModal } from "@/components/custom/Modal/Category/UpdateCategoryModal";
import { ConfirmDialog } from "@/components/custom/Modal/ConfirmDialog";
import {
  useCategories,
  useDeleteCategory,
} from "@/components/hooks/useCategories";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Icon, IconName } from "@/components/ui/icon-picker";
import { Category } from "@/lib/models/category";
import { Tag, Trash2 } from "lucide-react";
import { useState } from "react";

interface ViewCategoriesModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ViewCategoriesModal({
  isOpen,
  onOpenChange,
}: ViewCategoriesModalProps) {
  const { data: categories = [], isLoading } = useCategories();
  const deleteCategoryMutation = useDeleteCategory();

  const [isAddCategoryModalOpen, setIsAddCategoryModalOpen] = useState(false);
  const [selectedCategory, setSelectedCategory] = useState<Category | null>(
    null
  );
  const [confirmDeleteCategory, setConfirmDeleteCategory] =
    useState<Category | null>(null);

  const handleDeleteCategory = async (category: Category) => {
    deleteCategoryMutation.mutate(category.id, {
      onSuccess: () => {
        setConfirmDeleteCategory(null);
      },
    });
  };

  return (
    <>
      <Dialog open={isOpen} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-[425px] max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Tag className="h-5 w-5" />
              Categories
            </DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            {isLoading ? (
              <div className="space-y-2">
                {[...Array(3)].map((_, i) => (
                  <div
                    key={i}
                    className="h-16 bg-muted rounded animate-pulse"
                  />
                ))}
              </div>
            ) : (
              <div className="space-y-2">
                {categories.map((category) => (
                  <div
                    key={category.id}
                    className="flex items-center justify-between p-3 border rounded-lg hover:bg-muted/50"
                  >
                    <div className="flex items-center gap-3">
                      <Icon
                        name={(category.icon as IconName) || "Tag"}
                        className="h-5 w-5"
                      />
                      <span className="font-medium">{category.name}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setSelectedCategory(category)}
                      >
                        Edit
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setConfirmDeleteCategory(category)}
                        disabled={deleteCategoryMutation.isPending}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                ))}
                {categories.length === 0 && (
                  <div className="text-center py-8 text-muted-foreground">
                    No categories found. Add your first category to get started.
                  </div>
                )}
              </div>
            )}
            <Button
              onClick={() => setIsAddCategoryModalOpen(true)}
              className="w-full"
            >
              Add New Category
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      <AddCategoryModal
        isOpen={isAddCategoryModalOpen}
        onOpenChange={setIsAddCategoryModalOpen}
      />

      {selectedCategory && (
        <UpdateCategoryModal
          isOpen={true}
          onOpenChange={(open) => !open && setSelectedCategory(null)}
          category={selectedCategory}
        />
      )}

      <ConfirmDialog
        isOpen={!!confirmDeleteCategory}
        onOpenChange={(open) => !open && setConfirmDeleteCategory(null)}
        title="Delete Category"
        description={`Are you sure you want to delete "${confirmDeleteCategory?.name}"? This action cannot be undone.`}
        onConfirm={() => {
          if (confirmDeleteCategory) {
            handleDeleteCategory(confirmDeleteCategory);
          }
        }}
        loading={deleteCategoryMutation.isPending}
      />
    </>
  );
}
