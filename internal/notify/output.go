package notify

import (
	"chronix/pkg/typeutil"
	"fmt"
	"strings"

	"gorm.io/datatypes"
)

func formatOutputForEmail(output map[string]any) string {
	steps := asSlice(output["steps"])
	if len(steps) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\nJob Output:\n")
	for _, sVal := range steps {
		s, ok := asMap(sVal)
		if !ok {
			continue
		}
		name := ""
		if n, ok := s["name"].(string); ok {
			name = n
		}
		status := ""
		if st, ok := s["status"].(string); ok {
			status = st
		}
		fmt.Fprintf(&sb, "\n### Step: %s [%s]\n", name, status)

		if rc := typeutil.AsInt64(s["rows_count"]); rc > 0 {
			fmt.Fprintf(&sb, "- Rows Count: %d\n", rc)
		}
		if ra := typeutil.AsInt64(s["rows_affected"]); ra > 0 {
			fmt.Fprintf(&sb, "- Rows Affected: %d\n", ra)
		}

		if rl := asSlice(s["result_lines"]); len(rl) > 0 {
			sb.WriteString("\nResult Sample:\n")
			if firstRow, ok := asMap(rl[0]); ok {
				keys := make([]string, 0, len(firstRow))
				for k := range firstRow {
					keys = append(keys, k)
				}
				sb.WriteString("| " + strings.Join(keys, " | ") + " |\n")
				sb.WriteString("| " + strings.Repeat("--- | ", len(keys)) + "\n")
				for i, rowVal := range rl {
					if i >= 10 {
						fmt.Fprintf(&sb, "\n... (%d more rows)\n", len(rl)-10)
						break
					}
					if row, ok := asMap(rowVal); ok {
						vals := make([]string, 0, len(keys))
						for _, k := range keys {
							vals = append(vals, fmt.Sprintf("%v", row[k]))
						}
						sb.WriteString("| " + strings.Join(vals, " | ") + " |\n")
					}
				}
			}
		}

		if em, ok := s["expect_message"].(string); ok && em != "" && em != "ok" {
			fmt.Fprintf(&sb, "- Expectation: %s\n", em)
		}
		if erm, ok := s["error_message"].(string); ok && erm != "" {
			fmt.Fprintf(&sb, "- Error: %s\n", erm)
		}

		if stdout, ok := s["stdout"].(string); ok && stdout != "" {
			sb.WriteString("\nStdout:\n```\n")
			sb.WriteString(truncateString(stdout, 2000))
			sb.WriteString("\n```\n")
		}
		if stderr, ok := s["stderr"].(string); ok && stderr != "" {
			sb.WriteString("\nStderr:\n```\n")
			sb.WriteString(truncateString(stderr, 2000))
			sb.WriteString("\n```\n")
		}
		if responseBody, ok := s["response_body"].(string); ok && responseBody != "" {
			sb.WriteString("\nResponse Body:\n```\n")
			sb.WriteString(truncateString(responseBody, 2000))
			sb.WriteString("\n```\n")
		}
	}
	return sb.String()
}

func formatOutputForSMS(output map[string]any) string {
	steps := asSlice(output["steps"])
	if len(steps) == 0 {
		return ""
	}
	lastStep, ok := asMap(steps[len(steps)-1])
	if !ok {
		return ""
	}
	name := ""
	if n, ok := lastStep["name"].(string); ok {
		name = n
	}
	status := ""
	if st, ok := lastStep["status"].(string); ok {
		status = st
	}
	summary := fmt.Sprintf("Last Step: %s [%s]", name, status)
	if erm, ok := lastStep["error_message"].(string); ok && erm != "" {
		summary += fmt.Sprintf("\nErr: %s", truncateString(erm, 64))
	} else if ra := typeutil.AsInt64(lastStep["rows_affected"]); ra > 0 {
		summary += fmt.Sprintf("\nAffected: %d", ra)
	} else if rc := typeutil.AsInt64(lastStep["rows_count"]); rc > 0 {
		summary += fmt.Sprintf("\nRows: %d", rc)
	}
	return summary
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "\n... (truncated)"
}

func asMap(v any) (map[string]any, bool) {
	if v == nil {
		return nil, false
	}
	if m, ok := v.(map[string]any); ok {
		return m, true
	}
	if m, ok := v.(datatypes.JSONMap); ok {
		return map[string]any(m), true
	}
	return nil, false
}

func asSlice(v any) []any {
	if v == nil {
		return nil
	}
	if s, ok := v.([]any); ok {
		return s
	}
	if s, ok := v.([]map[string]any); ok {
		out := make([]any, len(s))
		for i, m := range s {
			out[i] = m
		}
		return out
	}
	return nil
}
