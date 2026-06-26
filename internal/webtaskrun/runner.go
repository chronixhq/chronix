package webtaskrun

import (
	"bytes"
	"chronix/internal/agentmux"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type WebTaskResult struct {
	StatusCode      int
	ResponseHeaders http.Header
	ResponseBody    []byte
	Latency         time.Duration
	RequestURL      string
	RequestMethod   string
	RequestHeaders  map[string]string
	RequestBody     []byte
}

type WebTaskRunner interface {
	Execute(ctx context.Context, method, url string, headers map[string]string, body []byte, timeout time.Duration) (*WebTaskResult, error)
}

type LocalRunner struct {
	Client *http.Client
}

func NewLocalRunner() *LocalRunner {
	return &LocalRunner{
		Client: &http.Client{},
	}
}

func (r *LocalRunner) Execute(ctx context.Context, method, url string, headers map[string]string, body []byte, timeout time.Duration) (*WebTaskResult, error) {
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	start := time.Now()
	resp, err := r.Client.Do(req)
	latency := time.Since(start)

	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	return &WebTaskResult{
		StatusCode:      resp.StatusCode,
		ResponseHeaders: resp.Header,
		ResponseBody:    respBody,
		Latency:         latency,
		RequestURL:      url,
		RequestMethod:   method,
		RequestHeaders:  headers,
		RequestBody:     body,
	}, nil
}

type AgentRunner struct {
	AgentID string
}

func (a *AgentRunner) Execute(ctx context.Context, method, url string, headers map[string]string, body []byte, timeout time.Duration) (*WebTaskResult, error) {
	c := agentmux.DefaultManager.Get(a.AgentID)
	if c == nil {
		return nil, fmt.Errorf("agent not connected")
	}

	toMs := int64(timeout.Milliseconds())
	if toMs == 0 {
		toMs = 30000
	}

	payload := map[string]any{
		"method":    method,
		"url":       url,
		"headers":   headers,
		"body":      body,
		"timeoutMs": toMs,
	}

	respType, respBytes, err := c.Request(ctx, "webtask.run", payload)
	if err != nil {
		return nil, err
	}

	switch respType {
	case "webtask.run.ok":
		var res struct {
			StatusCode      int               `json:"statusCode"`
			ResponseHeaders map[string]string `json:"responseHeaders"`
			ResponseBody    string            `json:"responseBody"`
			LatencyMs       int64             `json:"latencyMs"`
		}
		if err := json.Unmarshal(respBytes, &res); err != nil {
			return nil, err
		}

		h := http.Header{}
		for k, v := range res.ResponseHeaders {
			h.Set(k, v)
		}

		return &WebTaskResult{
			StatusCode:      res.StatusCode,
			ResponseHeaders: h,
			ResponseBody:    []byte(res.ResponseBody),
			Latency:         time.Duration(res.LatencyMs) * time.Millisecond,
			RequestURL:      url,
			RequestMethod:   method,
			RequestHeaders:  headers,
			RequestBody:     body,
		}, nil
	case "webtask.error":
		var e struct {
			Message string `json:"message"`
		}
		_ = json.Unmarshal(respBytes, &e)
		return nil, fmt.Errorf("agent error: %s", e.Message)
	}

	return nil, fmt.Errorf("unexpected response type: %s", respType)
}
