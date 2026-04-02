package rbac

import "github.com/pebblr/pebblr/internal/domain"

// canAccessTarget checks whether the actor has access to the given target
// based on their role. Both CanViewTarget and CanUpdateTarget share this logic
// today; they are kept as separate methods so they can diverge in the future.
func canAccessTarget(actor *domain.User, target *domain.Target) bool {
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

// canModifyActivity checks whether the actor can modify (update or delete) the
// given activity based on their role. Both CanUpdateActivity and CanDeleteActivity
// share this logic today; they are kept as separate methods so they can diverge
// in the future.
func canModifyActivity(actor *domain.User, activity *domain.Activity) bool {
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

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
