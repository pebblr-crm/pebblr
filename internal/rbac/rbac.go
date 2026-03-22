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

	// CanViewTarget returns true if the actor is permitted to view the given target.
	CanViewTarget(ctx context.Context, actor *domain.User, target *domain.Target) bool

	// CanUpdateTarget returns true if the actor is permitted to update the given target.
	CanUpdateTarget(ctx context.Context, actor *domain.User, target *domain.Target) bool

	// ScopeTargetQuery returns a TargetScope that restricts a query to the targets
	// the given actor is permitted to see.
	ScopeTargetQuery(ctx context.Context, actor *domain.User) TargetScope
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

// TargetScope encodes RBAC visibility constraints for a target list query.
type TargetScope struct {
	// AssigneeIDs restricts results to targets assigned to these user IDs.
	AssigneeIDs []string
	// TeamIDs restricts results to targets belonging to these team IDs.
	TeamIDs []string
	// AllTargets bypasses all scoping (admin only).
	AllTargets bool
}
