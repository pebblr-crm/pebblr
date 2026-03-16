package domain

// Role defines the access level of a user within pebblr.
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

// Permissions returns the set of named permissions granted to this role.
func (r Role) Permissions() []Permission {
	switch r {
	case RoleRep:
		return []Permission{PermViewOwnLeads, PermUpdateOwnLeads, PermAddNote}
	case RoleManager:
		return []Permission{
			PermViewOwnLeads, PermUpdateOwnLeads, PermAddNote,
			PermViewTeamLeads, PermAssignLeads, PermViewTeamMetrics,
		}
	case RoleAdmin:
		return []Permission{
			PermViewOwnLeads, PermUpdateOwnLeads, PermAddNote,
			PermViewTeamLeads, PermAssignLeads, PermViewTeamMetrics,
			PermViewAllLeads, PermManageUsers, PermManageTeams, PermViewAllMetrics,
		}
	}
	return nil
}

// Permission is a named capability within the system.
type Permission string

const (
	PermViewOwnLeads   Permission = "view:own_leads"
	PermUpdateOwnLeads Permission = "update:own_leads"
	PermAddNote        Permission = "add:note"
	PermViewTeamLeads  Permission = "view:team_leads"
	PermAssignLeads    Permission = "assign:leads"
	PermViewTeamMetrics Permission = "view:team_metrics"
	PermViewAllLeads   Permission = "view:all_leads"
	PermManageUsers    Permission = "manage:users"
	PermManageTeams    Permission = "manage:teams"
	PermViewAllMetrics Permission = "view:all_metrics"
)
