package domain

// User represents an authenticated pebblr user (rep, manager, or admin).
type User struct {
	ID string `json:"id"`
	// ExternalID is the Azure AD Object ID (OID) from the OIDC token subject claim.
	ExternalID string `json:"external_id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	Role       Role   `json:"role"`
	// TeamIDs lists the teams this user belongs to.
	TeamIDs []string `json:"team_ids"`
}
