package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Kubedoll-Heavy-Industries/mcp-helm/lib/helm_client"
)

func Register(s *mcp.Server, helmClient *helm_client.HelmClient) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_repository_charts",
		Description: "Lists all charts available in the repository",
	}, newListRepositoryChartsHandler(helmClient))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_latest_version_of_chart",
		Description: "Retrieves the latest version of the chart",
	}, newGetLatestVersionOfChartHandler(helmClient))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_chart_values",
		Description: "Retrieves values file for the chart",
	}, newGetChartValuesHandler(helmClient))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_chart_contents",
		Description: "Retrieves full chart contents",
	}, newGetChartContentsHandler(helmClient))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_chart_dependencies",
		Description: "Retrieves dependencies for the chart",
	}, newGetChartDependenciesHandler(helmClient))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_chart_versions",
		Description: "Lists available chart versions with metadata",
	}, newListChartVersionsHandler(helmClient))
}
