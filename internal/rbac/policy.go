package rbac

import (
	"context"

	"github.com/pebblr/pebblr/internal/domain"
)

// policyEnforcer implements Enforcer using role-based policy definitions.
type policyEnforcer struct{}

// NewEnforcer returns the default role-based policy Enforcer.
func NewEnforcer() Enforcer {
	return &policyEnforcer{}
}

func (e *policyEnforcer) CanViewLead(_ context.Context, actor *domain.User, lead *domain.Lead) bool {
	switch actor.Role {
	case domain.RoleAdmin:
		return true
	case domain.RoleManager:
		return containsString(actor.TeamIDs, lead.TeamID)
	case domain.RoleRep:
		return actor.ID == lead.AssigneeID
	}
	return false
}

func (e *policyEnforcer) CanAssignLead(_ context.Context, actor *domain.User, lead *domain.Lead) bool {
	switch actor.Role {
	case domain.RoleAdmin:
		return true
	case domain.RoleManager:
		return containsString(actor.TeamIDs, lead.TeamID)
	}
	return false
}

func (e *policyEnforcer) CanUpdateLead(_ context.Context, actor *domain.User, lead *domain.Lead) bool {
	switch actor.Role {
	case domain.RoleAdmin:
		return true
	case domain.RoleManager:
		return containsString(actor.TeamIDs, lead.TeamID)
	case domain.RoleRep:
		return actor.ID == lead.AssigneeID
	}
	return false
}

func (e *policyEnforcer) CanDeleteLead(_ context.Context, actor *domain.User, lead *domain.Lead) bool {
	switch actor.Role {
	case domain.RoleAdmin:
		return true
	case domain.RoleManager:
		return containsString(actor.TeamIDs, lead.TeamID)
	}
	return false
}

func (e *policyEnforcer) ScopeLeadQuery(_ context.Context, actor *domain.User) LeadScope {
	switch actor.Role {
	case domain.RoleAdmin:
		return LeadScope{AllLeads: true}
	case domain.RoleManager:
		return LeadScope{TeamIDs: actor.TeamIDs}
	case domain.RoleRep:
		return LeadScope{AssigneeIDs: []string{actor.ID}}
	}
	// Default: deny all — return an impossible scope.
	return LeadScope{AssigneeIDs: []string{""}}
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
