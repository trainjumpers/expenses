package models

import "time"

type Category struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedBy int64     `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

type SubCategory struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Color      string    `json:"color"`
	CreatedBy  int64     `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
}

type CategoryWithSubs struct {
	Category
	SubCategories []SubCategory `json:"sub_categories"`
}
