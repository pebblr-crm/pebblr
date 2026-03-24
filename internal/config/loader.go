package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Load reads and parses a tenant config JSON file, then validates
// its internal consistency (option refs resolve, status transitions
// reference valid statuses, etc.).
func Load(path string) (*TenantConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading tenant config: %w", err)
	}

	var cfg TenantConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing tenant config: %w", err)
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid tenant config: %w", err)
	}

	// Populate computed fields for the frontend.
	cfg.Activities.HoistedFields = HoistedFieldKeys()

	return &cfg, nil
}

// validateConfig checks internal consistency of the loaded config.
func validateConfig(cfg *TenantConfig) error {
	if cfg.Tenant.Name == "" {
		return fmt.Errorf("tenant.name is required")
	}

	statusKeys, err := validateStatuses(cfg)
	if err != nil {
		return err
	}

	if err := validateStatusTransitions(cfg, statusKeys); err != nil {
		return err
	}

	if err := validateAccountTypes(cfg); err != nil {
		return err
	}

	return validateActivityTypes(cfg)
}

// validateStatuses checks that statuses are well-formed and returns the set of known status keys.
func validateStatuses(cfg *TenantConfig) (map[string]bool, error) {
	if len(cfg.Activities.Statuses) == 0 {
		return nil, fmt.Errorf("activities.statuses must not be empty")
	}

	initialCount := 0
	statusKeys := make(map[string]bool)
	for _, s := range cfg.Activities.Statuses {
		if s.Key == "" {
			return nil, fmt.Errorf("status key must not be empty")
		}
		if statusKeys[s.Key] {
			return nil, fmt.Errorf("duplicate status key: %q", s.Key)
		}
		statusKeys[s.Key] = true
		if s.Initial {
			initialCount++
		}
	}
	if initialCount != 1 {
		return nil, fmt.Errorf("exactly one status must be marked initial, found %d", initialCount)
	}
	return statusKeys, nil
}

// validateStatusTransitions checks that all transition references point to known statuses.
func validateStatusTransitions(cfg *TenantConfig, statusKeys map[string]bool) error {
	for from, targets := range cfg.Activities.StatusTransitions {
		if !statusKeys[from] {
			return fmt.Errorf("status_transitions references unknown status %q", from)
		}
		for _, to := range targets {
			if !statusKeys[to] {
				return fmt.Errorf("status_transitions[%q] references unknown target status %q", from, to)
			}
		}
	}
	return nil
}

// validateAccountTypes checks that all account types have valid keys and fields.
func validateAccountTypes(cfg *TenantConfig) error {
	for _, at := range cfg.Accounts.Types {
		if at.Key == "" {
			return fmt.Errorf("account type key must not be empty")
		}
		if err := validateFieldConfigs(cfg, at.Fields, "accounts.types["+at.Key+"]"); err != nil {
			return err
		}
	}
	return nil
}

// validateActivityTypes checks that all activity types are well-formed.
func validateActivityTypes(cfg *TenantConfig) error {
	actTypeKeys := make(map[string]bool)
	for i := range cfg.Activities.Types {
		at := &cfg.Activities.Types[i]
		if err := validateSingleActivityType(cfg, at, actTypeKeys); err != nil {
			return err
		}
	}
	return nil
}

// validateSingleActivityType validates one activity type entry.
func validateSingleActivityType(cfg *TenantConfig, at *ActivityTypeConfig, seen map[string]bool) error {
	if at.Key == "" {
		return fmt.Errorf("activity type key must not be empty")
	}
	if seen[at.Key] {
		return fmt.Errorf("duplicate activity type key: %q", at.Key)
	}
	seen[at.Key] = true

	if at.Category != "field" && at.Category != "non_field" {
		return fmt.Errorf("activity type %q: category must be \"field\" or \"non_field\", got %q", at.Key, at.Category)
	}

	if err := validateFieldConfigs(cfg, at.Fields, "activities.types["+at.Key+"]"); err != nil {
		return err
	}

	return validateActivityTypeFieldRefs(at)
}

// validateActivityTypeFieldRefs checks submit_required and title_field references.
func validateActivityTypeFieldRefs(at *ActivityTypeConfig) error {
	fieldKeys := make(map[string]bool)
	for _, f := range at.Fields {
		fieldKeys[f.Key] = true
	}
	for _, sr := range at.SubmitRequired {
		if !fieldKeys[sr] {
			return fmt.Errorf("activity type %q: submit_required references unknown field %q", at.Key, sr)
		}
	}
	if at.TitleField != "" && !fieldKeys[at.TitleField] {
		return fmt.Errorf("activity type %q: title_field references unknown field %q", at.Key, at.TitleField)
	}
	return nil
}

// dbBackedRefs lists options_ref values that are resolved at runtime
// from database entities rather than from static config option lists.
var dbBackedRefs = map[string]bool{
	"users": true,
}

// validateFieldConfigs checks that field configs are valid and that
// options_ref values resolve to known option lists.
func validateFieldConfigs(cfg *TenantConfig, fields []FieldConfig, context string) error {
	validTypes := map[string]bool{
		"text": true, "select": true, "multi_select": true,
		"relation": true, "date": true,
	}

	for _, f := range fields {
		if f.Key == "" {
			return fmt.Errorf("%s: field key must not be empty", context)
		}
		if !validTypes[f.Type] {
			return fmt.Errorf("%s: field %q has invalid type %q", context, f.Key, f.Type)
		}
		if f.OptionsRef != "" && !dbBackedRefs[f.OptionsRef] && cfg.ResolveOptions(f.OptionsRef) == nil {
			return fmt.Errorf("%s: field %q references unknown options_ref %q", context, f.Key, f.OptionsRef)
		}
	}
	return nil
}
