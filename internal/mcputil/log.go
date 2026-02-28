package mcputil

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCP logging levels (RFC-5424 syslog levels as strings)
const (
	LogLevelDebug   mcp.LoggingLevel = "debug"
	LogLevelInfo    mcp.LoggingLevel = "info"
	LogLevelWarning mcp.LoggingLevel = "warning"
	LogLevelError   mcp.LoggingLevel = "error"
)

// getSession safely extracts the session from a request, returning nil if unavailable.
func getSession(req *mcp.CallToolRequest) *mcp.ServerSession {
	if req == nil {
		return nil
	}
	return req.Session
}

// SessionLog sends a log message to the MCP client.
// If request or session is nil, or logging fails, the error is silently ignored.
// This allows logging to be optional and non-blocking.
func SessionLog(ctx context.Context, req *mcp.CallToolRequest, level mcp.LoggingLevel, msg string, data map[string]any) {
	session := getSession(req)
	if session == nil {
		return
	}

	logData := data
	if logData == nil {
		logData = make(map[string]any)
	}
	logData["message"] = msg

	// Log errors are ignored - logging is best-effort
	_ = session.Log(ctx, &mcp.LoggingMessageParams{
		Level:  level,
		Logger: "mcp-helm",
		Data:   logData,
	})
}

// SessionLogInfo sends an info-level log to the client.
func SessionLogInfo(ctx context.Context, req *mcp.CallToolRequest, msg string, data map[string]any) {
	SessionLog(ctx, req, LogLevelInfo, msg, data)
}

// SessionLogDebug sends a debug-level log to the client.
func SessionLogDebug(ctx context.Context, req *mcp.CallToolRequest, msg string, data map[string]any) {
	SessionLog(ctx, req, LogLevelDebug, msg, data)
}

// SessionLogWarning sends a warning-level log to the client.
func SessionLogWarning(ctx context.Context, req *mcp.CallToolRequest, msg string, data map[string]any) {
	SessionLog(ctx, req, LogLevelWarning, msg, data)
}

// SessionLogError sends an error-level log to the client.
func SessionLogError(ctx context.Context, req *mcp.CallToolRequest, msg string, data map[string]any) {
	SessionLog(ctx, req, LogLevelError, msg, data)
}
