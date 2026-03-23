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

// PrepareForResponse injects hoisted column values back into the Fields map
// so the frontend sees them as dynamic fields driven by the tenant config.
func (a *Activity) PrepareForResponse() {
	if a.Fields == nil {
		a.Fields = map[string]any{}
	}
	if a.JointVisitUID != "" {
		a.Fields["joint_visit_user_id"] = a.JointVisitUID
	}
}

// IsSubmitted returns true if the activity has been submitted and is locked.
func (a *Activity) IsSubmitted() bool {
	return a.SubmittedAt != nil
}

// ActivityPatch holds a partial update payload for server-side apply PATCH semantics.
// Nil pointer fields mean "not provided — leave the existing value untouched".
// Non-nil pointer fields (including pointers to zero values) are applied to the existing record.
// Fields uses merge semantics: when FieldsPresent is true, keys in Fields are applied
// individually (nil values clear the key; non-nil values overwrite it). Absent sub-keys
// in Fields are left untouched.
type ActivityPatch struct {
	Status        *string
	DueDate       *time.Time
	Duration      *string
	Routing       *string
	Fields        map[string]any
	FieldsPresent bool // true when the "fields" key appeared in the PATCH body
	TargetID      *string
	JointVisitUID *string
}

// ApplyTo merges the patch into dst in-place, respecting server-side apply semantics.
// Only non-nil pointer fields are applied. When FieldsPresent is true the Fields map
// is merged key-by-key: nil values remove a key, non-nil values overwrite it.
func (p *ActivityPatch) ApplyTo(dst *Activity) {
	if p.Status != nil {
		dst.Status = *p.Status
	}
	if p.DueDate != nil {
		dst.DueDate = *p.DueDate
	}
	if p.Duration != nil {
		dst.Duration = *p.Duration
	}
	if p.Routing != nil {
		dst.Routing = *p.Routing
	}
	if p.FieldsPresent {
		if dst.Fields == nil {
			dst.Fields = map[string]any{}
		}
		for k, v := range p.Fields {
			if v == nil {
				delete(dst.Fields, k)
			} else {
				dst.Fields[k] = v
			}
		}
	}
	if p.TargetID != nil {
		dst.TargetID = *p.TargetID
	}
	if p.JointVisitUID != nil {
		dst.JointVisitUID = *p.JointVisitUID
	}
}
