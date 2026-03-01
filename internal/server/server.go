// Package server provides HTTP server infrastructure for mcp-helm.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"

	"github.com/Kubedoll-Heavy-Industries/mcp-helm/internal/config"
)

// Server wraps the MCP server with HTTP transport and lifecycle management.
type Server struct {
	cfg       *config.Config
	logger    *zap.Logger
	mcpServer *mcp.Server
}

// New creates a new Server.
func New(cfg *config.Config, logger *zap.Logger, mcpServer *mcp.Server) *Server {
	return &Server{
		cfg:       cfg,
		logger:    logger,
		mcpServer: mcpServer,
	}
}

// Run starts the server and blocks until the context is cancelled.
// It handles graceful shutdown automatically.
func (s *Server) Run(ctx context.Context) error {
	switch s.cfg.Transport {
	case "stdio":
		return s.runStdio(ctx)
	case "http":
		return s.runHTTP(ctx)
	default:
		return fmt.Errorf("unsupported transport: %s", s.cfg.Transport)
	}
}

// runStdio runs the MCP server over stdio.
func (s *Server) runStdio(ctx context.Context) error {
	s.logger.Info("starting MCP server",
		zap.String("transport", "stdio"),
		zap.String("version", s.cfg.Version),
	)

	return s.mcpServer.Run(ctx, &mcp.StdioTransport{})
}

// runHTTP runs the MCP server over HTTP with Streamable HTTP transport.
func (s *Server) runHTTP(ctx context.Context) error {
	// Create MCP HTTP handler
	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(_ *http.Request) *mcp.Server { return s.mcpServer },
		&mcp.StreamableHTTPOptions{},
	)

	// Wrap MCP handler to convert 202 empty responses to 200 with body.
	// Cloudflare Containers proxy cannot handle 202 with empty body, causing
	// internal errors that break MCP session establishment.
	wrappedMCP := &rewrite202Handler{next: mcpHandler}

	// Build router
	mux := http.NewServeMux()
	mux.Handle("/mcp", wrappedMCP)
	mux.Handle("/mcp/", wrappedMCP)
	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/readyz", s.handleReadyz)

	// Apply middleware
	var handler http.Handler = &jsonNotFoundMux{mux: mux}
	handler = RecoveryMiddleware(s.logger)(handler)
	handler = LoggingMiddleware(s.logger)(handler)

	// Create HTTP server with timeouts
	srv := &http.Server{
		Addr:              s.cfg.Listen,
		Handler:           handler,
		ReadTimeout:       s.cfg.ReadTimeout,
		WriteTimeout:      s.cfg.WriteTimeout,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("starting MCP server",
			zap.String("transport", "http"),
			zap.String("addr", s.cfg.Listen),
			zap.String("version", s.cfg.Version),
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	// Wait for shutdown signal or error
	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		s.logger.Info("shutting down server")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown error: %w", err)
		}

		s.logger.Info("server stopped gracefully")
		return nil
	}
}

// rewrite202Handler wraps an http.Handler so that 202 Accepted responses (used
// by the MCP SDK for JSON-RPC notifications) get a minimal JSON body. Cloudflare
// Containers' internal proxy throws "internal error" when it receives a 202 with
// an empty body, breaking the MCP session handshake.
type rewrite202Handler struct {
	next http.Handler
}

func (h *rewrite202Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rec := &statusRecorder{ResponseWriter: w}
	h.next.ServeHTTP(rec, r)
	if rec.status == http.StatusAccepted && !rec.wroteBody {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}
}

// statusRecorder captures the status code and whether a body was written,
// forwarding everything else to the underlying ResponseWriter.
type statusRecorder struct {
	http.ResponseWriter
	status    int
	wroteBody bool
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	if code != http.StatusAccepted {
		r.ResponseWriter.WriteHeader(code)
	}
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	r.wroteBody = true
	if r.status == http.StatusAccepted {
		// Buffer — don't forward the write for 202, let the wrapper handle it.
		return len(b), nil
	}
	return r.ResponseWriter.Write(b)
}

func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// jsonNotFoundMux wraps an http.ServeMux so that unmatched routes return a JSON
// 404 instead of the default plain-text "404 page not found". This is necessary
// because MCP clients (e.g. Claude Code) probe /.well-known/oauth-authorization-server
// and expect a JSON response even on 404.
type jsonNotFoundMux struct {
	mux *http.ServeMux
}

func (m *jsonNotFoundMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check if the mux has a handler for this path.
	_, pattern := m.mux.Handler(r)
	if pattern == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		pathJSON, _ := json.Marshal(r.URL.Path)
		_, _ = fmt.Fprintf(w, `{"error":"not_found","error_description":"no route for %s"}`, string(pathJSON))
		return
	}
	m.mux.ServeHTTP(w, r)
}

// handleHealthz handles liveness probe requests.
func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `{"status":"ok","timestamp":"%s"}`, time.Now().UTC().Format(time.RFC3339))
}

// handleReadyz handles readiness probe requests.
func (s *Server) handleReadyz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `{"status":"ready","timestamp":"%s"}`, time.Now().UTC().Format(time.RFC3339))
}
