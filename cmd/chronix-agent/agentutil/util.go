package agentutil

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/url"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func SafeStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func SafeInt(p *int, def int) int {
	if p == nil {
		return def
	}
	return *p
}

func RandFloat() float64 {
	var b [1]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0.5
	}
	return float64(b[0]) / 255.0
}

func ParseServerFlag(in string, defaultPort int) (host string, port int) {
	in = strings.TrimSpace(in)
	if in == "" {
		in = "localhost"
	}
	if strings.Contains(in, ":") {
		parts := strings.Split(in, ":")
		host = parts[0]
		_, _ = fmt.Sscanf(parts[1], "%d", &port)
		if port == 0 {
			port = defaultPort
		}
	} else {
		host = in
		port = defaultPort
	}
	return
}

func URLQueryEscape(s string) string {
	return url.QueryEscape(s)
}

func WaitWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func MaskDSN(driver, dsn string) string {
	d := strings.ToLower(strings.TrimSpace(driver))
	if d == "sqlite" {
		base := filepath.Base(strings.TrimSpace(dsn))
		if base == "." || base == "" {
			return "sqlite:(path)"
		}
		return "sqlite:" + base
	}
	if strings.Contains(dsn, "://") {
		if u, err := url.Parse(dsn); err == nil && u.Scheme != "" {
			if u.User != nil {
				username := u.User.Username()
				u.User = url.UserPassword(username, "***")
			}
			q := u.Query()
			for _, k := range []string{"password", "pwd", "pass", "secret", "token"} {
				if q.Has(k) {
					q.Set(k, "***")
				}
			}
			u.RawQuery = q.Encode()
			return strings.ReplaceAll(u.String(), "%2A%2A%2A", "***")
		}
	}
	out := dsn
	if i := strings.Index(out, "@"); i > 0 {
		start := 0
		if j := strings.Index(out, "://"); j >= 0 {
			start = j + 3
		}
		cred := out[start:i]
		if k := strings.Index(cred, ":"); k >= 0 {
			cred = cred[:k] + ":***"
			out = out[:start] + cred + out[i:]
		}
	}
	lower := strings.ToLower(out)
	for _, key := range []string{"password", "pwd", "pass", "secret", "token"} {
		needle := key + "="
		offset := 0
		for {
			idx := strings.Index(lower[offset:], needle)
			if idx == -1 {
				break
			}
			idx += offset
			start := idx + len(needle)
			end := len(out)
			for i := start; i < len(out); i++ {
				if out[i] == ';' || out[i] == ' ' || out[i] == '\n' || out[i] == '\t' {
					end = i
					break
				}
			}
			out = out[:idx] + key + "=***" + out[end:]
			lower = strings.ToLower(out)
			offset = idx + len(key) + 4
		}
	}
	return out
}

func SQLOpAndPreview(sqlText string) (op, preview string) {
	s := strings.TrimSpace(sqlText)
	low := strings.ToLower(s)
	op = "sql"
	for _, k := range []string{"select", "with", "insert", "update", "delete", "show", "pragma", "alter", "create", "drop"} {
		if strings.HasPrefix(low, k) {
			op = k
			break
		}
	}
	preview = compactSpaces(s)
	if len(preview) > 120 {
		preview = preview[:120] + "…"
	}
	return
}

func compactSpaces(s string) string {
	parts := strings.Fields(s)
	return strings.Join(parts, " ")
}

func ArgTypes(args []any) []string {
	if len(args) == 0 {
		return nil
	}
	res := make([]string, 0, len(args))
	for _, a := range args {
		if a == nil {
			res = append(res, "null")
			continue
		}
		t := reflect.TypeOf(a)
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		if t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8 {
			res = append(res, "[]byte")
			continue
		}
		res = append(res, t.String())
	}
	return res
}

func WriteJSON(conn *websocket.Conn, v any) {
	if conn == nil {
		return
	}
	_ = conn.WriteJSON(v)
}

func SendAgentError(conn *websocket.Conn, msgType, id, code, message string) {
	WriteJSON(conn, struct {
		Type    string `json:"type"`
		ID      string `json:"id"`
		Payload any    `json:"payload"`
	}{
		Type: msgType,
		ID:   id,
		Payload: struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}{Code: code, Message: message},
	})
}
