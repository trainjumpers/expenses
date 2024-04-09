package models

import "time"

type Expense struct {
	ID          int        `json:"id"`
	Description string     `json:"description"`
	PayerID     int        `json:"payer_id"`
	Amount      float64    `json:"amount"`
	CreatedBy   int        `json:"created_by"`
	CreatedAt   *time.Time `json:"created_at"`
}

type ExpenseInput struct {
	Description string  `json:"description"`
	PayerID     int     `json:"payer_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required"`
}

type UpdateExpenseInput struct {
	Description string  `json:"description"`
	PayerID     int     `json:"payer_id"`
	Amount      float64 `json:"amount"`
}
