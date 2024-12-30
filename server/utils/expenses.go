package utils

import (
	"encoding/base64"
	"expenses/models"
	"fmt"
	"time"
)

func UniqueIdentifierExpense(expense models.Expense) string {
	if expense.CreatedAt == nil {
		now := time.Now()
		expense.CreatedAt = &now
	}
	year, month, day := expense.CreatedAt.Date()
	return base64.StdEncoding.EncodeToString([]byte(
		fmt.Sprintf("%d_%02d_%02d_%d_%s_%s_%f",
			year, month, day, expense.PayerID,
			expense.Name,
			expense.Description,
			expense.Amount,
		),
	))
}