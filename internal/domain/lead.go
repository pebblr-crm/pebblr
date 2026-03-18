package domain

import "time"

// Lead represents a sales opportunity that a rep visits at a customer location.
type Lead struct {
	ID           string       `json:"id"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	Status       LeadStatus   `json:"status"`
	AssigneeID   string       `json:"assigneeId"` // User.ID of the assigned rep
	TeamID       string       `json:"teamId"`     // Team.ID that owns this lead
	CustomerID   string       `json:"customerId"` // Customer.ID this lead is associated with
	CustomerType CustomerType `json:"customerType"`
	CreatedAt    time.Time    `json:"createdAt"`
	UpdatedAt    time.Time    `json:"updatedAt"`
	// DeletedAt is set when the lead is soft-deleted. Nil means active.
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}
