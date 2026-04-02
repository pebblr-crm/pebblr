package config

import "fmt"

// FieldError describes a validation failure on a single field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e FieldError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// HoistedFieldKeys lists config field keys that map to top-level Activity
// columns in the database rather than living in the JSONB fields map.
// These are validated separately (e.g. ValidateDuration) and must be skipped
// by ValidateActivity.
//
// If you add a new hoisted column to the activities table, you must also add
// it here, in the Activity struct, and in the postgres repository scan.
var HoistedFieldKeys = []string{
	"duration",
	"account_id",
	"routing",
	"joint_visit_user_id",
}

// hoistedFields is the set-form of HoistedFieldKeys for O(1) lookups.
var hoistedFields = func() map[string]bool {
	m := make(map[string]bool, len(HoistedFieldKeys))
	for _, k := range HoistedFieldKeys {
		m[k] = true
	}
	return m
}()

// ValidateActivity validates field values for an activity against the
// tenant config. phase is "save" or "submit" — submit enforces
// additional required fields defined in submit_required.
func ValidateActivity(cfg *TenantConfig, activityType string, fields map[string]any, phase string) []FieldError {
	at := cfg.ActivityType(activityType)
	if at == nil {
		return []FieldError{{Field: "activity_type", Message: fmt.Sprintf("unknown activity type %q", activityType)}}
	}

	submitRequired := buildSubmitRequired(at, phase)

	var errs []FieldError
	for _, fc := range at.Fields {
		if hoistedFields[fc.Key] {
			continue
		}
		errs = append(errs, validateField(cfg, fc, fields, submitRequired)...)
	}

	return errs
}

// buildSubmitRequired returns a set of field keys that are required during submit phase.
func buildSubmitRequired(at *ActivityTypeConfig, phase string) map[string]bool {
	required := make(map[string]bool)
	if phase == "submit" {
		for _, k := range at.SubmitRequired {
			required[k] = true
		}
	}
	return required
}

// validateField validates a single field config entry against the provided fields map.
func validateField(cfg *TenantConfig, fc FieldConfig, fields map[string]any, submitRequired map[string]bool) []FieldError {
	val, present := fields[fc.Key]

	isRequired := fc.Required || submitRequired[fc.Key]
	if isRequired && (!present || isEmpty(val)) {
		return []FieldError{{Field: fc.Key, Message: "field is required"}}
	}

	if !present || val == nil {
		return nil
	}

	switch fc.Type {
	case "select":
		return validateSelect(cfg, fc, val)
	case "multi_select":
		return validateMultiSelect(cfg, fc, val)
	}
	return nil
}

// ValidateStatus checks that a status value is valid per the config.
func ValidateStatus(cfg *TenantConfig, status string) *FieldError {
	if !cfg.IsValidStatus(status) {
		return &FieldError{Field: "status", Message: fmt.Sprintf("unknown status %q", status)}
	}
	return nil
}

// ValidateStatusTransition checks that a status transition is allowed.
func ValidateStatusTransition(cfg *TenantConfig, from, to string) *FieldError {
	if !cfg.IsValidTransition(from, to) {
		return &FieldError{
			Field:   "status",
			Message: fmt.Sprintf("transition from %q to %q is not allowed", from, to),
		}
	}
	return nil
}

// ValidateDuration checks that a duration value is valid per the config.
func ValidateDuration(cfg *TenantConfig, duration string) *FieldError {
	for _, d := range cfg.Activities.Durations {
		if d.Key == duration {
			return nil
		}
	}
	return &FieldError{Field: "duration", Message: fmt.Sprintf("unknown duration %q", duration)}
}

func validateSelect(cfg *TenantConfig, fc FieldConfig, val any) []FieldError {
	s, ok := val.(string)
	if !ok {
		return []FieldError{{Field: fc.Key, Message: "must be a string"}}
	}

	allowed := resolveAllowed(cfg, fc)
	if allowed != nil && !allowed[s] {
		return []FieldError{{Field: fc.Key, Message: fmt.Sprintf("invalid option %q", s)}}
	}
	return nil
}

func validateMultiSelect(cfg *TenantConfig, fc FieldConfig, val any) []FieldError {
	arr, ok := val.([]any)
	if !ok {
		return []FieldError{{Field: fc.Key, Message: "must be an array"}}
	}

	allowed := resolveAllowed(cfg, fc)
	if allowed == nil {
		return nil
	}

	for _, item := range arr {
		s, ok := item.(string)
		if !ok {
			return []FieldError{{Field: fc.Key, Message: "array items must be strings"}}
		}
		if !allowed[s] {
			return []FieldError{{Field: fc.Key, Message: fmt.Sprintf("invalid option %q", s)}}
		}
	}
	return nil
}

// resolveAllowed builds a set of allowed option keys for a field config.
func resolveAllowed(cfg *TenantConfig, fc FieldConfig) map[string]bool {
	// Inline options take precedence.
	if len(fc.Options) > 0 {
		m := make(map[string]bool, len(fc.Options))
		for _, o := range fc.Options {
			m[o] = true
		}
		return m
	}
	// Try options_ref.
	if fc.OptionsRef != "" {
		opts := cfg.ResolveOptions(fc.OptionsRef)
		if opts != nil {
			m := make(map[string]bool, len(opts))
			for _, o := range opts {
				m[o.Key] = true
			}
			return m
		}
	}
	return nil
}

func isEmpty(val any) bool {
	if val == nil {
		return true
	}
	switch v := val.(type) {
	case string:
		return v == ""
	case []any:
		return len(v) == 0
	}
	return false
}
