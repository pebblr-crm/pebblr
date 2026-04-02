package postgres

import "testing"

func TestNullIfEmpty(t *testing.T) {
	t.Parallel()
	if nullIfEmpty("") != nil {
		t.Error("expected nil for empty string")
	}
	p := nullIfEmpty("abc")
	if p == nil || *p != "abc" {
		t.Errorf("expected pointer to 'abc', got %v", p)
	}
}

func TestEscapeILIKE(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no special chars", "hello", "hello"},
		{"percent", "100%", `100\%`},
		{"underscore", "a_b", `a\_b`},
		{"backslash", `a\b`, `a\\b`},
		{"all special", `%_\`, `\%\_\\`},
		{"mixed", `foo%bar_baz\qux`, `foo\%bar\_baz\\qux`},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := escapeILIKE(tt.input)
			if got != tt.want {
				t.Errorf("escapeILIKE(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
