package domain

import (
	"fmt"
	"time"
)

// Territory represents a geographic region assigned to a team for coverage tracking.
type Territory struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	TeamID    string         `json:"teamId"`
	Region    string         `json:"region,omitempty"`
	Boundary  map[string]any `json:"boundary,omitempty"` // GeoJSON geometry
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

// ValidGeoJSONTypes lists the GeoJSON geometry types accepted for boundaries.
var ValidGeoJSONTypes = map[string]bool{
	"Point":              true,
	"MultiPoint":         true,
	"LineString":         true,
	"MultiLineString":    true,
	"Polygon":            true,
	"MultiPolygon":       true,
	"GeometryCollection": true,
}

// ValidateBoundary checks that Boundary, if present, looks like a valid
// GeoJSON geometry object (has "type" and "coordinates" keys with sane values).
// It does not validate coordinate arrays in depth.
func (t *Territory) ValidateBoundary() error {
	if len(t.Boundary) == 0 {
		return nil // boundary is optional
	}

	typeVal, ok := t.Boundary["type"]
	if !ok {
		return fmt.Errorf("boundary: missing \"type\" field")
	}
	typeStr, ok := typeVal.(string)
	if !ok {
		return fmt.Errorf("boundary: \"type\" must be a string")
	}
	if !ValidGeoJSONTypes[typeStr] {
		return fmt.Errorf("boundary: unknown GeoJSON type %q", typeStr)
	}

	// GeometryCollection uses "geometries" instead of "coordinates".
	if typeStr == "GeometryCollection" {
		if _, ok := t.Boundary["geometries"]; !ok {
			return fmt.Errorf("boundary: GeometryCollection requires \"geometries\" field")
		}
		return nil
	}

	if _, ok := t.Boundary["coordinates"]; !ok {
		return fmt.Errorf("boundary: missing \"coordinates\" field")
	}
	return nil
}
