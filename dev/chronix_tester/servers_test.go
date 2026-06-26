package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestAPIRouterProvidesFixtureEndpoints(t *testing.T) {
	dataDir := t.TempDir()
	paths := AppPaths{
		DataDir:   dataDir,
		ResultsDB: filepath.Join(dataDir, "results.db"),
		TargetDB:  filepath.Join(dataDir, "target.db"),
	}

	store, err := OpenStore(paths)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	router := newAPIRouter(store.Results)

	req := httptest.NewRequest(http.MethodGet, "/json", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected /json status: %d", rec.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode /json body: %v", err)
	}
	slideshow, ok := payload["slideshow"].(map[string]any)
	if !ok || slideshow["title"] != "Chronix Fixture API" {
		t.Fatalf("unexpected /json payload: %#v", payload)
	}

	req = httptest.NewRequest(http.MethodGet, "/response-headers?X-Test=ABC", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Header().Get("X-Test") != "ABC" {
		t.Fatalf("expected X-Test header to round-trip")
	}

	req = httptest.NewRequest(http.MethodGet, "/html", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if !strings.Contains(rec.Body.String(), "<h1>Chronix Test Fixture</h1>") {
		t.Fatalf("unexpected /html response: %s", rec.Body.String())
	}
}
