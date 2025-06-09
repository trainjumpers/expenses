package models

import (
	"time"
)

// CreateTransactionInput is used for creating a new transaction
type CreateTransactionInput struct {
	Name        string    `json:"name" binding:"required,min=1,max=200"`
	Description string    `json:"description" binding:"max=1000"`
	Amount      *float64  `json:"amount" binding:"required"`
	Date        time.Time `json:"date" binding:"required"`
	CreatedBy   int64     `json:"created_by"`
}

// UpdateTransactionInput is used for updating an existing transaction
type UpdateTransactionInput struct {
	Name        string    `json:"name" binding:"omitempty,min=1,max=200"`
	Description *string   `json:"description" binding:"omitempty,max=1000"`
	Amount      *float64  `json:"amount" binding:"omitempty"`
	Date        time.Time `json:"date" binding:"omitempty"`
}

// TransactionResponse is the response model for a transaction
type TransactionResponse struct {
	Id          int64     `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Amount      float64   `json:"amount"`
	Date        time.Time `json:"date"`
	CreatedBy   int64     `json:"created_by"`
}
