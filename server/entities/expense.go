package entities

import "time"

type ExpenseInput struct {
	Name          string            `json:"name" binding:"required"`
	Description   string            `json:"description"`
	PayerID       int64             `json:"payer_id" binding:"required"`
	Amount        float64           `json:"amount" binding:"required"`
	Contributions map[int64]float64 `json:"contributions" binding:"required"`
	SubcategoryID int64             `json:"subcategory_id"`
}

type ExpenseOutput struct {
	Name          string                              `json:"name" binding:"required"`
	Description   string                              `json:"description"`
	PayerID       int64                               `json:"payer_id" binding:"required"`
	Amount        float64                             `json:"amount" binding:"required"`
	Contributions map[int64]ExpenseOutputContribution `json:"contributions" binding:"required"`
	ID            int64                               `json:"id"`
	CreatedBy     int64                               `json:"created_by"`
	CreatedAt     *time.Time                          `json:"created_at"`
}

type ExpenseOutputContribution struct {
	Amount float64 `json:"amount"`
	Name   string  `json:"name"`
}

type UpdateExpenseContributionsInput struct {
	Contributions map[int64]float64 `json:"contributions" binding:"required"`
}

type UpdateExpenseBasicInput struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	PayerID     int64  `json:"payer_id,omitempty"`
}
