package agentregister

import (
	cfgpkg "chronix-agent/agentconfig"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func testConfigForServer(t *testing.T, srv *httptest.Server) *cfgpkg.Config {
	t.Helper()
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse server URL: %v", err)
	}
	host, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		t.Fatalf("split host/port: %v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("parse port: %v", err)
	}
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate keypair: %v", err)
	}
	return &cfgpkg.Config{
		ServerHost: host,
		WSPort:     port,
		UUID:       "agent-1",
		Name:       "Agent One",
		PubKeyB64:  base64.StdEncoding.EncodeToString(pub),
		PrivKeyB64: base64.StdEncoding.EncodeToString(priv),
	}
}

func serverPinB64(t *testing.T, srv *httptest.Server) string {
	t.Helper()
	cert := srv.TLS.Certificates[0]
	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatalf("parse server cert: %v", err)
	}
	hash := sha256.Sum256(leaf.RawSubjectPublicKeyInfo)
	return base64.StdEncoding.EncodeToString(hash[:])
}

func TestRepairRegistration_UsesAuthenticatedRepairWhenAvailable(t *testing.T) {
	var repairCalls int32
	var registerCalls int32

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/agent/self/repair":
			atomic.AddInt32(&repairCalls, 1)
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPost && r.URL.Path == "/agent/register":
			atomic.AddInt32(&registerCalls, 1)
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	cfg := testConfigForServer(t, srv)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	res, err := RepairRegistration(ctx, cfg, "v1.2.3")
	if err != nil {
		t.Fatalf("RepairRegistration() error = %v", err)
	}
	if res.Mode != RepairModeAuthenticated {
		t.Fatalf("RepairRegistration() mode = %q, want %q", res.Mode, RepairModeAuthenticated)
	}
	if got := atomic.LoadInt32(&repairCalls); got != 1 {
		t.Fatalf("repair calls = %d, want 1", got)
	}
	if got := atomic.LoadInt32(&registerCalls); got != 0 {
		t.Fatalf("register calls = %d, want 0", got)
	}
	if cfg.ServerSPKIB64 == "" {
		t.Fatal("expected repair to learn server pin")
	}
}

func TestRepairRegistration_FallsBackToApprovalFlow(t *testing.T) {
	var repairCalls int32
	var registerCalls int32
	var statusCalls int32

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/agent/self/repair":
			atomic.AddInt32(&repairCalls, 1)
			w.WriteHeader(http.StatusNotFound)
		case r.Method == http.MethodPost && r.URL.Path == "/agent/register":
			atomic.AddInt32(&registerCalls, 1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"requestId": "req-1"})
		case r.Method == http.MethodGet && r.URL.Path == "/agent/register/status/req-1":
			atomic.AddInt32(&statusCalls, 1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "approved"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	cfg := testConfigForServer(t, srv)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	res, err := RepairRegistration(ctx, cfg, "v1.2.3")
	if err != nil {
		t.Fatalf("RepairRegistration() error = %v", err)
	}
	if res.Mode != RepairModeApproval {
		t.Fatalf("RepairRegistration() mode = %q, want %q", res.Mode, RepairModeApproval)
	}
	if got := atomic.LoadInt32(&repairCalls); got != 1 {
		t.Fatalf("repair calls = %d, want 1", got)
	}
	if got := atomic.LoadInt32(&registerCalls); got != 1 {
		t.Fatalf("register calls = %d, want 1", got)
	}
	if got := atomic.LoadInt32(&statusCalls); got != 1 {
		t.Fatalf("status calls = %d, want 1", got)
	}
}

func TestProbeServer_VerifiesConfiguredPin(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "expired"})
	}))
	defer srv.Close()

	cfg := testConfigForServer(t, srv)
	cfg.ServerSPKIB64 = serverPinB64(t, srv)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	probe, err := ProbeServer(ctx, cfg)
	if err != nil {
		t.Fatalf("ProbeServer() error = %v", err)
	}
	if probe == nil || !probe.Reachable {
		t.Fatalf("ProbeServer() = %#v, want reachable result", probe)
	}
}

func TestProbeServer_DetectsPinMismatch(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := testConfigForServer(t, srv)
	cfg.ServerSPKIB64 = "not-the-right-pin"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if _, err := ProbeServer(ctx, cfg); err == nil {
		t.Fatal("ProbeServer() error = nil, want pin mismatch error")
	}
}
