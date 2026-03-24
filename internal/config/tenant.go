package config

// TenantConfig is the root configuration for a tenant deployment.
// It defines account types, activity types, option lists, and business rules.
// Loaded once at startup from a JSON file; served read-only to the frontend.
type TenantConfig struct {
	Tenant     TenantInfo              `json:"tenant"`
	Accounts   AccountsConfig          `json:"accounts"`
	Activities ActivitiesConfig        `json:"activities"`
	Options    map[string][]OptionDef  `json:"options"`
	Rules      RulesConfig             `json:"rules"`
	Recovery   *RecoveryRule           `json:"recovery,omitempty"`
}

// TenantInfo holds basic tenant metadata.
type TenantInfo struct {
	Name   string `json:"name"`
	Locale string `json:"locale"`
}

// AccountsConfig groups the account type definitions.
type AccountsConfig struct {
	Types []AccountTypeConfig `json:"types"`
}

// AccountTypeConfig defines a kind of account (e.g. doctor, pharmacy).
type AccountTypeConfig struct {
	Key    string        `json:"key"`
	Label  string        `json:"label"`
	Fields []FieldConfig `json:"fields"`
}

// ActivitiesConfig groups statuses, durations, activity type definitions,
// and routing options.
type ActivitiesConfig struct {
	Statuses          []StatusDef              `json:"statuses"`
	StatusTransitions map[string][]string      `json:"status_transitions"`
	Durations         []OptionDef              `json:"durations"`
	Types             []ActivityTypeConfig     `json:"types"`
	RoutingOptions    []OptionDef              `json:"routing_options"`
	HoistedFields     []string                 `json:"hoisted_fields"`
}

// StatusDef defines a valid activity status.
type StatusDef struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Initial     bool   `json:"initial,omitempty"`
	Submittable bool   `json:"submittable,omitempty"` // allows report submission
}

// OptionDef is a generic key/label pair used for select options,
// durations, routing options, etc.
type OptionDef struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// ActivityTypeConfig defines a kind of activity (e.g. visit, administrative).
type ActivityTypeConfig struct {
	Key                    string        `json:"key"`
	Label                  string        `json:"label"`
	Category               string        `json:"category"` // "field" or "non_field"
	TitleField             string        `json:"title_field,omitempty"`
	HasDuration            bool          `json:"has_duration,omitempty"` // whether this type uses duration
	Fields                 []FieldConfig `json:"fields"`
	SubmitRequired         []string      `json:"submit_required,omitempty"`
	BlocksFieldActivities  bool          `json:"blocks_field_activities,omitempty"`
}

// FieldConfig defines a single field on an account or activity type.
type FieldConfig struct {
	Key        string   `json:"key"`
	Label      string   `json:"label,omitempty"`
	Type       string   `json:"type"` // text, select, multi_select, relation, date
	Required   bool     `json:"required"`
	Editable   *bool    `json:"editable,omitempty"` // nil means editable
	Options    []string `json:"options,omitempty"`
	OptionsRef string   `json:"options_ref,omitempty"`
}

// RulesConfig holds business rules that govern activity planning.
type RulesConfig struct {
	Frequency                    map[string]int `json:"frequency"`
	MaxActivitiesPerDay          int            `json:"max_activities_per_day"`
	VisitCadenceDays             int            `json:"visit_cadence_days"` // min days between visits to the same target
	DefaultVisitDurationMinutes  map[string]int `json:"default_visit_duration_minutes"`
	VisitDurationStepMinutes     int            `json:"visit_duration_step_minutes"`
}

// RecoveryRule defines weekend/recovery day rules.
type RecoveryRule struct {
	WeekendActivityFlag  bool   `json:"weekend_activity_flag"`
	RecoveryWindowDays   int    `json:"recovery_window_days"`
	RecoveryType         string `json:"recovery_type"`
}

// AccountType returns the AccountTypeConfig for the given key, or nil.
func (c *TenantConfig) AccountType(key string) *AccountTypeConfig {
	for i := range c.Accounts.Types {
		if c.Accounts.Types[i].Key == key {
			return &c.Accounts.Types[i]
		}
	}
	return nil
}

// ActivityType returns the ActivityTypeConfig for the given key, or nil.
func (c *TenantConfig) ActivityType(key string) *ActivityTypeConfig {
	for i := range c.Activities.Types {
		if c.Activities.Types[i].Key == key {
			return &c.Activities.Types[i]
		}
	}
	return nil
}

// InitialStatus returns the key of the status marked as initial, or "".
func (c *TenantConfig) InitialStatus() string {
	for _, s := range c.Activities.Statuses {
		if s.Initial {
			return s.Key
		}
	}
	return ""
}

// IsValidStatus reports whether key is a defined status.
func (c *TenantConfig) IsValidStatus(key string) bool {
	for _, s := range c.Activities.Statuses {
		if s.Key == key {
			return true
		}
	}
	return false
}

// IsSubmittableStatus reports whether the given status allows report submission.
func (c *TenantConfig) IsSubmittableStatus(key string) bool {
	for _, s := range c.Activities.Statuses {
		if s.Key == key {
			return s.Submittable
		}
	}
	return false
}

// IsValidTransition reports whether moving from fromStatus to toStatus is allowed.
func (c *TenantConfig) IsValidTransition(fromStatus, toStatus string) bool {
	allowed, ok := c.Activities.StatusTransitions[fromStatus]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == toStatus {
			return true
		}
	}
	return false
}

// ResolveOptions returns the option list for an options_ref key.
func (c *TenantConfig) ResolveOptions(ref string) []OptionDef {
	// Check top-level options map first.
	if opts, ok := c.Options[ref]; ok {
		return opts
	}
	// Check durations as a special case.
	if ref == "durations" {
		return c.Activities.Durations
	}
	return nil
}
