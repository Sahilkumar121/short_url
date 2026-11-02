package service

import (
	"crypto/rand"
	"math/big"
)

// generateShortCode: generates the short code of length 7
func GenerateShortCode() (string, error) {

	const alphabets = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 7

	result := make([]byte, length)

	for i := range length {

		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabets))))
		if err != nil {
			return "", err
		}

		result[i] = alphabets[n.Int64()]
		i++
	}

	return string(result), nil
}
