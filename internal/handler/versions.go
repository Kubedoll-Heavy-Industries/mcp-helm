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

type getVersionsInput struct {
	RepositoryURL string `json:"repository_url" jsonschema:"Helm repository URL (e.g. https://charts.bitnami.com/bitnami) or OCI registry (e.g. oci://ghcr.io/traefik/helm)"`
	ChartName     string `json:"chart_name" jsonschema:"Chart name (e.g. postgresql)"`
	Limit         int    `json:"limit,omitempty" jsonschema:"Maximum results (default 20, max 100)"`
}

type versionInfo struct {
	Version    string `json:"version" jsonschema:"Chart version"`
	AppVersion string `json:"app_version,omitempty" jsonschema:"Application version"`
	Created    string `json:"created,omitempty" jsonschema:"Creation timestamp (RFC3339)"`
	Deprecated bool   `json:"deprecated" jsonschema:"Whether the version is deprecated"`
}

type getVersionsOutput struct {
	Versions []versionInfo `json:"versions" jsonschema:"Chart versions (newest first)"`
	Total    int           `json:"total" jsonschema:"Total versions available (may exceed returned results if limit applied)"`
}

// Handler implementations

func (h *Handler) getVersions() mcp.ToolHandlerFor[getVersionsInput, getVersionsOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in getVersionsInput) (*mcp.CallToolResult, getVersionsOutput, error) {
		emptyOutput := getVersionsOutput{Versions: []versionInfo{}}

		if err := validateRequired(map[string]string{
			"repository_url": in.RepositoryURL,
			"chart_name":     in.ChartName,
		}); err != nil {
			return mcputil.TextError(err.Error()), emptyOutput, nil
		}

		if in.Limit < 0 {
			return mcputil.TextError("limit must be >= 0"), emptyOutput, nil
		}

		repo := strings.TrimSpace(in.RepositoryURL)
		chart := strings.TrimSpace(in.ChartName)

		versions, err := h.svc.ListVersions(ctx, repo, chart)
		if err != nil {
			return mcputil.HandleOpError("list_versions", repo, chart, "", err), emptyOutput, nil
		}

		total := len(versions)

		// Apply limit (default 20, max 100)
		limit := in.Limit
		if limit == 0 {
			limit = defaultVersionListLimit
		}
		if limit > 100 {
			limit = 100
		}

		if limit < total {
			versions = versions[:limit]
		}

		// Convert to output format
		result := make([]versionInfo, 0, len(versions))
		for _, v := range versions {
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

		return nil, getVersionsOutput{
			Versions: result,
			Total:    total,
		}, nil
	}
}
