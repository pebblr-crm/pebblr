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

// DashboardService provides RBAC-scoped dashboard metrics.
type DashboardService struct {
	dashboard store.DashboardRepository
	enforcer  rbac.Enforcer
	cfg       *config.TenantConfig
}

// NewDashboardService constructs a DashboardService with the given dependencies.
func NewDashboardService(
	dashboard store.DashboardRepository,
	enforcer rbac.Enforcer,
	cfg *config.TenantConfig,
) *DashboardService {
	return &DashboardService{
		dashboard: dashboard,
		enforcer:  enforcer,
		cfg:       cfg,
	}
}

// DashboardStatsResponse is the response for the activity stats endpoint.
type DashboardStatsResponse struct {
	Period     string            `json:"period"`
	DateFrom   string            `json:"dateFrom"`
	DateTo     string            `json:"dateTo"`
	Stats      *store.ActivityStats `json:"stats"`
	ByCategory []CategoryCount   `json:"byCategory"`
}

// CategoryCount holds the count of activities for a given category (field/non_field).
type CategoryCount struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

// parsePeriod converts a "YYYY-MM" period string to dateFrom/dateTo.
func parsePeriod(period string) (dateFrom, dateTo time.Time, err error) {
	t, err := time.Parse("2006-01", period)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid period format (expected YYYY-MM): %w", err)
	}
	dateFrom = t
	dateTo = t.AddDate(0, 1, -1) // last day of month
	return dateFrom, dateTo, nil
}

// ActivityStats returns aggregated activity metrics for a period, scoped by RBAC.
func (s *DashboardService) ActivityStats(ctx context.Context, actor *domain.User, period string, filter store.DashboardFilter) (*DashboardStatsResponse, error) {
	dateFrom, dateTo, err := parsePeriod(period)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidInput, err)
	}
	filter.DateFrom = dateFrom
	filter.DateTo = dateTo

	scope := s.enforcer.ScopeActivityQuery(ctx, actor)
	stats, err := s.dashboard.ActivityStats(ctx, scope, filter)
	if err != nil {
		return nil, fmt.Errorf("querying activity stats: %w", err)
	}

	// Compute category breakdown from type counts using config.
	byCategory := s.computeCategoryBreakdown(stats.ByType)

	return &DashboardStatsResponse{
		Period:     period,
		DateFrom:   dateFrom.Format("2006-01-02"),
		DateTo:     dateTo.Format("2006-01-02"),
		Stats:      stats,
		ByCategory: byCategory,
	}, nil
}

// CoverageStats returns target coverage metrics for a period, scoped by RBAC.
func (s *DashboardService) CoverageStats(ctx context.Context, actor *domain.User, period string, filter store.DashboardFilter) (*store.CoverageStats, error) {
	dateFrom, dateTo, err := parsePeriod(period)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidInput, err)
	}
	filter.DateFrom = dateFrom
	filter.DateTo = dateTo

	scope := s.enforcer.ScopeActivityQuery(ctx, actor)
	targetScope := s.enforcer.ScopeTargetQuery(ctx, actor)

	fieldTypes := s.fieldActivityTypes()
	realizedStatus := s.realizedStatus()

	return s.dashboard.CoverageStats(ctx, scope, targetScope, filter, fieldTypes, realizedStatus)
}

// UserStats returns per-user activity breakdowns for a period, scoped by RBAC.
func (s *DashboardService) UserStats(ctx context.Context, actor *domain.User, period string, filter store.DashboardFilter) ([]store.UserActivityStats, error) {
	dateFrom, dateTo, err := parsePeriod(period)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidInput, err)
	}
	filter.DateFrom = dateFrom
	filter.DateTo = dateTo

	scope := s.enforcer.ScopeActivityQuery(ctx, actor)
	return s.dashboard.UserStats(ctx, scope, filter)
}

// computeCategoryBreakdown maps activity type counts to field/non_field categories.
func (s *DashboardService) computeCategoryBreakdown(byType []store.TypeCount) []CategoryCount {
	categoryMap := map[string]int{
		"field":     0,
		"non_field": 0,
	}
	for _, tc := range byType {
		cat := "non_field" // default
		if s.cfg != nil {
			at := s.cfg.ActivityType(tc.ActivityType)
			if at != nil {
				cat = at.Category
			}
		}
		categoryMap[cat] += tc.Count
	}

	result := make([]CategoryCount, 0, len(categoryMap))
	for cat, count := range categoryMap {
		if count > 0 {
			result = append(result, CategoryCount{Category: cat, Count: count})
		}
	}
	return result
}

// fieldActivityTypes returns the keys of all field-category activity types.
func (s *DashboardService) fieldActivityTypes() []string {
	if s.cfg == nil {
		return nil
	}
	var types []string
	for _, at := range s.cfg.Activities.Types {
		if at.Category == "field" {
			types = append(types, at.Key)
		}
	}
	return types
}

// realizedStatus returns the status key that represents "realized" activities.
// Falls back to "realizat" if config doesn't define one explicitly.
func (s *DashboardService) realizedStatus() string {
	if s.cfg == nil {
		return "realizat"
	}
	// Use the second non-initial status, or fall back to "realizat".
	for _, st := range s.cfg.Activities.Statuses {
		if !st.Initial && st.Key != "anulat" {
			return st.Key
		}
	}
	return "realizat"
}
