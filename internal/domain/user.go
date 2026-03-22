package domain

// OnlineStatus represents a user's current availability/presence state.
type OnlineStatus string

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
