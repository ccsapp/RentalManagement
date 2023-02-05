package util

import (
	"math/rand"
	"time"
)

func InitRandom() {
	rand.Seed(time.Now().UnixNano())
}

var alphanumericRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// GenerateRandomString generates a random alphanumeric string of the given length.
func GenerateRandomString(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = alphanumericRunes[rand.Intn(len(alphanumericRunes))]
	}
	return string(b)
}
