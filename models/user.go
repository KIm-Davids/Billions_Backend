package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	UserID   uint
	Username string `json:"username"`
	Email    string `json:"email" gorm:"unique"`
	Password string `json:"password"`
	//Address    string `json:"address"` //gorm:"unique"`
	//Role       string `json:"role"`
	ReferrerID *uint   `json:"referrerId"`
	Referer    string  `json:"referrer" gorm:"unique"`
	Balance    float64 `json:"balance"`
	Package    string  `json:"package,omitempty"`
}
