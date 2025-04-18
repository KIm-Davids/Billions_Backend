package models

import "time"

type Withdraw struct {
	//UserID uint `json:"user_id"`
	//SenderName    string    `json:"senderName"`
	WithdrawAddress string    `json:"withdrawAddress"`
	Email           string    `json:"email"`
	WithdrawID      uint      `gorm:"primaryKey;autoIncrement" json:"withdrawId"` // ✅ FIXED HERE
	Status          string    `json:"status"`
	WalletType      string    `json:"walletType"`
	Amount          float64   `json:"amount"`
	CreatedAt       time.Time `json:"-"`
	Description     string    `json:"description"`
	Source          string
	ProfitType      string `json:"profitType"`
	//WithdrawDate
}
