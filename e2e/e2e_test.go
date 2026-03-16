//go:build e2e

// Package e2e contains end-to-end tests that run against a live Kubernetes cluster.
// Tests in this package require a running Kind cluster with the pebblr chart installed.
//
// Run with: make e2e
package e2e

import "testing"

// TestPlaceholder is a placeholder for future E2E tests.
// Replace or extend with real test cases as the suite is built out.
func TestPlaceholder(t *testing.T) {
	t.Skip("placeholder: no E2E tests implemented yet")
}
