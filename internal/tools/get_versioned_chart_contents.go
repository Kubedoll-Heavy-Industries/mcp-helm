package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Kubedoll-Heavy-Industries/mcp-helm/lib/helm_client"
)

type getChartContentsInput struct {
	RepositoryURL string `json:"repository_url" jsonschema:"Helm repository URL"`
	ChartName     string `json:"chart_name" jsonschema:"Chart name"`
	ChartVersion  string `json:"chart_version,omitempty" jsonschema:"Chart version. If omitted the latest version will be used"`
	Recursive     bool   `json:"recursive,omitempty" jsonschema:"If true, retrieves all files in the chart recursively. Defaults to false"`
}

type getChartContentsOutput struct {
	Contents string `json:"contents" jsonschema:"The chart contents"`
}

func newGetChartContentsHandler(c *helm_client.HelmClient) mcp.ToolHandlerFor[getChartContentsInput, getChartContentsOutput] {
	return func(_ context.Context, _ *mcp.CallToolRequest, in getChartContentsInput) (*mcp.CallToolResult, getChartContentsOutput, error) {
		repositoryURL := strings.TrimSpace(in.RepositoryURL)
		if repositoryURL == "" {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "repository_url is required"}}}, getChartContentsOutput{}, nil
		}
		chartName := strings.TrimSpace(in.ChartName)
		if chartName == "" {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "chart_name is required"}}}, getChartContentsOutput{}, nil
		}

		chartVersion := strings.TrimSpace(in.ChartVersion)
		if chartVersion == "" {
			var err error
			chartVersion, err = c.GetChartLatestVersion(repositoryURL, chartName)
			if err != nil {
				return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to get the latest chart version: %v", err)}}}, getChartContentsOutput{}, nil
			}
		}

		contents, err := c.GetChartContents(repositoryURL, chartName, chartVersion, in.Recursive)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to get chart contents: %v", err)}}}, getChartContentsOutput{}, nil
		}

		return nil, getChartContentsOutput{Contents: contents}, nil
	}
}
