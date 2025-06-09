package models

import (
	"time"
)

// CreateBaseTransactionInput is used for DB insert (without mapping fields)
type CreateBaseTransactionInput struct {
	Name        string    `json:"name" binding:"required,min=1,max=200"`
	Description string    `json:"description" binding:"max=1000"`
	Amount      *float64  `json:"amount" binding:"required"`
	Date        time.Time `json:"date" binding:"required"`
	CreatedBy   int64     `json:"created_by"`
	AccountId   int64     `json:"account_id" binding:"required"`
}

// UpdateBaseTransactionInput is used for updating DB update (without mapping fields)
type UpdateBaseTransactionInput struct {
	Name        string    `json:"name" binding:"omitempty,min=1,max=200"`
	Description *string   `json:"description" binding:"omitempty,max=1000"`
	Amount      *float64  `json:"amount" binding:"omitempty"`
	Date        time.Time `json:"date" binding:"omitempty"`
	AccountId   *int64    `json:"account_id"`
}

// TransactionBaseResponse is the base response model for a transaction (without mappings)
type TransactionBaseResponse struct {
	Id          int64     `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Amount      float64   `json:"amount"`
	Date        time.Time `json:"date"`
	CreatedBy   int64     `json:"created_by"`
	AccountId   int64     `json:"account_id"`
}

// CreateTransactionInput is used for creating a new transaction
type CreateTransactionInput struct {
	CreateBaseTransactionInput
	CategoryIds []int64 `json:"category_ids"`
}

// UpdateTransactionInput is used for updating an existing transaction
type UpdateTransactionInput struct {
	UpdateBaseTransactionInput
	CategoryIds *[]int64 `json:"category_ids"`
}

// TransactionResponse is the response model for a transaction
type TransactionResponse struct {
	TransactionBaseResponse
	CategoryIds []int64 `json:"category_ids"`
}
