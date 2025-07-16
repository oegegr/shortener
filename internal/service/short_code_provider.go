package service

import (
	"crypto/rand"
	"math/big"
)

const (
    chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func GenerateShortCode(length int) string {
	b := make([]byte, length)
	max := big.NewInt(int64(len(chars)))
	
	for i := range b {
		n, _ := rand.Int(rand.Reader, max)
		b[i] = chars[n.Int64()]
	}
	return string(b)
}