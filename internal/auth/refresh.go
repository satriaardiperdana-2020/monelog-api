package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

const refreshSecretLength = 32

// NewRefreshSecret creates an opaque token suitable for transport to one client.
func NewRefreshSecret() (string, error) {
	bytes := make([]byte, refreshSecretLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate refresh secret: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// HashRefreshSecret returns the only value persisted for a refresh token.
func HashRefreshSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

// NewCSRFToken creates the browser double-submit token.
func NewCSRFToken() (string, error) {
	return NewRefreshSecret()
}

// EqualSecret compares transport secrets without an early-exit comparison.
func EqualSecret(left, right string) bool {
	return subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}
