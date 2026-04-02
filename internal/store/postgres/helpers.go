package postgres

import "strings"

// nullIfEmpty converts an empty string to nil (for nullable TEXT columns).
func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// escapeILIKE escapes the special LIKE/ILIKE wildcard characters (%, _, \)
// in user-supplied input so they are matched literally.
func escapeILIKE(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}
