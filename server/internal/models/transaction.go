package models

import (
	"time"
)

type CreateTransactionInput struct {
	Name        string    `json:"name" binding:"required,min=1,max=200"`
	Description string    `json:"description" binding:"max=1000"`
	Amount      float64   `json:"amount" binding:"required"`
	Date        time.Time `json:"date" binding:"required"`
	CreatedBy   int64     `json:"created_by"`
}

type UpdateTransactionInput struct {
	Name        string    `json:"name,omitempty" binding:"omitempty,min=1,max=200"`
	Description *string   `json:"description,omitempty" binding:"omitempty,max=1000"`
	Amount      *float64  `json:"amount,omitempty" binding:"omitempty"`
	Date        time.Time `json:"date,omitempty"`
}

type TransactionResponse struct {
	Id          int64     `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Amount      float64   `json:"amount"`
	Date        time.Time `json:"date"`
	CreatedBy   int64     `json:"created_by"`
}
