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
	// LeadStatusScheduled indicates a visit has been scheduled for the lead.
	LeadStatusScheduled LeadStatus = "scheduled"
	// LeadStatusDone indicates the lead interaction is complete.
	LeadStatusDone LeadStatus = "done"
	// LeadStatusHotLead indicates a high-priority lead requiring immediate attention.
	LeadStatusHotLead LeadStatus = "hot_lead"
	// LeadStatusFollowUp indicates the lead requires a follow-up action.
	LeadStatusFollowUp LeadStatus = "follow_up"
	// LeadStatusInquiry indicates an inbound inquiry lead.
	LeadStatusInquiry LeadStatus = "inquiry"
)

// Valid returns true if the status is a recognized LeadStatus value.
func (s LeadStatus) Valid() bool {
	switch s {
	case LeadStatusNew, LeadStatusAssigned, LeadStatusInProgress,
		LeadStatusVisited, LeadStatusClosedWon, LeadStatusClosedLost,
		LeadStatusScheduled, LeadStatusDone, LeadStatusHotLead,
		LeadStatusFollowUp, LeadStatusInquiry:
		return true
	}
	return false
}

// Terminal returns true if the status represents a final state.
func (s LeadStatus) Terminal() bool {
	return s == LeadStatusClosedWon || s == LeadStatusClosedLost || s == LeadStatusDone
}
