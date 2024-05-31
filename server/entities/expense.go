package entities

type ExpenseInput struct {
	Description   string            `json:"description"`
	PayerID       int64             `json:"payer_id" binding:"required"`
	Amount        float64           `json:"amount" binding:"required"`
	Contributions map[int64]float64 `json:"contributions" binding:"required"`
}

type UpdateExpenseInput struct {
	Description string  `json:"description"`
	PayerID     int64   `json:"payer_id"`
	Amount      float64 `json:"amount"`
}
