package domain

// LeadStatus represents the current state of a lead in the sales pipeline.
type LeadStatus string

const (
	// LeadStatusNew is the initial state when a lead is created.
	LeadStatusNew LeadStatus = "new"
	// LeadStatusAssigned indicates the lead has been assigned to a rep.
	LeadStatusAssigned LeadStatus = "assigned"
	// LeadStatusInProgress indicates the rep is actively working the lead.
	LeadStatusInProgress LeadStatus = "in_progress"
	// LeadStatusVisited indicates the rep has visited the customer location.
	LeadStatusVisited LeadStatus = "visited"
	// LeadStatusClosedWon indicates the lead resulted in a successful outcome.
	LeadStatusClosedWon LeadStatus = "closed_won"
	// LeadStatusClosedLost indicates the lead did not result in a successful outcome.
	LeadStatusClosedLost LeadStatus = "closed_lost"
)

// Valid returns true if the status is a recognized LeadStatus value.
func (s LeadStatus) Valid() bool {
	switch s {
	case LeadStatusNew, LeadStatusAssigned, LeadStatusInProgress,
		LeadStatusVisited, LeadStatusClosedWon, LeadStatusClosedLost:
		return true
	}
	return false
}

// Terminal returns true if the status represents a final state.
func (s LeadStatus) Terminal() bool {
	return s == LeadStatusClosedWon || s == LeadStatusClosedLost
}
