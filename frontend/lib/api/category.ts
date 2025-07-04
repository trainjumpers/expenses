import { apiRequest } from "@/lib/api/request";
import { API_BASE_URL } from "@/lib/constants/api";
import { Category, CreateCategoryInput } from "@/lib/models/category";

export async function listCategory(): Promise<Category[]> {
  return apiRequest<Category[]>(
    `${API_BASE_URL}/category`,
    {
      credentials: "include",
    },
    "category",
    [],
    "Failed to fetch categories"
  );
}

export async function getCategory(id: number): Promise<Category> {
  return apiRequest<Category>(
    `${API_BASE_URL}/category/${id}`,
    {
      credentials: "include",
    },
    "category",
    [],
    "Failed to fetch category"
  );
}

export async function createCategory(
  input: CreateCategoryInput
): Promise<Category> {
  return apiRequest<Category>(
    `${API_BASE_URL}/category`,
    {
      method: "POST",
      credentials: "include",
      body: JSON.stringify(input),
    },
    "category",
    [],
    "Failed to create category"
  );
}

export async function updateCategory(
  id: number,
  input: Partial<CreateCategoryInput>
): Promise<Category> {
  return apiRequest<Category>(
    `${API_BASE_URL}/category/${id}`,
    {
      method: "PATCH",
      credentials: "include",
      body: JSON.stringify(input),
    },
    "category",
    [],
    "Failed to update category"
  );
}

export async function deleteCategory(id: number): Promise<void> {
  return apiRequest<void>(
    `${API_BASE_URL}/category/${id}`,
    {
      method: "DELETE",
      credentials: "include",
    },
    "category",
    [],
    "Failed to delete category"
  );
}
