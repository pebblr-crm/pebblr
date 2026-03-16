package metrics

import "github.com/pebblr/pebblr/internal/domain"

// PipelineStats summarises the state of leads across the entire pipeline.
type PipelineStats struct {
	TotalLeads      int
	ByStatus        map[domain.LeadStatus]int
	WinRate         float64 // closed_won / (closed_won + closed_lost)
	AverageDaysOpen float64
}

// RepPerformance captures productivity metrics for a single sales rep.
type RepPerformance struct {
	UserID      string
	UserName    string
	TotalLeads  int
	VisitedCount int
	WonCount    int
	LostCount   int
	WinRate     float64
}

// TeamOverview aggregates performance across all reps in a team.
type TeamOverview struct {
	TeamID      string
	TeamName    string
	TotalLeads  int
	WinRate     float64
	RepCount    int
	Reps        []RepPerformance
}
