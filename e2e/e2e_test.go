//go:build e2e

// Package e2e contains end-to-end tests that run against a live Kubernetes cluster.
// Tests in this package require a running Kind cluster with the pebblr chart installed.
//
// Run with: make e2e
package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

const (
	defaultNamespace = "pebblr-e2e"
	defaultService   = "svc/pebblr-e2e"
	servicePort      = "8080"
)

// testEnv holds the port-forward connection for the test suite.
type testEnv struct {
	baseURL string
	cancel  context.CancelFunc
	cmd     *exec.Cmd
}

var env *testEnv

func TestMain(m *testing.M) {
	var err error
	env, err = setupPortForward()
	if err != nil {
		fmt.Fprintf(os.Stderr, "e2e setup failed: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	env.teardown()
	os.Exit(code)
}

func setupPortForward() (*testEnv, error) {
	ns := os.Getenv("E2E_NAMESPACE")
	if ns == "" {
		ns = defaultNamespace
	}
	svc := os.Getenv("E2E_SERVICE")
	if svc == "" {
		svc = defaultService
	}

	// If E2E_BASE_URL is set, skip port-forward (e.g., for CI with a fixed endpoint).
	if base := os.Getenv("E2E_BASE_URL"); base != "" {
		return &testEnv{baseURL: strings.TrimRight(base, "/"), cancel: func() {}}, nil
	}

	// Find a free local port.
	localPort, err := freePort()
	if err != nil {
		return nil, fmt.Errorf("finding free port: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	portArg := fmt.Sprintf("%d:%s", localPort, servicePort)
	cmd := exec.CommandContext(ctx, "kubectl", "port-forward", svc, portArg, "-n", ns)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.WaitDelay = 2 * time.Second

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("starting port-forward: %w", err)
	}

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", localPort)

	// Wait for the port-forward to be ready.
	if err := waitForReady(baseURL, 30*time.Second); err != nil {
		cancel()
		return nil, fmt.Errorf("waiting for port-forward: %w", err)
	}

	return &testEnv{baseURL: baseURL, cancel: cancel, cmd: cmd}, nil
}

func (e *testEnv) teardown() {
	e.cancel()
	if e.cmd != nil && e.cmd.Process != nil {
		_ = e.cmd.Process.Kill()
		_ = e.cmd.Wait()
	}
}

func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// waitForReady polls the health endpoint until it returns 200 or times out.
func waitForReady(baseURL string, timeout time.Duration) error {
	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		req, _ := http.NewRequest("GET", baseURL+"/api/v1/health", nil)
		req.Header.Set("Authorization", "Bearer e2e-test-token")
		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("service not ready after %v", timeout)
}

// doRequest is a test helper that makes an HTTP request to the API.
func doRequest(t *testing.T, method, path string, body string, headers map[string]string) *http.Response {
	t.Helper()
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	url := env.baseURL + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("executing request %s %s: %v", method, path, err)
	}
	return resp
}

// authedRequest makes a request with a valid Bearer token.
func authedRequest(t *testing.T, method, path, body string) *http.Response {
	t.Helper()
	headers := map[string]string{
		"Authorization": "Bearer e2e-test-token",
		"Content-Type":  "application/json",
	}
	return doRequest(t, method, path, body, headers)
}

// readBody reads and closes the response body.
func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading response body: %v", err)
	}
	return string(b)
}

// ── Health Endpoint Tests ──────────────────────────────────────────────────

func TestHealthEndpoint(t *testing.T) {
	resp := authedRequest(t, "GET", "/api/v1/health", "")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body := readBody(t, resp)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, body)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decoding health response: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", result["status"])
	}
}

func TestHealthContentType(t *testing.T) {
	resp := authedRequest(t, "GET", "/api/v1/health", "")
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type=application/json, got %q", ct)
	}
}

// ── Auth Enforcement Tests ─────────────────────────────────────────────────

func TestAuthMissingHeader(t *testing.T) {
	resp := doRequest(t, "GET", "/api/v1/health", "", nil)
	body := readBody(t, resp)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without auth header, got %d: %s", resp.StatusCode, body)
	}

	var errResp errorEnvelope
	if err := json.Unmarshal([]byte(body), &errResp); err != nil {
		t.Fatalf("decoding error response: %v", err)
	}
	if errResp.Error.Code != "UNAUTHORIZED" {
		t.Errorf("expected error code UNAUTHORIZED, got %q", errResp.Error.Code)
	}
}

func TestAuthInvalidFormat(t *testing.T) {
	headers := map[string]string{
		"Authorization": "Basic dXNlcjpwYXNz",
	}
	resp := doRequest(t, "GET", "/api/v1/health", "", headers)
	body := readBody(t, resp)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 with Basic auth, got %d: %s", resp.StatusCode, body)
	}
}

func TestAuthEmptyBearer(t *testing.T) {
	headers := map[string]string{
		"Authorization": "Bearer",
	}
	resp := doRequest(t, "GET", "/api/v1/health", "", headers)
	body := readBody(t, resp)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 with empty Bearer, got %d: %s", resp.StatusCode, body)
	}
}

// ── Placeholder Endpoint Tests (501 Not Implemented) ───────────────────────

func TestUsersEndpointNotImplemented(t *testing.T) {
	resp := authedRequest(t, "GET", "/api/v1/users", "")
	body := readBody(t, resp)

	if resp.StatusCode != http.StatusNotImplemented {
		t.Fatalf("expected 501 for /users, got %d: %s", resp.StatusCode, body)
	}

	var errResp errorEnvelope
	if err := json.Unmarshal([]byte(body), &errResp); err != nil {
		t.Fatalf("decoding error response: %v", err)
	}
	if errResp.Error.Code != "NOT_IMPLEMENTED" {
		t.Errorf("expected error code NOT_IMPLEMENTED, got %q", errResp.Error.Code)
	}
}

func TestTeamsEndpointNotImplemented(t *testing.T) {
	resp := authedRequest(t, "GET", "/api/v1/teams", "")
	body := readBody(t, resp)

	if resp.StatusCode != http.StatusNotImplemented {
		t.Fatalf("expected 501 for /teams, got %d: %s", resp.StatusCode, body)
	}
}

func TestMetricsPipelineNotImplemented(t *testing.T) {
	resp := authedRequest(t, "GET", "/api/v1/metrics/pipeline", "")
	body := readBody(t, resp)

	if resp.StatusCode != http.StatusNotImplemented {
		t.Fatalf("expected 501 for /metrics/pipeline, got %d: %s", resp.StatusCode, body)
	}
}

// ── Lead Endpoint Tests ────────────────────────────────────────────────────
// Lead endpoints are not yet implemented and return 501. Once implemented,
// these tests should be updated to verify auth context enforcement (401
// when user context is missing) and actual CRUD behavior.

func TestLeadListNotImplemented(t *testing.T) {
	resp := authedRequest(t, "GET", "/api/v1/leads", "")
	body := readBody(t, resp)

	if resp.StatusCode != http.StatusNotImplemented {
		t.Fatalf("expected 501 for /leads (not implemented), got %d: %s", resp.StatusCode, body)
	}
}

func TestLeadCreateNotImplemented(t *testing.T) {
	payload := `{"title":"Test Lead","teamId":"b0000000-0000-0000-0000-000000000001","customerId":"c0000000-0000-0000-0000-000000000001"}`
	resp := authedRequest(t, "POST", "/api/v1/leads", payload)
	body := readBody(t, resp)

	if resp.StatusCode != http.StatusNotImplemented {
		t.Fatalf("expected 501 for POST /leads (not implemented), got %d: %s", resp.StatusCode, body)
	}
}

func TestLeadGetNotImplemented(t *testing.T) {
	resp := authedRequest(t, "GET", "/api/v1/leads/d0000000-0000-0000-0000-000000000001", "")
	body := readBody(t, resp)

	if resp.StatusCode != http.StatusNotImplemented {
		t.Fatalf("expected 501 for GET /leads/:id (not implemented), got %d: %s", resp.StatusCode, body)
	}
}

func TestLeadDeleteNotImplemented(t *testing.T) {
	resp := authedRequest(t, "DELETE", "/api/v1/leads/d0000000-0000-0000-0000-000000000001", "")
	body := readBody(t, resp)

	if resp.StatusCode != http.StatusNotImplemented {
		t.Fatalf("expected 501 for DELETE /leads/:id (not implemented), got %d: %s", resp.StatusCode, body)
	}
}

func TestLeadPatchStatusNotImplemented(t *testing.T) {
	payload := `{"status":"in_progress"}`
	resp := authedRequest(t, "PATCH", "/api/v1/leads/d0000000-0000-0000-0000-000000000001/status", payload)
	body := readBody(t, resp)

	if resp.StatusCode != http.StatusNotImplemented {
		t.Fatalf("expected 501 for PATCH /leads/:id/status (not implemented), got %d: %s", resp.StatusCode, body)
	}
}

// ── Routing Tests ──────────────────────────────────────────────────────────

func TestNotFoundReturns404(t *testing.T) {
	resp := authedRequest(t, "GET", "/api/v1/nonexistent", "")
	defer resp.Body.Close()

	// chi returns 404 for unmatched routes
	if resp.StatusCode != http.StatusNotFound {
		body := readBody(t, resp)
		t.Fatalf("expected 404 for unknown route, got %d: %s", resp.StatusCode, body)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	// PATCH on /health is not defined, should return 405
	resp := authedRequest(t, "PATCH", "/api/v1/health", "")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed && resp.StatusCode != http.StatusNotFound {
		body := readBody(t, resp)
		t.Fatalf("expected 405 or 404 for PATCH /health, got %d: %s", resp.StatusCode, body)
	}
}

// ── Response Format Tests ──────────────────────────────────────────────────

func TestErrorResponseFormat(t *testing.T) {
	// No auth header → 401 with structured error
	resp := doRequest(t, "GET", "/api/v1/health", "", nil)
	body := readBody(t, resp)

	var errResp errorEnvelope
	if err := json.Unmarshal([]byte(body), &errResp); err != nil {
		t.Fatalf("error response is not valid JSON: %v\nbody: %s", err, body)
	}
	if errResp.Error.Code == "" {
		t.Error("error response missing 'code' field")
	}
	if errResp.Error.Message == "" {
		t.Error("error response missing 'message' field")
	}
}

// ── Types ──────────────────────────────────────────────────────────────────

type errorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
