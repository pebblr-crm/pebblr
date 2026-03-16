package domain

// Team represents a group of sales reps managed by a single manager.
type Team struct {
	ID        string
	Name      string
	ManagerID string // User.ID of the team's manager
}
