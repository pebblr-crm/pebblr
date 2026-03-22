package service

import "errors"

// Sentinel errors returned by service methods.
var (
	// ErrForbidden indicates the caller lacks permission for the requested operation.
	ErrForbidden = errors.New("forbidden")

	// ErrInvalidInput indicates the request contains invalid data.
	ErrInvalidInput = errors.New("invalid input")
)
