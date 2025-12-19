package handler

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Kubedoll-Heavy-Industries/mcp-helm/internal/mcputil"
)

// Default pagination limit for versions
const defaultVersionListLimit = 20

// Input/output types for version tools

type listChartVersionsInput struct {
	RepositoryURL string `json:"repository_url" jsonschema:"Helm repository URL"`
	ChartName     string `json:"chart_name" jsonschema:"Chart name"`
	Limit         int    `json:"limit,omitempty" jsonschema:"Maximum versions to return (default 20, 0 = unlimited)"`
	Offset        int    `json:"offset,omitempty" jsonschema:"Number of versions to skip"`
}

type versionInfo struct {
	Version    string `json:"version" jsonschema:"Chart version"`
	AppVersion string `json:"app_version,omitempty" jsonschema:"Application version"`
	Created    string `json:"created,omitempty" jsonschema:"Creation timestamp (RFC3339)"`
	Deprecated bool   `json:"deprecated" jsonschema:"Whether the version is deprecated"`
}

type listChartVersionsOutput struct {
	Versions []versionInfo `json:"versions" jsonschema:"Chart versions (newest first)"`
	Total    int           `json:"total" jsonschema:"Total versions before pagination"`
	Limit    int           `json:"limit" jsonschema:"Applied limit"`
	Offset   int           `json:"offset" jsonschema:"Applied offset"`
}

type getChartLatestVersionInput struct {
	RepositoryURL string `json:"repository_url" jsonschema:"Helm repository URL"`
	ChartName     string `json:"chart_name" jsonschema:"Chart name"`
}

type getChartLatestVersionOutput struct {
	Version string `json:"version" jsonschema:"Latest chart version"`
}

type refreshIndexInput struct {
	RepositoryURL string `json:"repository_url" jsonschema:"Helm repository URL"`
}

type refreshIndexOutput struct {
	Success bool   `json:"success" jsonschema:"Whether the refresh succeeded"`
	Message string `json:"message" jsonschema:"Status message"`
}

// Handler implementations

func (h *Handler) listChartVersions() mcp.ToolHandlerFor[listChartVersionsInput, listChartVersionsOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in listChartVersionsInput) (*mcp.CallToolResult, listChartVersionsOutput, error) {
		emptyOutput := listChartVersionsOutput{Versions: []versionInfo{}}

		if err := validateRequired(map[string]string{
			"repository_url": in.RepositoryURL,
			"chart_name":     in.ChartName,
		}); err != nil {
			return err, emptyOutput, nil
		}

		if in.Limit < 0 {
			return mcputil.TextError("limit must be >= 0"), emptyOutput, nil
		}
		if in.Offset < 0 {
			return mcputil.TextError("offset must be >= 0"), emptyOutput, nil
		}

		repo := strings.TrimSpace(in.RepositoryURL)
		chart := strings.TrimSpace(in.ChartName)

		versions, err := h.svc.ListVersions(ctx, repo, chart)
		if err != nil {
			return mcputil.HandleError(err), emptyOutput, nil
		}

		total := len(versions)

		// Apply pagination with default limit
		limit := in.Limit
		if limit == 0 {
			limit = defaultVersionListLimit
		}

		start := in.Offset
		if start > total {
			start = total
		}
		end := total
		if limit > 0 && start+limit < end {
			end = start + limit
		}

		// Convert to output format
		result := make([]versionInfo, 0, end-start)
		for _, v := range versions[start:end] {
			created := ""
			if !v.Created.IsZero() {
				created = v.Created.UTC().Format("2006-01-02T15:04:05Z")
			}
			result = append(result, versionInfo{
				Version:    v.Version,
				AppVersion: v.AppVersion,
				Created:    created,
				Deprecated: v.Deprecated,
			})
		}

		return nil, listChartVersionsOutput{
			Versions: result,
			Total:    total,
			Limit:    limit,
			Offset:   in.Offset,
		}, nil
	}
}

func (h *Handler) getChartLatestVersion() mcp.ToolHandlerFor[getChartLatestVersionInput, getChartLatestVersionOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in getChartLatestVersionInput) (*mcp.CallToolResult, getChartLatestVersionOutput, error) {
		if err := validateRequired(map[string]string{
			"repository_url": in.RepositoryURL,
			"chart_name":     in.ChartName,
		}); err != nil {
			return err, getChartLatestVersionOutput{}, nil
		}

		repo := strings.TrimSpace(in.RepositoryURL)
		chart := strings.TrimSpace(in.ChartName)

		version, err := h.svc.GetLatestVersion(ctx, repo, chart)
		if err != nil {
			return mcputil.HandleError(err), getChartLatestVersionOutput{}, nil
		}

		return nil, getChartLatestVersionOutput{Version: version}, nil
	}
}

func (h *Handler) refreshIndex() mcp.ToolHandlerFor[refreshIndexInput, refreshIndexOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in refreshIndexInput) (*mcp.CallToolResult, refreshIndexOutput, error) {
		if err := validateRequired(map[string]string{"repository_url": in.RepositoryURL}); err != nil {
			return err, refreshIndexOutput{}, nil
		}

		repo := strings.TrimSpace(in.RepositoryURL)

		if err := h.svc.RefreshIndex(ctx, repo); err != nil {
			return mcputil.HandleError(err), refreshIndexOutput{}, nil
		}

		return nil, refreshIndexOutput{
			Success: true,
			Message: "Repository index refreshed successfully",
		}, nil
	}
}
