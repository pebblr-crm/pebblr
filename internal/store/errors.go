package store

import "errors"

// ErrNotFound is returned when a requested record does not exist.
var ErrNotFound = errors.New("record not found")

// ErrConflict is returned when a create/update violates a uniqueness constraint.
var ErrConflict = errors.New("record conflict")
