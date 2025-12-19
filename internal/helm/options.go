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
	cacheSize       int
	maxOutputBytes  int
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
		cacheSize:      50,
		maxOutputBytes: 2 * 1024 * 1024,
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

// WithCacheSize sets the maximum number of charts to cache.
func WithCacheSize(n int) Option {
	return func(o *clientOptions) {
		if n > 0 {
			o.cacheSize = n
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
