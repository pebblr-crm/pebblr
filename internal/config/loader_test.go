package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_ValidConfig(t *testing.T) {
	t.Parallel()
	cfg := writeTestConfig(t, validConfigJSON())
	if cfg.Tenant.Name != "Test Tenant" {
		t.Fatalf("expected tenant name %q, got %q", "Test Tenant", cfg.Tenant.Name)
	}
	if len(cfg.Accounts.Types) != 1 {
		t.Fatalf("expected 1 account type, got %d", len(cfg.Accounts.Types))
	}
	if len(cfg.Activities.Types) != 2 {
		t.Fatalf("expected 2 activity types, got %d", len(cfg.Activities.Types))
	}
	if cfg.InitialStatus() != "planned" {
		t.Fatalf("expected initial status %q, got %q", "planned", cfg.InitialStatus())
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	t.Parallel()
	_, err := Load("/nonexistent/path.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoad_MissingTenantName(t *testing.T) {
	t.Parallel()
	j := validConfigJSON()
	j = replaceInJSON(j, `"name": "Test Tenant"`, `"name": ""`)
	_, err := loadFromString(t, j)
	if err == nil {
		t.Fatal("expected error for empty tenant name")
	}
}

func TestLoad_NoInitialStatus(t *testing.T) {
	t.Parallel()
	j := validConfigJSON()
	j = replaceInJSON(j, `"initial": true`, `"initial": false`)
	_, err := loadFromString(t, j)
	if err == nil {
		t.Fatal("expected error for no initial status")
	}
}

func TestLoad_DuplicateStatusKey(t *testing.T) {
	t.Parallel()
	j := `{
		"tenant": {"name": "T", "locale": "en"},
		"accounts": {"types": []},
		"activities": {
			"statuses": [
				{"key": "a", "label": "A", "initial": true},
				{"key": "a", "label": "A2"}
			],
			"status_transitions": {},
			"durations": [],
			"types": [],
			"routing_options": []
		},
		"options": {},
		"rules": {"frequency": {}, "max_activities_per_day": 10, "default_visit_duration_minutes": {}, "visit_duration_step_minutes": 15}
	}`
	_, err := loadFromString(t, j)
	if err == nil {
		t.Fatal("expected error for duplicate status key")
	}
}

func TestLoad_BadStatusTransitionRef(t *testing.T) {
	t.Parallel()
	j := validConfigJSON()
	j = replaceInJSON(j, `"planned": ["completed", "cancelled"]`, `"planned": ["nonexistent"]`)
	_, err := loadFromString(t, j)
	if err == nil {
		t.Fatal("expected error for unknown transition target")
	}
}

func TestLoad_BadOptionsRef(t *testing.T) {
	t.Parallel()
	j := `{
		"tenant": {"name": "T", "locale": "en"},
		"accounts": {"types": [
			{"key": "doc", "label": "Doc", "fields": [
				{"key": "spec", "type": "select", "required": false, "options_ref": "nonexistent"}
			]}
		]},
		"activities": {
			"statuses": [{"key": "a", "label": "A", "initial": true}],
			"status_transitions": {},
			"durations": [],
			"types": [],
			"routing_options": []
		},
		"options": {},
		"rules": {"frequency": {}, "max_activities_per_day": 10, "default_visit_duration_minutes": {}, "visit_duration_step_minutes": 15}
	}`
	_, err := loadFromString(t, j)
	if err == nil {
		t.Fatal("expected error for unresolvable options_ref")
	}
}

func TestLoad_InvalidFieldType(t *testing.T) {
	t.Parallel()
	j := `{
		"tenant": {"name": "T", "locale": "en"},
		"accounts": {"types": [
			{"key": "doc", "label": "Doc", "fields": [
				{"key": "f", "type": "bogus", "required": false}
			]}
		]},
		"activities": {
			"statuses": [{"key": "a", "label": "A", "initial": true}],
			"status_transitions": {},
			"durations": [],
			"types": [],
			"routing_options": []
		},
		"options": {},
		"rules": {"frequency": {}, "max_activities_per_day": 10, "default_visit_duration_minutes": {}, "visit_duration_step_minutes": 15}
	}`
	_, err := loadFromString(t, j)
	if err == nil {
		t.Fatal("expected error for invalid field type")
	}
}

func TestLoad_InvalidActivityCategory(t *testing.T) {
	t.Parallel()
	j := `{
		"tenant": {"name": "T", "locale": "en"},
		"accounts": {"types": []},
		"activities": {
			"statuses": [{"key": "a", "label": "A", "initial": true}],
			"status_transitions": {},
			"durations": [],
			"types": [{"key": "bad", "label": "Bad", "category": "invalid", "fields": []}],
			"routing_options": []
		},
		"options": {},
		"rules": {"frequency": {}, "max_activities_per_day": 10, "default_visit_duration_minutes": {}, "visit_duration_step_minutes": 15}
	}`
	_, err := loadFromString(t, j)
	if err == nil {
		t.Fatal("expected error for invalid category")
	}
}

func TestLoad_SubmitRequiredUnknownField(t *testing.T) {
	t.Parallel()
	j := `{
		"tenant": {"name": "T", "locale": "en"},
		"accounts": {"types": []},
		"activities": {
			"statuses": [{"key": "a", "label": "A", "initial": true}],
			"status_transitions": {},
			"durations": [],
			"types": [{
				"key": "visit", "label": "Visit", "category": "field",
				"fields": [{"key": "f1", "type": "text", "required": false}],
				"submit_required": ["nonexistent"]
			}],
			"routing_options": []
		},
		"options": {},
		"rules": {"frequency": {}, "max_activities_per_day": 10, "default_visit_duration_minutes": {}, "visit_duration_step_minutes": 15}
	}`
	_, err := loadFromString(t, j)
	if err == nil {
		t.Fatal("expected error for submit_required referencing unknown field")
	}
}

func TestLoad_SampleConfig(t *testing.T) {
	t.Parallel()
	path := filepath.Join("..", "..", "config", "tenant.json")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("sample config should be valid: %v", err)
	}
	if cfg.Tenant.Name != "Pillzilla Pharmaceuticals" {
		t.Fatalf("expected %q, got %q", "Pillzilla Pharmaceuticals", cfg.Tenant.Name)
	}
	if len(cfg.Activities.Types) != 11 {
		t.Fatalf("expected 11 activity types, got %d", len(cfg.Activities.Types))
	}
}

// --- helpers ---

func validConfigJSON() string {
	return `{
		"tenant": {"name": "Test Tenant", "locale": "en"},
		"accounts": {"types": [
			{"key": "doctor", "label": "Doctor", "fields": [
				{"key": "name", "type": "text", "required": true},
				{"key": "specialty", "type": "select", "required": false, "options_ref": "specialties"}
			]}
		]},
		"activities": {
			"statuses": [
				{"key": "planned", "label": "Planned", "initial": true},
				{"key": "completed", "label": "Completed"},
				{"key": "cancelled", "label": "Cancelled"}
			],
			"status_transitions": {
				"planned": ["completed", "cancelled"],
				"completed": [],
				"cancelled": []
			},
			"durations": [
				{"key": "full_day", "label": "Full Day"},
				{"key": "half_day", "label": "Half Day"}
			],
			"types": [
				{
					"key": "visit", "label": "Visit", "category": "field",
					"fields": [
						{"key": "account_id", "type": "relation", "required": true},
						{"key": "feedback", "type": "text", "required": false},
						{"key": "duration", "type": "select", "required": true, "options_ref": "durations"}
					],
					"submit_required": ["feedback"]
				},
				{
					"key": "admin", "label": "Admin", "category": "non_field",
					"fields": [
						{"key": "duration", "type": "select", "required": true, "options_ref": "durations"}
					]
				}
			],
			"routing_options": [
				{"key": "week_1", "label": "Week 1"}
			]
		},
		"options": {
			"specialties": [
				{"key": "cardiology", "label": "Cardiology"},
				{"key": "neurology", "label": "Neurology"}
			]
		},
		"rules": {
			"frequency": {"a": 4, "b": 2, "c": 1},
			"max_activities_per_day": 10,
			"default_visit_duration_minutes": {"doctor": 30},
			"visit_duration_step_minutes": 15
		}
	}`
}

func writeTestConfig(t *testing.T, json string) *TenantConfig {
	t.Helper()
	cfg, err := loadFromString(t, json)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return cfg
}

func loadFromString(t *testing.T, content string) (*TenantConfig, error) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "tenant.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return Load(path)
}

func replaceInJSON(json, old, replacement string) string {
	return strings.Replace(json, old, replacement, 1)
}
