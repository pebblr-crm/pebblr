package domain

import "time"

// Territory represents a geographic region assigned to a team for coverage tracking.
type Territory struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	TeamID    string         `json:"teamId"`
	Region    string         `json:"region,omitempty"`
	Boundary  map[string]any `json:"boundary,omitempty"` // GeoJSON polygon
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}
