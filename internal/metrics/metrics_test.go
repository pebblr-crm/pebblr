package metrics_test

import (
	"testing"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/metrics"
)

func TestPipelineStatsStructure(t *testing.T) {
	// Verify the struct can be constructed and fields are accessible.
	stats := &metrics.PipelineStats{
		TotalLeads: 10,
		ByStatus: map[domain.LeadStatus]int{
			domain.LeadStatusNew:        5,
			domain.LeadStatusClosedWon:  3,
			domain.LeadStatusClosedLost: 2,
		},
		WinRate:         0.6,
		AverageDaysOpen: 7.5,
	}

	if stats.TotalLeads != 10 {
		t.Errorf("expected TotalLeads=10, got %d", stats.TotalLeads)
	}
	if stats.WinRate != 0.6 {
		t.Errorf("expected WinRate=0.6, got %f", stats.WinRate)
	}
}
