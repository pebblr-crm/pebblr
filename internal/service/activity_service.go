package service

import (
	"context"
	"fmt"
	"time"

	"github.com/pebblr/pebblr/internal/config"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/store"
)

// ErrSubmitted indicates that the activity is already submitted and locked.
var ErrSubmitted = fmt.Errorf("activity is submitted and locked")

// ErrMaxActivities indicates that the maximum activities per day has been reached.
var ErrMaxActivities = fmt.Errorf("maximum activities per day reached")

// ErrBlockedDay indicates that a blocking activity (e.g. vacation) prevents field activities on this day.
var ErrBlockedDay = fmt.Errorf("day is blocked by a non-field activity")

// ErrTargetRequired indicates that a field activity requires a target.
var ErrTargetRequired = fmt.Errorf("target required")

// ValidationErrors wraps a slice of config.FieldError for returning from the service layer.
type ValidationErrors struct {
	Errors []config.FieldError
}

func (e *ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return "validation failed"
	}
	return fmt.Sprintf("validation failed: %s", e.Errors[0].Error())
}

// ActivityService handles activity business logic with RBAC enforcement and config validation.
type ActivityService struct {
	activities store.ActivityRepository
	audit      store.AuditRepository
	enforcer   rbac.Enforcer
	cfg        *config.TenantConfig
}

// NewActivityService constructs an ActivityService with the given dependencies.
func NewActivityService(
	activities store.ActivityRepository,
	audit store.AuditRepository,
	enforcer rbac.Enforcer,
	cfg *config.TenantConfig,
) *ActivityService {
	return &ActivityService{
		activities: activities,
		audit:      audit,
		enforcer:   enforcer,
		cfg:        cfg,
	}
}

// Create persists a new activity after validating against tenant config and RBAC.
func (s *ActivityService) Create(ctx context.Context, actor *domain.User, activity *domain.Activity) (*domain.Activity, error) {
	activity.CreatorID = actor.ID
	if len(actor.TeamIDs) > 0 {
		activity.TeamID = actor.TeamIDs[0]
	}

	if err := s.validateCore(activity); err != nil {
		return nil, err
	}

	if s.cfg != nil {
		if errs := config.ValidateActivity(s.cfg, activity.ActivityType, activity.Fields, "save"); len(errs) > 0 {
			return nil, &ValidationErrors{Errors: errs}
		}
	}

	// Business rule: target required for field activities.
	if err := s.checkTargetRequired(activity); err != nil {
		return nil, err
	}

	// Business rule: max activities per day.
	if err := s.checkMaxActivitiesPerDay(ctx, activity.CreatorID, activity.DueDate); err != nil {
		return nil, err
	}

	// Business rule: blocked days (vacation/holiday blocks field activities and vice versa).
	if err := s.checkBlockedDay(ctx, activity.CreatorID, activity.DueDate, activity.ActivityType); err != nil {
		return nil, err
	}

	created, err := s.activities.Create(ctx, activity)
	if err != nil {
		return nil, fmt.Errorf("creating activity: %w", err)
	}

	_ = s.audit.Record(ctx, &domain.AuditEntry{
		EntityType: "activity",
		EntityID:   created.ID,
		EventType:  "created",
		ActorID:    actor.ID,
		NewValue:   map[string]any{"activityType": created.ActivityType, "status": created.Status},
	})

	return created, nil
}

// Get retrieves an activity by ID with RBAC enforcement.
func (s *ActivityService) Get(ctx context.Context, actor *domain.User, id string) (*domain.Activity, error) {
	activity, err := s.activities.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting activity: %w", err)
	}
	if !s.enforcer.CanViewActivity(ctx, actor, activity) {
		return nil, ErrForbidden
	}
	return activity, nil
}

// List returns a paginated list of activities scoped to the actor's permissions.
func (s *ActivityService) List(ctx context.Context, actor *domain.User, filter store.ActivityFilter, page, limit int) (*store.ActivityPage, error) {
	scope := s.enforcer.ScopeActivityQuery(ctx, actor)
	result, err := s.activities.List(ctx, scope, filter, page, limit)
	if err != nil {
		return nil, fmt.Errorf("listing activities: %w", err)
	}
	return result, nil
}

// Update persists changes to an existing activity. Blocked if submitted.
func (s *ActivityService) Update(ctx context.Context, actor *domain.User, id string, activity *domain.Activity) (*domain.Activity, error) {
	existing, err := s.activities.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting activity: %w", err)
	}
	if !s.enforcer.CanUpdateActivity(ctx, actor, existing) {
		return nil, ErrForbidden
	}
	if existing.IsSubmitted() {
		return nil, ErrSubmitted
	}

	if err := s.validateCore(activity); err != nil {
		return nil, err
	}

	if s.cfg != nil {
		if errs := config.ValidateActivity(s.cfg, activity.ActivityType, activity.Fields, "save"); len(errs) > 0 {
			return nil, &ValidationErrors{Errors: errs}
		}
	}

	// Preserve immutable fields from the existing record.
	activity.ID = id
	activity.CreatorID = existing.CreatorID
	activity.TeamID = existing.TeamID
	activity.SubmittedAt = existing.SubmittedAt
	activity.CreatedAt = existing.CreatedAt

	updated, err := s.activities.Update(ctx, activity)
	if err != nil {
		return nil, fmt.Errorf("updating activity: %w", err)
	}

	_ = s.audit.Record(ctx, &domain.AuditEntry{
		EntityType: "activity",
		EntityID:   id,
		EventType:  "updated",
		ActorID:    actor.ID,
		OldValue:   map[string]any{"status": existing.Status},
		NewValue:   map[string]any{"status": updated.Status},
	})

	return updated, nil
}

// Delete soft-deletes an activity. Blocked if submitted.
func (s *ActivityService) Delete(ctx context.Context, actor *domain.User, id string) error {
	existing, err := s.activities.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("getting activity: %w", err)
	}
	if !s.enforcer.CanDeleteActivity(ctx, actor, existing) {
		return ErrForbidden
	}
	if existing.IsSubmitted() {
		return ErrSubmitted
	}

	if err := s.activities.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("deleting activity: %w", err)
	}

	_ = s.audit.Record(ctx, &domain.AuditEntry{
		EntityType: "activity",
		EntityID:   id,
		EventType:  "deleted",
		ActorID:    actor.ID,
	})

	return nil
}

// Submit marks an activity as submitted. Validates submit-required fields and
// sets SubmittedAt, locking the activity from further edits.
func (s *ActivityService) Submit(ctx context.Context, actor *domain.User, id string) (*domain.Activity, error) {
	existing, err := s.activities.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting activity: %w", err)
	}
	if !s.enforcer.CanUpdateActivity(ctx, actor, existing) {
		return nil, ErrForbidden
	}
	if existing.IsSubmitted() {
		return nil, ErrSubmitted
	}

	// Validate with submit-phase strictness.
	if s.cfg != nil {
		if errs := config.ValidateActivity(s.cfg, existing.ActivityType, existing.Fields, "submit"); len(errs) > 0 {
			return nil, &ValidationErrors{Errors: errs}
		}
	}

	now := time.Now()
	existing.SubmittedAt = &now

	updated, err := s.activities.Update(ctx, existing)
	if err != nil {
		return nil, fmt.Errorf("submitting activity: %w", err)
	}

	_ = s.audit.Record(ctx, &domain.AuditEntry{
		EntityType: "activity",
		EntityID:   id,
		EventType:  "submitted",
		ActorID:    actor.ID,
	})

	return updated, nil
}

// PartialUpdate applies a server-side apply PATCH to an existing activity.
// Only fields present in the patch are merged; absent fields are left untouched.
// Same RBAC and ErrSubmitted guards as Update.
func (s *ActivityService) PartialUpdate(ctx context.Context, actor *domain.User, id string, patch *domain.ActivityPatch) (*domain.Activity, error) {
	existing, err := s.activities.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting activity: %w", err)
	}
	if !s.enforcer.CanUpdateActivity(ctx, actor, existing) {
		return nil, ErrForbidden
	}
	if existing.IsSubmitted() {
		return nil, ErrSubmitted
	}

	// Merge patch onto a copy of the existing record.
	merged := *existing
	patch.ApplyTo(&merged)

	// Validate the merged result with save-mode strictness.
	if s.cfg != nil {
		if errs := config.ValidateActivity(s.cfg, merged.ActivityType, merged.Fields, "save"); len(errs) > 0 {
			return nil, &ValidationErrors{Errors: errs}
		}
	}

	updated, err := s.activities.Update(ctx, &merged)
	if err != nil {
		return nil, fmt.Errorf("updating activity: %w", err)
	}

	_ = s.audit.Record(ctx, &domain.AuditEntry{
		EntityType: "activity",
		EntityID:   id,
		EventType:  "updated",
		ActorID:    actor.ID,
		OldValue:   map[string]any{"status": existing.Status},
		NewValue:   map[string]any{"status": updated.Status},
	})

	return updated, nil
}

// PatchStatus transitions an activity's status, validated against config transitions.
func (s *ActivityService) PatchStatus(ctx context.Context, actor *domain.User, id, newStatus string) (*domain.Activity, error) {
	existing, err := s.activities.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting activity: %w", err)
	}
	if !s.enforcer.CanUpdateActivity(ctx, actor, existing) {
		return nil, ErrForbidden
	}
	if existing.IsSubmitted() {
		return nil, ErrSubmitted
	}

	// Validate the status and transition.
	if s.cfg != nil {
		if fe := config.ValidateStatus(s.cfg, newStatus); fe != nil {
			return nil, &ValidationErrors{Errors: []config.FieldError{*fe}}
		}
		if fe := config.ValidateStatusTransition(s.cfg, existing.Status, newStatus); fe != nil {
			return nil, &ValidationErrors{Errors: []config.FieldError{*fe}}
		}
	}

	oldStatus := existing.Status
	existing.Status = newStatus

	updated, err := s.activities.Update(ctx, existing)
	if err != nil {
		return nil, fmt.Errorf("updating activity status: %w", err)
	}

	_ = s.audit.Record(ctx, &domain.AuditEntry{
		EntityType: "activity",
		EntityID:   id,
		EventType:  "status_changed",
		ActorID:    actor.ID,
		OldValue:   map[string]any{"status": oldStatus},
		NewValue:   map[string]any{"status": newStatus},
	})

	return updated, nil
}

// validateCore checks that an activity has valid core fields (type, status, duration).
func (s *ActivityService) validateCore(activity *domain.Activity) error {
	if activity.ActivityType == "" {
		return ErrInvalidInput
	}
	if activity.DueDate.IsZero() {
		return ErrInvalidInput
	}

	if s.cfg == nil {
		return nil
	}

	if s.cfg.ActivityType(activity.ActivityType) == nil {
		return ErrInvalidInput
	}
	if activity.Status == "" {
		activity.Status = s.cfg.InitialStatus()
	}
	if fe := config.ValidateStatus(s.cfg, activity.Status); fe != nil {
		return ErrInvalidInput
	}
	if activity.Duration != "" {
		if fe := config.ValidateDuration(s.cfg, activity.Duration); fe != nil {
			return ErrInvalidInput
		}
	}

	return nil
}

// checkMaxActivitiesPerDay enforces the max_activities_per_day rule.
func (s *ActivityService) checkMaxActivitiesPerDay(ctx context.Context, creatorID string, date time.Time) error {
	if s.cfg == nil || s.cfg.Rules.MaxActivitiesPerDay <= 0 {
		return nil
	}
	count, err := s.activities.CountByDate(ctx, creatorID, date)
	if err != nil {
		return fmt.Errorf("counting activities by date: %w", err)
	}
	if count >= s.cfg.Rules.MaxActivitiesPerDay {
		return ErrMaxActivities
	}
	return nil
}

// checkTargetRequired enforces that field-category activities have a target.
func (s *ActivityService) checkTargetRequired(activity *domain.Activity) error {
	if s.cfg == nil {
		return nil
	}
	at := s.cfg.ActivityType(activity.ActivityType)
	if at == nil {
		return nil // already caught by validateCore
	}
	if at.Category == "field" && activity.TargetID == "" {
		return ErrTargetRequired
	}
	return nil
}

// checkBlockedDay enforces two rules:
//  1. If the new activity is a field activity, no blocking (e.g. vacation) activity may exist on that day.
//  2. If the new activity blocks field activities, no field activity may exist on that day.
func (s *ActivityService) checkBlockedDay(ctx context.Context, creatorID string, date time.Time, activityType string) error {
	if s.cfg == nil {
		return nil
	}
	at := s.cfg.ActivityType(activityType)
	if at == nil {
		return nil
	}

	if at.Category == "field" {
		// Check if any blocking activity exists on this day.
		blockingTypes := s.blockingActivityTypes()
		if len(blockingTypes) > 0 {
			blocked, err := s.activities.HasActivityWithTypes(ctx, creatorID, date, blockingTypes)
			if err != nil {
				return fmt.Errorf("checking blocked day: %w", err)
			}
			if blocked {
				return ErrBlockedDay
			}
		}
	}

	if at.BlocksFieldActivities {
		// Check if any field activity exists on this day.
		fieldTypes := s.fieldActivityTypes()
		if len(fieldTypes) > 0 {
			hasField, err := s.activities.HasActivityWithTypes(ctx, creatorID, date, fieldTypes)
			if err != nil {
				return fmt.Errorf("checking blocked day: %w", err)
			}
			if hasField {
				return ErrBlockedDay
			}
		}
	}

	return nil
}

// blockingActivityTypes returns the keys of all activity types that block field activities.
func (s *ActivityService) blockingActivityTypes() []string {
	var types []string
	for _, at := range s.cfg.Activities.Types {
		if at.BlocksFieldActivities {
			types = append(types, at.Key)
		}
	}
	return types
}

// fieldActivityTypes returns the keys of all field-category activity types.
func (s *ActivityService) fieldActivityTypes() []string {
	var types []string
	for _, at := range s.cfg.Activities.Types {
		if at.Category == "field" {
			types = append(types, at.Key)
		}
	}
	return types
}
