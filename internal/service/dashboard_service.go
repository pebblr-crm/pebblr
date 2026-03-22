package service

import (
	"context"
	"fmt"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/store"
)

// DashboardStats holds aggregate metrics for the dashboard.
type DashboardStats struct {
	TotalLeads      int     `json:"totalLeads"`
	ConversionRate  float64 `json:"conversionRate"`
	UnassignedCount int     `json:"unassignedCount"`
}

// DashboardService computes dashboard statistics from the lead store.
type DashboardService struct {
	leads store.LeadRepository
}

// NewDashboardService constructs a DashboardService with the given repository.
func NewDashboardService(leads store.LeadRepository) *DashboardService {
	return &DashboardService{leads: leads}
}

// Stats returns aggregate dashboard statistics.
func (s *DashboardService) Stats(ctx context.Context) (*DashboardStats, error) {
	// Use an admin-scope query to count all leads.
	scope := rbac.LeadScope{AllLeads: true}

	allPage, err := s.leads.List(ctx, scope, store.LeadFilter{}, 1, 1)
	if err != nil {
		return nil, fmt.Errorf("counting leads: %w", err)
	}
	total := allPage.Total

	// Count closed_won leads for conversion rate numerator.
	wonStatus := domain.LeadStatusClosedWon
	wonPage, err := s.leads.List(ctx, scope, store.LeadFilter{Status: &wonStatus}, 1, 1)
	if err != nil {
		return nil, fmt.Errorf("counting won leads: %w", err)
	}

	// Count closed_lost leads for conversion rate denominator.
	lostStatus := domain.LeadStatusClosedLost
	lostPage, err := s.leads.List(ctx, scope, store.LeadFilter{Status: &lostStatus}, 1, 1)
	if err != nil {
		return nil, fmt.Errorf("counting lost leads: %w", err)
	}

	var conversionRate float64
	closedTotal := wonPage.Total + lostPage.Total
	if closedTotal > 0 {
		conversionRate = float64(wonPage.Total) / float64(closedTotal)
	}

	// Count unassigned leads (no assignee).
	emptyAssignee := ""
	unassignedPage, err := s.leads.List(ctx, scope, store.LeadFilter{Assignee: &emptyAssignee}, 1, 1)
	if err != nil {
		// Non-fatal: unassigned count defaults to 0 on error.
		unassignedPage = &store.LeadPage{Total: 0}
	}

	return &DashboardStats{
		TotalLeads:      total,
		ConversionRate:  conversionRate,
		UnassignedCount: unassignedPage.Total,
	}, nil
}
