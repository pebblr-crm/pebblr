package auth

import "github.com/pebblr/pebblr/internal/domain"

// UserClaims holds the extracted claims from a validated Azure AD JWT.
type UserClaims struct {
	// Sub is the subject (Azure AD object ID / OID) of the authenticated user.
	Sub string
	// Email is the user's email address from the token.
	Email string
	// Name is the display name of the user.
	Name string
	// Roles contains the application roles assigned to the user in Azure AD.
	Roles []domain.Role
	// TeamIDs contains the team identifiers the user belongs to.
	TeamIDs []string
}
