package models

import "time"

type Withdraw struct {
	//UserID uint `json:"user_id"`
	//SenderName    string    `json:"senderName"`
	//SenderAddress string    `json:"senderAddress"`
	Email       string    `json:"email"`
	WithdrawID  uint      `gorm:"primaryKey;autoIncrement" json:"-"`
	Status      string    `json:"status"`
	WalletType  string    `json:"walletType"`
	Amount      float64   `json:"amount"`
	CreatedAt   time.Time `json:"-"`
	Description string    `json:"description"`
	//WithdrawDate
}
