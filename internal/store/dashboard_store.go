package store

import (
	"context"
	"time"

	"github.com/pebblr/pebblr/internal/rbac"
)

// StatusCount holds the count of activities for a given status.
type StatusCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

// TypeCount holds the count of activities for a given activity type.
type TypeCount struct {
	ActivityType string `json:"activityType"`
	Count        int    `json:"count"`
}

// ActivityStats holds aggregated activity metrics for a period.
type ActivityStats struct {
	Total          int           `json:"total"`
	SubmittedCount int           `json:"submittedCount"`
	ByStatus       []StatusCount `json:"byStatus"`
	ByType         []TypeCount   `json:"byType"`
}

// CoverageStats holds target coverage metrics for a period.
type CoverageStats struct {
	TotalTargets    int     `json:"totalTargets"`
	VisitedTargets  int     `json:"visitedTargets"`
	CoveragePercent float64 `json:"coveragePercent"`
}

// UserActivityStats holds per-user activity metrics for a period.
type UserActivityStats struct {
	UserID   string        `json:"userId"`
	UserName string        `json:"userName"`
	Total    int           `json:"total"`
	ByStatus []StatusCount `json:"byStatus"`
}

// DashboardFilter specifies filter criteria for dashboard queries.
type DashboardFilter struct {
	DateFrom  time.Time
	DateTo    time.Time
	CreatorID *string
	TeamID    *string
}

// DashboardRepository provides aggregate query access for dashboard metrics.
type DashboardRepository interface {
	// ActivityStats returns aggregated activity counts by status and type for the given scope and filter.
	ActivityStats(ctx context.Context, scope rbac.ActivityScope, filter DashboardFilter) (*ActivityStats, error)

	// CoverageStats returns target coverage (visited vs total) for the given scope and filter.
	// Only counts field-category activity types (those whose keys are in fieldTypes).
	CoverageStats(ctx context.Context, scope rbac.ActivityScope, targetScope rbac.TargetScope, filter DashboardFilter, fieldTypes []string, realizedStatus string) (*CoverageStats, error)

	// UserStats returns per-user activity breakdowns for the given scope and filter.
	UserStats(ctx context.Context, scope rbac.ActivityScope, filter DashboardFilter) ([]UserActivityStats, error)
}
