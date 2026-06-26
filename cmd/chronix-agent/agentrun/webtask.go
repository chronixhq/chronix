package agentrun

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"
)

type WebTaskResult struct {
	StatusCode      int               `json:"statusCode"`
	ResponseHeaders map[string]string `json:"responseHeaders"`
	ResponseBody    string            `json:"responseBody"`
	LatencyMs       int64             `json:"latencyMs"`
}

func RunWebTask(ctx context.Context, method, url string, headers map[string]string, body []byte) (*WebTaskResult, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(resp.Body)

	respHeaders := map[string]string{}
	for k, v := range resp.Header {
		if len(v) > 0 {
			respHeaders[k] = v[0]
		}
	}

	return &WebTaskResult{
		StatusCode:      resp.StatusCode,
		ResponseHeaders: respHeaders,
		ResponseBody:    string(respBody),
		LatencyMs:       latency.Milliseconds(),
	}, nil
}
