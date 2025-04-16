package models

import "time"

type Profit struct {
	ProfitID  uint    `gorm:"primaryKey"`
	Email     string  `gorm:"not null"`
	Amount    float64 `gorm:"not null"`
	Source    string  `gorm:"not null"` // e.g., "daily profit"
	CreatedAt time.Time
	Date      time.Time `gorm:"not null"` // Store the date profits were generated
}
