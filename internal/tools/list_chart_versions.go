package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Kubedoll-Heavy-Industries/mcp-helm/lib/helm_client"
)

type listChartVersionsInput struct {
	RepositoryURL string `json:"repository_url" jsonschema:"Helm repository URL"`
	ChartName     string `json:"chart_name" jsonschema:"Chart name"`
	Limit         int    `json:"limit,omitempty" jsonschema:"Max number of results to return"`
	Offset        int    `json:"offset,omitempty" jsonschema:"Number of results to skip"`
}

type listChartVersionsVersion struct {
	Version    string `json:"version" jsonschema:"Chart version"`
	AppVersion string `json:"app_version,omitempty" jsonschema:"Chart appVersion"`
	Created    string `json:"created,omitempty" jsonschema:"Creation timestamp (RFC3339)"`
	Deprecated bool   `json:"deprecated" jsonschema:"Whether the chart version is deprecated"`
}

type listChartVersionsOutput struct {
	Versions []listChartVersionsVersion `json:"versions" jsonschema:"Chart versions (newest first)"`
	Total    int                        `json:"total" jsonschema:"Total number of versions available before pagination"`
	Limit    int                        `json:"limit" jsonschema:"Applied limit"`
	Offset   int                        `json:"offset" jsonschema:"Applied offset"`
}

func newListChartVersionsHandler(c *helm_client.HelmClient) mcp.ToolHandlerFor[listChartVersionsInput, listChartVersionsOutput] {
	return func(_ context.Context, _ *mcp.CallToolRequest, in listChartVersionsInput) (*mcp.CallToolResult, listChartVersionsOutput, error) {
		repositoryURL := strings.TrimSpace(in.RepositoryURL)
		if repositoryURL == "" {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "repository_url is required"}}}, listChartVersionsOutput{}, nil
		}
		chartName := strings.TrimSpace(in.ChartName)
		if chartName == "" {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "chart_name is required"}}}, listChartVersionsOutput{}, nil
		}
		if in.Limit < 0 {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "limit must be >= 0"}}}, listChartVersionsOutput{}, nil
		}
		if in.Offset < 0 {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "offset must be >= 0"}}}, listChartVersionsOutput{}, nil
		}

		meta, err := c.ListChartVersionsWithMetadata(repositoryURL, chartName)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to list chart versions: %v", err)}}}, listChartVersionsOutput{}, nil
		}

		total := len(meta)
		start := in.Offset
		if start > total {
			start = total
		}
		end := total
		if in.Limit > 0 && start+in.Limit < end {
			end = start + in.Limit
		}

		versions := make([]listChartVersionsVersion, 0, end-start)
		for _, v := range meta[start:end] {
			versions = append(versions, listChartVersionsVersion{
				Version:    v.Version,
				AppVersion: v.AppVersion,
				Created:    v.Created,
				Deprecated: v.Deprecated,
			})
		}

		return nil, listChartVersionsOutput{
			Versions: versions,
			Total:    total,
			Limit:    in.Limit,
			Offset:   in.Offset,
		}, nil
	}
}
