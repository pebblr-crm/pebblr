package postgres

// nullIfEmpty converts an empty string to nil (for nullable TEXT columns).
func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
