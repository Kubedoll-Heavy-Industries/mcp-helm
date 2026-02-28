package helm

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

// Option configures a Client.
type Option func(*clientOptions)

// clientOptions holds the configuration for a Client.
type clientOptions struct {
	timeout         time.Duration
	indexTTL        time.Duration
	indexCacheSize  int
	chartCacheSize  int
	maxOutputBytes  int
	maxChartBytes   int64
	allowPrivateIPs bool
	allowedHosts    []string
	deniedHosts     []string
	cacheDir        string
	logger          *zap.Logger
}

// defaultOptions returns the default client options.
func defaultOptions() *clientOptions {
	return &clientOptions{
		timeout:        30 * time.Second,
		indexTTL:       5 * time.Minute,
		indexCacheSize: 100,
		chartCacheSize: 50,
		maxOutputBytes: 2 * 1024 * 1024,
		maxChartBytes:  50 * 1024 * 1024, // 50 MB
		cacheDir:       filepath.Join(os.TempDir(), "mcp-helm-cache"),
		logger:         zap.NewNop(),
	}
}

// WithTimeout sets the timeout for Helm operations.
func WithTimeout(d time.Duration) Option {
	return func(o *clientOptions) {
		if d > 0 {
			o.timeout = d
		}
	}
}

// WithIndexTTL sets the TTL for cached repository indexes.
func WithIndexTTL(d time.Duration) Option {
	return func(o *clientOptions) {
		if d > 0 {
			o.indexTTL = d
		}
	}
}

// WithIndexCacheSize sets the maximum number of repository indexes to cache.
func WithIndexCacheSize(n int) Option {
	return func(o *clientOptions) {
		if n > 0 {
			o.indexCacheSize = n
		}
	}
}

// WithChartCacheSize sets the maximum number of charts to cache.
func WithChartCacheSize(n int) Option {
	return func(o *clientOptions) {
		if n > 0 {
			o.chartCacheSize = n
		}
	}
}

// WithMaxOutputBytes sets the maximum output size for tool responses.
func WithMaxOutputBytes(n int) Option {
	return func(o *clientOptions) {
		if n > 0 {
			o.maxOutputBytes = n
		}
	}
}

// WithMaxChartBytes sets the maximum allowed size for downloaded chart files.
func WithMaxChartBytes(n int64) Option {
	return func(o *clientOptions) {
		if n > 0 {
			o.maxChartBytes = n
		}
	}
}

// WithAllowPrivateIPs allows URLs that resolve to private IP addresses.
func WithAllowPrivateIPs(allow bool) Option {
	return func(o *clientOptions) {
		o.allowPrivateIPs = allow
	}
}

// WithAllowedHosts sets the allowlist of hostnames.
func WithAllowedHosts(hosts []string) Option {
	return func(o *clientOptions) {
		o.allowedHosts = hosts
	}
}

// WithDeniedHosts sets the denylist of hostnames.
func WithDeniedHosts(hosts []string) Option {
	return func(o *clientOptions) {
		o.deniedHosts = hosts
	}
}

// WithCacheDir sets the directory for Helm caches.
func WithCacheDir(dir string) Option {
	return func(o *clientOptions) {
		if dir != "" {
			o.cacheDir = dir
		}
	}
}

// WithLogger sets the logger for the client.
func WithLogger(l *zap.Logger) Option {
	return func(o *clientOptions) {
		if l != nil {
			o.logger = l
		}
	}
}
