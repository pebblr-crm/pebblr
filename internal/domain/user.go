package domain

// User represents an authenticated pebblr user (rep, manager, or admin).
type User struct {
	ID string
	// ExternalID is the Azure AD Object ID (OID) from the OIDC token subject claim.
	ExternalID string
	Email      string
	Name       string
	Role       Role
	// TeamIDs lists the teams this user belongs to.
	TeamIDs []string
}
