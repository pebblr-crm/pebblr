package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/pebblr/pebblr/internal/config"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/store"
)

const dateFormat = "2006-01-02"

// errFmtGettingActivity is the error format used when retrieving an activity fails.
const errFmtGettingActivity = "getting activity: %w"

// ErrSubmitted indicates that the activity is already submitted and locked.
var ErrSubmitted = errors.New("activity is submitted and locked")

// ErrMaxActivities indicates that the maximum activities per day has been reached.
var ErrMaxActivities = errors.New("maximum activities per day reached")

// ErrBlockedDay indicates that a blocking activity (e.g. vacation) prevents field activities on this day.
var ErrBlockedDay = errors.New("day is blocked by a non-field activity")

// ErrTargetRequired indicates that a field activity requires a target.
var ErrTargetRequired = errors.New("target required")

// ErrInvalidJointVisitor indicates that the joint visit user ID is invalid
// (e.g. self-reference or non-existent user).
var ErrInvalidJointVisitor = errors.New("invalid joint visit user")

// ErrStatusNotSubmittable indicates the activity's current status does not allow submission.
var ErrStatusNotSubmittable = errors.New("current status does not allow submission")

// ErrNoRecoveryBalance indicates that no recovery day is available to claim.
var ErrNoRecoveryBalance = errors.New("no recovery day balance available")

// ErrDuplicateActivity indicates that an activity for this target on this date already exists.
var ErrDuplicateActivity = errors.New("activity for this target on this date already exists")

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
	targets    store.TargetRepository
	users      store.UserRepository
	audit      store.AuditRepository
	dashboard  store.DashboardRepository
	enforcer   rbac.Enforcer
	cfg        *config.TenantConfig
}

// NewActivityService constructs an ActivityService with the given dependencies.
func NewActivityService(
	activities store.ActivityRepository,
	targets store.TargetRepository,
	users store.UserRepository,
	audit store.AuditRepository,
	enforcer rbac.Enforcer,
	cfg *config.TenantConfig,
	opts ...ActivityServiceOption,
) *ActivityService {
	s := &ActivityService{
		activities: activities,
		targets:    targets,
		users:      users,
		audit:      audit,
		enforcer:   enforcer,
		cfg:        cfg,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// ActivityServiceOption configures optional dependencies on ActivityService.
type ActivityServiceOption func(*ActivityService)

// WithDashboard injects a DashboardRepository for recovery balance validation.
func WithDashboard(d store.DashboardRepository) ActivityServiceOption {
	return func(s *ActivityService) {
		s.dashboard = d
	}
}

// Create persists a new activity after validating against tenant config and RBAC.
func (s *ActivityService) Create(ctx context.Context, actor *domain.User, activity *domain.Activity) (*domain.Activity, error) {
	activity.CreatorID = actor.ID
	if len(actor.TeamIDs) > 0 {
		activity.TeamID = actor.TeamIDs[0]
	}

	if err := s.validateForCreate(ctx, actor, activity); err != nil {
		return nil, err
	}

	created, err := s.activities.Create(ctx, activity)
	if err != nil {
		return nil, fmt.Errorf("creating activity: %w", err)
	}

	s.recordAudit(ctx, &domain.AuditEntry{
		EntityType: "activity",
		EntityID:   created.ID,
		EventType:  "created",
		ActorID:    actor.ID,
		NewValue:   map[string]any{"activityType": created.ActivityType, "status": created.Status},
	})

	return created, nil
}

// validateForCreate runs all validation and business rule checks for creating an activity.
func (s *ActivityService) validateForCreate(ctx context.Context, actor *domain.User, activity *domain.Activity) error {
	if err := s.validateCore(activity); err != nil {
		return err
	}

	if s.cfg != nil {
		if errs := config.ValidateActivity(s.cfg, activity.ActivityType, activity.Fields, "save"); len(errs) > 0 {
			return &ValidationErrors{Errors: errs}
		}
	}

	if err := s.checkJointVisitUser(ctx, activity.CreatorID, activity.JointVisitUID); err != nil {
		return err
	}
	if err := s.checkTargetRequired(activity); err != nil {
		return err
	}
	if err := s.checkTargetAccess(ctx, actor, activity.TargetID); err != nil {
		return err
	}
	if err := s.checkDuplicateActivity(ctx, activity); err != nil {
		return err
	}
	if err := s.checkMaxActivitiesPerDay(ctx, activity.CreatorID, activity.DueDate); err != nil {
		return err
	}
	if err := s.checkBlockedDay(ctx, activity.CreatorID, activity.DueDate, activity.ActivityType); err != nil {
		return err
	}
	return s.checkRecoveryBalance(ctx, actor, activity)
}

// checkDuplicateActivity ensures no activity for the same target on the same date exists.
func (s *ActivityService) checkDuplicateActivity(ctx context.Context, activity *domain.Activity) error {
	if activity.TargetID == "" {
		return nil
	}
	exists, err := s.activities.ExistsForTargetOnDate(ctx, activity.CreatorID, activity.TargetID, activity.DueDate)
	if err != nil {
		return fmt.Errorf("checking duplicate activity: %w", err)
	}
	if exists {
		return ErrDuplicateActivity
	}
	return nil
}

// Get retrieves an activity by ID with RBAC enforcement.
func (s *ActivityService) Get(ctx context.Context, actor *domain.User, id string) (*domain.Activity, error) {
	activity, err := s.activities.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf(errFmtGettingActivity, err)
	}
	if !s.enforcer.CanViewActivity(ctx, actor, activity) {
		return nil, ErrForbidden
	}
	return activity, nil
}

// List returns a paginated list of activities scoped to the actor's permissions.
// Non-field activities whose due date has passed are auto-completed.
func (s *ActivityService) List(ctx context.Context, actor *domain.User, filter store.ActivityFilter, page, limit int) (*store.ActivityPage, error) {
	scope := s.enforcer.ScopeActivityQuery(ctx, actor)
	result, err := s.activities.List(ctx, scope, filter, page, limit)
	if err != nil {
		return nil, fmt.Errorf("listing activities: %w", err)
	}
	s.autoCompleteNonFieldActivities(ctx, result.Activities)
	return result, nil
}

// nonFieldTypeSet returns a set of activity type keys that are in the "non_field" category.
func (s *ActivityService) nonFieldTypeSet() map[string]bool {
	types := make(map[string]bool)
	for _, at := range s.cfg.Activities.Types {
		if at.Category == "non_field" {
			types[at.Key] = true
		}
	}
	return types
}

// shouldAutoComplete returns true if the activity is a past non-field activity
// in the initial status that has not been submitted.
func shouldAutoComplete(a *domain.Activity, nonFieldTypes map[string]bool, initialStatus string, today time.Time) bool {
	if !nonFieldTypes[a.ActivityType] || a.Status != initialStatus || a.SubmittedAt != nil {
		return false
	}
	return a.DueDate.Before(today)
}

// autoCompleteNonFieldActivities transitions non-field activities whose due date
// is in the past from the initial status to "completed". This is fire-and-forget
// — errors are logged but don't block the response.
func (s *ActivityService) autoCompleteNonFieldActivities(ctx context.Context, activities []*domain.Activity) {
	if s.cfg == nil {
		return
	}
	initialStatus := s.cfg.InitialStatus()
	if initialStatus == "" {
		return
	}

	nonFieldTypes := s.nonFieldTypeSet()
	today := time.Now().Truncate(24 * time.Hour)

	for _, a := range activities {
		if !shouldAutoComplete(a, nonFieldTypes, initialStatus, today) {
			continue
		}
		a.Status = "completed"
		if _, err := s.activities.Update(ctx, a); err != nil {
			slog.Default().Warn("auto-complete non-field activity failed", "id", a.ID, "err", err)
		}
	}
}

// Update persists changes to an existing activity. Blocked if submitted.
func (s *ActivityService) Update(ctx context.Context, actor *domain.User, id string, activity *domain.Activity) (*domain.Activity, error) {
	existing, err := s.getEditableActivity(ctx, actor, id)
	if err != nil {
		return nil, err
	}

	if err := s.validateForUpdate(ctx, actor, existing, activity); err != nil {
		return nil, err
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

	s.recordAudit(ctx, &domain.AuditEntry{
		EntityType: "activity",
		EntityID:   id,
		EventType:  "updated",
		ActorID:    actor.ID,
		OldValue:   map[string]any{"status": existing.Status},
		NewValue:   map[string]any{"status": updated.Status},
	})

	return updated, nil
}

// getEditableActivity retrieves an activity and checks that the actor can update it
// and that it is not already submitted.
func (s *ActivityService) getEditableActivity(ctx context.Context, actor *domain.User, id string) (*domain.Activity, error) {
	existing, err := s.activities.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf(errFmtGettingActivity, err)
	}
	if !s.enforcer.CanUpdateActivity(ctx, actor, existing) {
		return nil, ErrForbidden
	}
	if existing.IsSubmitted() {
		return nil, ErrSubmitted
	}
	return existing, nil
}

// validateForUpdate runs validation and business rule checks for updating an activity.
func (s *ActivityService) validateForUpdate(ctx context.Context, actor *domain.User, existing, activity *domain.Activity) error {
	if err := s.validateCore(activity); err != nil {
		return err
	}
	if s.cfg != nil {
		if errs := config.ValidateActivity(s.cfg, activity.ActivityType, activity.Fields, "save"); len(errs) > 0 {
			return &ValidationErrors{Errors: errs}
		}
	}
	if err := s.checkJointVisitUser(ctx, existing.CreatorID, activity.JointVisitUID); err != nil {
		return err
	}
	if activity.TargetID != existing.TargetID {
		if err := s.checkTargetAccess(ctx, actor, activity.TargetID); err != nil {
			return err
		}
	}
	return nil
}

// Delete soft-deletes an activity. Blocked if submitted.
func (s *ActivityService) Delete(ctx context.Context, actor *domain.User, id string) error {
	existing, err := s.activities.Get(ctx, id)
	if err != nil {
		return fmt.Errorf(errFmtGettingActivity, err)
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

	s.recordAudit(ctx, &domain.AuditEntry{
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
		return nil, fmt.Errorf(errFmtGettingActivity, err)
	}
	if !s.enforcer.CanUpdateActivity(ctx, actor, existing) {
		return nil, ErrForbidden
	}
	if existing.IsSubmitted() {
		return nil, ErrSubmitted
	}

	// Business rule: only closed statuses (e.g. completed, cancelled) allow submission.
	if s.cfg != nil && !s.cfg.IsSubmittableStatus(existing.Status) {
		return nil, ErrStatusNotSubmittable
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

	s.recordAudit(ctx, &domain.AuditEntry{
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
		return nil, fmt.Errorf(errFmtGettingActivity, err)
	}
	if !s.enforcer.CanUpdateActivity(ctx, actor, existing) {
		return nil, ErrForbidden
	}
	if existing.IsSubmitted() {
		return nil, ErrSubmitted
	}

	// Business rule: non-admin users may only target their accessible targets.
	if patch.TargetID != nil && *patch.TargetID != existing.TargetID {
		if err := s.checkTargetAccess(ctx, actor, *patch.TargetID); err != nil {
			return nil, err
		}
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

	s.recordAudit(ctx, &domain.AuditEntry{
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
		return nil, fmt.Errorf(errFmtGettingActivity, err)
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

	s.recordAudit(ctx, &domain.AuditEntry{
		EntityType: "activity",
		EntityID:   id,
		EventType:  "status_changed",
		ActorID:    actor.ID,
		OldValue:   map[string]any{"status": oldStatus},
		NewValue:   map[string]any{"status": newStatus},
	})

	return updated, nil
}

// CloneWeekResult holds the outcome of a clone-week operation.
type CloneWeekResult struct {
	Created int `json:"created"`
	Skipped int `json:"skipped"`
}

// CloneWeek duplicates all activities from sourceWeekStart (Mon) to targetWeekStart (Mon),
// preserving the weekday offset, resetting status to initial, and generating new IDs.
func (s *ActivityService) CloneWeek(ctx context.Context, actor *domain.User, sourceWeekStart, targetWeekStart time.Time) (*CloneWeekResult, error) {
	if err := validateCloneWeekInputs(sourceWeekStart, targetWeekStart); err != nil {
		return nil, err
	}

	sourceEnd := sourceWeekStart.AddDate(0, 0, 4) // Friday
	targetEnd := targetWeekStart.AddDate(0, 0, 4)

	scope := s.enforcer.ScopeActivityQuery(ctx, actor)
	sourcePage, err := s.activities.List(ctx, scope, store.ActivityFilter{
		DateFrom: &sourceWeekStart,
		DateTo:   &sourceEnd,
	}, 1, 200)
	if err != nil {
		return nil, fmt.Errorf("listing source week activities: %w", err)
	}

	if len(sourcePage.Activities) == 0 {
		return nil, fmt.Errorf("source week has no activities: %w", ErrInvalidInput)
	}

	existingTargets, err := s.buildExistingTargetIndex(ctx, scope, targetWeekStart, targetEnd)
	if err != nil {
		return nil, err
	}

	return s.cloneActivities(ctx, actor, sourcePage.Activities, existingTargets, targetWeekStart.Sub(sourceWeekStart))
}

// validateCloneWeekInputs checks that source and target weeks are valid Mondays and differ.
func validateCloneWeekInputs(sourceWeekStart, targetWeekStart time.Time) error {
	if sourceWeekStart.Weekday() != time.Monday {
		return fmt.Errorf("sourceWeekStart must be a Monday: %w", ErrInvalidInput)
	}
	if targetWeekStart.Weekday() != time.Monday {
		return fmt.Errorf("targetWeekStart must be a Monday: %w", ErrInvalidInput)
	}
	if sourceWeekStart.Equal(targetWeekStart) {
		return fmt.Errorf("target week must differ from source week: %w", ErrInvalidInput)
	}
	return nil
}

// buildExistingTargetIndex returns a date→targetID→exists map for the target week.
func (s *ActivityService) buildExistingTargetIndex(ctx context.Context, scope rbac.ActivityScope, from, to time.Time) (map[string]map[string]bool, error) {
	targetPage, err := s.activities.List(ctx, scope, store.ActivityFilter{
		DateFrom: &from,
		DateTo:   &to,
	}, 1, 200)
	if err != nil {
		return nil, fmt.Errorf("listing target week activities: %w", err)
	}
	index := make(map[string]map[string]bool)
	for _, a := range targetPage.Activities {
		dateStr := a.DueDate.Format(dateFormat)
		if index[dateStr] == nil {
			index[dateStr] = make(map[string]bool)
		}
		if a.TargetID != "" {
			index[dateStr][a.TargetID] = true
		}
	}
	return index, nil
}

// cloneActivities creates clones for each source activity, skipping duplicates.
func (s *ActivityService) cloneActivities(ctx context.Context, actor *domain.User, sources []*domain.Activity, existingTargets map[string]map[string]bool, dayOffset time.Duration) (*CloneWeekResult, error) {
	result := &CloneWeekResult{}

	for _, src := range sources {
		newDueDate := src.DueDate.Add(dayOffset)
		newDateStr := newDueDate.Format(dateFormat)

		if src.TargetID != "" && existingTargets[newDateStr][src.TargetID] {
			result.Skipped++
			continue
		}

		clone := &domain.Activity{
			ActivityType:  src.ActivityType,
			Label:         src.Label,
			Status:        "", // will be set to initial by validateCore
			DueDate:       newDueDate,
			Duration:      src.Duration,
			Routing:       src.Routing,
			Fields:        copyFields(src.Fields),
			TargetID:      src.TargetID,
			JointVisitUID: src.JointVisitUID,
		}

		if _, err := s.Create(ctx, actor, clone); err != nil {
			result.Skipped++
			continue
		}
		result.Created++
	}

	return result, nil
}

// copyFields returns a shallow copy of a fields map.
func copyFields(fields map[string]any) map[string]any {
	if fields == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(fields))
	for k, v := range fields {
		out[k] = v
	}
	return out
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
	if config.ValidateStatus(s.cfg, activity.Status) != nil {
		return ErrInvalidInput
	}
	if activity.Duration != "" {
		if config.ValidateDuration(s.cfg, activity.Duration) != nil {
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
		if err := s.checkDayNotBlockedByNonField(ctx, creatorID, date); err != nil {
			return err
		}
	}

	if at.BlocksFieldActivities {
		if err := s.checkDayHasNoFieldActivities(ctx, creatorID, date); err != nil {
			return err
		}
	}

	return nil
}

// checkDayNotBlockedByNonField checks if any blocking activity (e.g. vacation) exists on this day.
func (s *ActivityService) checkDayNotBlockedByNonField(ctx context.Context, creatorID string, date time.Time) error {
	blockingTypes := s.blockingActivityTypes()
	if len(blockingTypes) == 0 {
		return nil
	}
	blocked, err := s.activities.HasActivityWithTypes(ctx, creatorID, date, blockingTypes)
	if err != nil {
		return fmt.Errorf("checking blocked day: %w", err)
	}
	if blocked {
		return ErrBlockedDay
	}
	return nil
}

// checkDayHasNoFieldActivities checks if any field activity exists on this day.
func (s *ActivityService) checkDayHasNoFieldActivities(ctx context.Context, creatorID string, date time.Time) error {
	fieldTypes := s.fieldActivityTypes()
	if len(fieldTypes) == 0 {
		return nil
	}
	hasField, err := s.activities.HasActivityWithTypes(ctx, creatorID, date, fieldTypes)
	if err != nil {
		return fmt.Errorf("checking blocked day: %w", err)
	}
	if hasField {
		return ErrBlockedDay
	}
	return nil
}

// blockingActivityTypes returns the keys of all activity types that block field activities.
func (s *ActivityService) blockingActivityTypes() []string {
	var types []string
	for i := range s.cfg.Activities.Types {
		if s.cfg.Activities.Types[i].BlocksFieldActivities {
			types = append(types, s.cfg.Activities.Types[i].Key)
		}
	}
	return types
}

// fieldActivityTypes returns the keys of all field-category activity types.
func (s *ActivityService) fieldActivityTypes() []string {
	var types []string
	for i := range s.cfg.Activities.Types {
		if s.cfg.Activities.Types[i].Category == "field" {
			types = append(types, s.cfg.Activities.Types[i].Key)
		}
	}
	return types
}

// checkRecoveryBalance verifies that a recovery-type activity can only be created
// when the actor has available recovery balance and the activity date falls within
// a valid claim window.
func (s *ActivityService) checkRecoveryBalance(ctx context.Context, actor *domain.User, activity *domain.Activity) error {
	if !s.isRecoveryActivity(activity) {
		return nil
	}

	scope := s.enforcer.ScopeActivityQuery(ctx, actor)
	recoveryRule := s.cfg.Recovery

	lookback := time.Duration(recoveryRule.RecoveryWindowDays*2+7) * 24 * time.Hour
	filter := store.DashboardFilter{
		DateFrom: activity.DueDate.Add(-lookback),
		DateTo:   activity.DueDate,
	}

	weekendActivities, err := s.dashboard.WeekendFieldActivities(ctx, scope, s.fieldActivityTypes(), filter)
	if err != nil {
		return fmt.Errorf("checking recovery balance: %w", err)
	}

	recoveryDates, err := s.dashboard.RecoveryActivities(ctx, scope, recoveryRule.RecoveryType, filter)
	if err != nil {
		return fmt.Errorf("checking recovery balance: %w", err)
	}

	if hasUnclaimedWindow(activity.DueDate, weekendActivities, recoveryDates, recoveryRule.RecoveryWindowDays) {
		return nil
	}
	return ErrNoRecoveryBalance
}

// isRecoveryActivity returns true if the activity is a recovery-type activity
// and recovery balance checking is applicable.
func (s *ActivityService) isRecoveryActivity(activity *domain.Activity) bool {
	if s.cfg == nil || s.cfg.Recovery == nil || !s.cfg.Recovery.WeekendActivityFlag {
		return false
	}
	if activity.ActivityType != s.cfg.Recovery.RecoveryType {
		return false
	}
	return s.dashboard != nil
}

// hasUnclaimedWindow checks whether the given dueDate falls within at least one
// unclaimed recovery claim window among the weekend activities.
func hasUnclaimedWindow(dueDate time.Time, weekendActivities []store.WeekendActivity, recoveryDates []time.Time, windowDays int) bool {
	takenSet := make(map[string]bool)
	for _, rd := range recoveryDates {
		takenSet[rd.Format(dateFormat)] = true
	}

	for _, wa := range weekendActivities {
		claimFrom := nextBusinessDay(wa.DueDate)
		claimBy := addBusinessDays(claimFrom, windowDays-1)

		if dueDate.Before(claimFrom) || dueDate.After(claimBy) {
			continue
		}

		if !isWindowClaimed(claimFrom, claimBy, recoveryDates, takenSet) {
			return true
		}
	}
	return false
}

// isWindowClaimed checks if a claim window already has a recovery activity claimed against it.
func isWindowClaimed(claimFrom, claimBy time.Time, recoveryDates []time.Time, takenSet map[string]bool) bool {
	for _, rd := range recoveryDates {
		rdKey := rd.Format(dateFormat)
		if !rd.Before(claimFrom) && !rd.After(claimBy) && !takenSet[rdKey+"_used"] {
			takenSet[rdKey+"_used"] = true
			return true
		}
	}
	return false
}

// recordAudit persists an audit entry. Failures are logged but do not block the caller.
func (s *ActivityService) recordAudit(ctx context.Context, entry *domain.AuditEntry) {
	if err := s.audit.Record(ctx, entry); err != nil {
		slog.Default().Warn("audit record failed", "entity", entry.EntityID, "event", entry.EventType, "err", err)
	}
}

// checkTargetAccess verifies that the actor can view the referenced target.
// Non-admin users may only create/update activities against targets within their visible scope.
func (s *ActivityService) checkTargetAccess(ctx context.Context, actor *domain.User, targetID string) error {
	if targetID == "" {
		return nil // non-field activities have no target
	}
	if actor.Role == domain.RoleAdmin {
		return nil
	}
	if s.targets == nil {
		return nil // no target repo configured (e.g. in tests without targets)
	}
	target, err := s.targets.Get(ctx, targetID)
	if err != nil {
		return fmt.Errorf("checking target access: %w", err)
	}
	if !s.enforcer.CanViewTarget(ctx, actor, target) {
		return ErrTargetNotAccessible
	}
	return nil
}

// checkJointVisitUser validates the joint visit user ID when set.
// It must not be the creator themselves, and must reference an existing user.
func (s *ActivityService) checkJointVisitUser(ctx context.Context, creatorID, jointVisitUID string) error {
	if jointVisitUID == "" {
		return nil
	}
	if jointVisitUID == creatorID {
		return ErrInvalidJointVisitor
	}
	if s.users != nil {
		if _, err := s.users.GetByID(ctx, jointVisitUID); err != nil {
			return ErrInvalidJointVisitor
		}
	}
	return nil
}
