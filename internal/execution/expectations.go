package execution

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"chronix/pkg/sqlutil"

	"github.com/Knetic/govaluate"
)

func isSelectLike(sqlText string) bool {
	s := strings.TrimSpace(strings.ToLower(sqlText))
	return strings.HasPrefix(s, "select") || strings.HasPrefix(s, "with") || strings.HasPrefix(s, "show ")
}

// bindParams replaces ${var} or {{var}} placeholders with positional placeholders and collects args.
// engine-aware to use correct placeholders (e.g. ? for mysql/sqlite, $1 for postgres, @p1 for sqlserver).
func bindParams(driver string, sqlText string, vars map[string]any) (string, []any, error) {
	txt := sqlText
	args := make([]any, 0, 8)
	idx := 1
	for {
		brace1 := strings.Index(txt, "${")
		brace2 := strings.Index(txt, "{{")
		if brace1 == -1 && brace2 == -1 {
			break
		}
		var start int
		useDouble := false
		if brace1 != -1 && (brace2 == -1 || brace1 < brace2) {
			start = brace1
		} else {
			start = brace2
			useDouble = true
		}

		if !useDouble {
			endOffset := strings.Index(txt[start+2:], "}")
			if endOffset == -1 {
				return "", nil, fmt.Errorf("unclosed var placeholder")
			}
			end := start + 2 + endOffset
			key := strings.TrimSpace(txt[start+2 : end])
			val, ok := vars[key]
			if !ok {
				return "", nil, fmt.Errorf("missing variable: %s", key)
			}
			ph := sqlutil.Placeholder(driver, idx)
			args = append(args, val)
			txt = txt[:start] + ph + txt[end+1:]
			idx++
		} else {
			endOffset := strings.Index(txt[start+2:], "}}")
			if endOffset == -1 {
				return "", nil, fmt.Errorf("unclosed var placeholder")
			}
			end := start + 2 + endOffset
			key := strings.TrimSpace(txt[start+2 : end])
			val, ok := vars[key]
			if !ok {
				return "", nil, fmt.Errorf("missing variable: %s", key)
			}
			ph := sqlutil.Placeholder(driver, idx)
			args = append(args, val)
			txt = txt[:start] + ph + txt[end+2:]
			idx++
		}
	}
	return txt, args, nil
}

// evaluateExpectation checks step outcomes against the expectation map.
func evaluateExpectation(expMap map[string]any, rowsCount int, rowsAffected int64, resultLines []map[string]any, vars map[string]any) (bool, string, map[string]any, error) {
	if kRaw, ok := expMap["kind"]; ok {
		kind, _ := kRaw.(string)
		switch kind {
		case "none":
			return true, "", map[string]any{"expect_kind": "none"}, nil
		case "noError":
			return true, "", map[string]any{"expect_kind": "noError"}, nil
		case "rowExists":
			if rowsCount >= 1 {
				return true, "", map[string]any{"expect_kind": "rowExists"}, nil
			}
			return false, "At least one row was expected, but no rows were returned.", map[string]any{"expect_kind": "rowExists"}, nil
		case "noRowsReturned":
			if rowsCount == 0 {
				return true, "", map[string]any{"expect_kind": "noRowsReturned", "expect_actual": rowsCount}, nil
			}
			return false, fmt.Sprintf("Zero rows were expected, but the query returned %d row(s).", rowsCount), map[string]any{"expect_kind": "noRowsReturned", "expect_actual": rowsCount}, nil
		case "fieldEqualsFirst", "fieldEqualsLast", "fieldEquals":
			col, _ := expMap["column"].(string)
			exp, _ := expMap["expected"].(string)
			exp = substituteVars(exp, vars)
			rowSel := "first"
			if kind == "fieldEqualsLast" {
				rowSel = "last"
			}
			meta := map[string]any{"expect_kind": "fieldEquals", "expect_row": rowSel}
			if col == "" {
				return false, "Expectation requires a column name.", meta, nil
			}
			meta["expect_column"] = col
			meta["expect_expected"] = exp
			if rowsCount == 0 || len(resultLines) == 0 {
				return false, "No rows were returned; cannot evaluate the field equality.", meta, nil
			}
			var row map[string]any
			if rowSel == "last" {
				row = resultLines[len(resultLines)-1]
			} else {
				row = resultLines[0]
			}
			val, ok := row[col]
			if !ok {
				return false, fmt.Sprintf("The selected row does not have a column named '%s'.", col), meta, nil
			}
			var sval string
			switch t := val.(type) {
			case []byte:
				sval = string(t)
			default:
				sval = fmt.Sprint(t)
			}
			meta["expect_actual"] = sval
			if sval != exp {
				msg := fmt.Sprintf("The %s result row was expected to have:\n\n%s = \"%s\"\n\nWhat was received is:\n\n%s = \"%s\"", rowSel, col, exp, col, sval)
				return false, msg, meta, nil
			}
			return true, "", meta, nil
		case "rowsAffected":
			op, _ := expMap["op"].(string)
			valStr, _ := expMap["value"].(string)
			valStr = substituteVars(valStr, vars)
			var n int64
			if _, err := fmt.Sscan(valStr, &n); err != nil {
				return false, "Invalid rowsAffected expected value.", map[string]any{"expect_kind": "rowsAffected"}, nil
			}
			meta := map[string]any{"expect_kind": "rowsAffected", "expect_op": op, "expect_expected": n, "expect_actual": rowsAffected}
			if rowsAffected < 0 {
				return false, "The database driver did not report the number of rows affected for this statement. Try using a 'Success (No Error)' expectation instead.", meta, nil
			}
			switch op {
			case ">=":
				if rowsAffected >= n {
					return true, "", meta, nil
				}
				return false, fmt.Sprintf("Rows affected was %d; expected >= %d.", rowsAffected, n), meta, nil
			case "==":
				if rowsAffected == n {
					return true, "", meta, nil
				}
				return false, fmt.Sprintf("Rows affected was %d; expected == %d.", rowsAffected, n), meta, nil
			case "<=":
				if rowsAffected <= n {
					return true, "", meta, nil
				}
				return false, fmt.Sprintf("Rows affected was %d; expected <= %d.", rowsAffected, n), meta, nil
			default:
				return false, "Invalid rowsAffected operator.", meta, nil
			}
		}
	}

	if exp, ok := expMap["assert"]; ok {
		if s, ok := exp.(string); ok {
			s = substituteVars(s, vars)
			params := map[string]any{"rows_count": rowsCount, "rows_affected": rowsAffected}
			expr, err := govaluate.NewEvaluableExpression(s)
			if err != nil {
				return false, "", nil, err
			}
			v, err := expr.Evaluate(params)
			if err != nil {
				return false, "", nil, err
			}
			b, ok := v.(bool)
			if !ok {
				return false, "assert expression did not return bool", map[string]any{"expect_kind": "assert"}, nil
			}
			if !b {
				return false, fmt.Sprintf("assert failed: %s", s), map[string]any{"expect_kind": "assert"}, nil
			}
			return true, "", map[string]any{"expect_kind": "assert"}, nil
		}
	}
	if v, ok := expMap["rows_min"]; ok {
		vSub := v
		if s, ok := v.(string); ok {
			vSub = substituteVars(s, vars)
		}
		minVal := intFromAny(vSub)
		if rowsCount < minVal {
			return false, fmt.Sprintf("rows_count < %d", minVal), map[string]any{"expect_kind": "rows_min", "expect_expected": minVal, "expect_actual": rowsCount}, nil
		}
	}
	if v, ok := expMap["rows_max"]; ok {
		vSub := v
		if s, ok := v.(string); ok {
			vSub = substituteVars(s, vars)
		}
		maxVal := intFromAny(vSub)
		if rowsCount > maxVal {
			return false, fmt.Sprintf("rows_count > %d", maxVal), map[string]any{"expect_kind": "rows_max", "expect_expected": maxVal, "expect_actual": rowsCount}, nil
		}
	}
	if v, ok := expMap["affected_min"]; ok {
		vSub := v
		if s, ok := v.(string); ok {
			vSub = substituteVars(s, vars)
		}
		minVal := intFromAny(vSub)
		if rowsAffected < 0 {
			return false, "The database driver did not report the number of rows affected for this statement.", map[string]any{"expect_kind": "affected_min"}, nil
		}
		if int(rowsAffected) < minVal {
			return false, fmt.Sprintf("rows_affected < %d", minVal), map[string]any{"expect_kind": "affected_min", "expect_expected": minVal, "expect_actual": rowsAffected}, nil
		}
	}
	return true, "", nil, nil
}

func evaluateShellExpectation(expMap map[string]any, exitCode int, stdout []byte, stderr []byte, outTruncated bool, errTruncated bool, truncationMode string, vars map[string]any) (bool, string, map[string]any) {
	if kRaw, ok := expMap["kind"]; ok {
		kind, _ := kRaw.(string)
		switch kind {
		case "none":
			return true, "", map[string]any{"expect_kind": "none"}
		case "noError":
			meta := map[string]any{"expect_kind": "noError", "expect_actual": exitCode}
			if exitCode == 0 {
				return true, "", meta
			}
			return false, fmt.Sprintf("Expected exit code 0, but got %d", exitCode), meta
		case "exitCodeEquals":
			valRaw := expMap["value"]
			valStr := ""
			if s, ok := valRaw.(string); ok {
				valStr = substituteVars(s, vars)
			} else {
				valStr = fmt.Sprint(valRaw)
			}
			expectedVal := intFromAny(valStr)
			meta := map[string]any{"expect_kind": "exitCodeEquals", "expect_expected": expectedVal, "expect_actual": exitCode}
			if exitCode == expectedVal {
				return true, "", meta
			}
			return false, fmt.Sprintf("Expected exit code %d, but got %d", expectedVal, exitCode), meta
		case "contains":
			expected, _ := expMap["value"].(string)
			expected = substituteVars(expected, vars)
			meta := map[string]any{"expect_kind": "contains", "expect_expected": expected, "expect_truncated": outTruncated || errTruncated}
			combined := string(stdout) + string(stderr)
			if strings.Contains(combined, expected) {
				return true, "", meta
			}
			return false, fmt.Sprintf("Output did not contain expected string: %q", expected), meta
		case "notContains":
			expected, _ := expMap["value"].(string)
			expected = substituteVars(expected, vars)
			meta := map[string]any{"expect_kind": "notContains", "expect_expected": expected, "expect_truncated": outTruncated || errTruncated}
			combined := string(stdout) + string(stderr)
			if !strings.Contains(combined, expected) {
				return true, "", meta
			}
			return false, fmt.Sprintf("Output contained forbidden string: %q", expected), meta
		case "firstLineEquals":
			expected, _ := expMap["value"].(string)
			expected = substituteVars(expected, vars)
			meta := map[string]any{"expect_kind": "firstLineEquals", "expect_expected": expected, "expect_truncated": outTruncated}
			outStr := strings.TrimSpace(string(stdout))
			if outStr == "" {
				if outTruncated {
					return true, "", meta
				}
				return false, "Output is empty", meta
			}
			lines := strings.Split(outStr, "\n")
			actual := strings.TrimSpace(lines[0])
			expected = strings.TrimSpace(expected)
			if outTruncated && truncationMode == "head" && len(lines) == 1 {
				if strings.HasPrefix(expected, actual) {
					return true, "", meta
				}
			} else if actual == expected {
				return true, "", meta
			}
			return false, fmt.Sprintf("First line mismatch.\nExpected: %q\nGot: %q", expected, actual), meta
		case "lastLineEquals":
			expected, _ := expMap["value"].(string)
			expected = substituteVars(expected, vars)
			meta := map[string]any{"expect_kind": "lastLineEquals", "expect_expected": expected, "expect_truncated": outTruncated}
			outStr := strings.TrimSpace(string(stdout))
			if outStr == "" {
				if outTruncated {
					return true, "", meta
				}
				return false, "Output is empty", meta
			}
			lines := strings.Split(outStr, "\n")
			actual := strings.TrimSpace(lines[len(lines)-1])
			expected = strings.TrimSpace(expected)
			if outTruncated && truncationMode == "tail" && len(lines) == 1 {
				if strings.HasSuffix(expected, actual) {
					return true, "", meta
				}
			} else if actual == expected {
				return true, "", meta
			}
			return false, fmt.Sprintf("Last line mismatch.\nExpected: %q\nGot: %q", expected, actual), meta
		case "regexMatch":
			pattern, _ := expMap["value"].(string)
			pattern = substituteVars(pattern, vars)
			groupRaw := expMap["group"]
			expectedValRaw := expMap["expected"]
			meta := map[string]any{"expect_kind": "regexMatch", "expect_expected": pattern, "expect_truncated": outTruncated || errTruncated}
			re, err := regexp.Compile(pattern)
			if err != nil {
				return false, fmt.Sprintf("Invalid regex pattern: %v", err), meta
			}
			combined := string(stdout) + string(stderr)
			matches := re.FindStringSubmatch(combined)
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
					return true, "", meta
				}
				return false, fmt.Sprintf("Regex matched but group %d not found", group), meta
			}
			return false, fmt.Sprintf("Output did not match regex: %q", pattern), meta
		}
	}
	return true, "", nil
}

func intFromAny(v any) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	case json.Number:
		i, _ := t.Int64()
		return int(i)
	case string:
		var i int
		if _, err := fmt.Sscan(t, &i); err != nil {
			return 0
		}
		return i
	default:
		return 0
	}
}
