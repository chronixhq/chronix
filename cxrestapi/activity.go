package cxrestapi

import (
	"encoding/csv"
	"html/template"
	"time"

	cxuserpkg "chronix/internal/cxuser"
	"chronix/internal/db"
	"fmt"
	"strconv"
	"strings"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"
)

// activityRouter wires the unified activity endpoint.
func activityRouter(app *gin.Engine) {
	app.GET("/activity", getUnifiedActivity)
	app.GET("/activity/export", exportActivityReport)
}

type unifiedActivityItem struct {
	ID      string  `json:"id"`
	When    string  `json:"when"`
	Action  string  `json:"action"`
	Details *string `json:"details,omitempty"`
	UserID  *int64  `json:"userId,omitempty"`
	User    *string `json:"user,omitempty"`
}

// aggregateUnifiedActivity composes a merged, sorted timeline. Returns up to fetchCap items.
func aggregateUnifiedActivity(current cxuserpkg.CxUser, fetchCap int) []unifiedActivityItem {
	items, _, _ := getUnifiedActivityInternal(current, fetchCap, 0, "", "", "", "", "")
	return items
}

func getUnifiedActivityInternal(current cxuserpkg.CxUser, limit, offset int, q, action, userFilter, from, to string) ([]unifiedActivityItem, int64, error) {
	// Base queries
	jobPart := `SELECT 'run:' || jr.run_uid AS id, COALESCE(jr.finished_at, jr.started_at, jr.queued_at) AS "when", COALESCE(jr.job_name, 'job') || ' ' || COALESCE(jr.status, '') AS action, NULL AS details, jr.triggered_by AS user_id, COALESCE(u.name, CASE WHEN jr.triggered_by = 0 THEN 'Chronix System' ELSE 'User ' || jr.triggered_by END) AS user 
	            FROM job_runs jr LEFT JOIN cx_users u ON u.id = jr.triggered_by`
	userPart := `SELECT 'ua:' || ua.id AS id, ua.created_at AS "when", ua.action, ua.details, ua.user_id, COALESCE(u.name, CASE WHEN ua.user_id = 0 THEN 'Chronix System' ELSE 'User ' || ua.user_id END) AS user 
	             FROM user_activity ua LEFT JOIN cx_users u ON u.id = ua.user_id`

	// For non-admins, limit user activity to their own
	var userPartArgs []interface{}
	if !current.Admin {
		userPart += " WHERE ua.user_id = ?"
		userPartArgs = append(userPartArgs, current.ID)
	}

	combinedSQL := fmt.Sprintf("%s UNION ALL %s", jobPart, userPart)

	var whereClauses []string
	var args []interface{}
	args = append(args, userPartArgs...)

	if q != "" {
		whereClauses = append(whereClauses, "(combined.action LIKE ? OR combined.details LIKE ? OR combined.user LIKE ?)")
		term := "%" + q + "%"
		args = append(args, term, term, term)
	}
	if action != "" {
		whereClauses = append(whereClauses, "combined.action LIKE ?")
		args = append(args, "%"+action+"%")
	}
	if userFilter != "" {
		whereClauses = append(whereClauses, "(combined.user LIKE ? OR CAST(combined.user_id AS TEXT) = ?)")
		args = append(args, "%"+userFilter+"%", userFilter)
	}
	if from != "" {
		whereClauses = append(whereClauses, "combined.\"when\" >= ?")
		args = append(args, from)
	}
	if to != "" {
		whereClauses = append(whereClauses, "combined.\"when\" <= ?")
		args = append(args, to)
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Count total
	var total int64
	countSQL := "SELECT COUNT(*) FROM (" + combinedSQL + ") AS combined" + whereSQL
	if err := db.DB.Raw(countSQL, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch items
	var items []unifiedActivityItem
	fetchSQL := "SELECT * FROM (" + combinedSQL + ") AS combined" + whereSQL + " ORDER BY combined.\"when\" DESC"
	if limit > 0 {
		fetchSQL += " LIMIT ? OFFSET ?"
		args = append(args, limit, offset)
	}
	if err := db.DB.Raw(fetchSQL, args...).Scan(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// getUnifiedActivity aggregates system/job events and user activity into a single timeline.
// Query: limit, offset, q, action, user, from, to
// Auth:
//   - Admins see system/job events + all users' activity
//   - Non-admins see system/job events + their own activity only
func getUnifiedActivity(c *gin.Context) {
	current := userFromGinContext(c)

	limit := 100
	if ls := c.Query("limit"); ls != "" {
		if n, err := strconv.Atoi(ls); err == nil && n > 0 {
			limit = n
		}
	}
	offset := 0
	if os := c.Query("offset"); os != "" {
		if n, err := strconv.Atoi(os); err == nil && n >= 0 {
			offset = n
		}
	}
	if limit > 500 {
		limit = 500
	}

	q := c.Query("q")
	action := c.Query("action")
	userFilter := c.Query("user")
	from := c.Query("from")
	to := c.Query("to")

	items, total, err := getUnifiedActivityInternal(*current, limit, offset, q, action, userFilter, from, to)
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to fetch activity", err.Error())
		return
	}

	restresponse.RestSuccess(c, gin.H{
		"items": items,
		"total": total,
	})
}

func exportActivityReport(c *gin.Context) {
	current := userFromGinContext(c)
	format := strings.ToLower(c.DefaultQuery("format", "csv"))

	switch format {
	case "csv", "html", "pdf":
	default:
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Unsupported format")
		return
	}

	q := c.Query("q")
	action := c.Query("action")
	userFilter := c.Query("user")
	from := c.Query("from")
	to := c.Query("to")

	items, _, err := getUnifiedActivityInternal(*current, 0, 0, q, action, userFilter, from, to)
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to fetch activity for report", err.Error())
		return
	}

	switch format {
	case "csv":
		generateCSVReport(c, items)
	case "html":
		generateHTMLReport(c, items)
	case "pdf":
		generatePDFReport(c, items)
	}
}

func generatePDFReport(c *gin.Context, items []unifiedActivityItem) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Activity Report")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 10, "Generated on: "+time.Now().Format(time.RFC1123))
	pdf.Ln(12)

	// Table header
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(45, 10, "When", "1", 0, "L", true, 0, "")
	pdf.CellFormat(35, 10, "User", "1", 0, "L", true, 0, "")
	pdf.CellFormat(50, 10, "Action", "1", 0, "L", true, 0, "")
	pdf.CellFormat(60, 10, "Details", "1", 0, "L", true, 0, "")
	pdf.Ln(-1)

	// Table rows
	pdf.SetFont("Arial", "", 9)
	for _, it := range items {
		user := ""
		if it.User != nil {
			user = *it.User
		}
		details := ""
		if it.Details != nil {
			details = *it.Details
		}

		// Simple truncation for PDF cells to avoid overflow issues in basic implementation
		if len(details) > 40 {
			details = details[:37] + "..."
		}

		pdf.CellFormat(45, 8, it.When, "1", 0, "L", false, 0, "")
		pdf.CellFormat(35, 8, user, "1", 0, "L", false, 0, "")
		pdf.CellFormat(50, 8, it.Action, "1", 0, "L", false, 0, "")
		pdf.CellFormat(60, 8, details, "1", 0, "L", false, 0, "")
		pdf.Ln(-1)
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment;filename=activity_report.pdf")
	_ = pdf.Output(c.Writer)
}

func generateCSVReport(c *gin.Context, items []unifiedActivityItem) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment;filename=activity_report.csv")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	_ = writer.Write([]string{"When", "User", "Action", "Details"})

	for _, it := range items {
		user := ""
		if it.User != nil {
			user = *it.User
		}
		details := ""
		if it.Details != nil {
			details = *it.Details
		}
		_ = writer.Write([]string{it.When, user, it.Action, details})
	}
}

func generateHTMLReport(c *gin.Context, items []unifiedActivityItem) {
	const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Activity Report</title>
    <style>
        body { font-family: sans-serif; margin: 20px; }
        table { width: 100%; border-collapse: collapse; margin-top: 20px; }
        th, td { border: 1px solid #ccc; padding: 8px; text-align: left; }
        th { background-color: #f4f4f4; }
        tr:nth-child(even) { background-color: #f9f9f9; }
    </style>
</head>
<body>
    <h1>Activity Report</h1>
    <p>Generated on: {{.GeneratedAt}}</p>
    <table>
        <thead>
            <tr>
                <th>When</th>
                <th>User</th>
                <th>Action</th>
                <th>Details</th>
            </tr>
        </thead>
        <tbody>
            {{range .Items}}
            <tr>
                <td>{{.When}}</td>
                <td>{{if .User}}{{.User}}{{end}}</td>
                <td>{{.Action}}</td>
                <td>{{if .Details}}{{.Details}}{{end}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>
</body>
</html>`

	tmpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to parse template")
		return
	}

	c.Header("Content-Type", "text/html")
	c.Header("Content-Disposition", "attachment;filename=activity_report.html")

	data := struct {
		GeneratedAt string
		Items       []unifiedActivityItem
	}{
		GeneratedAt: time.Now().Format(time.RFC1123),
		Items:       items,
	}

	_ = tmpl.Execute(c.Writer, data)
}
