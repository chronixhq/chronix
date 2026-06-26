package cxrestapi

import (
	webtaskconnspkg "chronix/cxrestapi/webtaskconnectionsapi"
	"chronix/internal/db/models"
	"testing"
	"time"
)

func ptr[T any](v T) *T {
	return &v
}

func TestAnyToI64(t *testing.T) {
	tests := []struct {
		input    any
		def      int64
		expected int64
	}{
		{int(10), 0, 10},
		{int64(20), 0, 20},
		{float64(30.0), 0, 30},
		{true, 0, 1},
		{false, 0, 0},
		{"40", 0, 40},
		{"invalid", 99, 99},
		{nil, 55, 55},
	}

	for _, tt := range tests {
		got := webtaskconnspkg.AnyToI64(tt.input, tt.def)
		if got != tt.expected {
			t.Errorf("AnyToI64(%v, %d) = %d; want %d", tt.input, tt.def, got, tt.expected)
		}
	}
}

func TestPtr(t *testing.T) {
	v := "hello"
	p := ptr(v)
	if *p != v {
		t.Errorf("ptr(%v) = %v; want %v", v, *p, v)
	}
}

func TestMapWebtaskConnection(t *testing.T) {
	it := &models.WebtaskConnection{
		ID:                       ptr(int64(1)),
		Name:                     "Test",
		BaseURL:                  ptr("http://localhost"),
		AutoCheckEnabled:         ptr(int64(1)),
		AutoCheckIntervalSeconds: ptr(int64(60)),
		AuthType:                 "none",
		LastStatus:               ptr("ok"),
		LastError:                ptr(""),
		AgentUUID:                ptr("agent-1"),
		CreatedAt:                ptr(time.Now()),
		UpdatedAt:                ptr(time.Now()),
	}

	m := webtaskconnspkg.MapWebtaskConnection(it)
	if *m["id"].(*int64) != int64(1) {
		t.Errorf("Expected id 1, got %v", m["id"])
	}
	if m["name"] != "Test" {
		t.Errorf("Expected name Test, got %v", m["name"])
	}
	if m["autoCheckEnabled"] != true {
		t.Errorf("Expected autoCheckEnabled true, got %v", m["autoCheckEnabled"])
	}
	if *m["autoCheckSeconds"].(*int64) != int64(60) {
		t.Errorf("Expected autoCheckSeconds 60, got %v", m["autoCheckSeconds"])
	}
}
