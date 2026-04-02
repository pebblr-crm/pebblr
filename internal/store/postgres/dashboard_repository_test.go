package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/store"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func newDashboardRepo(pool pgxmock.PgxPoolIface) *dashboardRepository {
	return &dashboardRepository{pool: pool}
}

func defaultFilter() store.DashboardFilter {
	return store.DashboardFilter{
		DateFrom: testTime(),
		DateTo:   testTime().Add(30 * 24 * time.Hour),
	}
}

// --- ActivityStats tests ---

func TestActivityStats_AllActivities(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}
	filter := defaultFilter()

	// AllActivities + 2 date filters = 2 args
	mock.ExpectQuery("SELECT status, COUNT\\(\\*\\) FROM activities").
		WithArgs(filter.DateFrom, filter.DateTo).
		WillReturnRows(mock.NewRows([]string{"status", "count"}).
			AddRow("planned", 5).
			AddRow("completed", 3))

	mock.ExpectQuery("SELECT activity_type, COUNT\\(\\*\\) FROM activities").
		WithArgs(filter.DateFrom, filter.DateTo).
		WillReturnRows(mock.NewRows([]string{"activity_type", "count"}).
			AddRow("visit", 6).
			AddRow("admin", 2))

	stats, err := repo.ActivityStats(ctx, scope, filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Total != 8 {
		t.Errorf("expected total 8, got %d", stats.Total)
	}
	if stats.ByStatus["planned"] != 5 {
		t.Errorf("expected planned=5, got %d", stats.ByStatus["planned"])
	}
	if stats.ByCategory["visit"] != 6 {
		t.Errorf("expected visit=6, got %d", stats.ByCategory["visit"])
	}
}

func TestActivityStats_EmptyScope(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()

	// Non-all scope with no creator/team IDs returns empty stats immediately.
	scope := rbac.ActivityScope{}
	filter := defaultFilter()

	stats, err := repo.ActivityStats(ctx, scope, filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Total != 0 {
		t.Errorf("expected total 0, got %d", stats.Total)
	}
	if len(stats.ByStatus) != 0 {
		t.Errorf("expected empty ByStatus, got %v", stats.ByStatus)
	}
}

func TestActivityStats_ScopedByCreatorIDs(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{CreatorIDs: []string{"user-1"}}
	filter := defaultFilter()

	// 1 scope arg + 2 date args = 3
	mock.ExpectQuery("SELECT status, COUNT\\(\\*\\) FROM activities").
		WithArgs(anyArgs(3)...).
		WillReturnRows(mock.NewRows([]string{"status", "count"}).
			AddRow("planned", 2))

	mock.ExpectQuery("SELECT activity_type, COUNT\\(\\*\\) FROM activities").
		WithArgs(anyArgs(3)...).
		WillReturnRows(mock.NewRows([]string{"activity_type", "count"}).
			AddRow("visit", 2))

	stats, err := repo.ActivityStats(ctx, scope, filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Total != 2 {
		t.Errorf("expected total 2, got %d", stats.Total)
	}
}

func TestActivityStats_WithUserAndTeamFilter(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}
	filter := store.DashboardFilter{
		DateFrom: testTime(),
		DateTo:   testTime().Add(24 * time.Hour),
		UserID:   strPtr("user-1"),
		TeamID:   strPtr("team-1"),
	}

	// 2 date args + 1 user + 1 team = 4
	mock.ExpectQuery("SELECT status, COUNT\\(\\*\\) FROM activities").
		WithArgs(anyArgs(4)...).
		WillReturnRows(mock.NewRows([]string{"status", "count"}).
			AddRow("planned", 1))

	mock.ExpectQuery("SELECT activity_type, COUNT\\(\\*\\) FROM activities").
		WithArgs(anyArgs(4)...).
		WillReturnRows(mock.NewRows([]string{"activity_type", "count"}).
			AddRow("visit", 1))

	stats, err := repo.ActivityStats(ctx, scope, filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Total != 1 {
		t.Errorf("expected total 1, got %d", stats.Total)
	}
}

func TestActivityStats_StatusQueryError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}

	mock.ExpectQuery("SELECT status, COUNT\\(\\*\\) FROM activities").
		WithArgs(anyArgs(2)...).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.ActivityStats(ctx, scope, defaultFilter())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestActivityStats_TypeQueryError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}
	filter := defaultFilter()

	mock.ExpectQuery("SELECT status, COUNT\\(\\*\\) FROM activities").
		WithArgs(filter.DateFrom, filter.DateTo).
		WillReturnRows(mock.NewRows([]string{"status", "count"}))

	mock.ExpectQuery("SELECT activity_type, COUNT\\(\\*\\) FROM activities").
		WithArgs(filter.DateFrom, filter.DateTo).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.ActivityStats(ctx, scope, filter)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- CoverageStats tests ---

func TestCoverageStats_AllScope(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}
	targetScope := rbac.TargetScope{AllTargets: true}
	filter := defaultFilter()

	// Count total targets: no conditions for AllTargets.
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM targets").
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(100))

	// Count visited targets: dateFrom + dateTo = 2 args
	mock.ExpectQuery("SELECT COUNT\\(DISTINCT a.target_id\\) FROM activities a").
		WithArgs(filter.DateFrom, filter.DateTo).
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(40))

	stats, err := repo.CoverageStats(ctx, scope, targetScope, filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalTargets != 100 {
		t.Errorf("expected total targets 100, got %d", stats.TotalTargets)
	}
	if stats.VisitedTargets != 40 {
		t.Errorf("expected visited targets 40, got %d", stats.VisitedTargets)
	}
}

func TestCoverageStats_EmptyTargetScope(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}
	targetScope := rbac.TargetScope{} // empty, not AllTargets

	stats, err := repo.CoverageStats(ctx, scope, targetScope, defaultFilter())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalTargets != 0 {
		t.Errorf("expected 0 total targets for empty scope, got %d", stats.TotalTargets)
	}
}

func TestCoverageStats_EmptyActivityScope(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{} // empty, not AllActivities
	targetScope := rbac.TargetScope{AllTargets: true}
	filter := defaultFilter()

	// Count total targets.
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM targets").
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(50))

	stats, err := repo.CoverageStats(ctx, scope, targetScope, filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalTargets != 50 {
		t.Errorf("expected 50, got %d", stats.TotalTargets)
	}
	if stats.VisitedTargets != 0 {
		t.Errorf("expected 0 visited for empty activity scope, got %d", stats.VisitedTargets)
	}
}

func TestCoverageStats_WithUserFilter(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}
	targetScope := rbac.TargetScope{AllTargets: true}
	filter := store.DashboardFilter{
		DateFrom: testTime(),
		DateTo:   testTime().Add(24 * time.Hour),
		UserID:   strPtr("user-1"),
	}

	// Target count: 1 user filter arg
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM targets").
		WithArgs("user-1").
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(20))

	// Visited count: dateFrom + dateTo + userID = 3 args
	mock.ExpectQuery("SELECT COUNT\\(DISTINCT a.target_id\\) FROM activities a").
		WithArgs(anyArgs(3)...).
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(10))

	stats, err := repo.CoverageStats(ctx, scope, targetScope, filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalTargets != 20 {
		t.Errorf("expected 20, got %d", stats.TotalTargets)
	}
	if stats.VisitedTargets != 10 {
		t.Errorf("expected 10, got %d", stats.VisitedTargets)
	}
}

func TestCoverageStats_WithTeamFilter(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}
	targetScope := rbac.TargetScope{AllTargets: true}
	filter := store.DashboardFilter{
		DateFrom: testTime(),
		DateTo:   testTime().Add(24 * time.Hour),
		TeamID:   strPtr("team-1"),
	}

	// Target count: 1 team filter arg
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM targets").
		WithArgs("team-1").
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(30))

	// Visited count: dateFrom + dateTo + teamID = 3 args
	mock.ExpectQuery("SELECT COUNT\\(DISTINCT a.target_id\\) FROM activities a").
		WithArgs(anyArgs(3)...).
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(15))

	stats, err := repo.CoverageStats(ctx, scope, targetScope, filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalTargets != 30 {
		t.Errorf("expected 30, got %d", stats.TotalTargets)
	}
	if stats.VisitedTargets != 15 {
		t.Errorf("expected 15, got %d", stats.VisitedTargets)
	}
}

func TestCoverageStats_TargetCountError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}
	targetScope := rbac.TargetScope{AllTargets: true}

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM targets").
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.CoverageStats(ctx, scope, targetScope, defaultFilter())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCoverageStats_VisitedCountError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}
	targetScope := rbac.TargetScope{AllTargets: true}
	filter := defaultFilter()

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM targets").
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(100))

	mock.ExpectQuery("SELECT COUNT\\(DISTINCT a.target_id\\) FROM activities a").
		WithArgs(filter.DateFrom, filter.DateTo).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.CoverageStats(ctx, scope, targetScope, filter)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- FrequencyStats tests ---

func TestFrequencyStats_AllScope(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}
	targetScope := rbac.TargetScope{AllTargets: true}
	filter := defaultFilter()

	// dateFrom + dateTo = 2 args
	mock.ExpectQuery("SELECT t.fields").
		WithArgs(filter.DateFrom, filter.DateTo).
		WillReturnRows(mock.NewRows([]string{"classification", "target_count", "total_visits"}).
			AddRow("A", 10, 50).
			AddRow("B", 20, 30))

	result, err := repo.FrequencyStats(ctx, scope, targetScope, filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(result))
	}
	if result[0].Classification != "A" {
		t.Errorf("expected classification A, got %s", result[0].Classification)
	}
	if result[0].TotalVisits != 50 {
		t.Errorf("expected 50 visits, got %d", result[0].TotalVisits)
	}
}

func TestFrequencyStats_EmptyTargetScope(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}
	targetScope := rbac.TargetScope{}

	result, err := repo.FrequencyStats(ctx, scope, targetScope, defaultFilter())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result for empty target scope, got %d", len(result))
	}
}

func TestFrequencyStats_QueryError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}
	targetScope := rbac.TargetScope{AllTargets: true}
	filter := defaultFilter()

	mock.ExpectQuery("SELECT t.fields").
		WithArgs(filter.DateFrom, filter.DateTo).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.FrequencyStats(ctx, scope, targetScope, filter)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFrequencyStats_WithFilters(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}
	targetScope := rbac.TargetScope{AllTargets: true}
	filter := store.DashboardFilter{
		DateFrom: testTime(),
		DateTo:   testTime().Add(24 * time.Hour),
		UserID:   strPtr("user-1"),
		TeamID:   strPtr("team-1"),
	}

	// userID + teamID on target + dateFrom + dateTo on activity = 4 args
	mock.ExpectQuery("SELECT t.fields").
		WithArgs(anyArgs(4)...).
		WillReturnRows(mock.NewRows([]string{"classification", "target_count", "total_visits"}).
			AddRow("A", 5, 20))

	result, err := repo.FrequencyStats(ctx, scope, targetScope, filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result))
	}
}

func TestFrequencyStats_ScanError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}
	targetScope := rbac.TargetScope{AllTargets: true}
	filter := defaultFilter()

	// Return wrong column types to trigger scan error
	mock.ExpectQuery("SELECT t.fields").
		WithArgs(filter.DateFrom, filter.DateTo).
		WillReturnRows(mock.NewRows([]string{"classification", "target_count", "total_visits"}).
			AddRow("A", "not-an-int", 50))

	_, err := repo.FrequencyStats(ctx, scope, targetScope, filter)
	if err == nil {
		t.Fatal("expected scan error, got nil")
	}
}

// --- WeekendFieldActivities tests ---

func TestWeekendFieldActivities_AllScope(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}
	filter := defaultFilter()
	fieldTypes := []string{"visit", "call"}

	sat := time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)

	// fieldTypes (1) + dateFrom + dateTo = 3 args
	mock.ExpectQuery("SELECT DISTINCT due_date FROM activities").
		WithArgs(anyArgs(3)...).
		WillReturnRows(mock.NewRows([]string{"due_date"}).AddRow(sat))

	result, err := repo.WeekendFieldActivities(ctx, scope, fieldTypes, filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	if !result[0].DueDate.Equal(sat) {
		t.Errorf("expected %v, got %v", sat, result[0].DueDate)
	}
}

func TestWeekendFieldActivities_EmptyScope(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{}

	result, err := repo.WeekendFieldActivities(ctx, scope, []string{"visit"}, defaultFilter())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for empty scope, got %v", result)
	}
}

func TestWeekendFieldActivities_QueryError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}

	mock.ExpectQuery("SELECT DISTINCT due_date FROM activities").
		WithArgs(anyArgs(3)...).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.WeekendFieldActivities(ctx, scope, []string{"visit"}, defaultFilter())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWeekendFieldActivities_WithUserFilter(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}
	filter := store.DashboardFilter{
		DateFrom: testTime(),
		DateTo:   testTime().Add(24 * time.Hour),
		UserID:   strPtr("user-1"),
	}

	// fieldTypes (1) + dateFrom + dateTo + userID = 4 args
	mock.ExpectQuery("SELECT DISTINCT due_date FROM activities").
		WithArgs(anyArgs(4)...).
		WillReturnRows(mock.NewRows([]string{"due_date"}))

	result, err := repo.WeekendFieldActivities(ctx, scope, []string{"visit"}, filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0, got %d", len(result))
	}
}

// --- RecoveryActivities tests ---

func TestRecoveryActivities_AllScope(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}
	filter := defaultFilter()
	date := testTime()

	// recoveryType (1) + dateFrom + dateTo = 3 args
	mock.ExpectQuery("SELECT due_date FROM activities").
		WithArgs(anyArgs(3)...).
		WillReturnRows(mock.NewRows([]string{"due_date"}).AddRow(date))

	result, err := repo.RecoveryActivities(ctx, scope, "recovery", filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 date, got %d", len(result))
	}
}

func TestRecoveryActivities_EmptyScope(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{}

	result, err := repo.RecoveryActivities(ctx, scope, "recovery", defaultFilter())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for empty scope, got %v", result)
	}
}

func TestRecoveryActivities_QueryError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}

	mock.ExpectQuery("SELECT due_date FROM activities").
		WithArgs(anyArgs(3)...).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.RecoveryActivities(ctx, scope, "recovery", defaultFilter())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRecoveryActivities_ScopedByCreatorIDs(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{CreatorIDs: []string{"user-1"}}
	filter := defaultFilter()

	// recoveryType (1) + 1 scope arg + dateFrom + dateTo = 4 args
	mock.ExpectQuery("SELECT due_date FROM activities").
		WithArgs(anyArgs(4)...).
		WillReturnRows(mock.NewRows([]string{"due_date"}).AddRow(testTime()))

	result, err := repo.RecoveryActivities(ctx, scope, "recovery", filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 date, got %d", len(result))
	}
}

func TestRecoveryActivities_MultipleResults(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newDashboardRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}
	filter := defaultFilter()
	d1 := testTime()
	d2 := testTime().Add(24 * time.Hour)

	mock.ExpectQuery("SELECT due_date FROM activities").
		WithArgs(anyArgs(3)...).
		WillReturnRows(mock.NewRows([]string{"due_date"}).AddRow(d1).AddRow(d2))

	result, err := repo.RecoveryActivities(ctx, scope, "recovery", filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 dates, got %d", len(result))
	}
}

// The old appendDashboardFilter, activityScopeConditions, and
// targetScopeConditions free functions have been replaced by the queryBuilder
// type (query_builder.go). Their coverage is now in query_builder_test.go.
