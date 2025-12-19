package mcputil

import (
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Kubedoll-Heavy-Industries/mcp-helm/internal/helm"
)

// HandleError converts a Helm error to an MCP error result.
// Returns nil if err is nil, indicating success.
func HandleError(err error) *mcp.CallToolResult {
	if err == nil {
		return nil
	}

	// Map specific error types to user-friendly messages
	switch {
	case helm.IsChartNotFound(err):
		return TextError(fmt.Sprintf("Chart not found: %v", err))
	case helm.IsRepositoryError(err):
		return TextError(fmt.Sprintf("Repository error: %v", err))
	case helm.IsURLValidationError(err):
		return TextError(fmt.Sprintf("Invalid URL: %v", err))
	case helm.IsOutputTooLarge(err):
		return TextError(fmt.Sprintf("Output too large: %v", err))
	default:
		return TextError(fmt.Sprintf("Error: %v", err))
	}
}
