package models

import "gorm.io/gorm"

type Admin struct {
	gorm.Model
	AdminID  uint
	Email    string `json:"email" gorm:"unique"`
	Password string `json:"password"`
}
