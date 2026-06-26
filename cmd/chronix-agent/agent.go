// Package main provides the Chronix Agent implementation.
package main

import (
	regpkg "chronix-agent/agentregister"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gorilla/websocket"
)

func runAgent(ctx context.Context, cfg *agentConfig) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("recovered from panic in runAgent", "agent_id", cfg.UUID, "agent_name", cfg.Name, "error", r, "stack", string(debug.Stack()))
			time.Sleep(5 * time.Second)
			go runAgent(ctx, cfg)
		}
	}()

	ping := 30 * time.Second
	backoff := 500 * time.Millisecond
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		wsURL := fmt.Sprintf("wss://%s:%d/agent/connect", cfg.ServerHost, cfg.WSPort)
		jwtStr, err := regpkg.BuildJWT(cfg, Version)
		if err != nil {
			slog.Error("jwt generate failed", "component", "agent", "op", "connect", "error", err)
			if err := waitWithContext(ctx, backoff); err != nil {
				return
			}
			backoff = min(backoff*2, 60*time.Second)
			continue
		}

		dialer := websocket.Dialer{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				VerifyConnection: func(cs tls.ConnectionState) error {
					if len(cs.PeerCertificates) == 0 {
						return fmt.Errorf("no peer certificates")
					}
					leaf := cs.PeerCertificates[0]
					spki := leaf.RawSubjectPublicKeyInfo
					hash := sha256.Sum256(spki)
					learned := base64.StdEncoding.EncodeToString(hash[:])

					if cfg.ServerSPKIB64 == "" {
						cfg.ServerSPKIB64 = learned
						if err := saveConfig(defaultConfigPath(), cfg); err != nil {
							slog.Error("failed to save config with learned pin", "error", err)
						}
						slog.Info("learned server pin", "pin", learned)
						return nil
					}

					if cfg.ServerSPKIB64 != learned {
						return fmt.Errorf("server pin mismatch")
					}
					return nil
				},
			},
			Proxy: http.ProxyFromEnvironment,
		}
		headers := http.Header{}
		headers.Set("Authorization", "Bearer "+jwtStr)
		slog.Info("connecting", "component", "agent", "op", "connect", "url", wsURL, "id", cfg.UUID, "agent", cfg.Name)
		conn, resp, err := dialer.DialContext(ctx, wsURL, headers)
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		if err != nil {
			slog.Error("connect failed", "component", "agent", "op", "connect", "error", err)
			retryTime := backoff + time.Duration(float64(backoff)*0.2*(2*randFloat()-1))
			slog.Error("will retry in "+retryTime.Truncate(time.Second).String(), "component", "agent", "op", "connect", "url", wsURL, "id", cfg.UUID, "agent", cfg.Name)
			if err := waitWithContext(ctx, retryTime); err != nil {
				return
			}
			backoff = min(backoff*2, 60*time.Second)
			continue
		}

		backoff = 500 * time.Millisecond
		slog.Info("connected", "component", "agent", "op", "connected", "url", wsURL, "id", cfg.UUID, "agent", cfg.Name)

		_ = conn.SetReadDeadline(time.Now().Add(ping * 3))
		conn.SetPongHandler(func(string) error { return conn.SetReadDeadline(time.Now().Add(ping * 3)) })

		stop := make(chan error, 1)
		inflight := newAgentInflight()

		go func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("recovered from panic in agent read goroutine", "error", r, "stack", string(debug.Stack()))
					stop <- fmt.Errorf("panic: %v", r)
				}
			}()
			for {
				t, data, err := conn.ReadMessage()
				if err != nil {
					stop <- err
					return
				}
				if t != websocket.TextMessage {
					continue
				}
				var env agentEnvelope
				if err := json.Unmarshal(data, &env); err != nil {
					continue
				}
				handleAgentMessage(ctx, conn, cfg, env, inflight)
			}
		}()

		go func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("recovered from panic in agent ping goroutine", "error", r, "stack", string(debug.Stack()))
				}
			}()
			ticker := time.NewTicker(ping)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
						return
					}
				case <-stop:
					return
				case <-ctx.Done():
					return
				}
			}
		}()

		select {
		case <-ctx.Done():
		case err = <-stop:
			slog.Error("disconnected", "component", "agent", "op", "disconnected", "error", err, "url", wsURL, "id", cfg.UUID, "agent", cfg.Name)
		}

		_ = conn.Close()
		inflight.cancelAll()

		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
			backoff = min(backoff*2, 60*time.Second)
		}
	}
}
