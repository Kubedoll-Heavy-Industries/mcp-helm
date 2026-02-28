// Package main provides the entry point for the mcp-helm server.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/Kubedoll-Heavy-Industries/mcp-helm/internal/config"
	"github.com/Kubedoll-Heavy-Industries/mcp-helm/internal/handler"
	"github.com/Kubedoll-Heavy-Industries/mcp-helm/internal/helm"
	"github.com/Kubedoll-Heavy-Industries/mcp-helm/internal/server"
)

// Build information, set by goreleaser.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Set build info
	cfg.Version = version
	cfg.Commit = commit
	cfg.Date = date

	// Create logger
	logger, err := newLogger(cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}
	defer func() { _ = logger.Sync() }()

	// Create Helm client
	helmClient := helm.NewClient(
		helm.WithTimeout(cfg.HelmTimeout),
		helm.WithIndexTTL(cfg.IndexTTL),
		helm.WithChartCacheSize(cfg.CacheSize),
		helm.WithMaxOutputBytes(cfg.MaxOutputBytes),
		helm.WithAllowPrivateIPs(cfg.AllowPrivateIPs),
		helm.WithAllowedHosts(cfg.AllowedHosts),
		helm.WithDeniedHosts(cfg.DeniedHosts),
		helm.WithLogger(logger),
	)

	// Create MCP server with capabilities
	mcpServer := mcp.NewServer(
		&mcp.Implementation{
			Name:    "Helm MCP Server",
			Version: fmt.Sprintf("v%s (commit: %s, date: %s)", version, commit, date),
		},
		&mcp.ServerOptions{
			Instructions: "Access Helm chart repositories to fetch values.yaml, schemas, dependencies, and chart contents. Use progressive disclosure (depth, max_array_items) to manage response sizes.",
			HasTools:     true,
			InitializedHandler: func(_ context.Context, req *mcp.InitializedRequest) {
				if req.Session != nil {
					logger.Info("MCP client connected",
						zap.String("session_id", req.Session.ID()),
					)
				}
			},
		},
	)

	// Register handlers
	h := handler.New(helmClient, logger)
	h.Register(mcpServer)

	// Create and run server
	srv := server.New(cfg, logger, mcpServer)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	return srv.Run(ctx)
}

// newLogger creates a zap logger with the specified level and format.
func newLogger(level, format string) (*zap.Logger, error) {
	var lvl zapcore.Level
	switch level {
	case "debug":
		lvl = zap.DebugLevel
	case "info":
		lvl = zap.InfoLevel
	case "warn":
		lvl = zap.WarnLevel
	case "error":
		lvl = zap.ErrorLevel
	default:
		lvl = zap.InfoLevel
	}

	var cfg zap.Config
	if format == "console" {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}
	cfg.Level = zap.NewAtomicLevelAt(lvl)

	return cfg.Build()
}
