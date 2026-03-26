// Package demo implements the auth.Authenticator interface for self-service
// demo environments. It issues and validates HMAC-SHA256 signed JWTs so
// prospects can explore the app without an external identity provider.
package demo

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/pebblr/pebblr/internal/auth"
	"github.com/pebblr/pebblr/internal/domain"
)

// DefaultTokenTTL is how long demo tokens remain valid.
const DefaultTokenTTL = 24 * time.Hour

// Authenticator issues and validates demo JWTs signed with HMAC-SHA256.
type Authenticator struct {
	signingKey []byte
	tokenTTL   time.Duration
}

// New creates a demo Authenticator. If signingKey is nil, a random 256-bit
// key is generated (appropriate for ephemeral demo instances).
func New(signingKey []byte) (*Authenticator, error) {
	if len(signingKey) == 0 {
		signingKey = make([]byte, 32)
		if _, err := rand.Read(signingKey); err != nil {
			return nil, fmt.Errorf("demo: generating signing key: %w", err)
		}
	}
	return &Authenticator{
		signingKey: signingKey,
		tokenTTL:   DefaultTokenTTL,
	}, nil
}

// IssueToken creates a signed JWT for a demo persona.
func (a *Authenticator) IssueToken(persona Persona) (string, error) {
	if !persona.Role.Valid() {
		return "", fmt.Errorf("demo: invalid role %q", persona.Role)
	}

	now := time.Now()
	claims := demoClaims{
		Sub:     persona.ID,
		Email:   persona.Email,
		Name:    persona.Name,
		Role:    string(persona.Role),
		TeamIDs: persona.TeamIDs,
		Iat:     now.Unix(),
		Exp:     now.Add(a.tokenTTL).Unix(),
	}

	return a.sign(claims)
}

// ValidateToken validates a demo JWT and returns the extracted claims.
func (a *Authenticator) ValidateToken(_ context.Context, token string) (*auth.UserClaims, error) {
	claims, err := a.verify(token)
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	if claims.Exp <= now {
		return nil, errors.New("demo: token expired")
	}

	role := domain.Role(claims.Role)
	if !role.Valid() {
		role = domain.RoleRep
	}

	teamIDs := claims.TeamIDs
	if teamIDs == nil {
		teamIDs = []string{}
	}
	return &auth.UserClaims{
		Sub:     claims.Sub,
		Email:   claims.Email,
		Name:    claims.Name,
		Roles:   []domain.Role{role},
		TeamIDs: teamIDs,
	}, nil
}

// Persona describes a demo user identity.
type Persona struct {
	ID      string      `json:"id"`
	Email   string      `json:"email"`
	Name    string      `json:"name"`
	Role    domain.Role `json:"role"`
	TeamIDs []string    `json:"teamIds,omitempty"`
}

// demoClaims are the JWT payload fields for demo tokens.
type demoClaims struct {
	Sub     string   `json:"sub"`
	Email   string   `json:"email"`
	Name    string   `json:"name"`
	Role    string   `json:"role"`
	TeamIDs []string `json:"team_ids,omitempty"`
	Iat     int64    `json:"iat"`
	Exp     int64    `json:"exp"`
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

func (a *Authenticator) sign(claims demoClaims) (string, error) {
	header := jwtHeader{Alg: "HS256", Typ: "JWT"}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)
	signingInput := headerB64 + "." + claimsB64

	mac := hmac.New(sha256.New, a.signingKey)
	mac.Write([]byte(signingInput))
	sig := mac.Sum(nil)

	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func (a *Authenticator) verify(token string) (*demoClaims, error) {
	parts := splitToken(token)
	if parts == nil {
		return nil, errors.New("demo: malformed JWT")
	}

	signingInput := parts[0] + "." + parts[1]

	sigBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("demo: decoding signature: %w", err)
	}

	mac := hmac.New(sha256.New, a.signingKey)
	mac.Write([]byte(signingInput))
	expected := mac.Sum(nil)

	if !hmac.Equal(sigBytes, expected) {
		return nil, errors.New("demo: invalid signature")
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("demo: decoding payload: %w", err)
	}

	var claims demoClaims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, fmt.Errorf("demo: parsing claims: %w", err)
	}

	return &claims, nil
}

// splitToken splits a JWT into its three parts. Returns nil if malformed.
func splitToken(token string) []string {
	var parts [3]string
	idx := 0
	start := 0
	for i := range len(token) {
		if token[i] == '.' {
			if idx >= 2 {
				return nil
			}
			parts[idx] = token[start:i]
			idx++
			start = i + 1
		}
	}
	if idx != 2 {
		return nil
	}
	parts[2] = token[start:]
	if parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return nil
	}
	return parts[:]
}
