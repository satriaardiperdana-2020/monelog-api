package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const MaxAccessTokenLifetime = 24 * time.Hour

var ErrInvalidAccessToken = errors.New("invalid access token")

// AccessClaims are the complete claims accepted for an access token.
type AccessClaims struct {
	TokenType string `json:"typ"`
	// Role is intentionally never issued. Authorization always loads the
	// actor's current role from the database, so a role-bearing token is not
	// accepted as an access token.
	Role string `json:"role,omitempty"`
	jwt.RegisteredClaims
}

// JWT issues and validates only HS256 access tokens.
type JWT struct {
	key      []byte
	issuer   string
	audience string
	lifetime time.Duration
	now      func() time.Time
}

func NewJWT(key []byte, issuer, audience string, lifetime time.Duration) (*JWT, error) {
	if len(key) < 32 || issuer == "" || audience == "" || lifetime <= 0 || lifetime > MaxAccessTokenLifetime {
		return nil, errors.New("invalid JWT configuration")
	}
	return &JWT{key: append([]byte(nil), key...), issuer: issuer, audience: audience, lifetime: lifetime, now: time.Now}, nil
}

func (j *JWT) Issue(subject uuid.UUID) (string, time.Time, error) {
	issuedAt := j.now().UTC()
	expiresAt := issuedAt.Add(j.lifetime)
	claims := AccessClaims{
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   subject.String(),
			Audience:  jwt.ClaimStrings{j.audience},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ID:        uuid.NewString(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(j.key)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}
	return signed, expiresAt, nil
}

func (j *JWT) Validate(raw string) (uuid.UUID, error) {
	claims := AccessClaims{}
	token, err := jwt.ParseWithClaims(raw, &claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 || token.Header["alg"] != jwt.SigningMethodHS256.Alg() {
			return nil, ErrInvalidAccessToken
		}
		return j.key, nil
	}, jwt.WithIssuer(j.issuer), jwt.WithAudience(j.audience), jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithLeeway(30*time.Second), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || token == nil || !token.Valid || claims.TokenType != "access" || claims.Role != "" || claims.ExpiresAt == nil || claims.IssuedAt == nil || claims.ID == "" {
		return uuid.Nil, ErrInvalidAccessToken
	}
	subject, err := uuid.Parse(claims.Subject)
	if err != nil || subject == uuid.Nil {
		return uuid.Nil, ErrInvalidAccessToken
	}
	if id, err := uuid.Parse(claims.ID); err != nil || id == uuid.Nil {
		return uuid.Nil, ErrInvalidAccessToken
	}
	issuedAt := claims.IssuedAt.Time
	expiresAt := claims.ExpiresAt.Time
	if expiresAt.Before(issuedAt) || expiresAt.Equal(issuedAt) || expiresAt.After(issuedAt.Add(j.lifetime+30*time.Second)) || issuedAt.After(j.now().Add(30*time.Second)) {
		return uuid.Nil, ErrInvalidAccessToken
	}
	return subject, nil
}
