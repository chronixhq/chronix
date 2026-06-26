package agentregister

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	ErrRegistrationExpired = errors.New("registration expired")
	ErrRegistrationDenied  = errors.New("registration denied")
)

type HTTPError struct {
	Op         string
	StatusCode int
	Code       string
	Message    string
	Body       string
}

func (e *HTTPError) Error() string {
	if e == nil {
		return ""
	}
	parts := make([]string, 0, 4)
	if e.Op != "" {
		parts = append(parts, e.Op)
	}
	if e.StatusCode > 0 {
		parts = append(parts, fmt.Sprintf("http %d", e.StatusCode))
	}
	if e.Code != "" {
		parts = append(parts, e.Code)
	}
	if e.Message != "" {
		parts = append(parts, e.Message)
	}
	if len(parts) == 0 {
		return "registration http error"
	}
	return strings.Join(parts, ": ")
}

func DecodeRestErrorResponse(body []byte) (code, message string, ok bool) {
	var r struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return "", "", false
	}
	r.Code = strings.TrimSpace(r.Code)
	r.Message = strings.TrimSpace(r.Message)
	if r.Code == "" && r.Message == "" {
		return "", "", false
	}
	return r.Code, r.Message, true
}

func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	if errors.Is(err, ErrRegistrationDenied) {
		return 3
	}
	if errors.Is(err, ErrRegistrationExpired) {
		return 4
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return 4
	}
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		if httpErr.StatusCode == http.StatusForbidden || httpErr.StatusCode == http.StatusUnauthorized {
			return 3
		}
	}
	return 1
}

func FormatError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, ErrRegistrationDenied) {
		return "Registration denied. An administrator rejected the request in the Chronix Web UI."
	}
	if errors.Is(err, ErrRegistrationExpired) {
		return "Registration expired. The approval window timed out. Please run the register command again."
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "Registration timed out while waiting for approval. Please try again."
	}
	if errors.Is(err, context.Canceled) {
		return "Registration canceled."
	}

	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		msg := strings.TrimSpace(httpErr.Message)
		code := strings.TrimSpace(httpErr.Code)
		if httpErr.StatusCode == http.StatusForbidden && strings.Contains(strings.ToLower(msg), "disabled") {
			return "Registration rejected: Agent connections are disabled on the server. Enable Agent Connections in the Chronix Admin UI (Network settings) and try again."
		}
		if httpErr.StatusCode == http.StatusTooManyRequests {
			return "Registration rejected: too many requests. Please wait a moment and try again."
		}
		if msg != "" {
			if code != "" {
				return fmt.Sprintf("Registration rejected by server (HTTP %d %s): %s", httpErr.StatusCode, code, msg)
			}
			return fmt.Sprintf("Registration rejected by server (HTTP %d): %s", httpErr.StatusCode, msg)
		}
		body := strings.TrimSpace(httpErr.Body)
		if body != "" {
			return fmt.Sprintf("Registration rejected by server (HTTP %d): %s", httpErr.StatusCode, body)
		}
		return fmt.Sprintf("Registration rejected by server (HTTP %d).", httpErr.StatusCode)
	}

	return fmt.Sprintf("Registration failed: %v", err)
}
