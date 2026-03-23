package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pebblr/pebblr/internal/config"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/service"
	"github.com/pebblr/pebblr/internal/store"
)

// --- stub activity repo ---

type stubActivityRepo struct {
	activity             *domain.Activity
	created              *domain.Activity
	updated              *domain.Activity
	getErr               error
	saveErr              error
	countByDate          int
	hasActivityWithTypes bool
}

func (r *stubActivityRepo) Get(_ context.Context, _ string) (*domain.Activity, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	// Return a copy so tests can mutate the original.
	a := *r.activity
	return &a, nil
}

func (r *stubActivityRepo) List(_ context.Context, _ rbac.ActivityScope, _ store.ActivityFilter, page, limit int) (*store.ActivityPage, error) {
	if r.activity != nil {
		return &store.ActivityPage{Activities: []*domain.Activity{r.activity}, Total: 1, Page: page, Limit: limit}, nil
	}
	return &store.ActivityPage{Page: page, Limit: limit}, nil
}

func (r *stubActivityRepo) Create(_ context.Context, activity *domain.Activity) (*domain.Activity, error) {
	if r.saveErr != nil {
		return nil, r.saveErr
	}
	activity.ID = "activity-1"
	activity.CreatedAt = time.Now().UTC()
	activity.UpdatedAt = time.Now().UTC()
	r.created = activity
	return activity, nil
}

func (r *stubActivityRepo) Update(_ context.Context, activity *domain.Activity) (*domain.Activity, error) {
	if r.saveErr != nil {
		return nil, r.saveErr
	}
	activity.UpdatedAt = time.Now().UTC()
	r.updated = activity
	return activity, nil
}

func (r *stubActivityRepo) SoftDelete(_ context.Context, _ string) error {
	return r.saveErr
}

func (r *stubActivityRepo) CountByDate(_ context.Context, _ string, _ time.Time) (int, error) {
	return r.countByDate, nil
}

func (r *stubActivityRepo) HasActivityWithTypes(_ context.Context, _ string, _ time.Time, _ []string) (bool, error) {
	return r.hasActivityWithTypes, nil
}

// --- stub audit repo ---

type stubAuditRepo struct {
	entries []*domain.AuditEntry
}

func (r *stubAuditRepo) Record(_ context.Context, entry *domain.AuditEntry) error {
	r.entries = append(r.entries, entry)
	return nil
}

func (r *stubAuditRepo) ListByEntity(_ context.Context, _, _ string) ([]*domain.AuditEntry, error) {
	return r.entries, nil
}

// --- test config with activities ---

func activityTestConfig() *config.TenantConfig {
	return &config.TenantConfig{
		Activities: config.ActivitiesConfig{
			Statuses: []config.StatusDef{
				{Key: "planificat", Label: "Planned", Initial: true},
				{Key: "realizat", Label: "Realized"},
				{Key: "anulat", Label: "Cancelled"},
			},
			StatusTransitions: map[string][]string{
				"planificat": {"realizat", "anulat"},
				"realizat":   {"anulat"},
			},
			Durations: []config.OptionDef{
				{Key: "full_day", Label: "Full Day"},
				{Key: "half_day", Label: "Half Day"},
			},
			Types: []config.ActivityTypeConfig{
				{
					Key:      "visit",
					Label:    "Visit",
					Category: "field",
					Fields: []config.FieldConfig{
						{Key: "notes", Type: "text"},
					},
					SubmitRequired: []string{"notes"},
				},
				{
					Key:                   "vacation",
					Label:                 "Vacation",
					Category:              "non_field",
					BlocksFieldActivities: true,
				},
			},
		},
		Rules: config.RulesConfig{
			MaxActivitiesPerDay: 10,
		},
	}
}

func sampleActivity() *domain.Activity {
	return &domain.Activity{
		ID:           "activity-1",
		ActivityType: "visit",
		Status:       "planificat",
		DueDate:      time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC),
		Duration:     "full_day",
		Fields:       map[string]any{},
		CreatorID:    "rep-1",
		TeamID:       "team-1",
	}
}

func newActivitySvc(repo *stubActivityRepo, audit *stubAuditRepo) *service.ActivityService {
	return service.NewActivityService(repo, audit, rbac.NewEnforcer(), activityTestConfig())
}

// --- Create tests ---

func TestActivityCreate_RepSucceeds(t *testing.T) {
	t.Parallel()
	repo := &stubActivityRepo{}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	activity := &domain.Activity{
		ActivityType: "visit",
		DueDate:      time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC),
		Duration:     "full_day",
		Fields:       map[string]any{},
		TargetID:     "target-1",
	}
	created, err := svc.Create(context.Background(), repUser(), activity)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID == "" {
		t.Error("expected ID to be set")
	}
	if created.CreatorID != "rep-1" {
		t.Errorf("expected creator rep-1, got %s", created.CreatorID)
	}
	if created.Status != "planificat" {
		t.Errorf("expected initial status planificat, got %s", created.Status)
	}
	if len(audit.entries) != 1 {
		t.Errorf("expected 1 audit entry, got %d", len(audit.entries))
	}
}

func TestActivityCreate_InvalidType(t *testing.T) {
	t.Parallel()
	repo := &stubActivityRepo{}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	activity := &domain.Activity{
		ActivityType: "unknown",
		DueDate:      time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC),
		Fields:       map[string]any{},
	}
	_, err := svc.Create(context.Background(), repUser(), activity)
	if !errors.Is(err, service.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestActivityCreate_MissingDueDate(t *testing.T) {
	t.Parallel()
	repo := &stubActivityRepo{}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	activity := &domain.Activity{
		ActivityType: "visit",
		Fields:       map[string]any{},
	}
	_, err := svc.Create(context.Background(), repUser(), activity)
	if !errors.Is(err, service.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestActivityCreate_InvalidDuration(t *testing.T) {
	t.Parallel()
	repo := &stubActivityRepo{}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	activity := &domain.Activity{
		ActivityType: "visit",
		DueDate:      time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC),
		Duration:     "invalid",
		Fields:       map[string]any{},
	}
	_, err := svc.Create(context.Background(), repUser(), activity)
	if !errors.Is(err, service.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestActivityCreate_MaxActivitiesPerDay(t *testing.T) {
	t.Parallel()
	repo := &stubActivityRepo{countByDate: 10}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	activity := &domain.Activity{
		ActivityType: "visit",
		DueDate:      time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC),
		Fields:       map[string]any{},
		TargetID:     "target-1",
	}
	_, err := svc.Create(context.Background(), repUser(), activity)
	if !errors.Is(err, service.ErrMaxActivities) {
		t.Errorf("expected ErrMaxActivities, got %v", err)
	}
}

func TestActivityCreate_BlockedByVacation(t *testing.T) {
	t.Parallel()
	repo := &stubActivityRepo{hasActivityWithTypes: true}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	activity := &domain.Activity{
		ActivityType: "visit",
		DueDate:      time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC),
		Fields:       map[string]any{},
		TargetID:     "target-1",
	}
	_, err := svc.Create(context.Background(), repUser(), activity)
	if !errors.Is(err, service.ErrBlockedDay) {
		t.Errorf("expected ErrBlockedDay, got %v", err)
	}
}

func TestActivityCreate_BlockingTypeBlockedByFieldActivity(t *testing.T) {
	t.Parallel()
	repo := &stubActivityRepo{hasActivityWithTypes: true}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	activity := &domain.Activity{
		ActivityType: "vacation",
		DueDate:      time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC),
		Fields:       map[string]any{},
	}
	_, err := svc.Create(context.Background(), repUser(), activity)
	if !errors.Is(err, service.ErrBlockedDay) {
		t.Errorf("expected ErrBlockedDay, got %v", err)
	}
}

func TestActivityCreate_VacationAllowedOnEmptyDay(t *testing.T) {
	t.Parallel()
	repo := &stubActivityRepo{hasActivityWithTypes: false}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	activity := &domain.Activity{
		ActivityType: "vacation",
		DueDate:      time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC),
		Fields:       map[string]any{},
	}
	created, err := svc.Create(context.Background(), repUser(), activity)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID == "" {
		t.Error("expected ID to be set")
	}
}

func TestActivityCreate_TargetRequiredForFieldActivity(t *testing.T) {
	t.Parallel()
	repo := &stubActivityRepo{}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	activity := &domain.Activity{
		ActivityType: "visit",
		DueDate:      time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC),
		Fields:       map[string]any{},
		// TargetID intentionally empty
	}
	_, err := svc.Create(context.Background(), repUser(), activity)
	if !errors.Is(err, service.ErrTargetRequired) {
		t.Errorf("expected ErrTargetRequired, got %v", err)
	}
}

func TestActivityCreate_TargetNotRequiredForNonField(t *testing.T) {
	t.Parallel()
	repo := &stubActivityRepo{}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	activity := &domain.Activity{
		ActivityType: "vacation",
		DueDate:      time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC),
		Fields:       map[string]any{},
		// No TargetID — should be fine for non_field
	}
	created, err := svc.Create(context.Background(), repUser(), activity)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID == "" {
		t.Error("expected ID to be set")
	}
}

// --- Get tests ---

func TestActivityGet_RepOwnsActivity(t *testing.T) {
	t.Parallel()
	repo := &stubActivityRepo{activity: sampleActivity()}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	activity, err := svc.Get(context.Background(), repUser(), "activity-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if activity.ID != "activity-1" {
		t.Errorf("expected activity-1, got %s", activity.ID)
	}
}

func TestActivityGet_RepForbiddenOther(t *testing.T) {
	t.Parallel()
	other := sampleActivity()
	other.CreatorID = "other-rep"
	other.TeamID = "team-2"
	repo := &stubActivityRepo{activity: other}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	_, err := svc.Get(context.Background(), repUser(), "activity-1")
	if !errors.Is(err, service.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestActivityGet_NotFound(t *testing.T) {
	t.Parallel()
	repo := &stubActivityRepo{getErr: store.ErrNotFound}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	_, err := svc.Get(context.Background(), adminUser(), "missing")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// --- List tests ---

func TestActivityList_RepScoped(t *testing.T) {
	t.Parallel()
	repo := &stubActivityRepo{activity: sampleActivity()}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	page, err := svc.List(context.Background(), repUser(), store.ActivityFilter{}, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Total != 1 {
		t.Errorf("expected 1 activity, got %d", page.Total)
	}
}

// --- Update tests ---

func TestActivityUpdate_RepOwnsActivity(t *testing.T) {
	t.Parallel()
	existing := sampleActivity()
	repo := &stubActivityRepo{activity: existing}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	update := &domain.Activity{
		ActivityType: "visit",
		Status:       "planificat",
		DueDate:      time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC),
		Fields:       map[string]any{},
	}
	updated, err := svc.Update(context.Background(), repUser(), "activity-1", update)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.ID != "activity-1" {
		t.Errorf("expected activity-1, got %s", updated.ID)
	}
	if updated.CreatorID != "rep-1" {
		t.Errorf("expected creator preserved as rep-1, got %s", updated.CreatorID)
	}
}

func TestActivityUpdate_RepForbiddenOther(t *testing.T) {
	t.Parallel()
	other := sampleActivity()
	other.CreatorID = "other-rep"
	other.TeamID = "team-2"
	repo := &stubActivityRepo{activity: other}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	update := &domain.Activity{
		ActivityType: "visit",
		Status:       "planificat",
		DueDate:      time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC),
		Fields:       map[string]any{},
	}
	_, err := svc.Update(context.Background(), repUser(), "activity-1", update)
	if !errors.Is(err, service.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestActivityUpdate_SubmittedBlocked(t *testing.T) {
	t.Parallel()
	submitted := sampleActivity()
	now := time.Now()
	submitted.SubmittedAt = &now
	repo := &stubActivityRepo{activity: submitted}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	update := &domain.Activity{
		ActivityType: "visit",
		Status:       "planificat",
		DueDate:      time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC),
		Fields:       map[string]any{},
	}
	_, err := svc.Update(context.Background(), repUser(), "activity-1", update)
	if !errors.Is(err, service.ErrSubmitted) {
		t.Errorf("expected ErrSubmitted, got %v", err)
	}
}

// --- Delete tests ---

func TestActivityDelete_RepOwns(t *testing.T) {
	t.Parallel()
	repo := &stubActivityRepo{activity: sampleActivity()}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	err := svc.Delete(context.Background(), repUser(), "activity-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(audit.entries) != 1 || audit.entries[0].EventType != "deleted" {
		t.Error("expected deleted audit entry")
	}
}

func TestActivityDelete_SubmittedBlocked(t *testing.T) {
	t.Parallel()
	submitted := sampleActivity()
	now := time.Now()
	submitted.SubmittedAt = &now
	repo := &stubActivityRepo{activity: submitted}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	err := svc.Delete(context.Background(), repUser(), "activity-1")
	if !errors.Is(err, service.ErrSubmitted) {
		t.Errorf("expected ErrSubmitted, got %v", err)
	}
}

// --- Submit tests ---

func TestActivitySubmit_Succeeds(t *testing.T) {
	t.Parallel()
	existing := sampleActivity()
	existing.Fields = map[string]any{"notes": "Visit completed successfully"}
	repo := &stubActivityRepo{activity: existing}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	submitted, err := svc.Submit(context.Background(), repUser(), "activity-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if submitted.SubmittedAt == nil {
		t.Error("expected SubmittedAt to be set")
	}
}

func TestActivitySubmit_MissingRequiredFields(t *testing.T) {
	t.Parallel()
	existing := sampleActivity()
	existing.Fields = map[string]any{} // "notes" is submit_required but missing
	repo := &stubActivityRepo{activity: existing}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	_, err := svc.Submit(context.Background(), repUser(), "activity-1")
	var ve *service.ValidationErrors
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationErrors, got %v", err)
	}
	if len(ve.Errors) == 0 {
		t.Error("expected at least one validation error")
	}
}

func TestActivitySubmit_AlreadySubmitted(t *testing.T) {
	t.Parallel()
	submitted := sampleActivity()
	now := time.Now()
	submitted.SubmittedAt = &now
	repo := &stubActivityRepo{activity: submitted}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	_, err := svc.Submit(context.Background(), repUser(), "activity-1")
	if !errors.Is(err, service.ErrSubmitted) {
		t.Errorf("expected ErrSubmitted, got %v", err)
	}
}

// --- PatchStatus tests ---

func TestActivityPatchStatus_ValidTransition(t *testing.T) {
	t.Parallel()
	existing := sampleActivity()
	repo := &stubActivityRepo{activity: existing}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	updated, err := svc.PatchStatus(context.Background(), repUser(), "activity-1", "realizat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != "realizat" {
		t.Errorf("expected realizat, got %s", updated.Status)
	}
	// Check audit recorded status change.
	found := false
	for _, e := range audit.entries {
		if e.EventType == "status_changed" {
			found = true
		}
	}
	if !found {
		t.Error("expected status_changed audit entry")
	}
}

func TestActivityPatchStatus_InvalidTransition(t *testing.T) {
	t.Parallel()
	existing := sampleActivity()
	existing.Status = "realizat"
	repo := &stubActivityRepo{activity: existing}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	_, err := svc.PatchStatus(context.Background(), repUser(), "activity-1", "planificat")
	var ve *service.ValidationErrors
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationErrors, got %v", err)
	}
}

func TestActivityPatchStatus_SubmittedBlocked(t *testing.T) {
	t.Parallel()
	submitted := sampleActivity()
	now := time.Now()
	submitted.SubmittedAt = &now
	repo := &stubActivityRepo{activity: submitted}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	_, err := svc.PatchStatus(context.Background(), repUser(), "activity-1", "realizat")
	if !errors.Is(err, service.ErrSubmitted) {
		t.Errorf("expected ErrSubmitted, got %v", err)
	}
}
