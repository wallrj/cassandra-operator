package e2e

import (
	"math/rand"
	"time"
)

var (
	chars = "abcdefghijklmnopqrstuvwxyz1234567890"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
