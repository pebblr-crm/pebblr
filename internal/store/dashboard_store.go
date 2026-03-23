package store

import (
	"context"
	"time"

	"github.com/pebblr/pebblr/internal/rbac"
)

// ActivityStats holds counts of activities grouped by status and category.
type ActivityStats struct {
	ByStatus   map[string]int `json:"byStatus"`
	ByCategory map[string]int `json:"byCategory"`
	Total      int            `json:"total"`
}

// CoverageStats holds target coverage data for a period.
type CoverageStats struct {
	TotalTargets   int `json:"totalTargets"`
	VisitedTargets int `json:"visitedTargets"`
}

// FrequencyRow holds visit counts vs required frequency for one classification.
type FrequencyRow struct {
	Classification string `json:"classification"`
	TargetCount    int    `json:"targetCount"`
	TotalVisits    int    `json:"totalVisits"`
	Required       int    `json:"required"`
}

// DashboardFilter specifies the period and optional user/team scope for dashboard queries.
type DashboardFilter struct {
	DateFrom time.Time
	DateTo   time.Time
	UserID   *string
	TeamID   *string
}

// DashboardRepository provides aggregation queries for the dashboard.
type DashboardRepository interface {
	// ActivityStats returns activity counts grouped by status and category for the given scope and filter.
	ActivityStats(ctx context.Context, scope rbac.ActivityScope, filter DashboardFilter) (*ActivityStats, error)

	// CoverageStats returns the total vs visited target counts for the given scope and filter.
	CoverageStats(ctx context.Context, scope rbac.ActivityScope, targetScope rbac.TargetScope, filter DashboardFilter) (*CoverageStats, error)

	// FrequencyStats returns per-classification visit counts for the given scope and filter.
	FrequencyStats(ctx context.Context, scope rbac.ActivityScope, targetScope rbac.TargetScope, filter DashboardFilter) ([]FrequencyRow, error)
}
