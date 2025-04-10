package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `json:"username"`
	Email    string `json:"email" gorm:"unique"`
	Password string `json:"password"`
	Address  string `json:"address"` //gorm:"unique"`
	Role     string `json:"role"`
}
