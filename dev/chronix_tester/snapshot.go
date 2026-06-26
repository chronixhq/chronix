package main

import (
	"database/sql"
	"fmt"
)

type SnapshotSummary struct {
	DataDir             string `json:"dataDir"`
	ResultsDB           string `json:"resultsDB"`
	TargetDB            string `json:"targetDB"`
	TargetActivityCount int    `json:"targetActivityCount"`
	APILogCount         int    `json:"apiLogCount"`
	WebhookLogCount     int    `json:"webhookLogCount"`
	IMAPLogCount        int    `json:"imapLogCount"`
	ShellLogCount       int    `json:"shellLogCount"`
}

type ResultsSnapshot struct {
	Summary        SnapshotSummary     `json:"summary"`
	TargetActivity []TargetActivityLog `json:"targetActivity"`
	APILogs        []RequestLog        `json:"apiLogs"`
	WebhookLogs    []WebhookLog        `json:"webhookLogs"`
	IMAPLogs       []IMAPLog           `json:"imapLogs"`
	ShellLogs      []ShellLog          `json:"shellLogs"`
}

type TargetActivityLog struct {
	ID        int    `json:"id"`
	TableName string `json:"tableName"`
	Operation string `json:"operation"`
	OldData   string `json:"oldData"`
	NewData   string `json:"newData"`
	CreatedAt string `json:"createdAt"`
}

type RequestLog struct {
	ID        int    `json:"id"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Headers   string `json:"headers"`
	Body      string `json:"body"`
	CreatedAt string `json:"createdAt"`
}

type WebhookLog struct {
	ID        int    `json:"id"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Headers   string `json:"headers"`
	Body      string `json:"body"`
	CreatedAt string `json:"createdAt"`
}

type IMAPLog struct {
	ID         int    `json:"id"`
	Subject    string `json:"subject"`
	FromAddr   string `json:"fromAddr"`
	Body       string `json:"body"`
	ReceivedAt string `json:"receivedAt"`
	CreatedAt  string `json:"createdAt"`
}

type ShellLog struct {
	ID        int    `json:"id"`
	Args      string `json:"args"`
	Output    string `json:"output"`
	CreatedAt string `json:"createdAt"`
}

func (s *Store) LoadSnapshot(limit int) (ResultsSnapshot, error) {
	limit = clampLimit(limit, 50)

	summary, err := s.loadSummary()
	if err != nil {
		return ResultsSnapshot{}, err
	}

	targetActivity, err := queryTargetActivity(s.Target, limit)
	if err != nil {
		return ResultsSnapshot{}, err
	}
	apiLogs, err := queryAPILogs(s.Results, limit)
	if err != nil {
		return ResultsSnapshot{}, err
	}
	webhookLogs, err := queryWebhookLogs(s.Results, limit)
	if err != nil {
		return ResultsSnapshot{}, err
	}
	imapLogs, err := queryIMAPLogs(s.Results, limit)
	if err != nil {
		return ResultsSnapshot{}, err
	}
	shellLogs, err := queryShellLogs(s.Results, limit)
	if err != nil {
		return ResultsSnapshot{}, err
	}

	return ResultsSnapshot{
		Summary:        summary,
		TargetActivity: targetActivity,
		APILogs:        apiLogs,
		WebhookLogs:    webhookLogs,
		IMAPLogs:       imapLogs,
		ShellLogs:      shellLogs,
	}, nil
}

func (s *Store) loadSummary() (SnapshotSummary, error) {
	summary := SnapshotSummary{
		DataDir:   s.Paths.DataDir,
		ResultsDB: s.Paths.ResultsDB,
		TargetDB:  s.Paths.TargetDB,
	}

	var err error
	summary.TargetActivityCount, err = countTable(s.Target, "target_activity")
	if err != nil {
		return SnapshotSummary{}, err
	}
	summary.APILogCount, err = countTable(s.Results, "api_logs")
	if err != nil {
		return SnapshotSummary{}, err
	}
	summary.WebhookLogCount, err = countTable(s.Results, "webhook_logs")
	if err != nil {
		return SnapshotSummary{}, err
	}
	summary.IMAPLogCount, err = countTable(s.Results, "imap_logs")
	if err != nil {
		return SnapshotSummary{}, err
	}
	summary.ShellLogCount, err = countTable(s.Results, "shell_logs")
	if err != nil {
		return SnapshotSummary{}, err
	}

	return summary, nil
}

func countTable(db *sql.DB, table string) (int, error) {
	var count int
	if err := db.QueryRow(fmt.Sprintf(`SELECT COUNT(*) FROM %s`, table)).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func queryTargetActivity(db *sql.DB, limit int) ([]TargetActivityLog, error) {
	rows, err := db.Query(`
SELECT id, table_name, operation, COALESCE(old_data, ''), COALESCE(new_data, ''), COALESCE(datetime(created_at), '')
FROM target_activity
ORDER BY id DESC
LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []TargetActivityLog
	for rows.Next() {
		var log TargetActivityLog
		if err := rows.Scan(&log.ID, &log.TableName, &log.Operation, &log.OldData, &log.NewData, &log.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

func queryAPILogs(db *sql.DB, limit int) ([]RequestLog, error) {
	rows, err := db.Query(`
SELECT id, method, path, COALESCE(headers, ''), COALESCE(body, ''), COALESCE(datetime(created_at), '')
FROM api_logs
ORDER BY id DESC
LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []RequestLog
	for rows.Next() {
		var log RequestLog
		if err := rows.Scan(&log.ID, &log.Method, &log.Path, &log.Headers, &log.Body, &log.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

func queryWebhookLogs(db *sql.DB, limit int) ([]WebhookLog, error) {
	rows, err := db.Query(`
SELECT id, COALESCE(method, ''), COALESCE(path, ''), COALESCE(headers, ''), COALESCE(body, ''), COALESCE(datetime(created_at), '')
FROM webhook_logs
ORDER BY id DESC
LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []WebhookLog
	for rows.Next() {
		var log WebhookLog
		if err := rows.Scan(&log.ID, &log.Method, &log.Path, &log.Headers, &log.Body, &log.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

func queryIMAPLogs(db *sql.DB, limit int) ([]IMAPLog, error) {
	rows, err := db.Query(`
SELECT id, COALESCE(subject, ''), COALESCE(from_addr, ''), COALESCE(body, ''), COALESCE(datetime(received_at), ''), COALESCE(datetime(created_at), '')
FROM imap_logs
ORDER BY id DESC
LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []IMAPLog
	for rows.Next() {
		var log IMAPLog
		if err := rows.Scan(&log.ID, &log.Subject, &log.FromAddr, &log.Body, &log.ReceivedAt, &log.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

func queryShellLogs(db *sql.DB, limit int) ([]ShellLog, error) {
	rows, err := db.Query(`
SELECT id, COALESCE(args, ''), COALESCE(output, ''), COALESCE(datetime(created_at), '')
FROM shell_logs
ORDER BY id DESC
LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []ShellLog
	for rows.Next() {
		var log ShellLog
		if err := rows.Scan(&log.ID, &log.Args, &log.Output, &log.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

func clampLimit(value int, fallback int) int {
	if value <= 0 {
		value = fallback
	}
	if value > 200 {
		value = 200
	}
	return value
}
