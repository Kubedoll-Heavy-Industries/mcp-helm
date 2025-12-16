package tools

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Kubedoll-Heavy-Industries/mcp-helm/lib/helm_client"
)

type listRepositoryChartsInput struct {
	RepositoryURL string `json:"repository_url" jsonschema:"Helm repository URL"`
}

type listRepositoryChartsOutput struct {
	Charts []string `json:"charts" jsonschema:"Charts available in the repository"`
}

func newListRepositoryChartsHandler(c *helm_client.HelmClient) mcp.ToolHandlerFor[listRepositoryChartsInput, listRepositoryChartsOutput] {
	return func(_ context.Context, _ *mcp.CallToolRequest, in listRepositoryChartsInput) (*mcp.CallToolResult, listRepositoryChartsOutput, error) {
		repositoryURL := strings.TrimSpace(in.RepositoryURL)
		if repositoryURL == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: "repository_url is required"}},
			}, listRepositoryChartsOutput{}, nil
		}

		charts, err := c.ListCharts(repositoryURL)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: "failed to list charts: " + err.Error()}},
			}, listRepositoryChartsOutput{}, nil
		}

		return nil, listRepositoryChartsOutput{Charts: charts}, nil
	}
}
