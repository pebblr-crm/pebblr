package azuread

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pebblr/pebblr/internal/domain"
)

// testKeyPair generates an RSA key pair and matching JWK for tests.
func testKeyPair(t *testing.T) (*rsa.PrivateKey, jwkKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating RSA key: %v", err)
	}

	jwk := jwkKey{
		Kty: "RSA",
		Use: "sig",
		Kid: "test-kid-1",
		N:   base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
	}
	return key, jwk
}

// signJWT creates a signed RS256 JWT for testing.
func signJWT(t *testing.T, key *rsa.PrivateKey, kid string, claims tokenClaims) string {
	t.Helper()

	header := jwtHeader{Alg: "RS256", Kid: kid, Typ: "JWT"}
	headerJSON, _ := json.Marshal(header)
	claimsJSON, _ := json.Marshal(claims)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signingInput := headerB64 + "." + claimsB64
	hash := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash[:])
	if err != nil {
		t.Fatalf("signing JWT: %v", err)
	}

	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig)
}

// newTestServer sets up a fake Azure AD OIDC discovery + JWKS server.
func newTestServer(t *testing.T, jwk jwkKey) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	var serverURL string

	mux.HandleFunc("GET /tenant-1/v2.0/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		doc := fmt.Sprintf(`{"jwks_uri":"%s/jwks"}`, serverURL)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(doc))
	})

	mux.HandleFunc("GET /jwks", func(w http.ResponseWriter, r *http.Request) {
		resp := jwksResponse{Keys: []jwkKey{jwk}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})

	srv := httptest.NewServer(mux)
	serverURL = srv.URL
	t.Cleanup(srv.Close)
	return srv
}

func TestValidateToken_Valid(t *testing.T) {
	t.Parallel()

	privKey, jwk := testKeyPair(t)
	srv := newTestServer(t, jwk)

	a, err := New(context.Background(), Config{
		TenantID:  "tenant-1",
		ClientID:  "app-client-id",
		IssuerURL: srv.URL + "/tenant-1/v2.0",
	})
	if err != nil {
		t.Fatalf("creating authenticator: %v", err)
	}

	now := time.Now().Unix()
	token := signJWT(t, privKey, "test-kid-1", tokenClaims{
		Iss:    srv.URL + "/tenant-1/v2.0",
		Sub:    "user-sub-1",
		Aud:    "app-client-id",
		Exp:    now + 3600,
		Nbf:    now - 60,
		Iat:    now,
		OID:    "azure-oid-abc",
		Email:  "rep@contoso.com",
		Name:   "Riley Rep",
		Roles:  []string{"rep"},
		Groups: []string{"team-west"},
	})

	claims, err := a.ValidateToken(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.Sub != "azure-oid-abc" {
		t.Errorf("expected sub azure-oid-abc, got %q", claims.Sub)
	}
	if claims.Email != "rep@contoso.com" {
		t.Errorf("expected email rep@contoso.com, got %q", claims.Email)
	}
	if len(claims.Roles) != 1 || claims.Roles[0] != domain.RoleRep {
		t.Errorf("expected [rep] role, got %v", claims.Roles)
	}
	if len(claims.TeamIDs) != 1 || claims.TeamIDs[0] != "team-west" {
		t.Errorf("expected [team-west], got %v", claims.TeamIDs)
	}
}

func TestValidateToken_Expired(t *testing.T) {
	t.Parallel()

	privKey, jwk := testKeyPair(t)
	srv := newTestServer(t, jwk)

	a, err := New(context.Background(), Config{
		TenantID:  "tenant-1",
		ClientID:  "app-client-id",
		IssuerURL: srv.URL + "/tenant-1/v2.0",
	})
	if err != nil {
		t.Fatalf("creating authenticator: %v", err)
	}

	now := time.Now().Unix()
	token := signJWT(t, privKey, "test-kid-1", tokenClaims{
		Iss: srv.URL + "/tenant-1/v2.0",
		Sub: "user-sub-1",
		Aud: "app-client-id",
		Exp: now - 60, // expired
		Nbf: now - 3600,
		Iat: now - 3600,
	})

	_, err = a.ValidateToken(context.Background(), token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestValidateToken_WrongAudience(t *testing.T) {
	t.Parallel()

	privKey, jwk := testKeyPair(t)
	srv := newTestServer(t, jwk)

	a, err := New(context.Background(), Config{
		TenantID:  "tenant-1",
		ClientID:  "app-client-id",
		IssuerURL: srv.URL + "/tenant-1/v2.0",
	})
	if err != nil {
		t.Fatalf("creating authenticator: %v", err)
	}

	now := time.Now().Unix()
	token := signJWT(t, privKey, "test-kid-1", tokenClaims{
		Iss: srv.URL + "/tenant-1/v2.0",
		Sub: "user-sub-1",
		Aud: "wrong-audience",
		Exp: now + 3600,
		Nbf: now - 60,
		Iat: now,
	})

	_, err = a.ValidateToken(context.Background(), token)
	if err == nil {
		t.Fatal("expected error for wrong audience")
	}
}

func TestValidateToken_WrongIssuer(t *testing.T) {
	t.Parallel()

	privKey, jwk := testKeyPair(t)
	srv := newTestServer(t, jwk)

	a, err := New(context.Background(), Config{
		TenantID:  "tenant-1",
		ClientID:  "app-client-id",
		IssuerURL: srv.URL + "/tenant-1/v2.0",
	})
	if err != nil {
		t.Fatalf("creating authenticator: %v", err)
	}

	now := time.Now().Unix()
	token := signJWT(t, privKey, "test-kid-1", tokenClaims{
		Iss: "https://evil.example.com",
		Sub: "user-sub-1",
		Aud: "app-client-id",
		Exp: now + 3600,
		Nbf: now - 60,
		Iat: now,
	})

	_, err = a.ValidateToken(context.Background(), token)
	if err == nil {
		t.Fatal("expected error for wrong issuer")
	}
}

func TestValidateToken_MalformedToken(t *testing.T) {
	t.Parallel()

	_, jwk := testKeyPair(t)
	srv := newTestServer(t, jwk)

	a, err := New(context.Background(), Config{
		TenantID:  "tenant-1",
		ClientID:  "app-client-id",
		IssuerURL: srv.URL + "/tenant-1/v2.0",
	})
	if err != nil {
		t.Fatalf("creating authenticator: %v", err)
	}

	_, err = a.ValidateToken(context.Background(), "not-a-jwt")
	if err == nil {
		t.Fatal("expected error for malformed token")
	}
}

func TestClaimsToUserClaims_DefaultsToRep(t *testing.T) {
	t.Parallel()

	c := &tokenClaims{
		Sub:   "sub-1",
		Email: "user@example.com",
		Name:  "Test User",
		Roles: []string{}, // no roles
	}
	uc := claimsToUserClaims(c)
	if len(uc.Roles) != 1 || uc.Roles[0] != domain.RoleRep {
		t.Errorf("expected default [rep] role, got %v", uc.Roles)
	}
}

func TestClaimsToUserClaims_PrefersOID(t *testing.T) {
	t.Parallel()

	c := &tokenClaims{
		Sub:   "sub-1",
		OID:   "oid-1",
		Email: "user@example.com",
		Roles: []string{"admin"},
	}
	uc := claimsToUserClaims(c)
	if uc.Sub != "oid-1" {
		t.Errorf("expected sub to be OID oid-1, got %q", uc.Sub)
	}
}
