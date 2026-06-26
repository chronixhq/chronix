package main

import (
	utilpkg "chronix-agent/agentutil"
	"context"
	"time"

	"github.com/gorilla/websocket"
)

func safeStr(p *string) string { return utilpkg.SafeStr(p) }

func safeInt(p *int, def int) int { return utilpkg.SafeInt(p, def) }

func randFloat() float64 { return utilpkg.RandFloat() }

func parseServerFlag(in string) (host string, port int) {
	return utilpkg.ParseServerFlag(in, DefaultAgentPort)
}

func waitWithContext(ctx context.Context, d time.Duration) error {
	return utilpkg.WaitWithContext(ctx, d)
}

func maskDSN(driver, dsn string) string { return utilpkg.MaskDSN(driver, dsn) }

func sqlOpAndPreview(sqlText string) (op, preview string) { return utilpkg.SQLOpAndPreview(sqlText) }

func argTypes(args []any) []string { return utilpkg.ArgTypes(args) }

func writeJSON(conn *websocket.Conn, v any) { utilpkg.WriteJSON(conn, v) }

func sendAgentError(conn *websocket.Conn, msgType, id, code, message string) {
	utilpkg.SendAgentError(conn, msgType, id, code, message)
}
