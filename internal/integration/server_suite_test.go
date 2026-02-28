//go:build integration

package integration

import (
	"context"
	"encoding/json"
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

// ServerSuite tests the MCP server layer.
// Tests verify MCP protocol contracts, not specific chart data.
type ServerSuite struct {
	suite.Suite
	helmClient helm.ChartService
	server     *httptest.Server
	mcpServer  *mcp.Server

	// Dynamic fixture
	sampleRepo  string
	sampleChart string
}

func (s *ServerSuite) SetupSuite() {
	logger := zap.NewNop()

	s.helmClient = helm.NewClient(
		helm.WithTimeout(60*time.Second),
		helm.WithIndexTTL(10*time.Minute),
		helm.WithChartCacheSize(100),
		helm.WithLogger(logger),
	)

	h := handler.New(s.helmClient, logger)

	s.mcpServer = mcp.NewServer(
		&mcp.Implementation{
			Name:    "mcp-helm-test",
			Version: "0.0.0-test",
		},
		nil,
	)
	h.Register(s.mcpServer)

	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(_ *http.Request) *mcp.Server { return s.mcpServer },
		&mcp.StreamableHTTPOptions{},
	)

	mux := http.NewServeMux()
	mux.Handle("/mcp", mcpHandler)
	mux.Handle("/mcp/", mcpHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	s.server = httptest.NewServer(mux)

	// Discover a working fixture
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	repos := []string{prometheusRepo, grafanaRepo, ingressRepo}
	for _, repo := range repos {
		charts, err := s.helmClient.ListCharts(ctx, repo)
		if err == nil && len(charts) > 0 {
			s.sampleRepo = repo
			s.sampleChart = charts[0]
			break
		}
	}

	if s.sampleRepo == "" {
		s.T().Fatal("Could not discover working fixture")
	}
}

func (s *ServerSuite) TearDownSuite() {
	if s.server != nil {
		s.server.Close()
	}
}

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

// =============================================================================
// HTTP Endpoint Tests
// =============================================================================

func (s *ServerSuite) TestHealthEndpoint_ReturnsOK() {
	resp, err := http.Get(s.server.URL + "/health")
	s.Require().NoError(err)
	defer resp.Body.Close()

	// Contract: Health endpoint returns 200 OK
	s.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	s.Equal("ok", string(body))
}

// =============================================================================
// MCP Protocol Tests
// =============================================================================

func (s *ServerSuite) TestListTools_ReturnsExpectedTools() {
	session := s.newSession()
	defer session.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := session.ListTools(ctx, nil)
	s.Require().NoError(err)

	// Contract: All registered tools should be present
	expectedTools := []string{
		"search_charts",
		"get_versions",
		"get_values",
		"get_dependencies",
		"get_notes",
	}

	toolNames := make(map[string]bool)
	for _, tool := range result.Tools {
		toolNames[tool.Name] = true

		// Contract: Every tool must have a description
		s.NotEmpty(tool.Description, "Tool %s should have description", tool.Name)
	}

	for _, expected := range expectedTools {
		s.True(toolNames[expected], "Missing expected tool: %s", expected)
	}
}

func (s *ServerSuite) TestListTools_ToolsHaveInputSchemas() {
	session := s.newSession()
	defer session.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := session.ListTools(ctx, nil)
	s.Require().NoError(err)

	for _, tool := range result.Tools {
		// Contract: Tools should have input schemas
		s.NotNil(tool.InputSchema, "Tool %s should have InputSchema", tool.Name)
	}
}

// =============================================================================
// Tool Call Contract Tests
// =============================================================================

func (s *ServerSuite) TestCallTool_SearchCharts_ReturnsStructuredResponse() {
	session := s.newSession()
	defer session.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "search_charts",
		Arguments: map[string]any{
			"repository_url": s.sampleRepo,
			"limit":          10,
		},
	})
	s.Require().NoError(err)

	// Contract: Successful call should not be an error
	s.False(result.IsError, "Valid call should succeed")
	s.Require().NotEmpty(result.Content, "Response should have content")

	// Contract: Content should be parseable
	textContent, ok := result.Content[0].(*mcp.TextContent)
	s.Require().True(ok, "Content should be TextContent")
	s.NotEmpty(textContent.Text, "Text content should not be empty")

	// Contract: Response should be valid JSON with expected structure
	var response struct {
		Charts []string `json:"charts"`
		Total  int      `json:"total"`
	}
	err = json.Unmarshal([]byte(textContent.Text), &response)
	s.Require().NoError(err, "Response should be valid JSON")
	s.NotEmpty(response.Charts, "Should return charts")
	s.LessOrEqual(len(response.Charts), 10, "Should respect limit")
	s.GreaterOrEqual(response.Total, len(response.Charts), "Total should be >= returned count")
}

func (s *ServerSuite) TestCallTool_GetVersions_ReturnsVersions() {
	session := s.newSession()
	defer session.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "get_versions",
		Arguments: map[string]any{
			"repository_url": s.sampleRepo,
			"chart_name":     s.sampleChart,
			"limit":          5,
		},
	})
	s.Require().NoError(err)
	s.False(result.IsError)

	textContent := result.Content[0].(*mcp.TextContent)

	// Contract: Response should have versions and total
	var response struct {
		Versions []struct {
			Version    string `json:"version"`
			AppVersion string `json:"app_version"`
		} `json:"versions"`
		Total int `json:"total"`
	}
	err = json.Unmarshal([]byte(textContent.Text), &response)
	s.Require().NoError(err)

	s.NotEmpty(response.Versions, "Should return versions")
	s.LessOrEqual(len(response.Versions), 5, "Should respect limit")
	s.GreaterOrEqual(response.Total, len(response.Versions), "Total should be >= returned count")
}

func (s *ServerSuite) TestCallTool_GetNotes_ReturnsNotes() {
	session := s.newSession()
	defer session.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "get_notes",
		Arguments: map[string]any{
			"repository_url": s.sampleRepo,
			"chart_name":     s.sampleChart,
		},
	})
	s.Require().NoError(err)
	s.False(result.IsError)

	textContent := result.Content[0].(*mcp.TextContent)

	// Contract: Response should be valid JSON with notes field
	var response struct {
		Notes string `json:"notes"`
	}
	err = json.Unmarshal([]byte(textContent.Text), &response)
	s.Require().NoError(err, "Response should be valid JSON")
	// Notes may be empty if chart doesn't have NOTES.txt, but field should exist
}

func (s *ServerSuite) TestCallTool_GetValues_WithIncludeSchema() {
	session := s.newSession()
	defer session.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Test with include_schema=true - should work or fail gracefully
	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "get_values",
		Arguments: map[string]any{
			"repository_url": s.sampleRepo,
			"chart_name":     s.sampleChart,
			"include_schema": true,
			"depth":          1, // Limit depth to keep response small
		},
	})
	s.Require().NoError(err)

	// Should either succeed or return a user-friendly error about size
	if result.IsError {
		textContent := result.Content[0].(*mcp.TextContent)
		// If it fails, it should be due to size limits, not crashes
		s.Contains(textContent.Text, "too large", "Error should mention size limit")
	} else {
		// If it succeeds, response should have values
		textContent := result.Content[0].(*mcp.TextContent)
		s.NotEmpty(textContent.Text)

		var response struct {
			Values string `json:"values"`
			Schema string `json:"schema"`
		}
		err = json.Unmarshal([]byte(textContent.Text), &response)
		s.Require().NoError(err, "Response should be valid JSON")
		s.NotEmpty(response.Values, "Should have values")
		// Schema may or may not be present depending on the chart
	}
}

// =============================================================================
// Error Handling Contract Tests
// =============================================================================

func (s *ServerSuite) TestCallTool_MissingRequiredField_ReturnsError() {
	session := s.newSession()
	defer session.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "search_charts",
		Arguments: map[string]any{
			"repository_url": "", // Empty required field
		},
	})

	// Contract: RPC should succeed, but tool returns error
	s.Require().NoError(err, "RPC should succeed even for validation errors")
	s.True(result.IsError, "Empty required field should return error")

	textContent := result.Content[0].(*mcp.TextContent)
	s.Contains(textContent.Text, "required", "Error should mention 'required'")
}

func (s *ServerSuite) TestCallTool_ChartNotFound_ReturnsErrorWithMessage() {
	session := s.newSession()
	defer session.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "get_versions",
		Arguments: map[string]any{
			"repository_url": s.sampleRepo,
			"chart_name":     "this-chart-does-not-exist-xyz123",
		},
	})

	s.Require().NoError(err, "RPC should succeed")
	s.True(result.IsError, "Non-existent chart should return error")

	textContent := result.Content[0].(*mcp.TextContent)
	s.Contains(textContent.Text, "not found", "Error should indicate chart not found")
}

func (s *ServerSuite) TestCallTool_InvalidLimit_ReturnsError() {
	session := s.newSession()
	defer session.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "search_charts",
		Arguments: map[string]any{
			"repository_url": s.sampleRepo,
			"limit":          -1, // Invalid
		},
	})

	s.Require().NoError(err)
	s.True(result.IsError, "Negative limit should return error")
}

// =============================================================================
// Session Management Tests
// =============================================================================

func (s *ServerSuite) TestMultipleSessions_Independent() {
	session1 := s.newSession()
	session2 := s.newSession()
	defer session1.Close()
	defer session2.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Both sessions should work independently
	result1, err := session1.ListTools(ctx, nil)
	s.Require().NoError(err)

	result2, err := session2.ListTools(ctx, nil)
	s.Require().NoError(err)

	// Contract: Both should return same tools
	s.Equal(len(result1.Tools), len(result2.Tools))
}

func TestServerSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests in short mode")
	}
	suite.Run(t, new(ServerSuite))
}
