package handler

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Kubedoll-Heavy-Industries/mcp-helm/internal/mcputil"
)

// Input/output types for version tools

type listVersionsInput struct {
	RepositoryURL string `json:"repository_url" jsonschema:"Helm repository URL"`
	ChartName     string `json:"chart_name" jsonschema:"Chart name"`
	Limit         int    `json:"limit,omitempty" jsonschema:"Maximum versions to return (0 = unlimited)"`
	Offset        int    `json:"offset,omitempty" jsonschema:"Number of versions to skip"`
}

type versionInfo struct {
	Version    string `json:"version" jsonschema:"Chart version"`
	AppVersion string `json:"app_version,omitempty" jsonschema:"Application version"`
	Created    string `json:"created,omitempty" jsonschema:"Creation timestamp (RFC3339)"`
	Deprecated bool   `json:"deprecated" jsonschema:"Whether the version is deprecated"`
}

type listVersionsOutput struct {
	Versions []versionInfo `json:"versions" jsonschema:"Chart versions (newest first)"`
	Total    int           `json:"total" jsonschema:"Total versions before pagination"`
	Limit    int           `json:"limit" jsonschema:"Applied limit"`
	Offset   int           `json:"offset" jsonschema:"Applied offset"`
}

type getLatestVersionInput struct {
	RepositoryURL string `json:"repository_url" jsonschema:"Helm repository URL"`
	ChartName     string `json:"chart_name" jsonschema:"Chart name"`
}

type getLatestVersionOutput struct {
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

func (h *Handler) listVersions() mcp.ToolHandlerFor[listVersionsInput, listVersionsOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in listVersionsInput) (*mcp.CallToolResult, listVersionsOutput, error) {
		emptyOutput := listVersionsOutput{Versions: []versionInfo{}}

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

		// Apply pagination
		start := in.Offset
		if start > total {
			start = total
		}
		end := total
		if in.Limit > 0 && start+in.Limit < end {
			end = start + in.Limit
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

		return nil, listVersionsOutput{
			Versions: result,
			Total:    total,
			Limit:    in.Limit,
			Offset:   in.Offset,
		}, nil
	}
}

func (h *Handler) getLatestVersion() mcp.ToolHandlerFor[getLatestVersionInput, getLatestVersionOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in getLatestVersionInput) (*mcp.CallToolResult, getLatestVersionOutput, error) {
		if err := validateRequired(map[string]string{
			"repository_url": in.RepositoryURL,
			"chart_name":     in.ChartName,
		}); err != nil {
			return err, getLatestVersionOutput{}, nil
		}

		repo := strings.TrimSpace(in.RepositoryURL)
		chart := strings.TrimSpace(in.ChartName)

		version, err := h.svc.GetLatestVersion(ctx, repo, chart)
		if err != nil {
			return mcputil.HandleError(err), getLatestVersionOutput{}, nil
		}

		return nil, getLatestVersionOutput{Version: version}, nil
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
