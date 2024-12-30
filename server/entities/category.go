package entities

type CreateCategoryInput struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
}

type CreateSubCategoryInput struct {
	Name       string `json:"name" binding:"required"`
	Color      string `json:"color"`
}

type UpdateCategoryInput struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type UpdateSubCategoryInput struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}
