package rbac

import "github.com/pebblr/pebblr/internal/domain"

// PolicyEnforcer evaluates access control decisions for pebblr resources.
// All data-layer queries must pass through a PolicyEnforcer to ensure RBAC
// is respected.
//
// This is the only enforcer implementation. An interface was removed here
// because there is exactly one implementation and context.Context parameters
// were removed because they were unused — the enforcer makes pure, in-memory
// decisions based on the actor's role and the resource's ownership.
type PolicyEnforcer struct{}

// NewEnforcer returns the default role-based PolicyEnforcer.
func NewEnforcer() *PolicyEnforcer {
	return &PolicyEnforcer{}
}

// CanViewTarget returns true if the actor is permitted to view the given target.
func (e *PolicyEnforcer) CanViewTarget(actor *domain.User, target *domain.Target) bool {
	return canAccessTarget(actor, target)
}

// CanUpdateTarget returns true if the actor is permitted to update the given target.
func (e *PolicyEnforcer) CanUpdateTarget(actor *domain.User, target *domain.Target) bool {
	return canAccessTarget(actor, target)
}

// ScopeTargetQuery returns a TargetScope that restricts a query to the targets
// the given actor is permitted to see.
func (e *PolicyEnforcer) ScopeTargetQuery(actor *domain.User) TargetScope {
	switch actor.Role {
	case domain.RoleAdmin:
		return TargetScope{AllTargets: true}
	case domain.RoleManager:
		return TargetScope{TeamIDs: actor.TeamIDs}
	case domain.RoleRep:
		return TargetScope{AssigneeIDs: []string{actor.ID}}
	}
	// Default: deny all — unrecognized role gets zero results.
	return TargetScope{DenyAll: true}
}

// CanViewActivity returns true if the actor is permitted to view the given activity.
func (e *PolicyEnforcer) CanViewActivity(actor *domain.User, activity *domain.Activity) bool {
	if actor == nil || activity == nil {
		return false
	}
	switch actor.Role {
	case domain.RoleAdmin:
		return true
	case domain.RoleManager:
		return containsString(actor.TeamIDs, activity.TeamID)
	case domain.RoleRep:
		return actor.ID == activity.CreatorID || actor.ID == activity.JointVisitUID
	}
	return false
}

// CanUpdateActivity returns true if the actor is permitted to update the given activity.
func (e *PolicyEnforcer) CanUpdateActivity(actor *domain.User, activity *domain.Activity) bool {
	return canModifyActivity(actor, activity)
}

// CanDeleteActivity returns true if the actor is permitted to soft-delete the given activity.
func (e *PolicyEnforcer) CanDeleteActivity(actor *domain.User, activity *domain.Activity) bool {
	return canModifyActivity(actor, activity)
}

// ScopeActivityQuery returns an ActivityScope that restricts a query to the activities
// the given actor is permitted to see.
func (e *PolicyEnforcer) ScopeActivityQuery(actor *domain.User) ActivityScope {
	switch actor.Role {
	case domain.RoleAdmin:
		return ActivityScope{AllActivities: true}
	case domain.RoleManager:
		return ActivityScope{TeamIDs: actor.TeamIDs}
	case domain.RoleRep:
		return ActivityScope{CreatorIDs: []string{actor.ID}}
	}
	// Default: deny all — unrecognized role gets zero results.
	return ActivityScope{DenyAll: true}
}

// TargetScope encodes RBAC visibility constraints for a target list query.
type TargetScope struct {
	// AssigneeIDs restricts results to targets assigned to these user IDs.
	AssigneeIDs []string
	// TeamIDs restricts results to targets belonging to these team IDs.
	TeamIDs []string
	// AllTargets bypasses all scoping (admin only).
	AllTargets bool
	// DenyAll rejects all results. Set when an unrecognized role is encountered.
	DenyAll bool
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
	// DenyAll rejects all results. Set when an unrecognized role is encountered.
	DenyAll bool
}
