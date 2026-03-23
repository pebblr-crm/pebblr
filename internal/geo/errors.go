package geo

import "errors"

// ErrNoResults indicates that the geocoder could not resolve the address.
var ErrNoResults = errors.New("no geocoding results for address")
