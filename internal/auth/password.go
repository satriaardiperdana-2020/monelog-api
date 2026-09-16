// Package auth contains cryptographic primitives shared by authentication flows.
package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/argon2"
)

const (
	minPasswordBytes = 12
	maxPasswordBytes = 128
	saltLength       = 16
	keyLength        = 32
	argonMemory      = 64 * 1024
	argonIterations  = 3
	argonParallelism = 1
)

var ErrInvalidPassword = errors.New("password does not meet policy")

// ValidatePassword enforces the password policy without altering its bytes.
func ValidatePassword(password string) error {
	if !utf8.ValidString(password) || len(password) < minPasswordBytes || len(password) > maxPasswordBytes {
		return ErrInvalidPassword
	}
	return nil
}

// HashPassword returns a PHC-formatted Argon2id hash.
func HashPassword(password string) (string, error) {
	if err := ValidatePassword(password); err != nil {
		return "", err
	}
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}
	hash := argon2.IDKey([]byte(password), salt, argonIterations, argonMemory, argonParallelism, keyLength)
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", argonMemory, argonIterations, argonParallelism,
		base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(hash)), nil
}

// VerifyPassword checks a password against the only PHC parameters supported by this service.
func VerifyPassword(encoded, password string) bool {
	if !utf8.ValidString(password) {
		return false
	}
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" || parts[2] != "v=19" || parts[3] != fmt.Sprintf("m=%d,t=%d,p=%d", argonMemory, argonIterations, argonParallelism) {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) != saltLength {
		return false
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(want) != keyLength {
		return false
	}
	got := argon2.IDKey([]byte(password), salt, argonIterations, argonMemory, argonParallelism, keyLength)
	return subtle.ConstantTimeCompare(got, want) == 1
}

// PasswordParameters exposes the fixed parameters for tests and diagnostics that contain no secrets.
func PasswordParameters() (memory, iterations uint32, parallelism uint8) {
	return argonMemory, argonIterations, argonParallelism
}

// ParsePasswordLength accepts decimal configuration only in tests that need the policy bounds.
func ParsePasswordLength(value string) (int, error) {
	return strconv.Atoi(value)
}
