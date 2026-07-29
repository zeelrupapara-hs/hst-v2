package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// TokenBytes is the entropy of an opaque refresh token.
const TokenBytes = 32

// GenerateToken returns a url safe 256 bit random token.
func GenerateToken() (string, error) {
	b := make([]byte, TokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// HashToken returns the sha256 of a token.
func HashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

// HashTokenHex is HashToken in the form used for redis keys.
func HashTokenHex(token string) string {
	return hex.EncodeToString(HashToken(token))
}
