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

// ActivityStatsResponse is the API-level response for activity statistics.
type ActivityStatsResponse struct {
	ByStatus   map[string]int `json:"byStatus"`
	ByCategory map[string]int `json:"byCategory"`
	Total      int            `json:"total"`
}

// CoverageResponse is the API-level response for target coverage.
type CoverageResponse struct {
	TotalTargets   int     `json:"totalTargets"`
	VisitedTargets int     `json:"visitedTargets"`
	Percentage     float64 `json:"percentage"`
}

// FrequencyResponse is the API-level response for frequency compliance.
type FrequencyResponse struct {
	Items []FrequencyItem `json:"items"`
}

// FrequencyItem represents one classification's visit compliance.
type FrequencyItem struct {
	Classification string  `json:"classification"`
	TargetCount    int     `json:"targetCount"`
	TotalVisits    int     `json:"totalVisits"`
	Required       int     `json:"required"`
	Compliance     float64 `json:"compliance"`
}

// DashboardService provides dashboard analytics with RBAC enforcement.
type DashboardService struct {
	dashboard store.DashboardRepository
	enforcer  rbac.Enforcer
	cfg       *config.TenantConfig
}

// NewDashboardService constructs a DashboardService.
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

// ActivityStats returns activity counts grouped by status and category for the given period.
func (s *DashboardService) ActivityStats(ctx context.Context, actor *domain.User, filter store.DashboardFilter) (*ActivityStatsResponse, error) {
	scope := s.enforcer.ScopeActivityQuery(ctx, actor)
	stats, err := s.dashboard.ActivityStats(ctx, scope, filter)
	if err != nil {
		return nil, fmt.Errorf("querying activity stats: %w", err)
	}

	// Map activity types to categories using config.
	byCategory := map[string]int{}
	if s.cfg != nil {
		for actType, count := range stats.ByCategory {
			at := s.cfg.ActivityType(actType)
			if at != nil {
				byCategory[at.Category] += count
			} else {
				byCategory["unknown"] += count
			}
		}
	} else {
		byCategory = stats.ByCategory
	}

	return &ActivityStatsResponse{
		ByStatus:   stats.ByStatus,
		ByCategory: byCategory,
		Total:      stats.Total,
	}, nil
}

// Coverage returns target coverage statistics for the given period.
func (s *DashboardService) Coverage(ctx context.Context, actor *domain.User, filter store.DashboardFilter) (*CoverageResponse, error) {
	actScope := s.enforcer.ScopeActivityQuery(ctx, actor)
	tgtScope := s.enforcer.ScopeTargetQuery(ctx, actor)

	stats, err := s.dashboard.CoverageStats(ctx, actScope, tgtScope, filter)
	if err != nil {
		return nil, fmt.Errorf("querying coverage stats: %w", err)
	}

	var pct float64
	if stats.TotalTargets > 0 {
		pct = float64(stats.VisitedTargets) / float64(stats.TotalTargets) * 100
	}

	return &CoverageResponse{
		TotalTargets:   stats.TotalTargets,
		VisitedTargets: stats.VisitedTargets,
		Percentage:     pct,
	}, nil
}

// Frequency returns visit frequency compliance per target classification.
func (s *DashboardService) Frequency(ctx context.Context, actor *domain.User, filter store.DashboardFilter) (*FrequencyResponse, error) {
	actScope := s.enforcer.ScopeActivityQuery(ctx, actor)
	tgtScope := s.enforcer.ScopeTargetQuery(ctx, actor)

	rows, err := s.dashboard.FrequencyStats(ctx, actScope, tgtScope, filter)
	if err != nil {
		return nil, fmt.Errorf("querying frequency stats: %w", err)
	}

	items := make([]FrequencyItem, 0, len(rows))
	for _, row := range rows {
		required := 0
		if s.cfg != nil {
			required = s.cfg.Rules.Frequency[row.Classification]
		}

		// Expected = required visits per target * number of targets * number of months in period.
		months := monthsInRange(filter.DateFrom, filter.DateTo)
		expected := required * row.TargetCount * months
		var compliance float64
		if expected > 0 {
			compliance = float64(row.TotalVisits) / float64(expected) * 100
			if compliance > 100 {
				compliance = 100
			}
		}

		items = append(items, FrequencyItem{
			Classification: row.Classification,
			TargetCount:    row.TargetCount,
			TotalVisits:    row.TotalVisits,
			Required:       required,
			Compliance:     compliance,
		})
	}

	return &FrequencyResponse{Items: items}, nil
}

// RecoveryClaimInterval represents a window in which a recovery day can be claimed.
type RecoveryClaimInterval struct {
	WeekendDate time.Time `json:"weekendDate"`
	ClaimFrom   time.Time `json:"claimFrom"`
	ClaimBy     time.Time `json:"claimBy"`
	Claimed     bool      `json:"claimed"`
}

// RecoveryBalanceResponse is the API response for recovery day balance.
type RecoveryBalanceResponse struct {
	Earned    int                     `json:"earned"`
	Taken     int                     `json:"taken"`
	Balance   int                     `json:"balance"`
	Intervals []RecoveryClaimInterval `json:"intervals"`
}

// RecoveryBalance returns the recovery day balance for the actor in the given period.
func (s *DashboardService) RecoveryBalance(ctx context.Context, actor *domain.User, filter store.DashboardFilter) (*RecoveryBalanceResponse, error) {
	if s.cfg == nil || s.cfg.Rules.Recovery == nil || !s.cfg.Rules.Recovery.WeekendActivityFlag {
		return &RecoveryBalanceResponse{Intervals: []RecoveryClaimInterval{}}, nil
	}

	scope := s.enforcer.ScopeActivityQuery(ctx, actor)
	recoveryRule := s.cfg.Rules.Recovery

	// Get field activity types from config.
	var fieldTypes []string
	for i := range s.cfg.Activities.Types {
		if s.cfg.Activities.Types[i].Category == "field" {
			fieldTypes = append(fieldTypes, s.cfg.Activities.Types[i].Key)
		}
	}

	weekendActivities, err := s.dashboard.WeekendFieldActivities(ctx, scope, fieldTypes, filter)
	if err != nil {
		return nil, fmt.Errorf("querying weekend activities: %w", err)
	}

	recoveryDates, err := s.dashboard.RecoveryActivities(ctx, scope, recoveryRule.RecoveryType, filter)
	if err != nil {
		return nil, fmt.Errorf("querying recovery activities: %w", err)
	}

	takenSet := make(map[string]bool)
	for _, d := range recoveryDates {
		takenSet[d.Format("2006-01-02")] = true
	}

	// Build claim intervals: each weekend activity earns one recovery day
	// claimable starting the next business day, within recovery_window_days business days.
	intervals := make([]RecoveryClaimInterval, 0, len(weekendActivities))
	for _, wa := range weekendActivities {
		claimFrom := nextBusinessDay(wa.DueDate)
		claimBy := addBusinessDays(claimFrom, recoveryRule.RecoveryWindowDays-1)

		// Check if any recovery was taken in this window.
		claimed := false
		for _, rd := range recoveryDates {
			rdStr := rd.Format("2006-01-02")
			if !rd.Before(claimFrom) && !rd.After(claimBy) && !takenSet[rdStr+"_used"] {
				claimed = true
				takenSet[rdStr+"_used"] = true
				break
			}
		}

		intervals = append(intervals, RecoveryClaimInterval{
			WeekendDate: wa.DueDate,
			ClaimFrom:   claimFrom,
			ClaimBy:     claimBy,
			Claimed:     claimed,
		})
	}

	earned := len(weekendActivities)
	claimedCount := 0
	for _, iv := range intervals {
		if iv.Claimed {
			claimedCount++
		}
	}

	return &RecoveryBalanceResponse{
		Earned:    earned,
		Taken:     claimedCount,
		Balance:   earned - claimedCount,
		Intervals: intervals,
	}, nil
}

// nextBusinessDay returns the next Monday–Friday after the given date.
func nextBusinessDay(d time.Time) time.Time {
	d = d.AddDate(0, 0, 1)
	for d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
		d = d.AddDate(0, 0, 1)
	}
	return d
}

// addBusinessDays adds n business days (Mon–Fri) to a date.
func addBusinessDays(d time.Time, n int) time.Time {
	for n > 0 {
		d = d.AddDate(0, 0, 1)
		if d.Weekday() != time.Saturday && d.Weekday() != time.Sunday {
			n--
		}
	}
	return d
}

// monthsInRange returns the number of calendar months spanned by the date range (minimum 1).
func monthsInRange(from, to time.Time) int {
	if to.Before(from) {
		return 1
	}
	months := (to.Year()-from.Year())*12 + int(to.Month()) - int(from.Month()) + 1
	if months < 1 {
		return 1
	}
	return months
}
