package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/Kubedoll-Heavy-Industries/mcp-helm/internal/tools"
	"github.com/Kubedoll-Heavy-Industries/mcp-helm/lib/helm_client"
	"github.com/Kubedoll-Heavy-Industries/mcp-helm/lib/logger"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type config struct {
	mode               string
	httpListenAddr     string
	heartbeatInterval  time.Duration
	sseKeepAlive       time.Duration
	helmTimeout        time.Duration
	allowPrivateIPs    bool
	allowedHosts       string
	deniedHosts        string
	maxToolOutputBytes int
}

var cfg config

func main() {
	rootCmd := &cobra.Command{
		Use:     "mcp-helm",
		Short:   "MCP server for Helm repositories and charts",
		Long:    "An MCP (Model Context Protocol) server that provides tools for interacting with Helm repositories and charts.",
		Version: fmt.Sprintf("%s (commit: %s, date: %s)", version, commit, date),
		Run:     run,
	}

	rootCmd.Flags().StringVarP(&cfg.mode, "mode", "m", "stdio", "Transport mode: stdio, http, or sse (deprecated)")
	rootCmd.Flags().StringVarP(&cfg.httpListenAddr, "addr", "a", ":8012", "Listen address for http/sse modes")
	rootCmd.Flags().DurationVar(&cfg.heartbeatInterval, "heartbeat", 30*time.Second, "Heartbeat interval for http mode")
	rootCmd.Flags().DurationVar(&cfg.sseKeepAlive, "sse-keepalive", 30*time.Second, "Keep-alive interval for sse mode (deprecated)")
	rootCmd.Flags().DurationVar(&cfg.helmTimeout, "timeout", 30*time.Second, "Timeout for Helm operations")
	rootCmd.Flags().BoolVar(&cfg.allowPrivateIPs, "allow-private-ips", false, "Allow URLs resolving to private/loopback IPs")
	rootCmd.Flags().StringVar(&cfg.allowedHosts, "allowed-hosts", "", "Comma-separated allowlist of hostnames")
	rootCmd.Flags().StringVar(&cfg.deniedHosts, "denied-hosts", "", "Comma-separated denylist of hostnames")
	rootCmd.Flags().IntVar(&cfg.maxToolOutputBytes, "max-output", 2*1024*1024, "Maximum tool output size in bytes")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	logger.Init()
	defer logger.Stop()

	switch cfg.mode {
	case "stdio", "http", "sse":
	default:
		logger.Error("Invalid mode", zap.String("mode", cfg.mode), zap.String("valid", "stdio, http, sse"))
		os.Exit(1)
	}

	if (cfg.mode == "http" || cfg.mode == "sse") && cfg.httpListenAddr == "" {
		logger.Error("Listen address required for http/sse mode")
		os.Exit(1)
	}

	if cfg.mode == "sse" {
		logger.Warn("SSE transport is deprecated, consider using http (Streamable HTTP) instead")
	}

	ctx := context.Background()

	s := mcp.NewServer(
		&mcp.Implementation{
			Name:    "Helm MCP Server",
			Version: fmt.Sprintf("v%s (commit: %s, date: %s)", version, commit, date),
		},
		nil,
	)

	helmClient := helm_client.NewClientWithOptions(helm_client.Options{
		Timeout:            cfg.helmTimeout,
		AllowPrivateIPs:    cfg.allowPrivateIPs,
		AllowedHosts:       splitCSVHosts(cfg.allowedHosts),
		DeniedHosts:        splitCSVHosts(cfg.deniedHosts),
		MaxToolOutputBytes: cfg.maxToolOutputBytes,
	})
	tools.Register(s, helmClient)

	logger.Info("Starting MCP Helm server",
		zap.String("version", version),
		zap.String("commit", commit),
		zap.String("date", date),
		zap.String("mode", cfg.mode),
		zap.String("addr", cfg.httpListenAddr),
	)

	switch cfg.mode {
	case "stdio":
		if err := s.Run(ctx, &mcp.StdioTransport{}); err != nil {
			logger.Error("Failed to start stdio transport", zap.Error(err))
		}
	case "http":
		h := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server { return s }, &mcp.StreamableHTTPOptions{})
		mux := http.NewServeMux()
		mux.Handle("/mcp", h)
		mux.Handle("/mcp/", h)

		srv := &http.Server{Addr: cfg.httpListenAddr, Handler: mux}
		if err := srv.ListenAndServe(); err != nil {
			logger.Error("Failed to start HTTP server", zap.Error(err))
		}
	case "sse":
		h := mcp.NewSSEHandler(func(_ *http.Request) *mcp.Server { return s }, &mcp.SSEOptions{})
		mux := http.NewServeMux()
		mux.Handle("/sse", h)
		mux.Handle("/message", h)

		srv := &http.Server{Addr: cfg.httpListenAddr, Handler: mux}
		if err := srv.ListenAndServe(); err != nil {
			logger.Error("Failed to start SSE server", zap.Error(err))
		}
	}
}

func splitCSVHosts(in string) []string {
	trimmed := strings.TrimSpace(in)
	if trimmed == "" {
		return nil
	}
	parts := strings.Split(trimmed, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if pp := strings.TrimSpace(p); pp != "" {
			out = append(out, pp)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
