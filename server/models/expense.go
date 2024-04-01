package models

type Expense struct {
	ID          int     `json:"id"`
	Description string  `json:"description"`
	PayerID     int     `json:"payer_id"`
	Amount      float64 `json:"amount"`
}

type ExpenseInput struct {
	Description string  `json:"description"`
	PayerID     int     `json:"payer_id"`
	Amount      float64 `json:"amount"`
}
