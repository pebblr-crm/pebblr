package rbac

import (
	"context"

	"github.com/pebblr/pebblr/internal/domain"
)

// Enforcer evaluates access control decisions for pebblr resources.
// All data-layer queries must pass through an Enforcer to ensure RBAC is respected.
type Enforcer interface {
	// CanViewTarget returns true if the actor is permitted to view the given target.
	CanViewTarget(ctx context.Context, actor *domain.User, target *domain.Target) bool

	// CanUpdateTarget returns true if the actor is permitted to update the given target.
	CanUpdateTarget(ctx context.Context, actor *domain.User, target *domain.Target) bool

	// ScopeTargetQuery returns a TargetScope that restricts a query to the targets
	// the given actor is permitted to see.
	ScopeTargetQuery(ctx context.Context, actor *domain.User) TargetScope
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
