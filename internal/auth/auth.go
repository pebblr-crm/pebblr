package auth

import "context"

// Authenticator validates bearer tokens and extracts user claims.
type Authenticator interface {
	// ValidateToken validates the given JWT token string and returns the parsed claims.
	// Returns an error if the token is expired, invalid, or from an untrusted issuer.
	ValidateToken(ctx context.Context, token string) (*UserClaims, error)
}

// Config holds configuration for Azure AD OIDC authentication.
type Config struct {
	// TenantID is the Azure AD tenant identifier.
	TenantID string
	// ClientID is the application (client) ID registered in Azure AD.
	ClientID string
	// Issuer is the expected token issuer URL (e.g. https://login.microsoftonline.com/{tenantID}/v2.0).
	Issuer string
}
