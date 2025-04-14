package utils

import (
	"math/rand"
	"time"
)

func GenerateReferralCode() *uint {
	rand.Seed(time.Now().UnixNano())
	code := uint(rand.Intn(900000) + 100000) // Generates a 6-digit uint
	return &code
}
