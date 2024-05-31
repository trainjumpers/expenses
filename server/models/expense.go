package models

import "time"

type Expense struct {
	ID          int64      `json:"id"`
	Description string     `json:"description"`
	PayerID     int64      `json:"payer_id"`
	Amount      float64    `json:"amount"`
	CreatedBy   int64      `json:"created_by"`
	CreatedAt   *time.Time `json:"created_at"`
}

type ExpenseUserMapping struct {
	ID        int64 `json:"id"`
	ExpenseID int64 `json:"expense_id"`
	UserID    int64 `json:"user_id"`
	Amount    int64 `json:"amount"`
}
