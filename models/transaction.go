package models

import "time"

type Transaction struct {
	UserID        uint      `json:"user_id"`
	SenderName    string    `json:"senderName"`
	SenderAddress string    `json:"senderAddress"`
	Type          string    `json:"transactionType"`
	Status        string    `json:"status"`
	PackageType   string    `json:"packageType"`
	Amount        float64   `json:"amount"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}
