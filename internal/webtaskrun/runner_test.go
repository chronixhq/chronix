package webtaskrun

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLocalRunner_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "Value")
		w.WriteHeader(http.StatusCreated)
		body, _ := io.ReadAll(r.Body)
		_, _ = w.Write(body)
	}))
	defer server.Close()

	runner := NewLocalRunner()
	ctx := context.Background()
	headers := map[string]string{"X-Req": "ReqValue"}
	body := []byte("hello")

	result, err := runner.Execute(ctx, "POST", server.URL, headers, body, time.Second)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", result.StatusCode)
	}
	if string(result.ResponseBody) != "hello" {
		t.Errorf("Expected body 'hello', got %s", string(result.ResponseBody))
	}
	if result.ResponseHeaders.Get("X-Test") != "Value" {
		t.Errorf("Expected response header X-Test=Value")
	}
	if result.RequestURL != server.URL {
		t.Errorf("Expected request URL %s, got %s", server.URL, result.RequestURL)
	}
}

func TestLocalRunner_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	runner := NewLocalRunner()
	ctx := context.Background()

	_, err := runner.Execute(ctx, "GET", server.URL, nil, nil, 10*time.Millisecond)
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}
}
