package helpers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"math/big"
)

func Contains(slice []string, target string) bool {
	for _, element := range slice {
		if element == target {
			return true // Element found
		}
	}
	return false // Element not found
}

func Ternary[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

func GenerateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
