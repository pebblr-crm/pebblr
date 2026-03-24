package auth

import "context"

// Authenticator validates bearer tokens and extracts user claims.
// Implementations include azuread.Authenticator (production OIDC),
// demo.Authenticator (self-service demos), and StaticAuthenticator (local dev).
type Authenticator interface {
	// ValidateToken validates the given JWT token string and returns the parsed claims.
	// Returns an error if the token is expired, invalid, or from an untrusted issuer.
	ValidateToken(ctx context.Context, token string) (*UserClaims, error)
}
