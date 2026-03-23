package domain

import "time"

// Collection represents a user-created saved group of targets.
type Collection struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatorID string    `json:"creatorId"`
	TeamID    string    `json:"teamId"`
	TargetIDs []string  `json:"targetIds"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
