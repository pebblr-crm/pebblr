// Package rbac implements row-level access control for the pebblr CRM.
// Access is enforced at the data layer: reps see only their own leads,
// managers see their team's leads, admins see everything.
package rbac
