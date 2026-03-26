package store

import (
	"context"
	"time"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
)

// ActivityFilter specifies optional filter criteria for activity list queries.
type ActivityFilter struct {
	ActivityType *string
	Status       *string
	CreatorID    *string
	TargetID     *string
	TeamID       *string
	DateFrom     *time.Time
	DateTo       *time.Time
}

// ActivityPage holds a paginated result set of activities.
type ActivityPage struct {
	Activities []*domain.Activity
	Total      int
	Page       int
	Limit      int
}

// ActivityRepository provides CRUD and scoped query access for activities.
type ActivityRepository interface {
	// Get retrieves a single activity by ID.
	// Returns ErrNotFound if no activity exists with that ID (or it is soft-deleted).
	Get(ctx context.Context, id string) (*domain.Activity, error)

	// List returns a paginated, scoped list of activities matching the filter.
	List(ctx context.Context, scope rbac.ActivityScope, filter ActivityFilter, page, limit int) (*ActivityPage, error)

	// Create persists a new activity and returns it with its generated ID.
	Create(ctx context.Context, activity *domain.Activity) (*domain.Activity, error)

	// Update persists changes to an existing activity.
	Update(ctx context.Context, activity *domain.Activity) (*domain.Activity, error)

	// SoftDelete marks an activity as deleted.
	SoftDelete(ctx context.Context, id string) error

	// CountByDate returns the number of non-deleted activities for a creator on a given date.
	CountByDate(ctx context.Context, creatorID string, date time.Time) (int, error)

	// HasActivityWithTypes reports whether the creator has any non-deleted activity
	// on the given date whose activity_type is one of the provided types.
	HasActivityWithTypes(ctx context.Context, creatorID string, date time.Time, types []string) (bool, error)

	// ExistsForTargetOnDate reports whether a non-deleted activity already exists
	// for the given creator, target, and date.
	ExistsForTargetOnDate(ctx context.Context, creatorID, targetID string, date time.Time) (bool, error)
}
