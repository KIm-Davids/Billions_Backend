package utils

import (
	"JWTProject/initializers"
	"JWTProject/models"
	"crypto/rand"
	"fmt"
	"math/big"
)

const addressCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateUniqueReferralID(length int) (string, error) {
	maxAttempts := 10

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Generate random string
		address := make([]byte, length)
		for i := range address {
			num, err := rand.Int(rand.Reader, big.NewInt(int64(len(addressCharset))))
			if err != nil {
				return "", err
			}
			address[i] = addressCharset[num.Int64()]
		}
		id := string(address)

		// Check for duplicate in DB
		var count int64
		err := initializers.DB.Model(&models.User{}).Where("referral_id = ?", id).Count(&count).Error
		if err != nil {
			return "", err
		}

		if count == 0 {
			// Unique ID found
			return id, nil
		}
		// Else: duplicate found, retry
	}

	return "", fmt.Errorf("could not generate a unique referral ID after %d attempts", maxAttempts)
}
