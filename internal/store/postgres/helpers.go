package postgres

import (
	"encoding/json"
	"strings"
)

// nullIfEmpty converts an empty string to nil (for nullable TEXT columns).
func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// nullJSONIfNil returns nil if the map is nil (stores NULL in db), otherwise the marshalled JSON.
func nullJSONIfNil(m map[string]any, marshalled []byte) []byte {
	if m == nil {
		return nil
	}
	return marshalled
}

// marshalJSONField marshals a map to JSON, returning nil for nil/empty maps.
func marshalJSONField(m map[string]any) ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// escapeILIKE escapes the special LIKE/ILIKE wildcard characters (%, _, \)
// in user-supplied input so they are matched literally.
func escapeILIKE(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}
