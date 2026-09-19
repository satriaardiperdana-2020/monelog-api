package auth

import (
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestJWTIssuesAndStrictlyValidatesAccessTokens(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	issuer := "monelog-api"
	audience := "monelog-app"
	validator, err := NewJWT(key, issuer, audience, MaxAccessTokenLifetime)
	if err != nil {
		t.Fatalf("NewJWT() error = %v", err)
	}
	var subject int64 = 42
	raw, expiresAt, err := validator.Issue(subject)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if got, err := validator.Validate(raw); err != nil || got != subject {
		t.Fatalf("Validate() = %v, %v; want %v, nil", got, err, subject)
	}
	if remaining := time.Until(expiresAt); remaining < 24*time.Hour-2*time.Second || remaining > 24*time.Hour+time.Second {
		t.Fatalf("access lifetime = %s, want 24 hours", remaining)
	}

	baseClaims := AccessClaims{TokenType: "access", RegisteredClaims: jwt.RegisteredClaims{Issuer: issuer, Subject: strconv.FormatInt(subject, 10), Audience: jwt.ClaimStrings{audience}, ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)), IssuedAt: jwt.NewNumericDate(time.Now()), ID: uuid.NewString()}}
	tests := []struct {
		name   string
		claims AccessClaims
		method jwt.SigningMethod
		key    []byte
	}{
		{name: "invalid signature", claims: baseClaims, method: jwt.SigningMethodHS256, key: []byte("other-signing-key-which-is-at-least-32-bytes")},
		{name: "wrong algorithm", claims: baseClaims, method: jwt.SigningMethodHS384, key: key},
		{name: "wrong issuer", claims: withIssuer(baseClaims, "other-issuer"), method: jwt.SigningMethodHS256, key: key},
		{name: "wrong audience", claims: withAudience(baseClaims, "other-audience"), method: jwt.SigningMethodHS256, key: key},
		{name: "expired", claims: withExpiration(baseClaims, time.Now().Add(-time.Minute)), method: jwt.SigningMethodHS256, key: key},
		{name: "role claim", claims: withRole(baseClaims, "admin"), method: jwt.SigningMethodHS256, key: key},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw, err := jwt.NewWithClaims(test.method, test.claims).SignedString(test.key)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := validator.Validate(raw); err == nil {
				t.Fatal("Validate() accepted an invalid token")
			}
		})
	}
}

func withIssuer(claims AccessClaims, issuer string) AccessClaims {
	claims.Issuer = issuer
	return claims
}

func withAudience(claims AccessClaims, audience string) AccessClaims {
	claims.Audience = jwt.ClaimStrings{audience}
	return claims
}

func withExpiration(claims AccessClaims, expiration time.Time) AccessClaims {
	claims.ExpiresAt = jwt.NewNumericDate(expiration)
	return claims
}

func withRole(claims AccessClaims, role string) AccessClaims {
	claims.Role = role
	return claims
}
