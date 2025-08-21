package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// tokenClaims holds custom fields used in JWT serialization for access tokens
// It avoids duplicating "sub" by using RegisteredClaims.Subject for the subject
type tokenClaims struct {
	TenantID string   `json:"tenant_id,omitempty"`
	Roles    []string `json:"roles,omitempty"`
	jwt.RegisteredClaims
}

// MintAccess mints a signed JWT access token and returns the token string and its expiry
func (c Config) MintAccess(userID, tenantID string, roles []string) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(c.AccessTTL)

	cl := tokenClaims{
		TenantID: tenantID,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    c.JWTIssuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, cl)
	signed, err := tok.SignedString(c.JWTSecret)
	return signed, exp, err
}

// NewOpaque returns a random hex string and its SHA-256 hex digest
func NewOpaque(n int) (opaque, hashHex string, err error) {
	b := make([]byte, n)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	opaque = hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(opaque))
	return opaque, hex.EncodeToString(sum[:]), nil
}

// ParseAccess validates and parses a JWT access token into AccessClaims
func (c Config) ParseAccess(raw string) (*AccessClaims, error) {
	keyFunc := func(*jwt.Token) (interface{}, error) { return c.JWTSecret, nil }

	t, err := jwt.ParseWithClaims(raw, &tokenClaims{}, keyFunc)
	if err != nil || !t.Valid {
		return nil, errors.New("invalid token")
	}
	tc, ok := t.Claims.(*tokenClaims)
	if !ok {
		return nil, errors.New("invalid token")
	}

	// Map internal JWT claims to public
	return &AccessClaims{
		Sub:      tc.Subject,
		TenantID: tc.TenantID,
		Roles:    tc.Roles,
	}, nil
}
