package config

import "testing"

func testConfig(t *testing.T) *TenantConfig {
	t.Helper()
	return writeTestConfig(t, validConfigJSON())
}

func TestValidateActivity_ValidSave(t *testing.T) {
	t.Parallel()
	cfg := testConfig(t)
	fields := map[string]any{
		"account_id": "some-uuid",
		"duration":   "full_day",
	}
	errs := ValidateActivity(cfg, "visit", fields, "save")
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
}

func TestValidateActivity_MissingRequiredField(t *testing.T) {
	t.Parallel()
	cfg := testConfig(t)
	fields := map[string]any{
		"duration": "full_day",
	}
	errs := ValidateActivity(cfg, "visit", fields, "save")
	if len(errs) != 1 || errs[0].Field != "account_id" {
		t.Fatalf("expected 1 error on account_id, got %v", errs)
	}
}

func TestValidateActivity_SubmitEnforcesExtraRequired(t *testing.T) {
	t.Parallel()
	cfg := testConfig(t)
	fields := map[string]any{
		"account_id": "some-uuid",
		"duration":   "full_day",
	}

	errs := ValidateActivity(cfg, "visit", fields, "save")
	if len(errs) != 0 {
		t.Fatalf("save: expected no errors, got %v", errs)
	}

	errs = ValidateActivity(cfg, "visit", fields, "submit")
	if len(errs) != 1 || errs[0].Field != "feedback" {
		t.Fatalf("submit: expected 1 error on feedback, got %v", errs)
	}
}

func TestValidateActivity_UnknownActivityType(t *testing.T) {
	t.Parallel()
	cfg := testConfig(t)
	errs := ValidateActivity(cfg, "nonexistent", nil, "save")
	if len(errs) != 1 || errs[0].Field != "activity_type" {
		t.Fatalf("expected error on activity_type, got %v", errs)
	}
}

func TestValidateActivity_InvalidSelectOption(t *testing.T) {
	t.Parallel()
	cfg := testConfig(t)
	fields := map[string]any{
		"account_id": "some-uuid",
		"duration":   "invalid_duration",
	}
	errs := ValidateActivity(cfg, "visit", fields, "save")
	if len(errs) != 1 || errs[0].Field != "duration" {
		t.Fatalf("expected error on duration, got %v", errs)
	}
}

func TestValidateActivity_ValidSelectOption(t *testing.T) {
	t.Parallel()
	cfg := testConfig(t)
	fields := map[string]any{
		"account_id": "some-uuid",
		"duration":   "half_day",
	}
	errs := ValidateActivity(cfg, "visit", fields, "save")
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
}

func TestValidateActivity_InvalidMultiSelectOption(t *testing.T) {
	t.Parallel()
	cfg := testConfig(t)
	visitType := cfg.ActivityType("visit")
	visitType.Fields = append(visitType.Fields, FieldConfig{
		Key:        "products",
		Type:       "multi_select",
		Required:   false,
		OptionsRef: "specialties",
	})

	fields := map[string]any{
		"account_id": "some-uuid",
		"duration":   "full_day",
		"products":   []any{"cardiology", "bogus"},
	}
	errs := ValidateActivity(cfg, "visit", fields, "save")
	if len(errs) != 1 || errs[0].Field != "products" {
		t.Fatalf("expected error on products, got %v", errs)
	}
}

func TestValidateActivity_ValidMultiSelect(t *testing.T) {
	t.Parallel()
	cfg := testConfig(t)
	visitType := cfg.ActivityType("visit")
	visitType.Fields = append(visitType.Fields, FieldConfig{
		Key:        "products",
		Type:       "multi_select",
		Required:   false,
		OptionsRef: "specialties",
	})

	fields := map[string]any{
		"account_id": "some-uuid",
		"duration":   "full_day",
		"products":   []any{"cardiology", "neurology"},
	}
	errs := ValidateActivity(cfg, "visit", fields, "save")
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
}

func TestValidateStatus(t *testing.T) {
	t.Parallel()
	cfg := testConfig(t)

	if err := ValidateStatus(cfg, "planned"); err != nil {
		t.Fatalf("expected valid status, got %v", err)
	}
	if err := ValidateStatus(cfg, "nonexistent"); err == nil {
		t.Fatal("expected error for unknown status")
	}
}

func TestValidateStatusTransition(t *testing.T) {
	t.Parallel()
	cfg := testConfig(t)

	if err := ValidateStatusTransition(cfg, "planned", "completed"); err != nil {
		t.Fatalf("expected valid transition, got %v", err)
	}
	if err := ValidateStatusTransition(cfg, "completed", "planned"); err == nil {
		t.Fatal("expected error for disallowed transition")
	}
	if err := ValidateStatusTransition(cfg, "nonexistent", "completed"); err == nil {
		t.Fatal("expected error for unknown source status")
	}
}

func TestValidateDuration(t *testing.T) {
	t.Parallel()
	cfg := testConfig(t)

	if err := ValidateDuration(cfg, "full_day"); err != nil {
		t.Fatalf("expected valid duration, got %v", err)
	}
	if err := ValidateDuration(cfg, "bogus"); err == nil {
		t.Fatal("expected error for unknown duration")
	}
}

func TestTenantConfig_AccountType(t *testing.T) {
	t.Parallel()
	cfg := testConfig(t)

	if at := cfg.AccountType("doctor"); at == nil {
		t.Fatal("expected doctor account type")
	}
	if at := cfg.AccountType("nonexistent"); at != nil {
		t.Fatal("expected nil for unknown account type")
	}
}

func TestTenantConfig_ActivityType(t *testing.T) {
	t.Parallel()
	cfg := testConfig(t)

	if at := cfg.ActivityType("visit"); at == nil {
		t.Fatal("expected visit activity type")
	}
	if at := cfg.ActivityType("admin"); at == nil {
		t.Fatal("expected admin activity type")
	}
	if at := cfg.ActivityType("nonexistent"); at != nil {
		t.Fatal("expected nil for unknown activity type")
	}
}

func TestTenantConfig_ResolveOptions(t *testing.T) {
	t.Parallel()
	cfg := testConfig(t)

	opts := cfg.ResolveOptions("specialties")
	if len(opts) != 2 {
		t.Fatalf("expected 2 specialties, got %d", len(opts))
	}

	opts = cfg.ResolveOptions("durations")
	if len(opts) != 2 {
		t.Fatalf("expected 2 durations, got %d", len(opts))
	}

	opts = cfg.ResolveOptions("nonexistent")
	if opts != nil {
		t.Fatal("expected nil for unknown options ref")
	}
}
