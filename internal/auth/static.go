package auth

import (
	"context"
	"crypto/subtle"
	"errors"

	"github.com/pebblr/pebblr/internal/domain"
)

// StaticAuthenticator validates tokens by constant-time comparison against a
// shared secret. Used for local development and E2E testing.
type StaticAuthenticator struct {
	token []byte
}

// NewStaticAuthenticator creates an authenticator that accepts the given token.
func NewStaticAuthenticator(token string) *StaticAuthenticator {
	return &StaticAuthenticator{token: []byte(token)}
}

// ValidateToken returns fixed admin claims when the token matches the secret.
func (s *StaticAuthenticator) ValidateToken(_ context.Context, token string) (*UserClaims, error) {
	if subtle.ConstantTimeCompare([]byte(token), s.token) != 1 {
		return nil, errors.New("invalid token")
	}
	return &UserClaims{
		Sub:     "a0000000-0000-0000-0000-000000000001",
		Email:   "admin@pebblr.dev",
		Name:    "Alex Admin",
		Roles:   []domain.Role{domain.RoleAdmin},
		TeamIDs: []string{},
	}, nil
}
