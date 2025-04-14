package models

import "time"

type Profit struct {
	ID        uint    `gorm:"primaryKey"`
	UserID    uint    `gorm:"not null"`
	Amount    float64 `gorm:"not null"`
	Source    string  `gorm:"not null"` // e.g., "daily profit"
	CreatedAt time.Time
}
