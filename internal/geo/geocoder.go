// Package geo provides geocoding services for converting addresses to coordinates.
package geo

import "context"

// Result holds the geocoding output for a single address.
type Result struct {
	Lat              float64
	Lng              float64
	FormattedAddress string
}

// Geocoder converts a street address into geographic coordinates.
type Geocoder interface {
	// Geocode returns the lat/lng for the given address, or an error.
	// Implementations should return ErrNoResults when the address cannot be resolved.
	Geocode(ctx context.Context, address string) (*Result, error)
}
