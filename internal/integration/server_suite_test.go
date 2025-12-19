//go:build integration

package integration

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/Kubedoll-Heavy-Industries/mcp-helm/internal/handler"
	"github.com/Kubedoll-Heavy-Industries/mcp-helm/internal/helm"
)

// ServerSuite tests the MCP server with a shared Helm client.
// The server and client are created once in SetupSuite.
type ServerSuite struct {
	suite.Suite
	helmClient helm.ChartService
	server     *httptest.Server
	mcpServer  *mcp.Server
}

func (s *ServerSuite) SetupSuite() {
	logger := zap.NewNop()

	// Create shared Helm client with generous cache settings
	s.helmClient = helm.NewClient(
		helm.WithTimeout(60*time.Second),
		helm.WithIndexTTL(10*time.Minute),
		helm.WithCacheSize(100),
		helm.WithLogger(logger),
	)

	// Create handler
	h := handler.New(s.helmClient, logger)

	// Create MCP server
	s.mcpServer = mcp.NewServer(
		&mcp.Implementation{
			Name:    "mcp-helm-test",
			Version: "0.0.0-test",
		},
		nil,
	)
	h.Register(s.mcpServer)

	// Create MCP HTTP handler
	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(_ *http.Request) *mcp.Server { return s.mcpServer },
		&mcp.StreamableHTTPOptions{},
	)

	// Create HTTP mux
	mux := http.NewServeMux()
	mux.Handle("/mcp", mcpHandler)
	mux.Handle("/mcp/", mcpHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	s.server = httptest.NewServer(mux)

	// Warm the cache by fetching a version
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	_, _ = s.helmClient.GetLatestVersion(ctx, bitnamiRepo, "nginx")
}

func (s *ServerSuite) TearDownSuite() {
	if s.server != nil {
		s.server.Close()
	}
}

// newSession creates a fresh MCP client session for the test
func (s *ServerSuite) newSession() *mcp.ClientSession {
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.0.0-test",
	}, nil)

	transport := &mcp.StreamableClientTransport{
		Endpoint: s.server.URL + "/mcp",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	session, err := client.Connect(ctx, transport, nil)
	s.Require().NoError(err)

	return session
}

// TestHealthEndpoint tests the health endpoint
func (s *ServerSuite) TestHealthEndpoint() {
	resp, err := http.Get(s.server.URL + "/health")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	s.Equal("ok", string(body))
}

// TestToolsList tests listing available tools
func (s *ServerSuite) TestToolsList() {
	session := s.newSession()
	defer session.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := session.ListTools(ctx, nil)
	s.Require().NoError(err)

	// Collect tool names
	toolNames := make([]string, len(result.Tools))
	for i, tool := range result.Tools {
		toolNames[i] = tool.Name
	}

	// Verify expected tools
	s.Contains(toolNames, "list_repository_charts")
	s.Contains(toolNames, "list_chart_versions")
	s.Contains(toolNames, "get_chart_latest_version")
	s.Contains(toolNames, "get_chart_values")
	s.Contains(toolNames, "get_chart_values_schema")
	s.Contains(toolNames, "list_chart_contents")
	s.Contains(toolNames, "get_chart_content")
	s.Contains(toolNames, "get_chart_dependencies")
	s.Contains(toolNames, "refresh_repository_index")
}

// TestCallTool_ListCharts tests calling list_repository_charts
func (s *ServerSuite) TestCallTool_ListCharts() {
	session := s.newSession()
	defer session.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "list_repository_charts",
		Arguments: map[string]any{
			"repository_url": bitnamiRepo,
			"search":         "nginx",
		},
	})
	s.Require().NoError(err)

	s.False(result.IsError)
	s.Require().NotEmpty(result.Content)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	s.Require().True(ok, "expected TextContent")
	s.Contains(textContent.Text, "nginx")
}

// TestCallTool_GetLatestVersion tests calling get_chart_latest_version
func (s *ServerSuite) TestCallTool_GetLatestVersion() {
	session := s.newSession()
	defer session.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "get_chart_latest_version",
		Arguments: map[string]any{
			"repository_url": bitnamiRepo,
			"chart_name":     "nginx",
		},
	})
	s.Require().NoError(err)

	s.False(result.IsError)
	s.Require().NotEmpty(result.Content)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	s.Require().True(ok, "expected TextContent")
	s.Regexp(`\d+\.\d+\.\d+`, textContent.Text)
}

// TestCallTool_ValidationError tests error handling for invalid input
func (s *ServerSuite) TestCallTool_ValidationError() {
	session := s.newSession()
	defer session.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "list_repository_charts",
		Arguments: map[string]any{
			"repository_url": "",
		},
	})
	s.Require().NoError(err)

	// Tool call succeeds at RPC level but returns error in content
	s.True(result.IsError)
}

// TestCallTool_ChartNotFound tests error handling for missing chart
func (s *ServerSuite) TestCallTool_ChartNotFound() {
	session := s.newSession()
	defer session.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "list_chart_versions",
		Arguments: map[string]any{
			"repository_url": bitnamiRepo,
			"chart_name":     "this-chart-does-not-exist-xyz",
		},
	})
	s.Require().NoError(err)

	s.True(result.IsError)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	s.Require().True(ok, "expected TextContent")
	s.Contains(textContent.Text, "not found")
}

func TestServerSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests in short mode")
	}
	suite.Run(t, new(ServerSuite))
}
