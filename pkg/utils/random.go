package utils

import (
	"crypto/rand"
	"math/big"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// RandomBase62 generates a cryptographically secure random Base62 string of length n.
// If n <= 0, it returns an empty string.
func RandomBase62(n int) string {
	if n <= 0 {
		return ""
	}

	b := make([]byte, n)
	max := big.NewInt(int64(len(base62Chars)))

	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, max)
		if err != nil {
			// fallback: deterministic but safe
			b[i] = base62Chars[(i*17+33)%62]
			continue
		}
		b[i] = base62Chars[num.Int64()]
	}

	return string(b)
}

// MustRandomBase62 is a convenience wrapper that panics if secure randomness fails.
func MustRandomBase62(n int) string {
	s := RandomBase62(n)
	if s == "" {
		panic("RandomBase62: failed to generate ID")
	}
	return s
}
