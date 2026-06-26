package execution

import (
	"chronix/internal/webtaskrun"
	"fmt"
	"net/http"
	"testing"
	"time"

	"gorm.io/datatypes"
)

func TestEvaluateWebTaskExpectation(t *testing.T) {
	vars := map[string]any{"code": 200, "token": "abc"}

	t.Run("statusCode success", func(t *testing.T) {
		expect := datatypes.JSONMap{"kind": "statusCode", "value": "200", "op": "=="}
		res := &webtaskrun.WebTaskResult{StatusCode: 200}
		pass, msg, _ := evaluateWebTaskExpectation(expect, res, vars)
		if !pass {
			t.Errorf("Expected pass, got %v (msg: %s)", pass, msg)
		}
	})

	t.Run("statusCode variable success", func(t *testing.T) {
		expect := datatypes.JSONMap{"kind": "statusCode", "value": "${code}", "op": "=="}
		res := &webtaskrun.WebTaskResult{StatusCode: 200}
		pass, msg, _ := evaluateWebTaskExpectation(expect, res, vars)
		if !pass {
			t.Errorf("Expected pass, got %v (msg: %s)", pass, msg)
		}
	})

	t.Run("bodyContains success", func(t *testing.T) {
		expect := datatypes.JSONMap{"kind": "bodyContains", "value": "hello"}
		res := &webtaskrun.WebTaskResult{ResponseBody: []byte("hello world")}
		pass, msg, _ := evaluateWebTaskExpectation(expect, res, vars)
		if !pass {
			t.Errorf("Expected pass, got %v (msg: %s)", pass, msg)
		}
	})

	t.Run("jsonPath success", func(t *testing.T) {
		expect := datatypes.JSONMap{"kind": "jsonPath", "path": "$.status", "value": "ok"}
		res := &webtaskrun.WebTaskResult{ResponseBody: []byte(`{"status":"ok"}`)}
		pass, msg, _ := evaluateWebTaskExpectation(expect, res, vars)
		if !pass {
			t.Errorf("Expected pass, got %v (msg: %s)", pass, msg)
		}
	})

	t.Run("latency failure", func(t *testing.T) {
		expect := datatypes.JSONMap{"kind": "latency", "value": 100}
		res := &webtaskrun.WebTaskResult{Latency: 200 * time.Millisecond}
		pass, msg, _ := evaluateWebTaskExpectation(expect, res, vars)
		if pass {
			t.Errorf("Expected fail, got pass (msg: %s)", msg)
		}
	})

	t.Run("bodyRegex success", func(t *testing.T) {
		expect := datatypes.JSONMap{"kind": "bodyRegex", "value": "id\": (\\d+)", "group": "1", "expected": "123"}
		res := &webtaskrun.WebTaskResult{ResponseBody: []byte(`{"id": 123}`)}
		pass, msg, _ := evaluateWebTaskExpectation(expect, res, vars)
		if !pass {
			t.Errorf("Expected pass, got %v (msg: %s)", pass, msg)
		}
	})

	t.Run("bodyRegex failure", func(t *testing.T) {
		expect := datatypes.JSONMap{"kind": "bodyRegex", "value": "id\": (\\d+)", "group": "1", "expected": "456"}
		res := &webtaskrun.WebTaskResult{ResponseBody: []byte(`{"id": 123}`)}
		pass, _, _ := evaluateWebTaskExpectation(expect, res, vars)
		if pass {
			t.Errorf("Expected fail, got pass")
		}
	})
}

func TestCaptureWebTaskVariables(t *testing.T) {
	captures := datatypes.JSONMap{
		"v1": map[string]any{"source": "jsonpath", "path": "$.id"},
		"v2": map[string]any{"source": "header", "name": "X-ID"},
		"v3": map[string]any{"source": "regex", "pattern": `ID: (\d+)`, "group": float64(1)},
	}
	res := &webtaskrun.WebTaskResult{
		ResponseBody:    []byte(`{"id": 123, "msg": "ID: 456"}`),
		ResponseHeaders: http.Header{"X-Id": []string{"789"}},
	}
	vars := make(map[string]any)
	newVars := captureWebTaskVariables(&captures, res, vars)

	if fmt.Sprint(newVars["v1"]) != "123" {
		t.Errorf("Expected v1=123, got %v", newVars["v1"])
	}
	if newVars["v2"] != "789" {
		t.Errorf("Expected v2=789, got %v", newVars["v2"])
	}
	if newVars["v3"] != "456" {
		t.Errorf("Expected v3=456, got %v", newVars["v3"])
	}
}
