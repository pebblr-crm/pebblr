package domain

import "time"

// Activity represents a scheduled or completed field action — e.g. a visit,
// administrative task, or time-off. Activity types and their fields are
// driven by the tenant configuration.
type Activity struct {
	ID             string         `json:"id"`
	ActivityType   string         `json:"activityType"`           // key from tenant config (e.g. "visit", "administrative")
	Status         string         `json:"status"`                 // key from config statuses (e.g. "planificat", "realizat")
	DueDate        time.Time      `json:"dueDate"`                // the scheduled date
	Duration       string         `json:"duration"`               // key from config durations (e.g. "full_day", "half_day")
	Routing        string         `json:"routing,omitempty"`      // optional routing week
	Fields         map[string]any `json:"fields"`                 // dynamic fields defined in tenant config
	TargetID       string         `json:"targetId,omitempty"`     // linked target (required for visits, empty for time-off)
	CreatorID      string         `json:"creatorId"`              // the rep who created it
	JointVisitUID  string         `json:"jointVisitUserId,omitempty"` // optional co-visitor
	TeamID         string         `json:"teamId,omitempty"`
	SubmittedAt    *time.Time     `json:"submittedAt,omitempty"`  // when the report was submitted (locks editing)
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      *time.Time     `json:"deletedAt,omitempty"`    // soft delete
}

// IsSubmitted returns true if the activity has been submitted and is locked.
func (a *Activity) IsSubmitted() bool {
	return a.SubmittedAt != nil
}
