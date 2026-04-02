package domain

// Role defines the access level of a user within pebblr.
// Access control decisions are made by the rbac package based on these roles;
// do not add permission-list methods here — they become decorative when
// the enforcer uses role-based switching directly.
type Role string

const (
	// RoleRep is a field sales representative. Sees only their own assigned leads.
	RoleRep Role = "rep"
	// RoleManager manages a team of reps. Sees all leads within their team(s).
	RoleManager Role = "manager"
	// RoleAdmin has full visibility across all tenants and teams.
	RoleAdmin Role = "admin"
)

// Valid returns true if the role is a recognized value.
func (r Role) Valid() bool {
	switch r {
	case RoleRep, RoleManager, RoleAdmin:
		return true
	}
	return false
}
