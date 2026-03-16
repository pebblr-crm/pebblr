package rbac

import (
	"context"

	"github.com/pebblr/pebblr/internal/domain"
)

// Enforcer evaluates access control decisions for pebblr resources.
// All data-layer queries must pass through an Enforcer to ensure RBAC is respected.
type Enforcer interface {
	// CanViewLead returns true if the actor is permitted to view the given lead.
	CanViewLead(ctx context.Context, actor *domain.User, lead *domain.Lead) bool

	// CanAssignLead returns true if the actor is permitted to assign the given lead.
	CanAssignLead(ctx context.Context, actor *domain.User, lead *domain.Lead) bool

	// CanUpdateLead returns true if the actor is permitted to update the given lead.
	CanUpdateLead(ctx context.Context, actor *domain.User, lead *domain.Lead) bool

	// CanDeleteLead returns true if the actor is permitted to delete the given lead.
	CanDeleteLead(ctx context.Context, actor *domain.User, lead *domain.Lead) bool

	// ScopeLeadQuery returns a LeadScope that restricts a query to the leads
	// the given actor is permitted to see. Use this when building store queries.
	ScopeLeadQuery(ctx context.Context, actor *domain.User) LeadScope
}

// LeadScope encodes RBAC visibility constraints for a lead list query.
// The data layer applies these as query predicates.
type LeadScope struct {
	// AssigneeIDs restricts results to leads assigned to these user IDs.
	// Empty means no restriction on assignee (all assignees visible).
	AssigneeIDs []string
	// TeamIDs restricts results to leads belonging to these team IDs.
	// Empty means no restriction on team.
	TeamIDs []string
	// AllLeads bypasses all scoping (admin only).
	AllLeads bool
}
