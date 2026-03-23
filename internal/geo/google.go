package geo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// GoogleGeocoder implements Geocoder using the Google Maps Geocoding API.
type GoogleGeocoder struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string // overridable for testing
}

// NewGoogleGeocoder returns a Geocoder backed by the Google Geocoding API.
func NewGoogleGeocoder(apiKey string) *GoogleGeocoder {
	return &GoogleGeocoder{
		apiKey:     apiKey,
		httpClient: http.DefaultClient,
		baseURL:    "https://maps.googleapis.com/maps/api/geocode/json",
	}
}

type googleResponse struct {
	Status  string         `json:"status"`
	Results []googleResult `json:"results"`
}

type googleResult struct {
	FormattedAddress string         `json:"formatted_address"`
	Geometry         googleGeometry `json:"geometry"`
}

type googleGeometry struct {
	Location googleLatLng `json:"location"`
}

type googleLatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func (g *GoogleGeocoder) Geocode(ctx context.Context, address string) (*Result, error) {
	u := fmt.Sprintf("%s?address=%s&key=%s",
		g.baseURL,
		url.QueryEscape(address),
		url.QueryEscape(g.apiKey),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("creating geocode request: %w", err)
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("geocoding request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocoding API returned status %d", resp.StatusCode)
	}

	var gResp googleResponse
	if err := json.NewDecoder(resp.Body).Decode(&gResp); err != nil {
		return nil, fmt.Errorf("decoding geocode response: %w", err)
	}

	if len(gResp.Results) == 0 {
		return nil, ErrNoResults
	}

	r := gResp.Results[0]
	return &Result{
		Lat:              r.Geometry.Location.Lat,
		Lng:              r.Geometry.Location.Lng,
		FormattedAddress: r.FormattedAddress,
	}, nil
}
