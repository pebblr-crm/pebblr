package metrics

import (
	"context"
	"time"
)

// MetricsService provides business telemetry aggregations.
type MetricsService interface {
	// PipelineStats returns aggregate lead statistics for the given time window.
	PipelineStats(ctx context.Context, from, to time.Time) (*PipelineStats, error)

	// RepPerformance returns productivity metrics for the given rep.
	RepPerformance(ctx context.Context, userID string, from, to time.Time) (*RepPerformance, error)

	// TeamOverview returns an aggregate view of a team's performance.
	TeamOverview(ctx context.Context, teamID string, from, to time.Time) (*TeamOverview, error)
}
