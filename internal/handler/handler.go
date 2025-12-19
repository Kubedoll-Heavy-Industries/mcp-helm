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
	// Repository tools
	mcputil.RegisterTool(s, mcputil.ToolDef{
		Name:        "list_repository_charts",
		Description: "List charts in a Helm repository (paginated, default limit=50)",
		ReadOnly:    true,
		OpenWorld:   true,
	}, h.listRepositoryCharts())

	mcputil.RegisterTool(s, mcputil.ToolDef{
		Name:        "refresh_repository_index",
		Description: "Force refresh of cached repository index",
		ReadOnly:    false,
		OpenWorld:   true,
	}, h.refreshIndex())

	// Chart version tools
	mcputil.RegisterTool(s, mcputil.ToolDef{
		Name:        "list_chart_versions",
		Description: "List versions of a chart with metadata (paginated, default limit=20)",
		ReadOnly:    true,
		OpenWorld:   true,
	}, h.listChartVersions())

	mcputil.RegisterTool(s, mcputil.ToolDef{
		Name:        "get_chart_latest_version",
		Description: "Get the latest version of a chart",
		ReadOnly:    true,
		OpenWorld:   true,
	}, h.getChartLatestVersion())

	// Chart content tools
	mcputil.RegisterTool(s, mcputil.ToolDef{
		Name:        "get_chart_values",
		Description: "Get values.yaml with optional yq-style path extraction (default max_lines=200)",
		ReadOnly:    true,
		OpenWorld:   true,
	}, h.getChartValues())

	mcputil.RegisterTool(s, mcputil.ToolDef{
		Name:        "get_chart_values_schema",
		Description: "Get values.schema.json if present",
		ReadOnly:    true,
		OpenWorld:   true,
	}, h.getValuesSchema())

	mcputil.RegisterTool(s, mcputil.ToolDef{
		Name:        "list_chart_contents",
		Description: "List files in a chart with optional glob pattern filter",
		ReadOnly:    true,
		OpenWorld:   true,
	}, h.listChartContents())

	mcputil.RegisterTool(s, mcputil.ToolDef{
		Name:        "get_chart_content",
		Description: "Get contents of specific files in a chart",
		ReadOnly:    true,
		OpenWorld:   true,
	}, h.getChartContent())

	mcputil.RegisterTool(s, mcputil.ToolDef{
		Name:        "get_chart_dependencies",
		Description: "Get chart dependencies from Chart.yaml",
		ReadOnly:    true,
		OpenWorld:   true,
	}, h.getChartDependencies())
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
