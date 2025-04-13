package models

import "time"

type Deposit struct {
	UserID      uint      `json:"user_id"`
	SenderName  string    `json:"senderName"`
	Hash        string    `json:"hash"`
	Status      string    `json:"status"`
	PackageType string    `json:"packageType"`
	Amount      float64   `json:"amount"`
	CreatedAt   time.Time `json:"created_at"`
}
