package postgres

import (
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

// newMockPool creates a pgxmock pool for tests. It registers a cleanup to verify
// that all expectations were met.
func newMockPool(t *testing.T) pgxmock.PgxPoolIface {
	t.Helper()
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("creating pgxmock pool: %v", err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet pgxmock expectations: %v", err)
		}
	})
	return mock
}

// anyArgs returns a slice of n pgxmock.AnyArg() values for use in WithArgs.
func anyArgs(n int) []any {
	args := make([]any, n)
	for i := range args {
		args[i] = pgxmock.AnyArg()
	}
	return args
}

// testTime returns a fixed time for deterministic tests.
func testTime() time.Time {
	return time.Date(2025, 3, 15, 10, 0, 0, 0, time.UTC)
}

// strPtr returns a pointer to a string.
func strPtr(s string) *string {
	return &s
}

