package entities

import "time"

type Statement struct {
	TrasactionId string    `json:"transaction_id"`
	Date         time.Time `json:"date"`
	Description  string    `json:"description"`
	Amount       float64   `json:"amount"`
	Balance      float64   `json:"balance"`
}
