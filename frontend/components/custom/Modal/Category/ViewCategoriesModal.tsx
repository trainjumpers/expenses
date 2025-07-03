import { AddCategoryModal } from "@/components/custom/Modal/Category/AddCategoryModal";
import { UpdateCategoryModal } from "@/components/custom/Modal/Category/UpdateCategoryModal";
import { ConfirmDialog } from "@/components/custom/Modal/ConfirmDialog";
import { useCategories } from "@/components/custom/Provider/CategoryProvider";
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
  const {
    read: readCategories,
    delete: deleteCategory,
    refresh,
  } = useCategories();
  const [isAddCategoryModalOpen, setIsAddCategoryModalOpen] = useState(false);
  const [selectedCategory, setSelectedCategory] = useState<Category | null>(
    null
  );
  const [loadingId, setLoadingId] = useState<number | null>(null);
  const [confirmDeleteCategory, setConfirmDeleteCategory] =
    useState<Category | null>(null);
  const [confirmLoading, setConfirmLoading] = useState(false);
  const categories = readCategories();

  const handleCategoryUpdated = () => {
    setSelectedCategory(null);
  };

  const openDeleteDialog = (category: Category) => {
    setConfirmDeleteCategory(category);
    setConfirmLoading(false);
  };

  const handleConfirmDelete = async () => {
    if (!confirmDeleteCategory) return;
    setConfirmLoading(true);
    setLoadingId(confirmDeleteCategory.id);
    await deleteCategory(confirmDeleteCategory.id);
    refresh();
    setConfirmDeleteCategory(null);
    setConfirmLoading(false);
    setLoadingId(null);
  };

  return (
    <>
      <Dialog open={isOpen} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Tag className="h-5 w-5" />
              View Categories
            </DialogTitle>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            {categories.length === 0 ? (
              <p className="text-center text-muted-foreground">
                No categories found. Add one to get started!
              </p>
            ) : (
              <div className="grid gap-4">
                {categories.map((category) => (
                  <div
                    key={category.id}
                    className="flex items-center justify-between p-4 rounded-lg border border-border"
                  >
                    <div className="flex items-center gap-2">
                      <Icon
                        name={
                          (category.icon
                            ? category.icon
                            : "circle-dashed") as IconName
                        }
                        className="h-4 w-4"
                      />
                      <h3 className="font-medium">{category.name}</h3>
                    </div>
                    <div className="flex gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setSelectedCategory(category)}
                      >
                        Edit
                      </Button>
                      <Button
                        variant="destructive"
                        size="sm"
                        disabled={loadingId === category.id}
                        onClick={() => openDeleteDialog(category)}
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
              onClick={() => setIsAddCategoryModalOpen(true)}
              className="w-full"
            >
              Add Category
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
          isOpen={selectedCategory !== null}
          onOpenChange={() => setSelectedCategory(null)}
          category={selectedCategory}
          onCategoryUpdated={handleCategoryUpdated}
        />
      )}
      <ConfirmDialog
        isOpen={!!confirmDeleteCategory}
        onOpenChange={(open) => {
          if (!open) setConfirmDeleteCategory(null);
        }}
        title="Delete Category"
        description={
          confirmDeleteCategory
            ? `Are you sure you want to delete the category "${confirmDeleteCategory.name}"? This action cannot be undone.`
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
}
