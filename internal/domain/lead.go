package domain

import "time"

// Lead represents a sales opportunity that a rep visits at a customer location.
type Lead struct {
	ID           string
	Title        string
	Description  string
	Status       LeadStatus
	AssigneeID   string // User.ID of the assigned rep
	TeamID       string // Team.ID that owns this lead
	CustomerID   string // Customer.ID this lead is associated with
	CustomerType CustomerType
	CreatedAt    time.Time
	UpdatedAt    time.Time
	// DeletedAt is set when the lead is soft-deleted. Nil means active.
	DeletedAt *time.Time
}
