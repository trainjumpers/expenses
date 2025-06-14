export interface CreateCategoryInput {
  name: string;
  icon?: string;
}

export interface Category extends CreateCategoryInput {
  id: number;
  created_by: number;
}
