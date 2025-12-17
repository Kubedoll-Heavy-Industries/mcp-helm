//go:build e2e

package e2e

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	binaryPath string
	bitnamiURL = "https://charts.bitnami.com/bitnami"
)

// TestMain builds the binary once for all e2e tests
func TestMain(m *testing.M) {
	// Build the binary
	tmpDir, err := os.MkdirTemp("", "mcp-helm-e2e")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	binaryPath = filepath.Join(tmpDir, "mcp-helm")
	cmd := exec.Command("go", "build", "-o", binaryPath, "../cmd/mcp-helm")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build binary: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// jsonRPCRequest represents a JSON-RPC 2.0 request
type jsonRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// jsonRPCResponse represents a JSON-RPC 2.0 response
type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// sendStdioMessage sends a JSON-RPC message to the process and reads the response
func sendStdioMessage(stdin io.Writer, stdout *bufio.Reader, req jsonRPCRequest) (jsonRPCResponse, error) {
	// Encode and send request
	data, err := json.Marshal(req)
	if err != nil {
		return jsonRPCResponse{}, fmt.Errorf("marshal request: %w", err)
	}

	if _, err := stdin.Write(data); err != nil {
		return jsonRPCResponse{}, fmt.Errorf("write request: %w", err)
	}
	if _, err := stdin.Write([]byte("\n")); err != nil {
		return jsonRPCResponse{}, fmt.Errorf("write newline: %w", err)
	}

	// Read response (skip any log lines)
	for {
		line, err := stdout.ReadBytes('\n')
		if err != nil {
			return jsonRPCResponse{}, fmt.Errorf("read response: %w", err)
		}

		// Skip non-JSON lines (log output)
		line = bytes.TrimSpace(line)
		if len(line) == 0 || line[0] != '{' {
			continue
		}

		var resp jsonRPCResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			// Not a valid JSON-RPC response, skip
			continue
		}

		// Check if this is a response to our request
		if resp.ID == req.ID {
			return resp, nil
		}
	}
}

func TestE2E_StdioMode_Initialize(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath)
	stdin, err := cmd.StdinPipe()
	require.NoError(t, err)
	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)

	require.NoError(t, cmd.Start())
	defer func() {
		stdin.Close()
		cmd.Process.Kill()
		cmd.Wait()
	}()

	reader := bufio.NewReader(stdout)

	// Send initialize request
	initReq := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "e2e-test",
				"version": "1.0",
			},
		},
	}

	resp, err := sendStdioMessage(stdin, reader, initReq)
	require.NoError(t, err)

	assert.Nil(t, resp.Error)
	assert.NotEmpty(t, resp.Result)

	// Verify the response contains server info
	var result struct {
		ProtocolVersion string `json:"protocolVersion"`
		ServerInfo      struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"serverInfo"`
	}
	require.NoError(t, json.Unmarshal(resp.Result, &result))
	assert.Equal(t, "2024-11-05", result.ProtocolVersion)
	assert.Contains(t, result.ServerInfo.Name, "Helm")
}

func TestE2E_StdioMode_ListTools(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath)
	stdin, err := cmd.StdinPipe()
	require.NoError(t, err)
	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)

	require.NoError(t, cmd.Start())
	defer func() {
		stdin.Close()
		cmd.Process.Kill()
		cmd.Wait()
	}()

	reader := bufio.NewReader(stdout)

	// Initialize first
	initReq := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo":      map[string]interface{}{"name": "e2e-test", "version": "1.0"},
		},
	}
	_, err = sendStdioMessage(stdin, reader, initReq)
	require.NoError(t, err)

	// Send initialized notification
	notifReq := jsonRPCRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}
	data, _ := json.Marshal(notifReq)
	stdin.Write(data)
	stdin.Write([]byte("\n"))

	// List tools
	listReq := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	resp, err := sendStdioMessage(stdin, reader, listReq)
	require.NoError(t, err)

	assert.Nil(t, resp.Error)

	var result struct {
		Tools []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"tools"`
	}
	require.NoError(t, json.Unmarshal(resp.Result, &result))

	// Verify we have the expected tools
	toolNames := make([]string, len(result.Tools))
	for i, tool := range result.Tools {
		toolNames[i] = tool.Name
	}

	assert.Contains(t, toolNames, "list_repository_charts")
	assert.Contains(t, toolNames, "list_chart_versions")
	assert.Contains(t, toolNames, "get_latest_version_of_chart")
}

func TestE2E_HTTPMode_Health(t *testing.T) {
	// Find a free port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start server in HTTP mode
	cmd := exec.CommandContext(ctx, binaryPath,
		"--transport=http",
		fmt.Sprintf("--listen=127.0.0.1:%d", port),
	)
	cmd.Stderr = os.Stderr

	require.NoError(t, cmd.Start())
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	// Wait for server to be ready
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(baseURL + "/healthz")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Test health endpoint
	resp, err := http.Get(baseURL + "/healthz")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestE2E_HTTPMode_Ready(t *testing.T) {
	// Find a free port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start server in HTTP mode
	cmd := exec.CommandContext(ctx, binaryPath,
		"--transport=http",
		fmt.Sprintf("--listen=127.0.0.1:%d", port),
	)
	cmd.Stderr = os.Stderr

	require.NoError(t, cmd.Start())
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	// Wait for server to be ready
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(baseURL + "/readyz")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Test ready endpoint
	resp, err := http.Get(baseURL + "/readyz")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestE2E_StdioMode_CallTool_ListCharts(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath)
	stdin, err := cmd.StdinPipe()
	require.NoError(t, err)
	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)

	require.NoError(t, cmd.Start())
	defer func() {
		stdin.Close()
		cmd.Process.Kill()
		cmd.Wait()
	}()

	reader := bufio.NewReader(stdout)

	// Initialize
	initReq := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo":      map[string]interface{}{"name": "e2e-test", "version": "1.0"},
		},
	}
	_, err = sendStdioMessage(stdin, reader, initReq)
	require.NoError(t, err)

	// Send initialized notification
	notifReq := jsonRPCRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}
	data, _ := json.Marshal(notifReq)
	stdin.Write(data)
	stdin.Write([]byte("\n"))

	// Call list_repository_charts tool
	callReq := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "list_repository_charts",
			"arguments": map[string]interface{}{
				"repository_url": bitnamiURL,
			},
		},
	}

	resp, err := sendStdioMessage(stdin, reader, callReq)
	require.NoError(t, err)

	assert.Nil(t, resp.Error)

	// Parse result to verify charts are listed
	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		StructuredContent struct {
			Charts []string `json:"charts"`
		} `json:"structuredContent"`
		IsError bool `json:"isError"`
	}
	require.NoError(t, json.Unmarshal(resp.Result, &result))

	assert.False(t, result.IsError)

	// Check that we have charts either in text content or structured content
	if len(result.StructuredContent.Charts) > 0 {
		assert.Contains(t, result.StructuredContent.Charts, "nginx")
		assert.Contains(t, result.StructuredContent.Charts, "redis")
	} else if len(result.Content) > 0 {
		text := result.Content[0].Text
		assert.True(t, strings.Contains(text, "nginx") || strings.Contains(text, "redis"),
			"expected charts to be listed")
	}
}
