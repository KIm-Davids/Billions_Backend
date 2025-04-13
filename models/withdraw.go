package models

import "time"

type Withdraw struct {
	UserID        uint      `json:"user_id"`
	SenderName    string    `json:"senderName"`
	SenderAddress string    `json:"senderAddress"`
	Status        string    `json:"status"`
	WalletType    string    `json:"walletType"`
	Amount        float64   `json:"amount"`
	CreatedAt     time.Time `json:"created_at"`
	Description   string    `json:"description"`
}
