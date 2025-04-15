package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	//UserID   uint
	Username   string  `json:"username"`
	Email      string  `json:"email" gorm:"unique"`
	Password   string  `json:"password"`
	ReferID    string  `json:"referrerId"`
	ReferredBy string  `json:"referrer" gorm:"unique"`
	Balance    float64 `json:"balance"`
	Package    string  `json:"package,omitempty"`
}
