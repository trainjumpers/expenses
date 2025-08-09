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
import { Input } from "@/components/ui/input";
import {
  Pagination,
  PaginationContent,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination";
import { Category } from "@/lib/models/category";
import { Search, Tag, Trash2 } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

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

  // Frontend-only search + pagination
  const [searchTerm, setSearchTerm] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [currentPage, setCurrentPage] = useState(1);
  const pageSize = 5;

  useEffect(() => {
    const t = setTimeout(() => setDebouncedSearch(searchTerm.trim()), 300);
    return () => clearTimeout(t);
  }, [searchTerm]);

  const filtered = useMemo(() => {
    if (!debouncedSearch) return categories;
    const s = debouncedSearch.toLowerCase();
    return categories.filter((c) => c.name.toLowerCase().includes(s));
  }, [categories, debouncedSearch]);

  const totalPages = Math.ceil(filtered.length / pageSize) || 1;
  const pagedCategories = useMemo(() => {
    const start = (currentPage - 1) * pageSize;
    return filtered.slice(start, start + pageSize);
  }, [filtered, currentPage]);

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
        <DialogContent className="sm:max-w-[600px] max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Tag className="h-5 w-5" />
              Categories
            </DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                aria-label="Search categories"
                placeholder="Search categories by name..."
                value={searchTerm}
                onChange={(e) => {
                  setSearchTerm(e.target.value);
                  setCurrentPage(1);
                }}
                className="pl-10"
              />
            </div>
            {isLoading ? (
              <div className="space-y-2">
                {[...Array(3)].map((_, i) => (
                  <div
                    key={i}
                    className="h-16 bg-muted rounded animate-pulse"
                  />
                ))}
              </div>
            ) : pagedCategories.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                {debouncedSearch
                  ? "No categories match your search."
                  : "No categories found. Add your first category to get started."}
              </div>
            ) : (
              <div className="space-y-2">
                {pagedCategories.map((category) => (
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
                {totalPages > 1 && (
                  <div className="flex items-center justify-between text-sm text-muted-foreground pt-2">
                    <Pagination>
                      <PaginationContent>
                        <PaginationItem>
                          <PaginationPrevious
                            onClick={() =>
                              setCurrentPage((p) => Math.max(1, p - 1))
                            }
                            className={
                              currentPage <= 1
                                ? "pointer-events-none opacity-50"
                                : "cursor-pointer"
                            }
                          />
                        </PaginationItem>
                        {Array.from(
                          { length: totalPages },
                          (_, i) => i + 1
                        ).map((page) => (
                          <PaginationItem key={page}>
                            <PaginationLink
                              onClick={() => setCurrentPage(page)}
                              isActive={currentPage === page}
                              className="cursor-pointer"
                            >
                              {page}
                            </PaginationLink>
                          </PaginationItem>
                        ))}
                        <PaginationItem>
                          <PaginationNext
                            onClick={() =>
                              setCurrentPage((p) => Math.min(totalPages, p + 1))
                            }
                            className={
                              currentPage >= totalPages
                                ? "pointer-events-none opacity-50"
                                : "cursor-pointer"
                            }
                          />
                        </PaginationItem>
                      </PaginationContent>
                    </Pagination>
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
