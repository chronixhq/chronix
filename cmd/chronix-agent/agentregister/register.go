package agentregister

import (
	"bytes"
	cfgpkg "chronix-agent/agentconfig"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func RegisterWithServer(ctx context.Context, cfg *cfgpkg.Config, version string) error {
	payload := map[string]any{
		"uuid":      cfg.UUID,
		"name":      cfg.Name,
		"version":   version,
		"publicKey": cfg.PubKeyB64,
		"metadata":  CollectMetadata(),
	}
	b, _ := json.Marshal(payload)

	base := baseURL(cfg)
	client := newHTTPClient(30*time.Second, "")

	apiURL := fmt.Sprintf("%s/agent/register", base)
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	rememberServerPin(cfg, resp)

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		code, msg, _ := DecodeRestErrorResponse(body)
		return &HTTPError{
			Op:         "register",
			StatusCode: resp.StatusCode,
			Code:       code,
			Message:    msg,
			Body:       string(body),
		}
	}

	var res struct {
		RequestID string `json:"requestId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}
	requestID := strings.TrimSpace(res.RequestID)
	if requestID == "" {
		return errors.New("registration returned empty requestId")
	}

	fmt.Println()
	fmt.Println("Registration request sent. Please approve this agent in the Chronix Web UI.")
	fmt.Println("Waiting for approval (timeout: 5 minutes)...")
	return pollRegistration(ctx, client, base, requestID, cfg)
}

func pollRegistration(ctx context.Context, client *http.Client, base string, requestID string, cfg *cfgpkg.Config) error {
	pollURL := fmt.Sprintf("%s/agent/register/status/%s", base, url.QueryEscape(requestID))
	const pollInterval = 5 * time.Second

	var lastWarn time.Time
	var lastErrMsg string

	pollOnce := func() (done bool, err error) {
		req, _ := http.NewRequestWithContext(ctx, "GET", pollURL, nil)
		resp, err := client.Do(req)
		if err != nil {
			em := err.Error()
			if em != lastErrMsg || time.Since(lastWarn) >= 30*time.Second {
				_, _ = fmt.Fprintln(os.Stderr, "Still waiting for approval — unable to contact server (will retry):", err)
				lastWarn = time.Now()
				lastErrMsg = em
			}
			return false, nil
		}
		defer func() { _ = resp.Body.Close() }()
		rememberServerPin(cfg, resp)

		if resp.StatusCode == http.StatusOK {
			var res struct {
				Status string `json:"status"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&res); err == nil {
				status := strings.ToLower(res.Status)
				if status == "approved" {
					return true, nil
				}
				if status == "denied" {
					return true, ErrRegistrationDenied
				}
				if status == "expired" {
					return true, ErrRegistrationExpired
				}
			}
			return false, nil
		}
		if resp.StatusCode == http.StatusNotFound {
			return true, ErrRegistrationExpired
		}
		body, _ := io.ReadAll(resp.Body)
		code, msg, _ := DecodeRestErrorResponse(body)
		return true, &HTTPError{
			Op:         "register-status",
			StatusCode: resp.StatusCode,
			Code:       code,
			Message:    msg,
			Body:       string(body),
		}
	}

	if done, err := pollOnce(); done {
		return err
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if done, err := pollOnce(); done {
				return err
			}
		}
	}
}
