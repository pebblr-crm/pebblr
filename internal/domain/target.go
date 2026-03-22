package domain

import "time"

// Target represents an entity that a rep visits — e.g. a doctor, pharmacy.
// The target_type and dynamic fields are driven by the tenant configuration.
type Target struct {
	ID         string         `json:"id"`
	ExternalID string         `json:"externalId,omitempty"` // external system identifier for import upsert
	TargetType string         `json:"targetType"`           // key from tenant config (e.g. "doctor", "pharmacy")
	Name       string         `json:"name"`
	Fields     map[string]any `json:"fields"`     // dynamic fields defined in tenant config
	AssigneeID string         `json:"assigneeId"` // User.ID of the assigned rep (territory)
	TeamID     string         `json:"teamId"`
	ImportedAt *time.Time     `json:"importedAt,omitempty"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
}
