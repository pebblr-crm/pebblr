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
	// Company is the name of the company this lead is associated with.
	Company string `json:"company"`
	// Industry is the industry sector of the lead's company.
	Industry string `json:"industry"`
	// Location is the geographic location of the lead.
	Location string `json:"location"`
	// ValueCents is the estimated deal value in cents (use cents to avoid float precision issues).
	ValueCents int64 `json:"valueCents"`
	// Initials are the display initials for the lead contact (e.g. "JD" for John Doe).
	Initials  string    `json:"initials"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	// DeletedAt is set when the lead is soft-deleted. Nil means active.
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}
