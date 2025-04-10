package models

import "gorm.io/gorm"

type Client struct {
	gorm.Model
	UserID  uint
	User    User    `gorm:"foreignKey:UserID"`
	Referer string  `json:"referrer" gorm:"unique"`
	Balance float64 `json:"balance"`
	Package string  `json:"package,omitempty"`
}
