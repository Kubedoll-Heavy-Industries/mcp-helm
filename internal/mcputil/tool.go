// Package mcputil provides utilities for building MCP tools.
package mcputil

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolDef defines metadata for an MCP tool.
type ToolDef struct {
	Name        string
	Description string
	ReadOnly    bool
	OpenWorld   bool
}

// RegisterTool registers a tool with the MCP server using the given definition.
func RegisterTool[I, O any](s *mcp.Server, def ToolDef, handler mcp.ToolHandlerFor[I, O]) {
	annotations := &mcp.ToolAnnotations{
		ReadOnlyHint: def.ReadOnly,
	}

	if def.OpenWorld {
		t := true
		annotations.OpenWorldHint = &t
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        def.Name,
		Description: def.Description,
		Annotations: annotations,
	}, handler)
}

// TextError creates an MCP error result with a text message.
func TextError(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}
}
