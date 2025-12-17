// Package handler provides MCP tool handlers for Helm operations.
package handler

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"

	"github.com/Kubedoll-Heavy-Industries/mcp-helm/internal/helm"
	"github.com/Kubedoll-Heavy-Industries/mcp-helm/internal/mcputil"
)

// Handler provides MCP tool handlers backed by a Helm service.
type Handler struct {
	svc    helm.ChartService
	logger *zap.Logger
}

// New creates a new Handler.
func New(svc helm.ChartService, logger *zap.Logger) *Handler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Handler{
		svc:    svc,
		logger: logger,
	}
}

// Register registers all Helm tools with the MCP server.
func (h *Handler) Register(s *mcp.Server) {
	// Chart listing and discovery
	mcputil.RegisterTool(s, mcputil.ToolDef{
		Name:        "list_repository_charts",
		Description: "Lists all charts available in a Helm repository",
		ReadOnly:    true,
		OpenWorld:   true,
	}, h.listCharts())

	mcputil.RegisterTool(s, mcputil.ToolDef{
		Name:        "list_chart_versions",
		Description: "Lists all versions of a chart with metadata (appVersion, created, deprecated)",
		ReadOnly:    true,
		OpenWorld:   true,
	}, h.listVersions())

	mcputil.RegisterTool(s, mcputil.ToolDef{
		Name:        "get_latest_version_of_chart",
		Description: "Gets the latest version of a chart",
		ReadOnly:    true,
		OpenWorld:   true,
	}, h.getLatestVersion())

	// Chart content retrieval
	mcputil.RegisterTool(s, mcputil.ToolDef{
		Name:        "get_chart_values",
		Description: "Gets the values.yaml contents for a chart",
		ReadOnly:    true,
		OpenWorld:   true,
	}, h.getValues())

	mcputil.RegisterTool(s, mcputil.ToolDef{
		Name:        "get_values_schema",
		Description: "Gets the values.schema.json for a chart if present",
		ReadOnly:    true,
		OpenWorld:   true,
	}, h.getValuesSchema())

	mcputil.RegisterTool(s, mcputil.ToolDef{
		Name:        "get_chart_contents",
		Description: "Gets all files in a chart (templates, values, etc.)",
		ReadOnly:    true,
		OpenWorld:   true,
	}, h.getContents())

	mcputil.RegisterTool(s, mcputil.ToolDef{
		Name:        "get_chart_dependencies",
		Description: "Gets the dependencies defined in a chart",
		ReadOnly:    true,
		OpenWorld:   true,
	}, h.getDependencies())

	// Repository management
	mcputil.RegisterTool(s, mcputil.ToolDef{
		Name:        "refresh_repository_index",
		Description: "Forces a refresh of the cached repository index",
		ReadOnly:    false, // Modifies cache state
		OpenWorld:   true,
	}, h.refreshIndex())
}

// resolveVersion returns the given version if non-empty, otherwise fetches the latest.
func (h *Handler) resolveVersion(ctx context.Context, repo, chart, version string) (string, error) {
	version = strings.TrimSpace(version)
	if version != "" {
		return version, nil
	}
	return h.svc.GetLatestVersion(ctx, repo, chart)
}

// validateRequired checks that required string fields are non-empty.
func validateRequired(fields map[string]string) *mcp.CallToolResult {
	for name, value := range fields {
		if strings.TrimSpace(value) == "" {
			return mcputil.TextError(name + " is required")
		}
	}
	return nil
}
