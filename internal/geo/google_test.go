package geo

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGoogleGeocoder_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := googleResponse{
			Status: "OK",
			Results: []googleResult{
				{
					FormattedAddress: "Bucharest, Romania",
					Geometry: googleGeometry{
						Location: googleLatLng{Lat: 44.4268, Lng: 26.1025},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	g := &GoogleGeocoder{
		apiKey:     "test-key",
		httpClient: srv.Client(),
		baseURL:    srv.URL,
	}

	result, err := g.Geocode(context.Background(), "Bucharest")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Lat != 44.4268 {
		t.Errorf("expected lat 44.4268, got %f", result.Lat)
	}
	if result.Lng != 26.1025 {
		t.Errorf("expected lng 26.1025, got %f", result.Lng)
	}
	if result.FormattedAddress != "Bucharest, Romania" {
		t.Errorf("expected formatted address Bucharest, Romania, got %s", result.FormattedAddress)
	}
}

func TestGoogleGeocoder_NoResults(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := googleResponse{
			Status:  "ZERO_RESULTS",
			Results: []googleResult{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	g := &GoogleGeocoder{
		apiKey:     "test-key",
		httpClient: srv.Client(),
		baseURL:    srv.URL,
	}

	_, err := g.Geocode(context.Background(), "nonexistent address xyz123")
	if !errors.Is(err, ErrNoResults) {
		t.Fatalf("expected ErrNoResults, got %v", err)
	}
}

func TestGoogleGeocoder_HTTPError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	g := &GoogleGeocoder{
		apiKey:     "test-key",
		httpClient: srv.Client(),
		baseURL:    srv.URL,
	}

	_, err := g.Geocode(context.Background(), "Bucharest")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestGoogleGeocoder_InvalidJSON(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{not json"))
	}))
	defer srv.Close()

	g := &GoogleGeocoder{
		apiKey:     "test-key",
		httpClient: srv.Client(),
		baseURL:    srv.URL,
	}

	_, err := g.Geocode(context.Background(), "Bucharest")
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
}

func TestGoogleGeocoder_ConnectionError(t *testing.T) {
	t.Parallel()
	g := &GoogleGeocoder{
		apiKey:     "test-key",
		httpClient: http.DefaultClient,
		baseURL:    "http://localhost:1", // unreachable port
	}

	_, err := g.Geocode(context.Background(), "Bucharest")
	if err == nil {
		t.Fatal("expected error for connection failure")
	}
}

func TestNewGoogleGeocoder(t *testing.T) {
	t.Parallel()
	g := NewGoogleGeocoder("my-api-key")
	if g.apiKey != "my-api-key" {
		t.Errorf("expected apiKey my-api-key, got %s", g.apiKey)
	}
	if g.httpClient == nil {
		t.Error("expected non-nil httpClient")
	}
	if g.baseURL == "" {
		t.Error("expected non-empty baseURL")
	}
}
