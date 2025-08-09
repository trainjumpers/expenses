package models

type BankType string

const (
	BankTypeInvestment  BankType = "investment"
	BankTypeAxis        BankType = "axis"
	BankTypeSBI         BankType = "sbi"
	BankTypeHDFC        BankType = "hdfc"
	BankTypeICICI       BankType = "icici"
	BankTypeICICICredit BankType = "icici_credit"
	BankTypeOthers      BankType = "others"
)

const (
	CurrencyINR = "inr"
	CurrencyUSD = "usd"
)

type CreateAccountInput struct {
	Name      string   `json:"name" binding:"required"`
	BankType  BankType `json:"bank_type" binding:"required,oneof=investment axis sbi hdfc icici icici_credit others"`
	Currency  string   `json:"currency" binding:"required,oneof=inr usd"`
	Balance   *float64 `json:"balance"`
	CreatedBy int64    `json:"created_by" binding:"required"`
}

type UpdateAccountInput struct {
	Name     string   `json:"name,omitempty"`
	BankType BankType `json:"bank_type,omitempty" binding:"omitempty,oneof=investment axis sbi hdfc icici icici_credit others"`
	Currency string   `json:"currency,omitempty" binding:"omitempty,oneof=inr usd"`
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
