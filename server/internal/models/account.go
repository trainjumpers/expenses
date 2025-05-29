package models

type BankType string

const (
	BankTypeInvestment BankType = "investment"
	BankTypeAxis       BankType = "Axis"
	BankTypeSBI        BankType = "SBI"
	BankTypeHDFC       BankType = "HDFC"
	BankTypeICICI      BankType = "ICICI"
)

const (
	CurrencyINR = "INR"
	CurrencyUSD = "USD"
)

type CreateAccountInput struct {
	Name      string   `json:"name" binding:"required"`
	BankType  BankType `json:"bank_type" binding:"required,oneof=investment Axis SBI HDFC ICICI"`
	Currency  string   `json:"currency" binding:"required,oneof=INR USD"`
	Balance   *float64 `json:"balance"`
	CreatedBy int64    `json:"created_by"`
}

type UpdateAccountInput struct {
	Name     string   `json:"name,omitempty"`
	BankType BankType `json:"bank_type,omitempty"`
	Currency string   `json:"currency,omitempty"`
	Balance  *float64 `json:"balance,omitempty"`
}

type AccountResponse struct {
	Id        int64    `json:"id"`
	Name      string   `json:"name"`
	BankType  BankType `json:"bank_type"`
	Currency  string   `json:"currency"`
	Balance   float64  `json:"balance"`
	CreatedBy int64    `json:"created_by"`
}
