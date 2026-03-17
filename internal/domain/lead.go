package domain

import "time"

// Lead represents a sales opportunity that a rep visits at a customer location.
type Lead struct {
	ID           string       `json:"id"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	Status       LeadStatus   `json:"status"`
	AssigneeID   string       `json:"assignee_id"` // User.ID of the assigned rep
	TeamID       string       `json:"team_id"`     // Team.ID that owns this lead
	CustomerID   string       `json:"customer_id"` // Customer.ID this lead is associated with
	CustomerType CustomerType `json:"customer_type"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	// DeletedAt is set when the lead is soft-deleted. Nil means active.
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
