package notify

import (
	"chronix/pkg/typeutil"
	"fmt"
	"html"
	"sort"
	"strings"
	"time"

	"chronix/internal/db/models"
)

func generateEmailHTML(n *models.Notification) string {
	severity := strings.ToLower(n.Severity)
	headerColor := "#2196f3"
	switch severity {
	case "success":
		headerColor = "#4caf50"
	case "error":
		headerColor = "#f44336"
	case "warning":
		headerColor = "#ff9800"
	}

	var sb strings.Builder
	sb.WriteString("<!DOCTYPE html><html><head><style>")
	sb.WriteString("body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f4f7f9; }")
	sb.WriteString(".container { max-width: 800px; margin: 20px auto; padding: 20px; }")
	sb.WriteString(".card { background: #fff; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); overflow: hidden; }")
	sb.WriteString(".header { padding: 20px; color: #fff; font-size: 24px; font-weight: bold; }")
	sb.WriteString(".content { padding: 30px; }")
	sb.WriteString(".section-title { font-size: 18px; font-weight: bold; margin-top: 25px; margin-bottom: 10px; border-bottom: 2px solid #eee; padding-bottom: 5px; }")
	sb.WriteString(".details-table { width: 100%; border-collapse: collapse; margin-bottom: 20px; }")
	sb.WriteString(".details-table td { padding: 8px 12px; border-bottom: 1px solid #eee; }")
	sb.WriteString(".details-table td.label { font-weight: bold; width: 150px; color: #666; }")
	sb.WriteString(".step-card { border: 1px solid #e0e0e0; border-radius: 4px; margin-bottom: 15px; background: #fafafa; }")
	sb.WriteString(".step-header { padding: 10px 15px; border-bottom: 1px solid #e0e0e0; font-weight: bold; display: flex; justify-content: space-between; }")
	sb.WriteString(".step-body { padding: 15px; }")
	sb.WriteString(".status-badge { padding: 2px 8px; border-radius: 12px; font-size: 12px; text-transform: uppercase; font-weight: bold; }")
	sb.WriteString(".status-success { background: #e8f5e9; color: #2e7d32; }")
	sb.WriteString(".status-error { background: #ffebee; color: #c62828; }")
	sb.WriteString(".result-table { width: 100%; border-collapse: collapse; margin-top: 10px; font-size: 13px; }")
	sb.WriteString(".result-table th { background: #f0f0f0; padding: 8px; text-align: left; border: 1px solid #ddd; }")
	sb.WriteString(".result-table td { padding: 8px; border: 1px solid #ddd; }")
	sb.WriteString(".code-block { background: #2d2d2d; color: #f8f8f2; padding: 15px; border-radius: 4px; font-family: monospace; font-size: 13px; white-space: pre-wrap; overflow-x: auto; margin: 10px 0; }")
	sb.WriteString(".footer { text-align: center; padding: 20px; font-size: 12px; color: #888; }")
	sb.WriteString("</style></head><body>")

	sb.WriteString("<div class='container'><div class='card'>")

	title := n.Subject
	if n.Category == string(CategoryJob) && n.Data != nil {
		jobName, _ := (*n.Data)["job_name"].(string)
		status, _ := (*n.Data)["status"].(string)
		if status == "" {
			status = n.Severity
		}
		if jobName != "" && status != "" {
			title = fmt.Sprintf("Job: %s — %s", jobName, strings.ToUpper(status))
		}
	} else if n.Category == string(CategoryConnection) && n.Data != nil {
		connName, _ := (*n.Data)["connection_name"].(string)
		status, _ := (*n.Data)["status"].(string)
		if status == "" {
			status = n.Severity
		}
		if connName != "" && status != "" {
			title = fmt.Sprintf("Connection: %s — %s", connName, strings.ToUpper(status))
		}
	}

	fmt.Fprintf(&sb, "<div class='header' style='background-color: %s;'>%s</div>", headerColor, html.EscapeString(title))

	sb.WriteString("<div class='content'>")
	if n.Origin != nil {
		fmt.Fprintf(&sb, "<p><strong>Origin:</strong> %s</p>", html.EscapeString(*n.Origin))
	}

	if n.Data != nil {
		sb.WriteString("<div class='section-title'>Details</div>")
		sb.WriteString("<table class='details-table'>")

		keys := make([]string, 0, len(*n.Data))
		for k := range *n.Data {
			if k != "output" {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)

		for _, k := range keys {
			v := (*n.Data)[k]
			label := k
			switch k {
			case "job_id":
				label = "Job ID"
			case "job_name":
				label = "Job Name"
			case "run_id":
				label = "Run ID"
			case "status":
				label = "Status"
			case "message":
				label = "Message"
			case "finished_at":
				label = "Finished At"
			}
			fmt.Fprintf(&sb, "<tr><td class='label'>%s</td><td>%s</td></tr>", html.EscapeString(label), html.EscapeString(fmt.Sprintf("%v", v)))
		}
		sb.WriteString("</table>")

		if outputVal, ok := (*n.Data)["output"]; ok {
			if output, ok := asMap(outputVal); ok {
				sb.WriteString(formatOutputForEmailHTML(output))
			}
		}
	}

	sb.WriteString("</div></div>")
	sb.WriteString("<div class='footer'>Sent by Chronix &bull; " + time.Now().Format("2006-01-02 15:04:05") + "</div>")
	sb.WriteString("</div></body></html>")

	return sb.String()
}

func formatOutputForEmailHTML(output map[string]any) string {
	steps := asSlice(output["steps"])
	if len(steps) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("<div class='section-title'>Job Output</div>")

	for _, sVal := range steps {
		s, ok := asMap(sVal)
		if !ok {
			continue
		}
		name := ""
		if n, ok := s["name"].(string); ok {
			name = n
		}
		status := "unknown"
		if st, ok := s["status"].(string); ok {
			status = st
		}

		statusClass := "status-success"
		if status == "error" || status == "failed" {
			statusClass = "status-error"
		}

		sb.WriteString("<div class='step-card'>")
		sb.WriteString("<div class='step-header'>")
		fmt.Fprintf(&sb, "<span>%s</span>", html.EscapeString(name))
		fmt.Fprintf(&sb, "<span class='status-badge %s'>%s</span>", statusClass, html.EscapeString(status))
		sb.WriteString("</div>")
		sb.WriteString("<div class='step-body'>")

		if rc := typeutil.AsInt64(s["rows_count"]); rc > 0 {
			fmt.Fprintf(&sb, "<div><strong>Rows Count:</strong> %d</div>", rc)
		}
		if ra := typeutil.AsInt64(s["rows_affected"]); ra > 0 {
			fmt.Fprintf(&sb, "<div><strong>Rows Affected:</strong> %d</div>", ra)
		}

		if rl := asSlice(s["result_lines"]); len(rl) > 0 {
			sb.WriteString("<div style='margin-top: 10px; font-weight: bold;'>Result Sample:</div>")
			if firstRow, ok := asMap(rl[0]); ok {
				sb.WriteString("<table class='result-table'><thead><tr>")
				keys := make([]string, 0, len(firstRow))
				for k := range firstRow {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					fmt.Fprintf(&sb, "<th>%s</th>", html.EscapeString(k))
				}
				sb.WriteString("</tr></thead><tbody>")

				for i, rowVal := range rl {
					if i >= 10 {
						fmt.Fprintf(&sb, "<tr><td colspan='%d' style='text-align: center; color: #888;'>... (%d more rows)</td></tr>", len(keys), len(rl)-10)
						break
					}
					if row, ok := asMap(rowVal); ok {
						sb.WriteString("<tr>")
						for _, k := range keys {
							fmt.Fprintf(&sb, "<td>%s</td>", html.EscapeString(fmt.Sprintf("%v", row[k])))
						}
						sb.WriteString("</tr>")
					}
				}
				sb.WriteString("</tbody></table>")
			}
		}

		if em, ok := s["expect_message"].(string); ok && em != "" && em != "ok" {
			fmt.Fprintf(&sb, "<div style='margin-top: 5px;'><strong>Expectation:</strong> %s</div>", html.EscapeString(em))
		}
		if erm, ok := s["error_message"].(string); ok && erm != "" {
			fmt.Fprintf(&sb, "<div style='margin-top: 5px; color: #d32f2f;'><strong>Error:</strong> %s</div>", html.EscapeString(erm))
		}

		if stdout, ok := s["stdout"].(string); ok && stdout != "" {
			sb.WriteString("<div style='margin-top: 10px; font-weight: bold;'>Stdout:</div>")
			fmt.Fprintf(&sb, "<div class='code-block'>%s</div>", html.EscapeString(truncateString(stdout, 2000)))
		}
		if stderr, ok := s["stderr"].(string); ok && stderr != "" {
			sb.WriteString("<div style='margin-top: 10px; font-weight: bold;'>Stderr:</div>")
			fmt.Fprintf(&sb, "<div class='code-block'>%s</div>", html.EscapeString(truncateString(stderr, 2000)))
		}
		if responseBody, ok := s["response_body"].(string); ok && responseBody != "" {
			sb.WriteString("<div style='margin-top: 10px; font-weight: bold;'>Response Body:</div>")
			fmt.Fprintf(&sb, "<div class='code-block'>%s</div>", html.EscapeString(truncateString(responseBody, 2000)))
		}

		sb.WriteString("</div></div>")
	}

	return sb.String()
}
