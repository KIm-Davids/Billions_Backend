package utils

import (
	"crypto/rand"
	"math/big"
)

const addressCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateAddress(length int) (string, error) {
	address := make([]byte, length)
	for i := range address {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(addressCharset))))
		if err != nil {
			return "", err
		}
		address[i] = addressCharset[num.Int64()]
	}
	return string(address), nil
}
