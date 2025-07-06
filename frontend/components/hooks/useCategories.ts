"use client";

import {
  createCategory,
  deleteCategory,
  listCategory,
  updateCategory,
} from "@/lib/api/category";
import { Category, CreateCategoryInput } from "@/lib/models/category";
import { queryKeys } from "@/lib/query-client";
import { ApiErrorType, getErrorMessage } from "@/lib/types/errors";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { useSession } from "./useSession";

export function useCategories() {
  const { isAuthenticated } = useSession();

  return useQuery({
    queryKey: queryKeys.categories,
    queryFn: () => listCategory(),
    enabled: isAuthenticated,
    staleTime: 10 * 60 * 1000,
  });
}
export function useCategory(id: number) {
  const { data: categories } = useCategories();

  return useQuery({
    queryKey: queryKeys.category(id),
    queryFn: () => {
      const category = categories?.find(
        (category: Category) => category.id === id
      );
      if (!category) throw new Error("Category not found");
      return category;
    },
    enabled: !!id && !!categories,
  });
}

export function useCreateCategory() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (categoryData: CreateCategoryInput) =>
      createCategory(categoryData),
    onMutate: async (newCategory) => {
      await queryClient.cancelQueries({ queryKey: queryKeys.categories });

      const previousCategories = queryClient.getQueryData<Category[]>(
        queryKeys.categories
      );

      const tempId = Date.now();
      if (previousCategories) {
        const optimisticCategory: Category = {
          id: tempId,
          created_by: 0,
          ...newCategory,
        };
        queryClient.setQueryData<Category[]>(queryKeys.categories, [
          ...previousCategories,
          optimisticCategory,
        ]);
        return { previousCategories, tempId };
      }
      return { previousCategories, tempId };
    },
    onError: (error: ApiErrorType, variables, context) => {
      if (context?.previousCategories) {
        queryClient.setQueryData(
          queryKeys.categories,
          context.previousCategories
        );
      }
      const message = getErrorMessage(error);
      console.error(message || "Failed to create category");
    },
    onSuccess: (newCategory, variables, context) => {
      queryClient.setQueryData<Category[]>(queryKeys.categories, (old) => {
        if (!old) return [newCategory];
        const withoutOptimistic = old.filter(
          (category) => category.id !== context?.tempId
        );
        return [...withoutOptimistic, newCategory];
      });

      queryClient.invalidateQueries({ queryKey: ["transactions"] });

      toast.success("Category created successfully");
    },
  });
}

export function useUpdateCategory() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: number;
      data: Partial<CreateCategoryInput>;
    }) => updateCategory(id, data),
    onMutate: async ({ id, data }) => {
      await queryClient.cancelQueries({ queryKey: queryKeys.categories });

      const previousCategories = queryClient.getQueryData<Category[]>(
        queryKeys.categories
      );

      if (previousCategories) {
        queryClient.setQueryData<Category[]>(
          queryKeys.categories,
          previousCategories.map((category) =>
            category.id === id ? { ...category, ...data } : category
          )
        );
      }

      return { previousCategories };
    },
    onError: (error: ApiErrorType, _, context) => {
      if (context?.previousCategories) {
        queryClient.setQueryData(
          queryKeys.categories,
          context.previousCategories
        );
      }
      const message = getErrorMessage(error);
      console.error(message || "Failed to update category");
    },
    onSuccess: (updatedCategory) => {
      queryClient.setQueryData<Category[]>(queryKeys.categories, (old) => {
        if (!old) return [updatedCategory];
        return old.map((category) =>
          category.id === updatedCategory.id ? updatedCategory : category
        );
      });
      queryClient.invalidateQueries({ queryKey: ["transactions"] });
      toast.success("Category updated successfully");
    },
  });
}

export function useDeleteCategory() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => deleteCategory(id),
    onMutate: async (id) => {
      await queryClient.cancelQueries({ queryKey: queryKeys.categories });

      const previousCategories = queryClient.getQueryData<Category[]>(
        queryKeys.categories
      );

      if (previousCategories) {
        queryClient.setQueryData<Category[]>(
          queryKeys.categories,
          previousCategories.filter((category) => category.id !== id)
        );
      }

      return { previousCategories };
    },
    onError: (error: ApiErrorType, variables, context) => {
      if (context?.previousCategories) {
        queryClient.setQueryData(
          queryKeys.categories,
          context.previousCategories
        );
      }
      const message = getErrorMessage(error);
      console.error(message || "Failed to delete category");
    },
    onSuccess: (_, deletedId) => {
      queryClient.setQueryData<Category[]>(queryKeys.categories, (old) => {
        if (!old) return [];
        return old.filter((category) => category.id !== deletedId);
      });

      queryClient.invalidateQueries({ queryKey: ["transactions"] });

      toast.success("Category deleted successfully");
    },
  });
}
