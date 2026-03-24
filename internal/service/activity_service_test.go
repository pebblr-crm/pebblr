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
	listActivities       []*domain.Activity
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
	if r.listActivities != nil {
		return &store.ActivityPage{Activities: r.listActivities, Total: len(r.listActivities), Page: page, Limit: limit}, nil
	}
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

// --- stub user repo ---

type stubUserRepo struct {
	users map[string]*domain.User
}

func (r *stubUserRepo) GetByID(_ context.Context, id string) (*domain.User, error) {
	if u, ok := r.users[id]; ok {
		return u, nil
	}
	return nil, store.ErrNotFound
}

func (r *stubUserRepo) GetByExternalID(_ context.Context, _ string) (*domain.User, error) {
	return nil, store.ErrNotFound
}

func (r *stubUserRepo) List(_ context.Context) ([]*domain.User, error) {
	return nil, nil
}

func (r *stubUserRepo) Upsert(_ context.Context, u *domain.User) (*domain.User, error) {
	return u, nil
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
				{Key: "realizat", Label: "Realized", Submittable: true},
				{Key: "anulat", Label: "Cancelled", Submittable: true},
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
				{
					Key:                   "recovery",
					Label:                 "Recovery Day",
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

func activityTestConfigWithRecovery() *config.TenantConfig {
	cfg := activityTestConfig()
	cfg.Recovery = &config.RecoveryRule{
		WeekendActivityFlag: true,
		RecoveryWindowDays:  5,
		RecoveryType:        "recovery",
	}
	return cfg
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

func defaultUserRepo() *stubUserRepo {
	return &stubUserRepo{users: map[string]*domain.User{
		"rep-1":   {ID: "rep-1", Name: "Rep User", Role: domain.RoleRep, TeamIDs: []string{"team-1"}},
		"rep-2":   {ID: "rep-2", Name: "Rep Two", Role: domain.RoleRep, TeamIDs: []string{"team-1"}},
		"mgr-1":   {ID: "mgr-1", Name: "Manager User", Role: domain.RoleManager, TeamIDs: []string{"team-1"}},
		"admin-1": {ID: "admin-1", Name: "Admin User", Role: domain.RoleAdmin, TeamIDs: []string{"team-1"}},
	}}
}

func newActivitySvc(repo *stubActivityRepo, audit *stubAuditRepo) *service.ActivityService {
	return service.NewActivityService(repo, nil, defaultUserRepo(), audit, rbac.NewEnforcer(), activityTestConfig())
}

func newActivitySvcWithTargets(repo *stubActivityRepo, targets *stubTargetRepo, audit *stubAuditRepo) *service.ActivityService {
	return service.NewActivityService(repo, targets, defaultUserRepo(), audit, rbac.NewEnforcer(), activityTestConfig())
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
	existing.Status = "realizat" // closed status required for submission
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

func TestActivitySubmit_PlannedStatusRejected(t *testing.T) {
	t.Parallel()
	existing := sampleActivity() // status is "planificat" (not submittable)
	existing.Fields = map[string]any{"notes": "some notes"}
	repo := &stubActivityRepo{activity: existing}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	_, err := svc.Submit(context.Background(), repUser(), "activity-1")
	if !errors.Is(err, service.ErrStatusNotSubmittable) {
		t.Errorf("expected ErrStatusNotSubmittable, got %v", err)
	}
}

func TestActivitySubmit_MissingRequiredFields(t *testing.T) {
	t.Parallel()
	existing := sampleActivity()
	existing.Status = "realizat" // closed status, but missing required fields
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

// --- PartialUpdate tests ---

func TestActivityPartialUpdate_StatusOnly(t *testing.T) {
	t.Parallel()
	existing := sampleActivity()
	repo := &stubActivityRepo{activity: existing}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	newStatus := "realizat"
	patch := &domain.ActivityPatch{Status: &newStatus}
	updated, err := svc.PartialUpdate(context.Background(), repUser(), "activity-1", patch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != "realizat" {
		t.Errorf("expected realizat, got %s", updated.Status)
	}
	// DueDate should be unchanged.
	if !updated.DueDate.Equal(existing.DueDate) {
		t.Errorf("expected DueDate unchanged, got %v", updated.DueDate)
	}
}

func TestActivityPartialUpdate_DueDate(t *testing.T) {
	t.Parallel()
	existing := sampleActivity()
	repo := &stubActivityRepo{activity: existing}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	newDate := time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC)
	patch := &domain.ActivityPatch{DueDate: &newDate}
	updated, err := svc.PartialUpdate(context.Background(), repUser(), "activity-1", patch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updated.DueDate.Equal(newDate) {
		t.Errorf("expected %v, got %v", newDate, updated.DueDate)
	}
}

func TestActivityPartialUpdate_FieldsMerge(t *testing.T) {
	t.Parallel()
	existing := sampleActivity()
	existing.Fields = map[string]any{"notes": "old note", "keep": "me"}
	repo := &stubActivityRepo{activity: existing}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	patch := &domain.ActivityPatch{
		Fields:        map[string]any{"notes": "new note"},
		FieldsPresent: true,
	}
	updated, err := svc.PartialUpdate(context.Background(), repUser(), "activity-1", patch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Fields["notes"] != "new note" {
		t.Errorf("expected notes=new note, got %v", updated.Fields["notes"])
	}
	// Unpatched key must survive.
	if updated.Fields["keep"] != "me" {
		t.Errorf("expected keep=me, got %v", updated.Fields["keep"])
	}
}

func TestActivityPartialUpdate_FieldsNullClearsKey(t *testing.T) {
	t.Parallel()
	existing := sampleActivity()
	existing.Fields = map[string]any{"notes": "old note"}
	repo := &stubActivityRepo{activity: existing}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	patch := &domain.ActivityPatch{
		Fields:        map[string]any{"notes": nil},
		FieldsPresent: true,
	}
	updated, err := svc.PartialUpdate(context.Background(), repUser(), "activity-1", patch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, exists := updated.Fields["notes"]; exists {
		t.Error("expected notes key to be cleared")
	}
}

func TestActivityPartialUpdate_FieldsAbsentLeavesUntouched(t *testing.T) {
	t.Parallel()
	existing := sampleActivity()
	existing.Fields = map[string]any{"notes": "preserved"}
	repo := &stubActivityRepo{activity: existing}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	// Patch with no fields key.
	newStatus := "realizat"
	patch := &domain.ActivityPatch{Status: &newStatus}
	updated, err := svc.PartialUpdate(context.Background(), repUser(), "activity-1", patch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Fields["notes"] != "preserved" {
		t.Errorf("expected notes preserved, got %v", updated.Fields["notes"])
	}
}

func TestActivityPartialUpdate_ForbiddenOtherRep(t *testing.T) {
	t.Parallel()
	other := sampleActivity()
	other.CreatorID = "other-rep"
	other.TeamID = "team-2"
	repo := &stubActivityRepo{activity: other}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	newStatus := "realizat"
	patch := &domain.ActivityPatch{Status: &newStatus}
	_, err := svc.PartialUpdate(context.Background(), repUser(), "activity-1", patch)
	if !errors.Is(err, service.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestActivityPartialUpdate_SubmittedBlocked(t *testing.T) {
	t.Parallel()
	submitted := sampleActivity()
	now := time.Now()
	submitted.SubmittedAt = &now
	repo := &stubActivityRepo{activity: submitted}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	newStatus := "anulat"
	patch := &domain.ActivityPatch{Status: &newStatus}
	_, err := svc.PartialUpdate(context.Background(), repUser(), "activity-1", patch)
	if !errors.Is(err, service.ErrSubmitted) {
		t.Errorf("expected ErrSubmitted, got %v", err)
	}
}

func TestActivityPartialUpdate_AuditRecorded(t *testing.T) {
	t.Parallel()
	existing := sampleActivity()
	repo := &stubActivityRepo{activity: existing}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	newStatus := "realizat"
	patch := &domain.ActivityPatch{Status: &newStatus}
	_, err := svc.PartialUpdate(context.Background(), repUser(), "activity-1", patch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(audit.entries) != 1 || audit.entries[0].EventType != "updated" {
		t.Errorf("expected 1 updated audit entry, got %v", audit.entries)
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

// --- Joint visit tests ---

func TestActivityCreate_WithValidJointVisitor(t *testing.T) {
	t.Parallel()
	repo := &stubActivityRepo{}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	activity := &domain.Activity{
		ActivityType:  "visit",
		DueDate:       time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC),
		Duration:      "full_day",
		Fields:        map[string]any{},
		TargetID:      "target-1",
		JointVisitUID: "rep-2",
	}
	created, err := svc.Create(context.Background(), repUser(), activity)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.JointVisitUID != "rep-2" {
		t.Errorf("expected joint visit user rep-2, got %s", created.JointVisitUID)
	}
}

func TestActivityCreate_SelfJointVisitorRejected(t *testing.T) {
	t.Parallel()
	repo := &stubActivityRepo{}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	activity := &domain.Activity{
		ActivityType:  "visit",
		DueDate:       time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC),
		Duration:      "full_day",
		Fields:        map[string]any{},
		TargetID:      "target-1",
		JointVisitUID: "rep-1", // same as creator
	}
	_, err := svc.Create(context.Background(), repUser(), activity)
	if !errors.Is(err, service.ErrInvalidJointVisitor) {
		t.Errorf("expected ErrInvalidJointVisitor, got %v", err)
	}
}

func TestActivityCreate_NonExistentJointVisitorRejected(t *testing.T) {
	t.Parallel()
	repo := &stubActivityRepo{}
	audit := &stubAuditRepo{}
	svc := newActivitySvc(repo, audit)

	activity := &domain.Activity{
		ActivityType:  "visit",
		DueDate:       time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC),
		Duration:      "full_day",
		Fields:        map[string]any{},
		TargetID:      "target-1",
		JointVisitUID: "nonexistent-user",
	}
	_, err := svc.Create(context.Background(), repUser(), activity)
	if !errors.Is(err, service.ErrInvalidJointVisitor) {
		t.Errorf("expected ErrInvalidJointVisitor, got %v", err)
	}
}

func TestActivityGet_RepCanViewAsJointVisitor(t *testing.T) {
	t.Parallel()
	jointActivity := sampleActivity()
	jointActivity.CreatorID = "rep-2"
	jointActivity.JointVisitUID = "rep-1"
	repo := &stubActivityRepo{activity: jointActivity}
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

func TestActivityUpdate_JointVisitorCannotUpdate(t *testing.T) {
	t.Parallel()
	jointActivity := sampleActivity()
	jointActivity.CreatorID = "rep-2"
	jointActivity.JointVisitUID = "rep-1"
	repo := &stubActivityRepo{activity: jointActivity}
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
		t.Errorf("expected ErrForbidden for joint visitor update, got %v", err)
	}
}

func TestActivityCreate_EmptyJointVisitorAllowed(t *testing.T) {
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
	if created.JointVisitUID != "" {
		t.Errorf("expected empty joint visit user, got %s", created.JointVisitUID)
	}
}

// --- CloneWeek tests ---

func TestCloneWeek_Success(t *testing.T) {
	t.Parallel()
	monday := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)
	repo := &stubActivityRepo{
		listActivities: []*domain.Activity{
			{ID: "a1", ActivityType: "visit", DueDate: monday, Duration: "full_day", TargetID: "t1", Fields: map[string]any{"visit_type": "f2f"}},
			{ID: "a2", ActivityType: "visit", DueDate: monday.AddDate(0, 0, 2), Duration: "full_day", TargetID: "t2", Fields: map[string]any{"visit_type": "f2f"}},
		},
	}
	svc := newActivitySvc(repo, &stubAuditRepo{})

	targetMonday := monday.AddDate(0, 0, 21) // +3 weeks
	result, err := svc.CloneWeek(context.Background(), repUser(), monday, targetMonday)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// First List call returns source activities, second List call returns empty target week,
	// but our stub returns the same list for both calls. The duplicate detection uses
	// targetID+date, and since dates are shifted, no conflicts → all created.
	if result.Created == 0 {
		t.Error("expected at least 1 created activity")
	}
}

func TestCloneWeek_SourceNotMonday(t *testing.T) {
	t.Parallel()
	tuesday := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC) // Tuesday
	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	repo := &stubActivityRepo{}
	svc := newActivitySvc(repo, &stubAuditRepo{})

	_, err := svc.CloneWeek(context.Background(), repUser(), tuesday, monday)
	if err == nil {
		t.Fatal("expected error for non-Monday source")
	}
}

func TestCloneWeek_TargetNotMonday(t *testing.T) {
	t.Parallel()
	monday := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)
	wednesday := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC) // Wednesday
	repo := &stubActivityRepo{}
	svc := newActivitySvc(repo, &stubAuditRepo{})

	_, err := svc.CloneWeek(context.Background(), repUser(), monday, wednesday)
	if err == nil {
		t.Fatal("expected error for non-Monday target")
	}
}

func TestCloneWeek_SameWeek(t *testing.T) {
	t.Parallel()
	monday := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)
	repo := &stubActivityRepo{}
	svc := newActivitySvc(repo, &stubAuditRepo{})

	_, err := svc.CloneWeek(context.Background(), repUser(), monday, monday)
	if err == nil {
		t.Fatal("expected error for same source and target week")
	}
}

func TestCloneWeek_EmptySourceWeek(t *testing.T) {
	t.Parallel()
	monday := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)
	target := monday.AddDate(0, 0, 21)
	repo := &stubActivityRepo{} // no activities
	svc := newActivitySvc(repo, &stubAuditRepo{})

	_, err := svc.CloneWeek(context.Background(), repUser(), monday, target)
	if err == nil {
		t.Fatal("expected error for empty source week")
	}
}

// --- Recovery balance tests ---

func TestCreate_RecoveryWithBalance_Succeeds(t *testing.T) {
	t.Parallel()
	// Saturday weekend work earns a recovery day; claim on Monday (next business day).
	saturday := time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC) // Saturday
	monday := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)   // Monday

	dashRepo := &stubDashboardRepo{
		weekendActivities:  []store.WeekendActivity{{DueDate: saturday}},
		recoveryActivities: []time.Time{}, // none taken yet
	}
	actRepo := &stubActivityRepo{}
	cfg := activityTestConfigWithRecovery()
	svc := service.NewActivityService(actRepo, nil, defaultUserRepo(), &stubAuditRepo{}, rbac.NewEnforcer(), cfg, service.WithDashboard(dashRepo))

	activity := &domain.Activity{
		ActivityType: "recovery",
		DueDate:      monday,
		Duration:     "full_day",
		Fields:       map[string]any{},
	}
	_, err := svc.Create(context.Background(), repUser(), activity)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestCreate_RecoveryWithZeroBalance_Fails(t *testing.T) {
	t.Parallel()
	monday := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)

	// No weekend activities — zero earned balance.
	dashRepo := &stubDashboardRepo{
		weekendActivities:  []store.WeekendActivity{},
		recoveryActivities: []time.Time{},
	}
	actRepo := &stubActivityRepo{}
	cfg := activityTestConfigWithRecovery()
	svc := service.NewActivityService(actRepo, nil, defaultUserRepo(), &stubAuditRepo{}, rbac.NewEnforcer(), cfg, service.WithDashboard(dashRepo))

	activity := &domain.Activity{
		ActivityType: "recovery",
		DueDate:      monday,
		Duration:     "full_day",
		Fields:       map[string]any{},
	}
	_, err := svc.Create(context.Background(), repUser(), activity)
	if !errors.Is(err, service.ErrNoRecoveryBalance) {
		t.Fatalf("expected ErrNoRecoveryBalance, got: %v", err)
	}
}

func TestCreate_RecoveryAlreadyClaimed_Fails(t *testing.T) {
	t.Parallel()
	saturday := time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)
	monday := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)
	tuesday := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)

	// Weekend earned, but already claimed on Monday.
	dashRepo := &stubDashboardRepo{
		weekendActivities:  []store.WeekendActivity{{DueDate: saturday}},
		recoveryActivities: []time.Time{monday}, // already claimed
	}
	actRepo := &stubActivityRepo{}
	cfg := activityTestConfigWithRecovery()
	svc := service.NewActivityService(actRepo, nil, defaultUserRepo(), &stubAuditRepo{}, rbac.NewEnforcer(), cfg, service.WithDashboard(dashRepo))

	activity := &domain.Activity{
		ActivityType: "recovery",
		DueDate:      tuesday,
		Duration:     "full_day",
		Fields:       map[string]any{},
	}
	_, err := svc.Create(context.Background(), repUser(), activity)
	if !errors.Is(err, service.ErrNoRecoveryBalance) {
		t.Fatalf("expected ErrNoRecoveryBalance, got: %v", err)
	}
}

func TestCreate_RecoveryOutsideClaimWindow_Fails(t *testing.T) {
	t.Parallel()
	saturday := time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)
	// Claim window is Mon 23–Fri 27 (5 business days). The following Monday is outside.
	nextMonday := time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC)

	dashRepo := &stubDashboardRepo{
		weekendActivities:  []store.WeekendActivity{{DueDate: saturday}},
		recoveryActivities: []time.Time{},
	}
	actRepo := &stubActivityRepo{}
	cfg := activityTestConfigWithRecovery()
	svc := service.NewActivityService(actRepo, nil, defaultUserRepo(), &stubAuditRepo{}, rbac.NewEnforcer(), cfg, service.WithDashboard(dashRepo))

	activity := &domain.Activity{
		ActivityType: "recovery",
		DueDate:      nextMonday,
		Duration:     "full_day",
		Fields:       map[string]any{},
	}
	_, err := svc.Create(context.Background(), repUser(), activity)
	if !errors.Is(err, service.ErrNoRecoveryBalance) {
		t.Fatalf("expected ErrNoRecoveryBalance, got: %v", err)
	}
}

func TestCreate_NonRecoveryType_SkipsBalanceCheck(t *testing.T) {
	t.Parallel()
	// Even with no weekend activities, creating a visit should work fine.
	dashRepo := &stubDashboardRepo{
		weekendActivities:  []store.WeekendActivity{},
		recoveryActivities: []time.Time{},
	}
	actRepo := &stubActivityRepo{}
	cfg := activityTestConfigWithRecovery()
	svc := service.NewActivityService(actRepo, nil, defaultUserRepo(), &stubAuditRepo{}, rbac.NewEnforcer(), cfg, service.WithDashboard(dashRepo))

	activity := &domain.Activity{
		ActivityType: "visit",
		DueDate:      time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC),
		Duration:     "full_day",
		TargetID:     "target-1",
		Fields:       map[string]any{},
	}
	_, err := svc.Create(context.Background(), repUser(), activity)
	if err != nil {
		t.Fatalf("expected success for non-recovery type, got: %v", err)
	}
}

// --- Target access tests ---

func TestActivityCreate_RepOwnTargetSucceeds(t *testing.T) {
	t.Parallel()
	actRepo := &stubActivityRepo{}
	targetRepo := &stubTargetRepo{target: sampleTarget()} // assigned to rep-1
	audit := &stubAuditRepo{}
	svc := newActivitySvcWithTargets(actRepo, targetRepo, audit)

	activity := &domain.Activity{
		ActivityType: "visit",
		DueDate:      time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC),
		Duration:     "full_day",
		Fields:       map[string]any{},
		TargetID:     "target-1",
	}
	_, err := svc.Create(context.Background(), repUser(), activity)
	if err != nil {
		t.Fatalf("expected success for own target, got: %v", err)
	}
}

func TestActivityCreate_RepOtherTargetFails(t *testing.T) {
	t.Parallel()
	otherTarget := &domain.Target{ID: "target-2", TargetType: "doctor", Name: "Other Dr", AssigneeID: "rep-2", TeamID: "team-1", Fields: map[string]any{}}
	actRepo := &stubActivityRepo{}
	targetRepo := &stubTargetRepo{target: otherTarget}
	audit := &stubAuditRepo{}
	svc := newActivitySvcWithTargets(actRepo, targetRepo, audit)

	activity := &domain.Activity{
		ActivityType: "visit",
		DueDate:      time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC),
		Duration:     "full_day",
		Fields:       map[string]any{},
		TargetID:     "target-2",
	}
	_, err := svc.Create(context.Background(), repUser(), activity)
	if !errors.Is(err, service.ErrTargetNotAccessible) {
		t.Errorf("expected ErrTargetNotAccessible, got %v", err)
	}
}

func TestActivityCreate_AdminAnyTargetSucceeds(t *testing.T) {
	t.Parallel()
	otherTarget := &domain.Target{ID: "target-2", TargetType: "doctor", Name: "Other Dr", AssigneeID: "rep-2", TeamID: "team-2", Fields: map[string]any{}}
	actRepo := &stubActivityRepo{}
	targetRepo := &stubTargetRepo{target: otherTarget}
	audit := &stubAuditRepo{}
	svc := newActivitySvcWithTargets(actRepo, targetRepo, audit)

	activity := &domain.Activity{
		ActivityType: "visit",
		DueDate:      time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC),
		Duration:     "full_day",
		Fields:       map[string]any{},
		TargetID:     "target-2",
	}
	_, err := svc.Create(context.Background(), adminUser(), activity)
	if err != nil {
		t.Fatalf("expected admin to create activity with any target, got: %v", err)
	}
}

func TestActivityCreate_ManagerTeamTargetSucceeds(t *testing.T) {
	t.Parallel()
	teamTarget := &domain.Target{ID: "target-1", TargetType: "doctor", Name: "Dr. Smith", AssigneeID: "rep-1", TeamID: "team-1", Fields: map[string]any{}}
	actRepo := &stubActivityRepo{}
	targetRepo := &stubTargetRepo{target: teamTarget}
	audit := &stubAuditRepo{}
	svc := newActivitySvcWithTargets(actRepo, targetRepo, audit)

	activity := &domain.Activity{
		ActivityType: "visit",
		DueDate:      time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC),
		Duration:     "full_day",
		Fields:       map[string]any{},
		TargetID:     "target-1",
	}
	_, err := svc.Create(context.Background(), managerUser(), activity)
	if err != nil {
		t.Fatalf("expected manager to create activity with team target, got: %v", err)
	}
}

func TestActivityCreate_ManagerOtherTeamTargetFails(t *testing.T) {
	t.Parallel()
	otherTeamTarget := &domain.Target{ID: "target-3", TargetType: "doctor", Name: "Dr. Far", AssigneeID: "rep-9", TeamID: "team-9", Fields: map[string]any{}}
	actRepo := &stubActivityRepo{}
	targetRepo := &stubTargetRepo{target: otherTeamTarget}
	audit := &stubAuditRepo{}
	svc := newActivitySvcWithTargets(actRepo, targetRepo, audit)

	activity := &domain.Activity{
		ActivityType: "visit",
		DueDate:      time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC),
		Duration:     "full_day",
		Fields:       map[string]any{},
		TargetID:     "target-3",
	}
	_, err := svc.Create(context.Background(), managerUser(), activity)
	if !errors.Is(err, service.ErrTargetNotAccessible) {
		t.Errorf("expected ErrTargetNotAccessible, got %v", err)
	}
}

func TestActivityUpdate_TargetChangedToInaccessibleFails(t *testing.T) {
	t.Parallel()
	existing := sampleActivity()
	existing.TargetID = "target-1"
	otherTarget := &domain.Target{ID: "target-2", TargetType: "doctor", Name: "Other Dr", AssigneeID: "rep-2", TeamID: "team-1", Fields: map[string]any{}}
	actRepo := &stubActivityRepo{activity: existing}
	targetRepo := &stubTargetRepo{target: otherTarget}
	audit := &stubAuditRepo{}
	svc := newActivitySvcWithTargets(actRepo, targetRepo, audit)

	updated := *existing
	updated.ActivityType = "visit"
	updated.DueDate = existing.DueDate
	updated.Duration = "full_day"
	updated.Fields = map[string]any{}
	updated.TargetID = "target-2"

	_, err := svc.Update(context.Background(), repUser(), "activity-1", &updated)
	if !errors.Is(err, service.ErrTargetNotAccessible) {
		t.Errorf("expected ErrTargetNotAccessible, got %v", err)
	}
}
