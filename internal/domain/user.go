package domain

import "fmt"

// OnlineStatus represents a user's current availability/presence state.
type OnlineStatus string

// OnlineStatus values.
const (
	OnlineStatusOnline  OnlineStatus = "online"
	OnlineStatusAway    OnlineStatus = "away"
	OnlineStatusOffline OnlineStatus = "offline"
)

// Valid returns true if the OnlineStatus is a recognized value.
func (s OnlineStatus) Valid() bool {
	switch s {
	case OnlineStatusOnline, OnlineStatusAway, OnlineStatusOffline:
		return true
	}
	return false
}

// User represents an authenticated pebblr user (rep, manager, or admin).
type User struct {
	ID string `json:"id"`
	// ExternalID is the Azure AD Object ID (OID) from the OIDC token subject claim.
	ExternalID string `json:"externalId"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	Role       Role   `json:"role"`
	// TeamIDs lists the teams this user belongs to.
	TeamIDs []string `json:"teamIds"`
	// Avatar is the URL or path to the user's profile picture.
	Avatar string `json:"avatar"`
	// OnlineStatus indicates the user's current presence state.
	OnlineStatus OnlineStatus `json:"onlineStatus"`
}

// Validate checks that the User has valid required fields. This is a defense-in-depth
// check to catch malformed users before they reach RBAC decisions.
func (u *User) Validate() error {
	if u.ID == "" {
		return fmt.Errorf("user ID must not be empty")
	}
	if u.ExternalID == "" {
		return fmt.Errorf("user external ID must not be empty")
	}
	if !u.Role.Valid() {
		return fmt.Errorf("invalid user role %q", u.Role)
	}
	return nil
}
