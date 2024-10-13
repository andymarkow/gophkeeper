// Package auth provides JWT authentication.
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// JWTAuth represents JWT authentication.
type JWTAuth struct {
	secret   []byte
	issuer   string
	tokenTTL time.Duration
}

// Claims represents JWT claims.
type Claims struct {
	jwt.RegisteredClaims
}

// NewJWTAuth creates new JWTAuth.
func NewJWTAuth(secret []byte, opts ...Option) *JWTAuth {
	a := &JWTAuth{
		secret:   secret,
		tokenTTL: 24 * time.Hour,
		issuer:   "gophkeeper",
	}

	for _, opt := range opts {
		opt(a)
	}

	return a
}

// Option is a functional option type for JWTAuth.
type Option func(a *JWTAuth)

// WithIssuer sets JWT issuer.
func WithIssuer(issuer string) Option {
	return func(a *JWTAuth) {
		a.issuer = issuer
	}
}

// WithTokenTTL sets JWT token TTL.
func WithTokenTTL(ttl time.Duration) Option {
	return func(a *JWTAuth) {
		a.tokenTTL = ttl
	}
}

// CreateJWTString creates new JWT token.
func (a *JWTAuth) CreateJWTString(subject string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    a.issuer,
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.tokenTTL)),
		},
	})

	tokenString, err := token.SignedString(a.secret)
	if err != nil {
		return "", fmt.Errorf("token.SignedString: %w", err)
	}

	return tokenString, nil
}
