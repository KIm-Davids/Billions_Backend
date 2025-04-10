package models

import "time"

type Transaction struct {
	UserID        uint      `json:"user_id"`
	SenderName    string    `json:"sender-name"`
	SenderAddress string    `json:"sender-address"`
	Type          string    `json:"transaction-type"`
	Status        string    `json:"status"`
	Amount        float64   `json:"amount"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}
