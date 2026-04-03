package config

import "testing"

func newTestConfig() *TenantConfig {
	cfg := &TenantConfig{
		Tenant: TenantInfo{Name: "Test", Locale: "en"},
		Accounts: AccountsConfig{
			Types: []AccountTypeConfig{
				{Key: "doctor", Label: "Doctor"},
				{Key: "pharmacy", Label: "Pharmacy"},
			},
		},
		Activities: ActivitiesConfig{
			Statuses: []StatusDef{
				{Key: "planned", Label: "Planned", Initial: true},
				{Key: "completed", Label: "Completed", Submittable: true},
				{Key: "cancelled", Label: "Cancelled"},
			},
			StatusTransitions: map[string][]string{
				"planned":   {"completed", "cancelled"},
				"completed": {},
				"cancelled": {},
			},
			Durations: []OptionDef{
				{Key: "full_day", Label: "Full Day"},
				{Key: "half_day", Label: "Half Day"},
			},
			Types: []ActivityTypeConfig{
				{Key: "visit", Label: "Visit", Category: CategoryField, BlocksFieldActivities: false},
				{Key: "admin", Label: "Admin", Category: CategoryNonField, BlocksFieldActivities: false},
				{Key: "vacation", Label: "Vacation", Category: CategoryNonField, BlocksFieldActivities: true},
			},
		},
		Options: map[string][]OptionDef{
			"specialties": {{Key: "cardiology", Label: "Cardiology"}},
		},
	}
	cfg.buildIndexes()
	return cfg
}

func TestBuildIndexes(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig()

	if cfg.accountTypeIndex == nil {
		t.Fatal("expected accountTypeIndex to be built")
	}
	if cfg.activityTypeIndex == nil {
		t.Fatal("expected activityTypeIndex to be built")
	}
	if cfg.statusIndex == nil {
		t.Fatal("expected statusIndex to be built")
	}
	if len(cfg.accountTypeIndex) != 2 {
		t.Errorf("expected 2 account types, got %d", len(cfg.accountTypeIndex))
	}
}

func TestAccountType_WithIndex(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig()

	at := cfg.AccountType("doctor")
	if at == nil {
		t.Fatal("expected doctor account type")
	}
	if at.Key != "doctor" {
		t.Errorf("expected key doctor, got %s", at.Key)
	}

	if cfg.AccountType("unknown") != nil {
		t.Error("expected nil for unknown key")
	}
}

func TestAccountType_WithoutIndex(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig()
	cfg.accountTypeIndex = nil // Force fallback path

	at := cfg.AccountType("doctor")
	if at == nil {
		t.Fatal("expected doctor account type via fallback")
	}

	if cfg.AccountType("unknown") != nil {
		t.Error("expected nil for unknown key via fallback")
	}
}

func TestActivityType_WithIndex(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig()

	at := cfg.ActivityType("visit")
	if at == nil {
		t.Fatal("expected visit activity type")
	}

	if cfg.ActivityType("unknown") != nil {
		t.Error("expected nil for unknown key")
	}
}

func TestActivityType_WithoutIndex(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig()
	cfg.activityTypeIndex = nil

	at := cfg.ActivityType("visit")
	if at == nil {
		t.Fatal("expected visit activity type via fallback")
	}

	if cfg.ActivityType("unknown") != nil {
		t.Error("expected nil for unknown key via fallback")
	}
}

func TestInitialStatus(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig()

	if s := cfg.InitialStatus(); s != "planned" {
		t.Errorf("expected planned, got %s", s)
	}

	// No initial status
	cfg2 := &TenantConfig{
		Activities: ActivitiesConfig{
			Statuses: []StatusDef{
				{Key: "completed", Label: "Completed"},
			},
		},
	}
	if s := cfg2.InitialStatus(); s != "" {
		t.Errorf("expected empty string, got %s", s)
	}
}

func TestIsValidStatus_WithIndex(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig()

	if !cfg.IsValidStatus("planned") {
		t.Error("expected planned to be valid")
	}
	if cfg.IsValidStatus("unknown") {
		t.Error("expected unknown to be invalid")
	}
}

func TestIsValidStatus_WithoutIndex(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig()
	cfg.statusIndex = nil

	if !cfg.IsValidStatus("planned") {
		t.Error("expected planned to be valid via fallback")
	}
	if cfg.IsValidStatus("unknown") {
		t.Error("expected unknown to be invalid via fallback")
	}
}

func TestIsSubmittableStatus_WithIndex(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig()

	if !cfg.IsSubmittableStatus("completed") {
		t.Error("expected completed to be submittable")
	}
	if cfg.IsSubmittableStatus("planned") {
		t.Error("expected planned to not be submittable")
	}
	if cfg.IsSubmittableStatus("unknown") {
		t.Error("expected unknown to not be submittable")
	}
}

func TestIsSubmittableStatus_WithoutIndex(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig()
	cfg.statusIndex = nil

	if !cfg.IsSubmittableStatus("completed") {
		t.Error("expected completed to be submittable via fallback")
	}
	if cfg.IsSubmittableStatus("planned") {
		t.Error("expected planned to not be submittable via fallback")
	}
	if cfg.IsSubmittableStatus("unknown") {
		t.Error("expected unknown to not be submittable via fallback")
	}
}

func TestIsValidTransition(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig()

	if !cfg.IsValidTransition("planned", "completed") {
		t.Error("expected planned->completed to be valid")
	}
	if cfg.IsValidTransition("planned", "unknown") {
		t.Error("expected planned->unknown to be invalid")
	}
	if cfg.IsValidTransition("unknown", "completed") {
		t.Error("expected unknown->completed to be invalid")
	}
	if cfg.IsValidTransition("completed", "planned") {
		t.Error("expected completed->planned to be invalid (empty list)")
	}
}

func TestFieldActivityTypes(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig()

	types := cfg.FieldActivityTypes()
	if len(types) != 1 {
		t.Fatalf("expected 1 field type, got %d", len(types))
	}
	if types[0] != "visit" {
		t.Errorf("expected visit, got %s", types[0])
	}
}

func TestBlockingActivityTypes(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig()

	types := cfg.BlockingActivityTypes()
	if len(types) != 1 {
		t.Fatalf("expected 1 blocking type, got %d", len(types))
	}
	if types[0] != "vacation" {
		t.Errorf("expected vacation, got %s", types[0])
	}
}

func TestResolveOptions(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig()

	// Resolve from options map
	opts := cfg.ResolveOptions("specialties")
	if len(opts) != 1 {
		t.Fatalf("expected 1 option, got %d", len(opts))
	}
	if opts[0].Key != "cardiology" {
		t.Errorf("expected cardiology, got %s", opts[0].Key)
	}

	// Resolve durations special case
	opts = cfg.ResolveOptions("durations")
	if len(opts) != 2 {
		t.Fatalf("expected 2 durations, got %d", len(opts))
	}

	// Unknown ref returns nil
	opts = cfg.ResolveOptions("nonexistent")
	if opts != nil {
		t.Error("expected nil for unknown ref")
	}
}
