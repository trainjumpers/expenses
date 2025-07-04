"use client";

import { useSession } from "@/components/custom/Provider/SessionProvider";
import {
  createCategory,
  deleteCategory,
  listCategory,
  updateCategory,
} from "@/lib/api/category";
import { Category, CreateCategoryInput } from "@/lib/models/category";
import { createResource } from "@/lib/utils/suspense";
import React, {
  ReactNode,
  createContext,
  useContext,
  useEffect,
  useState,
} from "react";

export type CategoryResource = {
  read: () => Category[];
  refresh: () => void;
  create: (category: CreateCategoryInput) => Promise<Category>;
  update: (
    id: number,
    category: Partial<CreateCategoryInput>
  ) => Promise<Category>;
  delete: (id: number) => Promise<void>;
};

const CategoryContext = createContext<CategoryResource | null>(null);

export const CategoryProvider = ({ children }: { children: ReactNode }) => {
  const { isTokenAvailable } = useSession();
  const [abortController, setAbortController] =
    useState<AbortController | null>(null);

  const handleNewResource = () => {
    if (!isTokenAvailable) {
      return createResource<Category[]>(() => Promise.resolve([]));
    }
    if (abortController) {
      abortController.abort();
    }
    const controller = new AbortController();
    setAbortController(controller);
    return createResource<Category[]>(() => listCategory(), controller.signal);
  };

  const [resource, setResource] = useState(() => {
    return handleNewResource();
  });

  const refresh = () => {
    setResource(handleNewResource());
  };

  const create = async (category: CreateCategoryInput) => {
    try {
      const newCategory = await createCategory(category);
      return newCategory;
    } finally {
      refresh();
    }
  };
  const update = async (id: number, category: Partial<CreateCategoryInput>) => {
    try {
      const updated = await updateCategory(id, category);
      return updated;
    } finally {
      refresh();
    }
  };
  const del = async (id: number) => {
    try {
      await deleteCategory(id);
    } finally {
      refresh();
    }
  };

  useEffect(() => {
    refresh();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isTokenAvailable]);

  const read = () => {
    if (!resource) throw new Error("Resource not found");
    const result = resource.read();
    if (!result) return [];
    return result;
  };

  const value: CategoryResource = {
    read,
    refresh,
    create,
    update,
    delete: del,
  };

  return (
    <CategoryContext.Provider value={value}>
      {children}
    </CategoryContext.Provider>
  );
};

export function useCategories() {
  const ctx = useContext(CategoryContext);
  if (!ctx)
    throw new Error("useCategories must be used within a CategoryProvider");
  return ctx;
}
