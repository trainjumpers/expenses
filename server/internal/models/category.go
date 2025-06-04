package models

// CreateCategoryInput is used for creating a new category
type CreateCategoryInput struct {
	Name      string `json:"name" binding:"required" maxlength:"100"`
	Icon      string `json:"icon"`
	CreatedBy int64  `json:"created_by"`
}

// UpdateCategoryInput is used for updating an existing category
type UpdateCategoryInput struct {
	Name string  `json:"name,omitempty" maxlength:"100"`
	Icon *string `json:"icon,omitempty"`
}

// CategoryResponse is the response model for a category
type CategoryResponse struct {
	Id        int64   `json:"id"`
	Name      string  `json:"name"`
	Icon      *string `json:"icon"`
	CreatedBy int64   `json:"created_by"`
}
