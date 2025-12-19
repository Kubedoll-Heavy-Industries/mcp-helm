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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/Kubedoll-Heavy-Industries/mcp-helm/internal/handler"
	"github.com/Kubedoll-Heavy-Industries/mcp-helm/internal/helm"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	logger := zap.NewNop()

	// Create real Helm client
	helmClient := helm.NewClient(
		helm.WithTimeout(60*time.Second),
		helm.WithIndexTTL(5*time.Minute),
		helm.WithCacheSize(10),
		helm.WithLogger(logger),
	)

	// Create handler
	h := handler.New(helmClient, logger)

	// Create MCP server using the SDK
	mcpServer := mcp.NewServer(
		&mcp.Implementation{
			Name:    "mcp-helm-test",
			Version: "0.0.0-test",
		},
		nil,
	)
	h.Register(mcpServer)

	// Create MCP HTTP handler
	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(_ *http.Request) *mcp.Server { return mcpServer },
		&mcp.StreamableHTTPOptions{},
	)

	// Create a simple mux for testing
	mux := http.NewServeMux()
	mux.Handle("/mcp", mcpHandler)
	mux.Handle("/mcp/", mcpHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	return httptest.NewServer(mux)
}

// newMCPSession creates an MCP client session connected to the test server
func newMCPSession(t *testing.T, serverURL string) *mcp.ClientSession {
	t.Helper()

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.0.0-test",
	}, nil)

	transport := &mcp.StreamableClientTransport{
		Endpoint: serverURL + "/mcp",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	session, err := client.Connect(ctx, transport, nil)
	require.NoError(t, err)

	return session
}

func TestServerIntegration_HealthEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "ok", string(body))
}

func TestServerIntegration_ToolsList(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := newTestServer(t)
	defer ts.Close()

	session := newMCPSession(t, ts.URL)
	defer session.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := session.ListTools(ctx, nil)
	require.NoError(t, err)

	// Collect tool names
	toolNames := make([]string, len(result.Tools))
	for i, tool := range result.Tools {
		toolNames[i] = tool.Name
	}

	// Should have our defined tools
	assert.Contains(t, toolNames, "list_repository_charts")
	assert.Contains(t, toolNames, "list_chart_versions")
	assert.Contains(t, toolNames, "get_chart_latest_version")
	assert.Contains(t, toolNames, "get_chart_values")
	assert.Contains(t, toolNames, "get_chart_values_schema")
	assert.Contains(t, toolNames, "list_chart_contents")
	assert.Contains(t, toolNames, "get_chart_content")
	assert.Contains(t, toolNames, "get_chart_dependencies")
	assert.Contains(t, toolNames, "refresh_repository_index")
}

func TestServerIntegration_CallTool_ListCharts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := newTestServer(t)
	defer ts.Close()

	session := newMCPSession(t, ts.URL)
	defer session.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "list_repository_charts",
		Arguments: map[string]any{
			"repository_url": bitnamiRepo,
			"search":         "nginx", // Use search to find specific chart
		},
	})
	require.NoError(t, err)

	assert.False(t, result.IsError)
	require.NotEmpty(t, result.Content)

	// Get the text content
	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")

	assert.Contains(t, textContent.Text, "nginx")
}

func TestServerIntegration_CallTool_GetLatestVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := newTestServer(t)
	defer ts.Close()

	session := newMCPSession(t, ts.URL)
	defer session.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "get_chart_latest_version",
		Arguments: map[string]any{
			"repository_url": bitnamiRepo,
			"chart_name":     testChart,
		},
	})
	require.NoError(t, err)

	assert.False(t, result.IsError)
	require.NotEmpty(t, result.Content)

	// The result should contain a version number
	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Regexp(t, `\d+\.\d+\.\d+`, textContent.Text)
}

func TestServerIntegration_CallTool_ValidationError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := newTestServer(t)
	defer ts.Close()

	session := newMCPSession(t, ts.URL)
	defer session.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Call with empty URL which should fail validation
	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "list_repository_charts",
		Arguments: map[string]any{
			"repository_url": "",
		},
	})
	require.NoError(t, err)

	// The tool call should succeed at RPC level but return an error in content
	assert.True(t, result.IsError)
}

func TestServerIntegration_CallTool_ChartNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := newTestServer(t)
	defer ts.Close()

	session := newMCPSession(t, ts.URL)
	defer session.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "list_chart_versions",
		Arguments: map[string]any{
			"repository_url": bitnamiRepo,
			"chart_name":     "this-chart-does-not-exist-xyz",
		},
	})
	require.NoError(t, err)

	assert.True(t, result.IsError)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "not found")
}
