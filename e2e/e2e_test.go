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
	// Must match the jwt-secret created by scripts/cluster-db.sh.
	defaultToken = "local-jwt-secret-not-for-production"
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

// waitForReady polls the healthz endpoint (outside auth) until it returns 200.
func waitForReady(baseURL string, timeout time.Duration) error {
	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := client.Get(baseURL + "/healthz")
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

// apiRequest makes an authenticated request to an API endpoint.
func apiRequest(t *testing.T, method, path, body string) *http.Response {
	t.Helper()
	token := os.Getenv("E2E_TOKEN")
	if token == "" {
		token = defaultToken
	}
	headers := map[string]string{
		"Authorization": "Bearer " + token,
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
	resp := apiRequest(t, "GET", "/api/v1/health", "")
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
	resp := apiRequest(t, "GET", "/api/v1/health", "")
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
}

func TestAuthWrongToken(t *testing.T) {
	headers := map[string]string{
		"Authorization": "Bearer wrong-token",
	}
	resp := doRequest(t, "GET", "/api/v1/health", "", headers)
	body := readBody(t, resp)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 with wrong token, got %d: %s", resp.StatusCode, body)
	}
}

// ── Current User (Me) Endpoint ───────────────────────────────────────────

func TestMeEndpointReturnsDevUser(t *testing.T) {
	resp := apiRequest(t, "GET", "/api/v1/me", "")
	body := readBody(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for /me, got %d: %s", resp.StatusCode, body)
	}

	var user map[string]any
	if err := json.Unmarshal([]byte(body), &user); err != nil {
		t.Fatalf("decoding /me response: %v\nbody: %s", err, body)
	}
	if user["role"] != "admin" {
		t.Errorf("expected dev user role=admin, got %q", user["role"])
	}
}

// ── User Endpoint Tests ──────────────────────────────────────────────────

func TestUserListReturnsOK(t *testing.T) {
	resp := apiRequest(t, "GET", "/api/v1/users", "")
	body := readBody(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for GET /users, got %d: %s", resp.StatusCode, body)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		t.Fatalf("decoding user list response: %v\nbody: %s", err, body)
	}
	if _, ok := result["items"]; !ok {
		t.Error("expected 'items' key in user list response")
	}
}

// ── Team Endpoint Tests ──────────────────────────────────────────────────

func TestTeamListReturnsOK(t *testing.T) {
	resp := apiRequest(t, "GET", "/api/v1/teams", "")
	body := readBody(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for GET /teams, got %d: %s", resp.StatusCode, body)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		t.Fatalf("decoding team list response: %v\nbody: %s", err, body)
	}
	if _, ok := result["items"]; !ok {
		t.Error("expected 'items' key in team list response")
	}
}

// ── Activity Endpoint Tests ──────────────────────────────────────────────

func TestActivityListReturnsOK(t *testing.T) {
	resp := apiRequest(t, "GET", "/api/v1/activities", "")
	body := readBody(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for GET /activities, got %d: %s", resp.StatusCode, body)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		t.Fatalf("decoding activity list response: %v\nbody: %s", err, body)
	}
	if _, ok := result["items"]; !ok {
		t.Error("expected 'items' key in activity list response")
	}
}

// ── Routing Tests ────────────────────────────────────────────────────────

func TestMethodNotAllowed(t *testing.T) {
	// PATCH on /health is not defined, should return 405
	resp := apiRequest(t, "PATCH", "/api/v1/health", "")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed && resp.StatusCode != http.StatusNotFound {
		body := readBody(t, resp)
		t.Fatalf("expected 405 or 404 for PATCH /health, got %d: %s", resp.StatusCode, body)
	}
}

// ── Response Format Tests ────────────────────────────────────────────────

func TestErrorResponseFormat(t *testing.T) {
	// Request a non-existent target → 404 with structured error
	resp := apiRequest(t, "GET", "/api/v1/targets/00000000-0000-0000-0000-000000000000", "")
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

// ── Types ────────────────────────────────────────────────────────────────

type errorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
