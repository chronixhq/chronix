package main

import (
	regpkg "chronix-agent/agentregister"
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func TestFormatRegisterError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "denied",
			err:  regpkg.ErrRegistrationDenied,
			want: "Registration denied. An administrator rejected the request in the Chronix Web UI.",
		},
		{
			name: "expired",
			err:  regpkg.ErrRegistrationExpired,
			want: "Registration expired. The approval window timed out. Please run the register command again.",
		},
		{
			name: "deadline exceeded",
			err:  context.DeadlineExceeded,
			want: "Registration timed out while waiting for approval. Please try again.",
		},
		{
			name: "agent disabled forbidden",
			err: &regpkg.HTTPError{
				Op:         "register",
				StatusCode: http.StatusForbidden,
				Code:       "PermissionDenied",
				Message:    "agent connections are disabled",
			},
			want: "Registration rejected: Agent connections are disabled on the server. Enable Agent Connections in the Chronix Admin UI (Network settings) and try again.",
		},
		{
			name: "too many requests",
			err: &regpkg.HTTPError{
				Op:         "register",
				StatusCode: http.StatusTooManyRequests,
				Code:       "TooManyRequests",
				Message:    "too many requests",
			},
			want: "Registration rejected: too many requests. Please wait a moment and try again.",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := regpkg.FormatError(tt.err)
			if got != tt.want {
				t.Fatalf("FormatError() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRegisterExitCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "nil", err: nil, want: 0},
		{name: "denied", err: regpkg.ErrRegistrationDenied, want: 3},
		{name: "expired", err: regpkg.ErrRegistrationExpired, want: 4},
		{name: "deadline exceeded", err: context.DeadlineExceeded, want: 4},
		{name: "forbidden", err: &regpkg.HTTPError{StatusCode: http.StatusForbidden}, want: 3},
		{name: "unauthorized", err: &regpkg.HTTPError{StatusCode: http.StatusUnauthorized}, want: 3},
		{name: "generic", err: errors.New("boom"), want: 1},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := regpkg.ExitCode(tt.err)
			if got != tt.want {
				t.Fatalf("ExitCode() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestRegisterWithServer_UsesHTTPSWhenServerIsHTTPS(t *testing.T) {
	var registerCalls int32
	var statusCalls int32

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/agent/register":
			atomic.AddInt32(&registerCalls, 1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"requestId": "req-1"})
			return
		case r.Method == http.MethodGet && r.URL.Path == "/agent/register/status/req-1":
			atomic.AddInt32(&statusCalls, 1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "approved",
			})
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	t.Cleanup(srv.Close)

	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse httptest URL: %v", err)
	}
	host, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		t.Fatalf("split host/port: %v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("parse port: %v", err)
	}

	cfg := &agentConfig{
		UUID:       "agent-1",
		Name:       "Agent One",
		PubKeyB64:  "ignored",
		ServerHost: host,
		WSPort:     port,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)

	if err := regpkg.RegisterWithServer(ctx, cfg, Version); err != nil {
		t.Fatalf("RegisterWithServer() error = %v", err)
	}
	if got := atomic.LoadInt32(&registerCalls); got != 1 {
		t.Fatalf("register calls = %d, want %d", got, 1)
	}
	if got := atomic.LoadInt32(&statusCalls); got != 1 {
		t.Fatalf("status calls = %d, want %d", got, 1)
	}
	if cfg.ServerHost != host {
		t.Fatalf("cfg.ServerHost = %q, want %q", cfg.ServerHost, host)
	}
	if cfg.WSPort != port {
		t.Fatalf("cfg.WSPort = %d, want %d", cfg.WSPort, port)
	}
	if cfg.ServerSPKIB64 == "" {
		t.Fatal("expected RegisterWithServer to learn the server pin")
	}
}
