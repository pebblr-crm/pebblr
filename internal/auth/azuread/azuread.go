// Package azuread implements the auth.Authenticator interface using Azure AD
// (Entra ID) OIDC token validation. It fetches the JWKS from Azure AD's
// discovery endpoint and validates JWT signatures, audience, issuer, and expiry.
package azuread

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pebblr/pebblr/internal/auth"
	"github.com/pebblr/pebblr/internal/domain"
)

// Config holds Azure AD OIDC settings needed to validate tokens.
type Config struct {
	// TenantID is the Azure AD tenant identifier.
	TenantID string
	// ClientID is the application (client) ID — used as the expected audience.
	ClientID string
	// IssuerURL is the expected token issuer (e.g. "https://login.microsoftonline.com/{tenant}/v2.0").
	// If empty, it is derived from TenantID.
	IssuerURL string
}

// minRefreshInterval is the minimum time between JWKS refreshes to prevent
// an attacker from forcing continuous outbound requests with forged kid values.
const minRefreshInterval = 60 * time.Second

// Authenticator validates Azure AD OIDC bearer tokens.
type Authenticator struct {
	clientID  string
	issuerURL string
	jwksURL   string

	mu          sync.RWMutex
	keys        map[string]*rsa.PublicKey
	lastRefresh time.Time

	httpClient *http.Client
}

// New creates an Authenticator that validates tokens issued by Azure AD.
// It fetches the JWKS from Azure AD's OpenID Connect discovery endpoint.
func New(ctx context.Context, cfg Config) (*Authenticator, error) {
	if cfg.TenantID == "" {
		return nil, errors.New("azuread: TenantID is required")
	}
	if cfg.ClientID == "" {
		return nil, errors.New("azuread: ClientID is required")
	}

	issuer := cfg.IssuerURL
	if issuer == "" {
		issuer = "https://login.microsoftonline.com/" + cfg.TenantID + "/v2.0"
	}

	client := &http.Client{Timeout: 10 * time.Second}

	jwksURL, err := discoverJWKS(ctx, client, issuer)
	if err != nil {
		return nil, fmt.Errorf("azuread: discovering JWKS endpoint: %w", err)
	}

	a := &Authenticator{
		clientID:   cfg.ClientID,
		issuerURL:  issuer,
		jwksURL:    jwksURL,
		keys:       make(map[string]*rsa.PublicKey),
		httpClient: client,
	}

	if err := a.refreshKeys(ctx); err != nil {
		return nil, fmt.Errorf("azuread: initial key fetch: %w", err)
	}

	return a, nil
}

// ValidateToken validates a JWT bearer token against Azure AD's public keys
// and returns the extracted user claims.
func (a *Authenticator) ValidateToken(ctx context.Context, token string) (*auth.UserClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("azuread: malformed JWT")
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("azuread: decoding header: %w", err)
	}

	var header jwtHeader
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, fmt.Errorf("azuread: parsing header: %w", err)
	}

	if header.Alg != "RS256" {
		return nil, fmt.Errorf("azuread: unsupported algorithm %q", header.Alg)
	}

	key, err := a.getKey(ctx, header.Kid)
	if err != nil {
		return nil, err
	}

	if err := verifyRS256(parts[0]+"."+parts[1], parts[2], key); err != nil {
		return nil, fmt.Errorf("azuread: signature verification failed: %w", err)
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("azuread: decoding payload: %w", err)
	}

	var claims tokenClaims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, fmt.Errorf("azuread: parsing claims: %w", err)
	}

	now := time.Now().Unix()
	if claims.Exp <= now {
		return nil, errors.New("azuread: token expired")
	}
	if claims.Nbf > now {
		return nil, errors.New("azuread: token not yet valid")
	}
	if claims.Iss != a.issuerURL {
		return nil, fmt.Errorf("azuread: unexpected issuer %q", claims.Iss)
	}
	if claims.Aud != a.clientID {
		return nil, fmt.Errorf("azuread: unexpected audience %q", claims.Aud)
	}

	return claimsToUserClaims(&claims), nil
}

// getKey returns the RSA public key for the given key ID, refreshing the
// key set if the kid is not found and the cooldown has elapsed.
func (a *Authenticator) getKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	a.mu.RLock()
	key, ok := a.keys[kid]
	lastRefresh := a.lastRefresh
	a.mu.RUnlock()
	if ok {
		return key, nil
	}

	// Rate-limit JWKS refreshes to prevent amplification attacks where
	// an attacker sends tokens with random kid values.
	if time.Since(lastRefresh) < minRefreshInterval {
		return nil, fmt.Errorf("azuread: unknown key ID %q (refresh on cooldown)", kid)
	}

	// Key not found — refresh JWKS (key rotation).
	if err := a.refreshKeys(ctx); err != nil {
		return nil, fmt.Errorf("azuread: refreshing keys: %w", err)
	}

	a.mu.RLock()
	key, ok = a.keys[kid]
	a.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("azuread: unknown key ID %q", kid)
	}
	return key, nil
}

// refreshKeys fetches the JWKS from Azure AD and updates the key cache.
func (a *Authenticator) refreshKeys(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.jwksURL, http.NoBody)
	if err != nil {
		return err
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MB limit
	if err != nil {
		return err
	}

	var jwks jwksResponse
	if err := json.Unmarshal(body, &jwks); err != nil {
		return fmt.Errorf("parsing JWKS: %w", err)
	}

	keys := make(map[string]*rsa.PublicKey, len(jwks.Keys))
	for _, k := range jwks.Keys {
		if k.Kty != "RSA" || k.Use != "sig" {
			continue
		}
		pub, err := parseRSAPublicKey(k)
		if err != nil {
			continue
		}
		keys[k.Kid] = pub
	}

	a.mu.Lock()
	a.keys = keys
	a.lastRefresh = time.Now()
	a.mu.Unlock()

	return nil
}

// discoverJWKS fetches the OIDC discovery document and extracts the jwks_uri.
func discoverJWKS(ctx context.Context, client *http.Client, issuer string) (string, error) {
	discoveryURL := strings.TrimRight(issuer, "/") + "/.well-known/openid-configuration"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, http.NoBody)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("discovery endpoint returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}

	var doc struct {
		JWKSURI string `json:"jwks_uri"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return "", fmt.Errorf("parsing discovery document: %w", err)
	}
	if doc.JWKSURI == "" {
		return "", errors.New("discovery document missing jwks_uri")
	}

	return doc.JWKSURI, nil
}

// claimsToUserClaims maps Azure AD token claims to the application's UserClaims.
func claimsToUserClaims(c *tokenClaims) *auth.UserClaims {
	// Azure AD puts the object ID in the "oid" claim. Fall back to "sub".
	sub := c.OID
	if sub == "" {
		sub = c.Sub
	}

	// Map Azure AD app roles to domain roles.
	var roles []domain.Role
	for _, r := range c.Roles {
		role := domain.Role(r)
		if role.Valid() {
			roles = append(roles, role)
		}
	}
	if len(roles) == 0 {
		roles = []domain.Role{domain.RoleRep}
	}

	// Extract team IDs from the "groups" claim if present.
	teamIDs := make([]string, len(c.Groups))
	copy(teamIDs, c.Groups)

	return &auth.UserClaims{
		Sub:     sub,
		Email:   c.Email,
		Name:    c.Name,
		Roles:   roles,
		TeamIDs: teamIDs,
	}
}

// JWT and JWKS types.

type jwtHeader struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Typ string `json:"typ"`
}

type tokenClaims struct {
	Iss    string   `json:"iss"`
	Sub    string   `json:"sub"`
	Aud    string   `json:"aud"`
	Exp    int64    `json:"exp"`
	Nbf    int64    `json:"nbf"`
	Iat    int64    `json:"iat"`
	OID    string   `json:"oid"`
	Email  string   `json:"email"`
	Name   string   `json:"name"`
	Roles  []string `json:"roles"`
	Groups []string `json:"groups"`
}

type jwksResponse struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// parseRSAPublicKey constructs an RSA public key from a JWK.
func parseRSAPublicKey(k jwkKey) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("decoding modulus: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("decoding exponent: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}, nil
}
