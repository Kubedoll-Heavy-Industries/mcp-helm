package tools

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Kubedoll-Heavy-Industries/mcp-helm/lib/helm_client"
)

type getLatestVersionOfChartInput struct {
	RepositoryURL string `json:"repository_url" jsonschema:"Helm repository URL"`
	ChartName     string `json:"chart_name" jsonschema:"Chart name"`
}

type getLatestVersionOfChartOutput struct {
	Version string `json:"version" jsonschema:"Latest chart version"`
}

func newGetLatestVersionOfChartHandler(c *helm_client.HelmClient) mcp.ToolHandlerFor[getLatestVersionOfChartInput, getLatestVersionOfChartOutput] {
	return func(_ context.Context, _ *mcp.CallToolRequest, in getLatestVersionOfChartInput) (*mcp.CallToolResult, getLatestVersionOfChartOutput, error) {
		repositoryURL := strings.TrimSpace(in.RepositoryURL)
		if repositoryURL == "" {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "repository_url is required"}}}, getLatestVersionOfChartOutput{}, nil
		}
		chartName := strings.TrimSpace(in.ChartName)
		if chartName == "" {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "chart_name is required"}}}, getLatestVersionOfChartOutput{}, nil
		}

		version, err := c.GetChartLatestVersion(repositoryURL, chartName)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "failed to get latest version: " + err.Error()}}}, getLatestVersionOfChartOutput{}, nil
		}

		return nil, getLatestVersionOfChartOutput{Version: version}, nil
	}
}
