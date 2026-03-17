package domain

// Team represents a group of sales reps managed by a single manager.
type Team struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ManagerID string `json:"manager_id"` // User.ID of the team's manager
}
