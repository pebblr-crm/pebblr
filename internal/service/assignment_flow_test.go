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

const (
	testFlowTeamAlpha  = "team-alpha"
	testFlowTeamBeta   = "team-beta"
	testFlowTarget1    = "t-flow-1"
	testFlowReassign   = "t-reassign"
	testFlowCross      = "t-cross"
	testFlowActUpdate  = "act-update"
	testFlowActPatch   = "act-patch"
	testFlowAdmin      = "t-admin"
)

// ── Multi-user target registry ──────────────────────────────────────────────
// These stubs maintain a map of targets so we can simulate multiple users
// interacting with the same target pool.

type targetRegistry struct {
	targets map[string]*domain.Target
	saveErr error
}

func newTargetRegistry(targets ...*domain.Target) *targetRegistry {
	r := &targetRegistry{targets: make(map[string]*domain.Target)}
	for _, t := range targets {
		r.targets[t.ID] = t
	}
	return r
}

func (r *targetRegistry) Get(_ context.Context, id string) (*domain.Target, error) {
	t, ok := r.targets[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	cpy := *t
	return &cpy, nil
}

func (r *targetRegistry) List(_ context.Context, _ rbac.TargetScope, _ store.TargetFilter, _, _ int) (*store.TargetPage, error) {
	return &store.TargetPage{}, nil
}

func (r *targetRegistry) Create(_ context.Context, t *domain.Target) (*domain.Target, error) {
	r.targets[t.ID] = t
	return t, nil
}

func (r *targetRegistry) Update(_ context.Context, t *domain.Target) (*domain.Target, error) {
	if r.saveErr != nil {
		return nil, r.saveErr
	}
	r.targets[t.ID] = t
	return t, nil
}

func (r *targetRegistry) Upsert(_ context.Context, targets []*domain.Target) (*store.ImportResult, error) {
	return &store.ImportResult{}, nil
}

func (r *targetRegistry) VisitStatus(_ context.Context, _ rbac.TargetScope, _ []string) ([]store.TargetVisitStatus, error) {
	return nil, nil
}

func (r *targetRegistry) FrequencyStatus(_ context.Context, _ rbac.TargetScope, _ []string, _, _ time.Time) ([]store.TargetFrequencyStatus, error) {
	return nil, nil
}

// ── Users ───────────────────────────────────────────────────────────────────

func repA() *domain.User {
	return &domain.User{ID: "rep-a", Name: "Alice (Rep)", Role: domain.RoleRep, TeamIDs: []string{testFlowTeamAlpha}}
}

func repB() *domain.User {
	return &domain.User{ID: "rep-b", Name: "Bob (Rep)", Role: domain.RoleRep, TeamIDs: []string{testFlowTeamAlpha}}
}

func repC() *domain.User {
	return &domain.User{ID: "rep-c", Name: "Charlie (Rep)", Role: domain.RoleRep, TeamIDs: []string{testFlowTeamBeta}}
}

func mgrAlpha() *domain.User {
	return &domain.User{ID: "mgr-alpha", Name: "Manager Alpha", Role: domain.RoleManager, TeamIDs: []string{testFlowTeamAlpha}}
}

func mgrBeta() *domain.User {
	return &domain.User{ID: "mgr-beta", Name: "Manager Beta", Role: domain.RoleManager, TeamIDs: []string{testFlowTeamBeta}}
}

func admin() *domain.User {
	return &domain.User{ID: "admin-x", Name: "Admin", Role: domain.RoleAdmin, TeamIDs: []string{testFlowTeamAlpha, testFlowTeamBeta}}
}

func flowUserRepo() *stubUserRepo {
	return &stubUserRepo{users: map[string]*domain.User{
		"rep-a":     repA(),
		"rep-b":     repB(),
		"rep-c":     repC(),
		"mgr-alpha": mgrAlpha(),
		"mgr-beta":  mgrBeta(),
		"admin-x":   admin(),
	}}
}

func flowConfig() *config.TenantConfig {
	return activityTestConfig()
}

// ── Flow 1: Full dual-visit lifecycle ───────────────────────────────────────
// Admin assigns target → Rep A creates visit with joint visitor Rep B →
// Rep A can view activity → Rep B can view activity → Rep B cannot see target.

func TestFlow_DualVisitLifecycle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	enforcer := rbac.NewEnforcer()
	userRepo := flowUserRepo()
	auditRepo := &stubAuditRepo{}

	// Target initially assigned to Rep A on team-alpha.
	target := &domain.Target{
		ID: testFlowTarget1, TargetType: "doctor", Name: "Dr. Lifecycle",
		AssigneeID: "rep-a", TeamID: testFlowTeamAlpha,
		Fields: map[string]any{"city": "Bucharest"},
	}
	targetReg := newTargetRegistry(target)
	actRepo := &stubActivityRepo{}

	targetSvc := service.NewTargetService(targetReg, enforcer, flowConfig(),
		service.WithUsers(userRepo), service.WithAudit(auditRepo))
	actSvc := service.NewActivityService(actRepo, targetReg, userRepo, auditRepo,
		enforcer, flowConfig())

	// ── Step 1: Admin assigns target to Rep A (already assigned, but verify the endpoint works).
	assigned, err := targetSvc.Assign(ctx, admin(), testFlowTarget1, "rep-a", testFlowTeamAlpha)
	if err != nil {
		t.Fatalf("step 1: admin assign failed: %v", err)
	}
	if assigned.AssigneeID != "rep-a" {
		t.Fatalf("step 1: expected assignee rep-a, got %s", assigned.AssigneeID)
	}

	// ── Step 2: Rep A creates a visit with joint visitor Rep B.
	activity := &domain.Activity{
		ActivityType:  "visit",
		DueDate:       time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		Duration:      "full_day",
		Fields:        map[string]any{},
		TargetID:      testFlowTarget1,
		JointVisitUID: "rep-b",
	}
	created, err := actSvc.Create(ctx, repA(), activity)
	if err != nil {
		t.Fatalf("step 2: rep A create visit failed: %v", err)
	}
	if created.CreatorID != "rep-a" {
		t.Errorf("step 2: expected creator rep-a, got %s", created.CreatorID)
	}
	if created.JointVisitUID != "rep-b" {
		t.Errorf("step 2: expected joint visitor rep-b, got %s", created.JointVisitUID)
	}

	// ── Step 3: Rep A can view the activity (they created it).
	// Simulate by checking RBAC directly.
	if !enforcer.CanViewActivity(repA(), created) {
		t.Error("step 3: rep A should be able to view their own activity")
	}

	// ── Step 4: Rep B can view the activity (they are the joint visitor).
	if !enforcer.CanViewActivity(repB(), created) {
		t.Error("step 4: rep B should be able to view activity as joint visitor")
	}

	// ── Step 5: Rep B CANNOT see the target in the target list (not assigned to them).
	if enforcer.CanViewTarget(repB(), target) {
		t.Error("step 5: rep B should NOT be able to view the target (not their assignee)")
	}

	// ── Step 6: Rep B can see the target name through the activity response.
	if created.TargetID != testFlowTarget1 {
		t.Error("step 6: activity should carry the target ID for response enrichment")
	}
}

// ── Flow 2: Rep blocked from creating activity on another rep's target ──────

func TestFlow_RepBlockedOnOtherRepsTarget(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	enforcer := rbac.NewEnforcer()
	userRepo := flowUserRepo()
	auditRepo := &stubAuditRepo{}

	// Target assigned to Rep A.
	target := &domain.Target{
		ID: "t-flow-2", TargetType: "doctor", Name: "Dr. OffLimits",
		AssigneeID: "rep-a", TeamID: testFlowTeamAlpha,
		Fields: map[string]any{},
	}
	targetReg := newTargetRegistry(target)
	actRepo := &stubActivityRepo{}

	actSvc := service.NewActivityService(actRepo, targetReg, userRepo, auditRepo,
		enforcer, flowConfig())

	// Rep B tries to create a visit on Rep A's target → should fail.
	activity := &domain.Activity{
		ActivityType: "visit",
		DueDate:      time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		Duration:     "full_day",
		Fields:       map[string]any{},
		TargetID:     "t-flow-2",
	}
	_, err := actSvc.Create(ctx, repB(), activity)
	if !errors.Is(err, service.ErrTargetNotAccessible) {
		t.Errorf("expected ErrTargetNotAccessible, got %v", err)
	}

	// Rep C (different team) also blocked.
	_, err = actSvc.Create(ctx, repC(), activity)
	if !errors.Is(err, service.ErrTargetNotAccessible) {
		t.Errorf("expected ErrTargetNotAccessible for cross-team rep, got %v", err)
	}
}

// ── Flow 3: Manager can create on team target but not cross-team ────────────

func TestFlow_ManagerTeamBoundary(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	enforcer := rbac.NewEnforcer()
	userRepo := flowUserRepo()
	auditRepo := &stubAuditRepo{}

	alphaTarget := &domain.Target{
		ID: "t-alpha", TargetType: "doctor", Name: "Dr. Alpha",
		AssigneeID: "rep-a", TeamID: testFlowTeamAlpha,
		Fields: map[string]any{},
	}
	betaTarget := &domain.Target{
		ID: "t-beta", TargetType: "doctor", Name: "Dr. Beta",
		AssigneeID: "rep-c", TeamID: testFlowTeamBeta,
		Fields: map[string]any{},
	}
	targetReg := newTargetRegistry(alphaTarget, betaTarget)
	actRepo := &stubActivityRepo{}

	actSvc := service.NewActivityService(actRepo, targetReg, userRepo, auditRepo,
		enforcer, flowConfig())

	// Manager Alpha can create on their team's target.
	activity := &domain.Activity{
		ActivityType: "visit",
		DueDate:      time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		Duration:     "full_day",
		Fields:       map[string]any{},
		TargetID:     "t-alpha",
	}
	_, err := actSvc.Create(ctx, mgrAlpha(), activity)
	if err != nil {
		t.Fatalf("manager alpha should create on team target, got: %v", err)
	}

	// Manager Alpha cannot create on team-beta's target.
	crossTeamActivity := &domain.Activity{
		ActivityType: "visit",
		DueDate:      time.Date(2026, 4, 2, 0, 0, 0, 0, time.UTC),
		Duration:     "full_day",
		Fields:       map[string]any{},
		TargetID:     "t-beta",
	}
	_, err = actSvc.Create(ctx, mgrAlpha(), crossTeamActivity)
	if !errors.Is(err, service.ErrTargetNotAccessible) {
		t.Errorf("manager alpha should be blocked on cross-team target, got %v", err)
	}

	// Manager Beta CAN create on team-beta's target.
	_, err = actSvc.Create(ctx, mgrBeta(), crossTeamActivity)
	if err != nil {
		t.Fatalf("manager beta should create on own team target, got: %v", err)
	}
}

// ── Flow 4: Admin reassigns target, old rep loses access ───────────────────

func TestFlow_ReassignmentRevokesOldRepAccess(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	enforcer := rbac.NewEnforcer()
	userRepo := flowUserRepo()
	auditRepo := &stubAuditRepo{}

	// Target starts assigned to Rep A.
	target := &domain.Target{
		ID: testFlowReassign, TargetType: "doctor", Name: "Dr. Reassign",
		AssigneeID: "rep-a", TeamID: testFlowTeamAlpha,
		Fields: map[string]any{},
	}
	targetReg := newTargetRegistry(target)
	actRepo := &stubActivityRepo{}

	targetSvc := service.NewTargetService(targetReg, enforcer, flowConfig(),
		service.WithUsers(userRepo), service.WithAudit(auditRepo))
	actSvc := service.NewActivityService(actRepo, targetReg, userRepo, auditRepo,
		enforcer, flowConfig())

	// ── Step 1: Rep A can create a visit on their target.
	activity := &domain.Activity{
		ActivityType: "visit",
		DueDate:      time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		Duration:     "full_day",
		Fields:       map[string]any{},
		TargetID:     testFlowReassign,
	}
	_, err := actSvc.Create(ctx, repA(), activity)
	if err != nil {
		t.Fatalf("step 1: rep A should create visit on own target, got: %v", err)
	}

	// ── Step 2: Admin reassigns the target to Rep B.
	updated, err := targetSvc.Assign(ctx, admin(), testFlowReassign, "rep-b", testFlowTeamAlpha)
	if err != nil {
		t.Fatalf("step 2: admin reassign failed: %v", err)
	}
	if updated.AssigneeID != "rep-b" {
		t.Fatalf("step 2: expected assignee rep-b, got %s", updated.AssigneeID)
	}

	// Verify audit was recorded for the reassignment.
	foundAssign := false
	for _, e := range auditRepo.entries {
		if e.EventType == "assigned" && e.EntityID == testFlowReassign {
			oldVal, _ := e.OldValue["assigneeId"].(string)
			newVal, _ := e.NewValue["assigneeId"].(string)
			if oldVal == "rep-a" && newVal == "rep-b" {
				foundAssign = true
			}
		}
	}
	if !foundAssign {
		t.Error("step 2: expected audit entry for reassignment from rep-a to rep-b")
	}

	// ── Step 3: Rep A can NO LONGER create a visit on the reassigned target.
	activity2 := &domain.Activity{
		ActivityType: "visit",
		DueDate:      time.Date(2026, 4, 2, 0, 0, 0, 0, time.UTC),
		Duration:     "full_day",
		Fields:       map[string]any{},
		TargetID:     testFlowReassign,
	}
	_, err = actSvc.Create(ctx, repA(), activity2)
	if !errors.Is(err, service.ErrTargetNotAccessible) {
		t.Errorf("step 3: rep A should lose access after reassignment, got %v", err)
	}

	// ── Step 4: Rep B CAN now create a visit on the target.
	_, err = actSvc.Create(ctx, repB(), activity2)
	if err != nil {
		t.Fatalf("step 4: rep B should create visit on newly assigned target, got: %v", err)
	}
}

// ── Flow 5: Cross-team reassignment ────────────────────────────────────────
// Admin reassigns target from team-alpha to team-beta. Manager Alpha loses
// access, Manager Beta gains it.

func TestFlow_CrossTeamReassignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	enforcer := rbac.NewEnforcer()
	userRepo := flowUserRepo()
	auditRepo := &stubAuditRepo{}

	target := &domain.Target{
		ID: testFlowCross, TargetType: "pharmacy", Name: "Central Pharmacy",
		AssigneeID: "rep-a", TeamID: testFlowTeamAlpha,
		Fields: map[string]any{},
	}
	targetReg := newTargetRegistry(target)
	actRepo := &stubActivityRepo{}

	cfg := flowConfig()
	cfg.Accounts = config.AccountsConfig{
		Types: []config.AccountTypeConfig{
			{Key: "doctor", Label: "Doctor"},
			{Key: "pharmacy", Label: "Pharmacy"},
		},
	}

	targetSvc := service.NewTargetService(targetReg, enforcer, cfg,
		service.WithUsers(userRepo), service.WithAudit(auditRepo))
	actSvc := service.NewActivityService(actRepo, targetReg, userRepo, auditRepo,
		enforcer, cfg)

	// ── Before reassignment: Manager Alpha can create, Manager Beta cannot.
	activity := &domain.Activity{
		ActivityType: "visit",
		DueDate:      time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		Duration:     "full_day",
		Fields:       map[string]any{},
		TargetID:     testFlowCross,
	}
	_, err := actSvc.Create(ctx, mgrAlpha(), activity)
	if err != nil {
		t.Fatalf("pre-reassign: manager alpha should create, got: %v", err)
	}

	activity2 := &domain.Activity{
		ActivityType: "visit",
		DueDate:      time.Date(2026, 4, 2, 0, 0, 0, 0, time.UTC),
		Duration:     "full_day",
		Fields:       map[string]any{},
		TargetID:     testFlowCross,
	}
	_, err = actSvc.Create(ctx, mgrBeta(), activity2)
	if !errors.Is(err, service.ErrTargetNotAccessible) {
		t.Errorf("pre-reassign: manager beta should be blocked, got %v", err)
	}

	// ── Admin reassigns to Rep C on team-beta.
	_, err = targetSvc.Assign(ctx, admin(), testFlowCross, "rep-c", testFlowTeamBeta)
	if err != nil {
		t.Fatalf("reassign failed: %v", err)
	}

	// ── After reassignment: Manager Alpha blocked, Manager Beta can create.
	activity3 := &domain.Activity{
		ActivityType: "visit",
		DueDate:      time.Date(2026, 4, 3, 0, 0, 0, 0, time.UTC),
		Duration:     "full_day",
		Fields:       map[string]any{},
		TargetID:     testFlowCross,
	}
	_, err = actSvc.Create(ctx, mgrAlpha(), activity3)
	if !errors.Is(err, service.ErrTargetNotAccessible) {
		t.Errorf("post-reassign: manager alpha should lose access, got %v", err)
	}

	_, err = actSvc.Create(ctx, mgrBeta(), activity3)
	if err != nil {
		t.Fatalf("post-reassign: manager beta should gain access, got: %v", err)
	}

	// Rep C (new assignee) can create too.
	activity4 := &domain.Activity{
		ActivityType: "visit",
		DueDate:      time.Date(2026, 4, 4, 0, 0, 0, 0, time.UTC),
		Duration:     "full_day",
		Fields:       map[string]any{},
		TargetID:     testFlowCross,
	}
	_, err = actSvc.Create(ctx, repC(), activity4)
	if err != nil {
		t.Fatalf("post-reassign: rep C should create on own target, got: %v", err)
	}
}

// ── Flow 6: Manager assigns within team, rep cannot assign ─────────────────

func TestFlow_ManagerAssignWithinTeam_RepCannotAssign(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	enforcer := rbac.NewEnforcer()
	userRepo := flowUserRepo()
	auditRepo := &stubAuditRepo{}

	target := &domain.Target{
		ID: "t-mgr", TargetType: "doctor", Name: "Dr. Team",
		AssigneeID: "rep-a", TeamID: testFlowTeamAlpha,
		Fields: map[string]any{},
	}
	targetReg := newTargetRegistry(target)

	targetSvc := service.NewTargetService(targetReg, enforcer, flowConfig(),
		service.WithUsers(userRepo), service.WithAudit(auditRepo))

	// Manager Alpha can reassign within their team.
	updated, err := targetSvc.Assign(ctx, mgrAlpha(), "t-mgr", "rep-b", "")
	if err != nil {
		t.Fatalf("manager assign failed: %v", err)
	}
	if updated.AssigneeID != "rep-b" {
		t.Errorf("expected assignee rep-b, got %s", updated.AssigneeID)
	}

	// Rep A cannot assign (even their own target back to themselves).
	_, err = targetSvc.Assign(ctx, repA(), "t-mgr", "rep-a", "")
	if !errors.Is(err, service.ErrForbidden) {
		t.Errorf("rep should not be able to assign, got %v", err)
	}

	// Manager Beta cannot assign on team-alpha's target.
	_, err = targetSvc.Assign(ctx, mgrBeta(), "t-mgr", "rep-c", "")
	if !errors.Is(err, service.ErrForbidden) {
		t.Errorf("cross-team manager should be blocked, got %v", err)
	}
}

// ── Flow 7: Activity update with target change ─────────────────────────────
// Rep A creates activity on own target, then tries to update targetId to
// another rep's target — should be blocked.

func TestFlow_ActivityUpdateTargetChange(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	enforcer := rbac.NewEnforcer()
	userRepo := flowUserRepo()
	auditRepo := &stubAuditRepo{}

	ownTarget := &domain.Target{
		ID: "t-own", TargetType: "doctor", Name: "Dr. Own",
		AssigneeID: "rep-a", TeamID: testFlowTeamAlpha,
		Fields: map[string]any{},
	}
	otherTarget := &domain.Target{
		ID: "t-other", TargetType: "doctor", Name: "Dr. Other",
		AssigneeID: "rep-b", TeamID: testFlowTeamAlpha,
		Fields: map[string]any{},
	}
	targetReg := newTargetRegistry(ownTarget, otherTarget)

	existing := &domain.Activity{
		ID: testFlowActUpdate, ActivityType: "visit", Status: "planificat",
		DueDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		Duration: "full_day", Fields: map[string]any{},
		TargetID: "t-own", CreatorID: "rep-a", TeamID: testFlowTeamAlpha,
	}
	actRepo := &stubActivityRepo{activity: existing}

	actSvc := service.NewActivityService(actRepo, targetReg, userRepo, auditRepo,
		enforcer, flowConfig())

	// Rep A tries to change targetId to t-other (Rep B's target) → blocked.
	updated := *existing
	updated.TargetID = "t-other"
	_, err := actSvc.Update(ctx, repA(), testFlowActUpdate, &updated)
	if !errors.Is(err, service.ErrTargetNotAccessible) {
		t.Errorf("expected ErrTargetNotAccessible on target change, got %v", err)
	}

	// Rep A can keep the same target → should succeed.
	same := *existing
	same.TargetID = "t-own"
	_, err = actSvc.Update(ctx, repA(), testFlowActUpdate, &same)
	if err != nil {
		t.Fatalf("keeping same target should succeed, got: %v", err)
	}
}

// ── Flow 8: Partial update (PATCH) target change ───────────────────────────

func TestFlow_PartialUpdateTargetChange(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	enforcer := rbac.NewEnforcer()
	userRepo := flowUserRepo()
	auditRepo := &stubAuditRepo{}

	ownTarget := &domain.Target{
		ID: "t-patch-own", TargetType: "doctor", Name: "Dr. Own",
		AssigneeID: "rep-a", TeamID: testFlowTeamAlpha,
		Fields: map[string]any{},
	}
	otherTarget := &domain.Target{
		ID: "t-patch-other", TargetType: "doctor", Name: "Dr. Other",
		AssigneeID: "rep-b", TeamID: testFlowTeamAlpha,
		Fields: map[string]any{},
	}
	targetReg := newTargetRegistry(ownTarget, otherTarget)

	existing := &domain.Activity{
		ID: testFlowActPatch, ActivityType: "visit", Status: "planificat",
		DueDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		Duration: "full_day", Fields: map[string]any{},
		TargetID: "t-patch-own", CreatorID: "rep-a", TeamID: testFlowTeamAlpha,
	}
	actRepo := &stubActivityRepo{activity: existing}

	actSvc := service.NewActivityService(actRepo, targetReg, userRepo, auditRepo,
		enforcer, flowConfig())

	// PATCH targetId to another rep's target → blocked.
	otherID := "t-patch-other"
	patch := &domain.ActivityPatch{TargetID: &otherID}
	_, err := actSvc.PartialUpdate(ctx, repA(), testFlowActPatch, patch)
	if !errors.Is(err, service.ErrTargetNotAccessible) {
		t.Errorf("expected ErrTargetNotAccessible on patch target change, got %v", err)
	}

	// PATCH without targetId → should succeed (no target change).
	statusPatch := "planificat"
	noTargetPatch := &domain.ActivityPatch{Status: &statusPatch}
	_, err = actSvc.PartialUpdate(ctx, repA(), testFlowActPatch, noTargetPatch)
	if err != nil {
		t.Fatalf("patch without target change should succeed, got: %v", err)
	}
}

// ── Flow 9: Admin bypasses all target access checks ────────────────────────

func TestFlow_AdminBypassesTargetAccess(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	enforcer := rbac.NewEnforcer()
	userRepo := flowUserRepo()
	auditRepo := &stubAuditRepo{}

	// Target on team-beta, assigned to Rep C — admin should still be able to
	// create activities and assign regardless.
	target := &domain.Target{
		ID: testFlowAdmin, TargetType: "doctor", Name: "Dr. Admin",
		AssigneeID: "rep-c", TeamID: testFlowTeamBeta,
		Fields: map[string]any{},
	}
	targetReg := newTargetRegistry(target)
	actRepo := &stubActivityRepo{}

	targetSvc := service.NewTargetService(targetReg, enforcer, flowConfig(),
		service.WithUsers(userRepo), service.WithAudit(auditRepo))
	actSvc := service.NewActivityService(actRepo, targetReg, userRepo, auditRepo,
		enforcer, flowConfig())

	// Admin creates activity on any target.
	activity := &domain.Activity{
		ActivityType: "visit",
		DueDate:      time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		Duration:     "full_day",
		Fields:       map[string]any{},
		TargetID:     testFlowAdmin,
	}
	_, err := actSvc.Create(ctx, admin(), activity)
	if err != nil {
		t.Fatalf("admin should create on any target, got: %v", err)
	}

	// Admin reassigns target from team-beta/rep-c to team-alpha/rep-a.
	updated, err := targetSvc.Assign(ctx, admin(), testFlowAdmin, "rep-a", testFlowTeamAlpha)
	if err != nil {
		t.Fatalf("admin should assign any target, got: %v", err)
	}
	if updated.AssigneeID != "rep-a" {
		t.Errorf("expected rep-a, got %s", updated.AssigneeID)
	}
	if updated.TeamID != testFlowTeamAlpha {
		t.Errorf("expected team-alpha, got %s", updated.TeamID)
	}

	// Admin can still create after reassignment.
	activity2 := &domain.Activity{
		ActivityType: "visit",
		DueDate:      time.Date(2026, 4, 2, 0, 0, 0, 0, time.UTC),
		Duration:     "full_day",
		Fields:       map[string]any{},
		TargetID:     testFlowAdmin,
	}
	_, err = actSvc.Create(ctx, admin(), activity2)
	if err != nil {
		t.Fatalf("admin should still create after reassignment, got: %v", err)
	}
}

// ── Flow 10: Non-field activity skips target access check ──────────────────

func TestFlow_NonFieldActivitySkipsTargetCheck(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	enforcer := rbac.NewEnforcer()
	userRepo := flowUserRepo()
	auditRepo := &stubAuditRepo{}

	// No targets at all — vacation shouldn't need one.
	targetReg := newTargetRegistry()
	actRepo := &stubActivityRepo{}

	actSvc := service.NewActivityService(actRepo, targetReg, userRepo, auditRepo,
		enforcer, flowConfig())

	// Rep A creates a vacation (non-field) with no target → should work.
	vacation := &domain.Activity{
		ActivityType: "vacation",
		DueDate:      time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		Duration:     "full_day",
		Fields:       map[string]any{},
		TargetID:     "", // no target
	}
	_, err := actSvc.Create(ctx, repA(), vacation)
	if err != nil {
		t.Fatalf("non-field activity without target should succeed, got: %v", err)
	}
}

// ── Flow 11: Non-existent target ID rejected ───────────────────────────────

func TestFlow_NonExistentTargetRejected(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	enforcer := rbac.NewEnforcer()
	userRepo := flowUserRepo()
	auditRepo := &stubAuditRepo{}

	// Empty registry — target doesn't exist.
	targetReg := newTargetRegistry()
	actRepo := &stubActivityRepo{}

	actSvc := service.NewActivityService(actRepo, targetReg, userRepo, auditRepo,
		enforcer, flowConfig())

	// Rep A tries to create visit on non-existent target.
	activity := &domain.Activity{
		ActivityType: "visit",
		DueDate:      time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		Duration:     "full_day",
		Fields:       map[string]any{},
		TargetID:     "t-does-not-exist",
	}
	_, err := actSvc.Create(ctx, repA(), activity)
	if err == nil {
		t.Fatal("expected error for non-existent target, got nil")
	}
	// Should bubble up as a wrapped ErrNotFound from the target lookup.
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound for non-existent target, got %v", err)
	}
}
