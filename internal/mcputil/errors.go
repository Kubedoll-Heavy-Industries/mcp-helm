package mcputil

import (
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Kubedoll-Heavy-Industries/mcp-helm/internal/helm"
)

// OperationError provides structured error context for handler operations.
type OperationError struct {
	Op         string // operation name (e.g., "get_values", "list_versions")
	Repository string // repository URL
	Chart      string // chart name (may be empty)
	Version    string // chart version (may be empty)
	Err        error  // underlying error
}

func (e *OperationError) Error() string {
	// Format: "op: chart@version: message" or "op: repo: message"
	if e.Chart != "" && e.Version != "" {
		return fmt.Sprintf("%s: %s@%s: %v", e.Op, e.Chart, e.Version, e.Err)
	}
	if e.Chart != "" {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Chart, e.Err)
	}
	if e.Repository != "" {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Repository, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *OperationError) Unwrap() error {
	return e.Err
}

// OpError creates a structured operation error with context.
func OpError(op, repo, chart, version string, err error) *OperationError {
	return &OperationError{
		Op:         op,
		Repository: repo,
		Chart:      chart,
		Version:    version,
		Err:        err,
	}
}

// HandleError converts a Helm error to an MCP error result.
// Returns nil if err is nil, indicating success.
func HandleError(err error) *mcp.CallToolResult {
	if err == nil {
		return nil
	}

	// Map specific error types to user-friendly messages
	switch {
	case helm.IsChartNotFound(err):
		return TextError(fmt.Sprintf("chart not found: %v", err))
	case helm.IsRepositoryError(err):
		return TextError(fmt.Sprintf("repository error: %v", err))
	case helm.IsURLValidationError(err):
		return TextError(fmt.Sprintf("invalid URL: %v", err))
	case helm.IsOutputTooLarge(err):
		return TextError(fmt.Sprintf("output too large: %v", err))
	default:
		return TextError(err.Error())
	}
}

// HandleOpError wraps an error with operation context and returns an MCP error result.
func HandleOpError(op, repo, chart, version string, err error) *mcp.CallToolResult {
	if err == nil {
		return nil
	}
	return HandleError(OpError(op, repo, chart, version, err))
}
