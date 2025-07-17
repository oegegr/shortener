package service

import (
	"crypto/rand"
	"math/big"
)

type ShortCodeProvider interface {
	Get(length int) string
}

type RandomShortCodeProvider struct{}

const (
	chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func (p *RandomShortCodeProvider) Get(length int) string {
	b := make([]byte, length)
	max := big.NewInt(int64(len(chars)))

	for i := range b {
		n, _ := rand.Int(rand.Reader, max)
		b[i] = chars[n.Int64()]
	}
	return string(b)
}
