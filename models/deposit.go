package models

import "time"

type Deposit struct {
	//UserID      string    `json:"user_id"`
	Email       string    `gorm:"primaryKey" json:"email"`
	Hash        string    `json:"hash" gorm:"unique"`
	Status      string    `json:"status"`
	PackageType string    `json:"packageType"`
	Amount      float64   `json:"amount"`
	CreatedAt   time.Time `json:"created_at"`
}
