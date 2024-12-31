package models

import "time"

type Expense struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	PayerID     int64      `json:"payer_id"`
	Amount      float64    `json:"amount"`
	CreatedBy   int64      `json:"created_by"`
	CreatedAt   *time.Time `json:"created_at"`
	UniqueId    string     `json:"unique_id"`
}

type ExpenseUserMapping struct {
	ID        int64   `json:"id"`
	ExpenseID int64   `json:"expense_id"`
	UserID    int64   `json:"user_id"`
	Amount    float64 `json:"amount"`
}

type ExpenseWithContribution struct {
	Expense
	UserAmount float64 `json:"user_amount"`
}

type ExpenseWithAllContributions struct {
	Expense
	ContributorId   int64   `json:"contributor_id"`
	ContributorName string  `json:"contributor_name"`
	Contribution    float64 `json:"contribution"`
}
