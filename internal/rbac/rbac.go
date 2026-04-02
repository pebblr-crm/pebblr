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

	// CanViewActivity returns true if the actor is permitted to view the given activity.
	CanViewActivity(ctx context.Context, actor *domain.User, activity *domain.Activity) bool

	// CanUpdateActivity returns true if the actor is permitted to update the given activity.
	CanUpdateActivity(ctx context.Context, actor *domain.User, activity *domain.Activity) bool

	// CanDeleteActivity returns true if the actor is permitted to soft-delete the given activity.
	CanDeleteActivity(ctx context.Context, actor *domain.User, activity *domain.Activity) bool

	// ScopeActivityQuery returns an ActivityScope that restricts a query to the activities
	// the given actor is permitted to see.
	ScopeActivityQuery(ctx context.Context, actor *domain.User) ActivityScope
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

// ActivityScope encodes RBAC visibility constraints for an activity list query.
type ActivityScope struct {
	// CreatorIDs restricts results to activities created by these user IDs.
	CreatorIDs []string
	// TeamIDs restricts results to activities belonging to these team IDs.
	TeamIDs []string
	// JointVisitUID, when non-empty, includes activities where this user is the
	// joint visitor (in addition to CreatorIDs). This ensures reps see activities
	// they participate in as co-visitors, matching CanViewActivity semantics.
	JointVisitUID string
	// AllActivities bypasses all scoping (admin only).
	AllActivities bool
}
