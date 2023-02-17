package util

import (
	"crypto/rand"
	"math/big"
)

var alphanumericRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// GenerateRandomString generates a cryptographically secure random alphanumeric string of the given length.
func GenerateRandomString(length int) string {
	b := make([]rune, length)
	for i := range b {
		randInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphanumericRunes))))
		if err != nil {
			panic(err)
		}
		b[i] = alphanumericRunes[randInt.Int64()]
	}
	return string(b)
}
