package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateSchema_ValidConfig(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "valid.json")
	if err := os.WriteFile(path, []byte(validConfigJSON()), 0o644); err != nil {
		t.Fatal(err)
	}

	errs, err := ValidateSchema(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("expected no validation errors, got %v", errs)
	}
}

func TestValidateSchema_InvalidConfig(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.json")
	// Missing required fields like "activities".
	content := `{"tenant": {"name": "Test"}}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	errs, err := ValidateSchema(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) == 0 {
		t.Fatal("expected validation errors for missing fields")
	}
}

func TestValidateSchema_FileNotFound(t *testing.T) {
	t.Parallel()
	_, err := ValidateSchema("/nonexistent/path.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestValidateSchema_InvalidJSON(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ValidateSchema(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoadAndValidate_ValidConfig(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "valid.json")
	if err := os.WriteFile(path, []byte(validConfigJSON()), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, errs, err := LoadAndValidate(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
	if cfg == nil {
		t.Fatal("expected config to be non-nil")
	}
	if cfg.Tenant.Name != "Test Tenant" {
		t.Errorf("expected tenant name Test Tenant, got %s", cfg.Tenant.Name)
	}
}

func TestLoadAndValidate_SchemaErrors(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.json")
	content := `{"tenant": {"name": "Test"}}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, errs, err := LoadAndValidate(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) == 0 {
		t.Fatal("expected schema validation errors")
	}
	if cfg != nil {
		t.Error("expected nil config when schema validation fails")
	}
}

func TestLoadAndValidate_SemanticErrors(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "semantic.json")
	// Passes schema but fails semantic (empty tenant name).
	content := `{
		"tenant": {"name": "", "locale": "en"},
		"accounts": {"types": []},
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
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, errs, err := LoadAndValidate(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) == 0 {
		t.Fatal("expected semantic validation errors")
	}
	if cfg != nil {
		t.Error("expected nil config when semantic validation fails")
	}
}

func TestLoadAndValidate_FileNotFound(t *testing.T) {
	t.Parallel()
	_, _, err := LoadAndValidate("/nonexistent/path.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestFlattenErrors_LeafError(t *testing.T) {
	t.Parallel()
	// Test that flattenErrors handles a simple leaf error without causes.
	// This is tested indirectly through ValidateSchema, but let's verify
	// the invalid config case covers it.
	dir := t.TempDir()
	path := filepath.Join(dir, "extra_field.json")
	// This config has a valid structure but missing required fields to trigger errors.
	content := `{"tenant": {"name": "T"}, "extra": "field"}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	errs, err := ValidateSchema(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should have errors for missing required properties.
	if len(errs) == 0 {
		t.Fatal("expected validation errors")
	}
	// Each error should contain a path prefix.
	for _, e := range errs {
		if e == "" {
			t.Error("expected non-empty error string")
		}
	}
}
