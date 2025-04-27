package models

import "time"

type Profit struct {
	ProfitID        uint   `gorm:"primaryKey"`
	Email           string `gorm:"not null"`
	Source          string `gorm:"not null"` // e.g., "daily profit"
	NetProfitStatus string `gorm:"not null"`
	ProfitDate      string
	CreatedAt       time.Time
	Date            time.Time `gorm:"not null"` // Store the date profits were generated
	NewProfit       float64   `gorm:"default:0"`
}
