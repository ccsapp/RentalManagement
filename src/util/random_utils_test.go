package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateRandomString_10(t *testing.T) {
	length := 10
	s := GenerateRandomString(length)
	assert.Equal(t, length, len(s))
}

func TestGenerateRandomString_empty(t *testing.T) {
	length := 0
	s := GenerateRandomString(length)
	assert.Equal(t, length, len(s))
}
