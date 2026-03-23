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

func (e *policyEnforcer) CanViewTarget(_ context.Context, actor *domain.User, target *domain.Target) bool {
	switch actor.Role {
	case domain.RoleAdmin:
		return true
	case domain.RoleManager:
		return containsString(actor.TeamIDs, target.TeamID)
	case domain.RoleRep:
		return actor.ID == target.AssigneeID
	}
	return false
}

func (e *policyEnforcer) CanUpdateTarget(_ context.Context, actor *domain.User, target *domain.Target) bool {
	switch actor.Role {
	case domain.RoleAdmin:
		return true
	case domain.RoleManager:
		return containsString(actor.TeamIDs, target.TeamID)
	case domain.RoleRep:
		return actor.ID == target.AssigneeID
	}
	return false
}

func (e *policyEnforcer) ScopeTargetQuery(_ context.Context, actor *domain.User) TargetScope {
	switch actor.Role {
	case domain.RoleAdmin:
		return TargetScope{AllTargets: true}
	case domain.RoleManager:
		return TargetScope{TeamIDs: actor.TeamIDs}
	case domain.RoleRep:
		return TargetScope{AssigneeIDs: []string{actor.ID}}
	}
	// Default: deny all.
	return TargetScope{AssigneeIDs: []string{""}}
}

func (e *policyEnforcer) CanViewActivity(_ context.Context, actor *domain.User, activity *domain.Activity) bool {
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

func (e *policyEnforcer) CanUpdateActivity(_ context.Context, actor *domain.User, activity *domain.Activity) bool {
	switch actor.Role {
	case domain.RoleAdmin:
		return true
	case domain.RoleManager:
		return containsString(actor.TeamIDs, activity.TeamID)
	case domain.RoleRep:
		return actor.ID == activity.CreatorID
	}
	return false
}

func (e *policyEnforcer) CanDeleteActivity(_ context.Context, actor *domain.User, activity *domain.Activity) bool {
	switch actor.Role {
	case domain.RoleAdmin:
		return true
	case domain.RoleManager:
		return containsString(actor.TeamIDs, activity.TeamID)
	case domain.RoleRep:
		return actor.ID == activity.CreatorID
	}
	return false
}

func (e *policyEnforcer) ScopeActivityQuery(_ context.Context, actor *domain.User) ActivityScope {
	switch actor.Role {
	case domain.RoleAdmin:
		return ActivityScope{AllActivities: true}
	case domain.RoleManager:
		return ActivityScope{TeamIDs: actor.TeamIDs}
	case domain.RoleRep:
		return ActivityScope{CreatorIDs: []string{actor.ID}}
	}
	// Default: deny all.
	return ActivityScope{CreatorIDs: []string{""}}
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
