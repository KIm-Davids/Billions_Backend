package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	//UserID   uint
	Username   string  `json:"username"`
	Email      string  `json:"email" gorm:"unique"`
	Password   string  `json:"password"`
	ReferID    string  `json:"referrerId" gorm:"unique"`
	ReferredBy string  `json:"referred_by"` // referrer's ID (not unique)
	Balance    float64 `json:"balance"`
	Package    string  `json:"package,omitempty"`
	Profit     float64 `gorm:"default:0"`
}
