// Package service содержит реализацию провайдера коротких кодов.
package service

import (
	"crypto/rand"
	"math/big"
)

// ShortCodeProvider представляет интерфейс для провайдера коротких кодов.
type ShortCodeProvider interface {
	// Get возвращает короткий код заданной длины.
	Get(length int) string
}

// RandomShortCodeProvider представляет реализацию провайдера коротких кодов на основе случайных чисел.
type RandomShortCodeProvider struct{}

// chars представляет строку с допустимыми символами для короткого кода.
const (
	chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// Get возвращает короткий код заданной длины, сгенерированный на основе случайных чисел.
func (p *RandomShortCodeProvider) Get(length int) string {
	b := make([]byte, length)
	max := big.NewInt(int64(len(chars)))

	for i := range b {
		n, _ := rand.Int(rand.Reader, max)
		b[i] = chars[n.Int64()]
	}
	return string(b)
}
