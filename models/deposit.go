package models

import "time"

type Deposit struct {
	DepositID uint `gorm:"primaryKey;autoIncrement"`
	//UserID      uint      `json:"user_id"`
	Email       string    `json:"email"  gorm:"unique"`
	Hash        string    `json:"hash" gorm:"unique"`
	Status      string    `json:"status"`
	PackageType string    `json:"packageType"`
	Amount      float64   `json:"amount"`
	CreatedAt   time.Time `json:"created_at"`
}
