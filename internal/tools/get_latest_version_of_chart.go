package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/Kubedoll-Heavy-Industries/mcp-helm/lib/helm_client"
)

func NewGetLatestVersionOfChartTool() mcp.Tool {
	return mcp.NewTool("get_latest_version_of_chart",
		mcp.WithDescription("Retrieves the latest version of the chart"),
		mcp.WithString("repository_url",
			mcp.Required(),
			mcp.Description("Helm repository URL"),
		),
		mcp.WithString("chart_name",
			mcp.Required(),
			mcp.Description("Chart name"),
		),
	)
}

func GetLatestVersionOfCharHandler(c *helm_client.HelmClient) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		repositoryURL, err := request.RequireString("repository_url")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		repositoryURL = strings.TrimSpace(repositoryURL)

		chartName, err := request.RequireString("chart_name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		chartName = strings.TrimSpace(chartName)

		version, err := c.GetChartLatestVersion(repositoryURL, chartName)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to list charts: %v", err)), nil
		}

		return mcp.NewToolResultText(version), nil
	}
}
