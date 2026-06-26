package execution

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"chronix/internal/webtaskrun"

	"github.com/oliveagle/jsonpath"
	"gorm.io/datatypes"
)

func evaluateWebTaskExpectation(expectation datatypes.JSONMap, result *webtaskrun.WebTaskResult, vars map[string]any) (bool, string, map[string]any) {
	kind, _ := expectation["kind"].(string)
	if kind == "" {
		return true, "ok", nil
	}
	meta := map[string]any{"expect_kind": kind}

	asFloat := func(v any) (float64, bool) {
		if f, ok := v.(float64); ok {
			return f, true
		}
		if s, ok := v.(string); ok {
			var f float64
			if _, err := fmt.Sscan(s, &f); err == nil {
				return f, true
			}
		}
		return 0, false
	}

	switch kind {
	case "statusCode":
		expected := float64(200)
		val := expectation["value"]
		if s, ok := val.(string); ok {
			val = substituteVars(s, vars)
		}
		if v, ok := asFloat(val); ok {
			expected = v
		}
		op, _ := expectation["op"].(string)
		if op == "" {
			op = "=="
		}
		actual := float64(result.StatusCode)
		meta["expect_expected"] = expected
		meta["expect_actual"] = actual
		meta["expect_op"] = op
		pass := false
		switch op {
		case "==":
			pass = actual == expected
		case "!=":
			pass = actual != expected
		case ">":
			pass = actual > expected
		case "<":
			pass = actual < expected
		case ">=":
			pass = actual >= expected
		case "<=":
			pass = actual <= expected
		}
		if pass {
			return true, "ok", meta
		}
		return false, fmt.Sprintf("Expected status code %s %v, but got %v", op, expected, actual), meta

	case "bodyContains":
		expected, _ := expectation["value"].(string)
		expected = substituteVars(expected, vars)
		meta["expect_expected"] = expected
		if strings.Contains(string(result.ResponseBody), expected) {
			return true, "ok", meta
		}
		return false, fmt.Sprintf("Body did not contain expected string: %q", expected), meta

	case "jsonPath":
		path, _ := expectation["path"].(string)
		path = substituteVars(path, vars)
		expected, _ := expectation["value"].(string)
		expected = substituteVars(expected, vars)
		meta["expect_path"] = path
		meta["expect_expected"] = expected
		var jsonData interface{}
		if err := json.Unmarshal(result.ResponseBody, &jsonData); err != nil {
			return false, "Invalid JSON in response body", meta
		}
		res, err := jsonpath.JsonPathLookup(jsonData, path)
		if err != nil {
			return false, fmt.Sprintf("JSONPath lookup failed: %v", err), meta
		}
		actual := fmt.Sprint(res)
		meta["expect_actual"] = actual
		if actual == expected {
			return true, "ok", meta
		}
		return false, fmt.Sprintf("JSONPath %q: expected %q, got %q", path, expected, actual), meta

	case "latency":
		val := expectation["value"]
		if s, ok := val.(string); ok {
			val = substituteVars(s, vars)
		}
		maxMs, _ := asFloat(val)
		actualMs := result.Latency.Milliseconds()
		meta["expect_expected"] = maxMs
		meta["expect_actual"] = actualMs
		if actualMs <= int64(maxMs) {
			return true, "ok", meta
		}
		return false, fmt.Sprintf("Latency %dms exceeded maximum %vms", actualMs, maxMs), meta

	case "bodyRegex":
		pattern, _ := expectation["value"].(string)
		pattern = substituteVars(pattern, vars)
		groupRaw := expectation["group"]
		expectedValRaw := expectation["expected"]

		meta["expect_expected"] = pattern
		re, err := regexp.Compile(pattern)
		if err != nil {
			return false, fmt.Sprintf("Invalid regex pattern: %v", err), meta
		}
		body := string(result.ResponseBody)
		matches := re.FindStringSubmatch(body)
		if len(matches) > 0 {
			group := 0
			if groupRaw != nil {
				group = intFromAny(groupRaw)
			}
			if len(matches) > group {
				actual := matches[group]
				if expectedValRaw != nil {
					expected := substituteVars(expectedValRaw.(string), vars)
					if actual != expected {
						return false, fmt.Sprintf("Regex group %d: expected %q, got %q", group, expected, actual), meta
					}
				}
				return true, "ok", meta
			}
			return false, fmt.Sprintf("Regex matched but group %d not found", group), meta
		}
		return false, fmt.Sprintf("Body did not match regex: %q", pattern), meta
	}
	return true, "ok", nil
}
