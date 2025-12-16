package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Kubedoll-Heavy-Industries/mcp-helm/lib/helm_client"
)

type getChartValuesInput struct {
	RepositoryURL string `json:"repository_url" jsonschema:"Helm repository URL"`
	ChartName     string `json:"chart_name" jsonschema:"Chart name"`
	ChartVersion  string `json:"chart_version,omitempty" jsonschema:"Chart version. If omitted the latest version will be used"`
}

type getChartValuesOutput struct {
	Values string `json:"values" jsonschema:"The chart values.yaml contents"`
}

func newGetChartValuesHandler(c *helm_client.HelmClient) mcp.ToolHandlerFor[getChartValuesInput, getChartValuesOutput] {
	return func(_ context.Context, _ *mcp.CallToolRequest, in getChartValuesInput) (*mcp.CallToolResult, getChartValuesOutput, error) {
		repositoryURL := strings.TrimSpace(in.RepositoryURL)
		if repositoryURL == "" {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "repository_url is required"}}}, getChartValuesOutput{}, nil
		}
		chartName := strings.TrimSpace(in.ChartName)
		if chartName == "" {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "chart_name is required"}}}, getChartValuesOutput{}, nil
		}

		chartVersion := strings.TrimSpace(in.ChartVersion)
		if chartVersion == "" {
			var err error
			chartVersion, err = c.GetChartLatestVersion(repositoryURL, chartName)
			if err != nil {
				return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to get the latest chart version: %v", err)}}}, getChartValuesOutput{}, nil
			}
		}

		values, err := c.GetChartValues(repositoryURL, chartName, chartVersion)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to get chart values: %v", err)}}}, getChartValuesOutput{}, nil
		}

		return nil, getChartValuesOutput{Values: values}, nil
	}
}
