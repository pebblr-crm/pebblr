package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/pebblr/pebblr/internal/config"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/service"
	"github.com/pebblr/pebblr/internal/store"
)

// --- stub dashboard repo ---

type stubDashboardRepo struct {
	activityStats *store.ActivityStats
	coverageStats *store.CoverageStats
	frequencyRows []store.FrequencyRow
}

func (r *stubDashboardRepo) ActivityStats(_ context.Context, _ rbac.ActivityScope, _ store.DashboardFilter) (*store.ActivityStats, error) {
	return r.activityStats, nil
}

func (r *stubDashboardRepo) CoverageStats(_ context.Context, _ rbac.ActivityScope, _ rbac.TargetScope, _ store.DashboardFilter) (*store.CoverageStats, error) {
	return r.coverageStats, nil
}

func (r *stubDashboardRepo) FrequencyStats(_ context.Context, _ rbac.ActivityScope, _ rbac.TargetScope, _ store.DashboardFilter) ([]store.FrequencyRow, error) {
	return r.frequencyRows, nil
}

func dashboardConfig() *config.TenantConfig {
	return &config.TenantConfig{
		Activities: config.ActivitiesConfig{
			Types: []config.ActivityTypeConfig{
				{Key: "visit", Label: "Visit", Category: "field"},
				{Key: "administrative", Label: "Administrative", Category: "non_field"},
				{Key: "vacation", Label: "Vacation", Category: "non_field"},
			},
		},
		Rules: config.RulesConfig{
			Frequency: map[string]int{"a": 4, "b": 2, "c": 1},
		},
	}
}

func newDashboardSvc(repo *stubDashboardRepo) *service.DashboardService {
	return service.NewDashboardService(repo, rbac.NewEnforcer(), dashboardConfig())
}

func marchFilter() store.DashboardFilter {
	return store.DashboardFilter{
		DateFrom: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		DateTo:   time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC),
	}
}

// --- ActivityStats tests ---

func TestDashboard_ActivityStats_GroupsByCategory(t *testing.T) {
	t.Parallel()
	repo := &stubDashboardRepo{
		activityStats: &store.ActivityStats{
			ByStatus:   map[string]int{"planned": 5, "completed": 3},
			ByCategory: map[string]int{"visit": 6, "administrative": 2},
			Total:      8,
		},
	}
	svc := newDashboardSvc(repo)

	resp, err := svc.ActivityStats(context.Background(), adminUser(), marchFilter())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Total != 8 {
		t.Errorf("total = %d, want 8", resp.Total)
	}
	if resp.ByStatus["planned"] != 5 {
		t.Errorf("byStatus[planned] = %d, want 5", resp.ByStatus["planned"])
	}
	if resp.ByCategory["field"] != 6 {
		t.Errorf("byCategory[field] = %d, want 6", resp.ByCategory["field"])
	}
	if resp.ByCategory["non_field"] != 2 {
		t.Errorf("byCategory[non_field] = %d, want 2", resp.ByCategory["non_field"])
	}
}

func TestDashboard_ActivityStats_RepScopedByRBAC(t *testing.T) {
	t.Parallel()
	repo := &stubDashboardRepo{
		activityStats: &store.ActivityStats{
			ByStatus:   map[string]int{"planned": 2},
			ByCategory: map[string]int{"visit": 2},
			Total:      2,
		},
	}
	svc := newDashboardSvc(repo)

	resp, err := svc.ActivityStats(context.Background(), repUser(), marchFilter())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 2 {
		t.Errorf("rep should get scoped results, got total = %d", resp.Total)
	}
}

// --- Coverage tests ---

func TestDashboard_Coverage_CalculatesPercentage(t *testing.T) {
	t.Parallel()
	repo := &stubDashboardRepo{
		coverageStats: &store.CoverageStats{
			TotalTargets:   20,
			VisitedTargets: 15,
		},
	}
	svc := newDashboardSvc(repo)

	resp, err := svc.Coverage(context.Background(), adminUser(), marchFilter())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.TotalTargets != 20 {
		t.Errorf("totalTargets = %d, want 20", resp.TotalTargets)
	}
	if resp.VisitedTargets != 15 {
		t.Errorf("visitedTargets = %d, want 15", resp.VisitedTargets)
	}
	if resp.Percentage != 75 {
		t.Errorf("percentage = %f, want 75", resp.Percentage)
	}
}

func TestDashboard_Coverage_ZeroTargets(t *testing.T) {
	t.Parallel()
	repo := &stubDashboardRepo{
		coverageStats: &store.CoverageStats{
			TotalTargets:   0,
			VisitedTargets: 0,
		},
	}
	svc := newDashboardSvc(repo)

	resp, err := svc.Coverage(context.Background(), adminUser(), marchFilter())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Percentage != 0 {
		t.Errorf("percentage = %f, want 0", resp.Percentage)
	}
}

// --- Frequency tests ---

func TestDashboard_Frequency_WithConfigRules(t *testing.T) {
	t.Parallel()
	repo := &stubDashboardRepo{
		frequencyRows: []store.FrequencyRow{
			{Classification: "a", TargetCount: 10, TotalVisits: 40},
			{Classification: "b", TargetCount: 5, TotalVisits: 5},
			{Classification: "c", TargetCount: 8, TotalVisits: 8},
		},
	}
	svc := newDashboardSvc(repo)

	resp, err := svc.Frequency(context.Background(), adminUser(), marchFilter())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Items) != 3 {
		t.Fatalf("items = %d, want 3", len(resp.Items))
	}

	// Class A: 10 targets * 4 required * 1 month = 40 expected, 40 actual = 100%
	a := resp.Items[0]
	if a.Classification != "a" {
		t.Errorf("item 0 classification = %s, want a", a.Classification)
	}
	if a.Required != 4 {
		t.Errorf("item 0 required = %d, want 4", a.Required)
	}
	if a.Compliance != 100 {
		t.Errorf("item 0 compliance = %f, want 100", a.Compliance)
	}

	// Class B: 5 targets * 2 required * 1 month = 10 expected, 5 actual = 50%
	b := resp.Items[1]
	if b.Compliance != 50 {
		t.Errorf("item 1 compliance = %f, want 50", b.Compliance)
	}

	// Class C: 8 targets * 1 required * 1 month = 8 expected, 8 actual = 100%
	c := resp.Items[2]
	if c.Compliance != 100 {
		t.Errorf("item 2 compliance = %f, want 100", c.Compliance)
	}
}

func TestDashboard_Frequency_MultiMonth(t *testing.T) {
	t.Parallel()
	repo := &stubDashboardRepo{
		frequencyRows: []store.FrequencyRow{
			{Classification: "a", TargetCount: 10, TotalVisits: 40},
		},
	}
	svc := newDashboardSvc(repo)

	// Q1 2026: Jan-Mar = 3 months
	filter := store.DashboardFilter{
		DateFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		DateTo:   time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC),
	}

	resp, err := svc.Frequency(context.Background(), adminUser(), filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 10 targets * 4 required * 3 months = 120 expected, 40 actual ≈ 33.33%
	a := resp.Items[0]
	expected := float64(40) / float64(120) * 100
	if a.Compliance < expected-0.1 || a.Compliance > expected+0.1 {
		t.Errorf("compliance = %f, want ~%f", a.Compliance, expected)
	}
}

func TestDashboard_Frequency_NoFrequencyRule(t *testing.T) {
	t.Parallel()
	repo := &stubDashboardRepo{
		frequencyRows: []store.FrequencyRow{
			{Classification: "x", TargetCount: 5, TotalVisits: 3},
		},
	}
	svc := newDashboardSvc(repo)

	resp, err := svc.Frequency(context.Background(), adminUser(), marchFilter())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No frequency rule for "x" means required=0, compliance=0
	x := resp.Items[0]
	if x.Required != 0 {
		t.Errorf("required = %d, want 0", x.Required)
	}
	if x.Compliance != 0 {
		t.Errorf("compliance = %f, want 0", x.Compliance)
	}
}
