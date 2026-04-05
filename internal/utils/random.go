package utils

import (
	"math/rand"
	"time"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec

// RandomSellerID возвращает случайный sellerID в диапазоне 111111–999999,
// чтобы не пересекаться с другими тестировщиками.
func RandomSellerID() int {
	const (
		minSellerID = 111111
		maxSellerID = 999999
	)
	return minSellerID + rng.Intn(maxSellerID-minSellerID+1)
}

// RandomString возвращает случайную строку из строчных латинских букв длиной n.
func RandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rng.Intn(len(letters))]
	}
	return string(b)
}

// RandomIntN возвращает случайное целое число в диапазоне [0, n).
func RandomIntN(n int) int {
	return rng.Intn(n)
}
