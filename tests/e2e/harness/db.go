package harness

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

// Count runs a SELECT count(*) style query and returns the number.
func (e *Env) Count(t *testing.T, sql string, args ...any) int {
	t.Helper()
	var n int
	if err := e.DB.QueryRow(context.Background(), sql, args...).Scan(&n); err != nil {
		t.Fatalf("count %q: %v", sql, err)
	}
	return n
}

// Row runs a single-row query; the caller scans it.
func (e *Env) Row(sql string, args ...any) pgx.Row {
	return e.DB.QueryRow(context.Background(), sql, args...)
}

// Scan runs a single-row query into dest and fails the test when it is missing.
func (e *Env) Scan(t *testing.T, sql string, args []any, dest ...any) {
	t.Helper()
	if err := e.DB.QueryRow(context.Background(), sql, args...).Scan(dest...); err != nil {
		t.Fatalf("query %q: %v", sql, err)
	}
}
