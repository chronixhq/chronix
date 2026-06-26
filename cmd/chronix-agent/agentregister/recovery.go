package agentregister

import (
	"bytes"
	cfgpkg "chronix-agent/agentconfig"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type RepairMode string

const (
	RepairModeAuthenticated RepairMode = "authenticated"
	RepairModeApproval      RepairMode = "approval"
)

type RepairResult struct {
	Mode RepairMode
}

type ProbeResult struct {
	PinVerified bool
	Reachable   bool
	StatusCode  int
}

func ProbeServer(ctx context.Context, cfg *cfgpkg.Config) (*ProbeResult, error) {
	client := newHTTPClient(5*time.Second, cfg.ServerSPKIB64)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/agent/register/status/%s", baseURL(cfg), "__probe__"), nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	return &ProbeResult{
		PinVerified: strings.TrimSpace(cfg.ServerSPKIB64) != "",
		Reachable:   true,
		StatusCode:  resp.StatusCode,
	}, nil
}

func RepairRegistration(ctx context.Context, cfg *cfgpkg.Config, version string) (*RepairResult, error) {
	payload := map[string]any{
		"metadata": CollectMetadata(),
	}
	b, _ := json.Marshal(payload)

	jwtStr, err := BuildJWT(cfg, version)
	if err != nil {
		return nil, err
	}

	client := newHTTPClient(30*time.Second, "")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/agent/self/repair", baseURL(cfg)), bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+jwtStr)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	rememberServerPin(cfg, resp)

	switch resp.StatusCode {
	case http.StatusOK, http.StatusNoContent:
		return &RepairResult{Mode: RepairModeAuthenticated}, nil
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
		if err := RegisterWithServer(ctx, cfg, version); err != nil {
			return nil, err
		}
		return &RepairResult{Mode: RepairModeApproval}, nil
	default:
		body, _ := io.ReadAll(resp.Body)
		code, msg, _ := DecodeRestErrorResponse(body)
		return nil, &HTTPError{
			Op:         "repair-register",
			StatusCode: resp.StatusCode,
			Code:       code,
			Message:    msg,
			Body:       string(body),
		}
	}
}

func UnregisterFromServer(ctx context.Context, cfg *cfgpkg.Config) error {
	jwtStr, err := BuildJWT(cfg, "")
	if err != nil {
		return err
	}

	payload := map[string]any{"reason": "cli unregister"}
	b, _ := json.Marshal(payload)

	client := newHTTPClient(15*time.Second, cfg.ServerSPKIB64)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/agent/self/unregister", baseURL(cfg)), bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+jwtStr)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	rememberServerPin(cfg, resp)

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	code, msg, _ := DecodeRestErrorResponse(body)
	return &HTTPError{
		Op:         "unregister",
		StatusCode: resp.StatusCode,
		Code:       code,
		Message:    msg,
		Body:       string(body),
	}
}
