package helm

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestDefaultOptions(t *testing.T) {
	opts := defaultOptions()

	assert.Equal(t, 30*time.Second, opts.timeout)
	assert.Equal(t, 5*time.Minute, opts.indexTTL)
	assert.Equal(t, 100, opts.indexCacheSize)
	assert.Equal(t, 50, opts.chartCacheSize)
	assert.Equal(t, 2*1024*1024, opts.maxOutputBytes)
	assert.Equal(t, filepath.Join(os.TempDir(), "mcp-helm-cache"), opts.cacheDir)
	assert.NotNil(t, opts.logger)
	assert.False(t, opts.allowPrivateIPs)
	assert.Nil(t, opts.allowedHosts)
	assert.Nil(t, opts.deniedHosts)
}

func TestWithTimeout(t *testing.T) {
	t.Run("positive duration", func(t *testing.T) {
		opts := defaultOptions()
		WithTimeout(60 * time.Second)(opts)

		assert.Equal(t, 60*time.Second, opts.timeout)
	})

	t.Run("zero duration ignored", func(t *testing.T) {
		opts := defaultOptions()
		WithTimeout(0)(opts)

		assert.Equal(t, 30*time.Second, opts.timeout) // default unchanged
	})

	t.Run("negative duration ignored", func(t *testing.T) {
		opts := defaultOptions()
		WithTimeout(-1 * time.Second)(opts)

		assert.Equal(t, 30*time.Second, opts.timeout) // default unchanged
	})
}

func TestWithIndexTTL(t *testing.T) {
	t.Run("positive duration", func(t *testing.T) {
		opts := defaultOptions()
		WithIndexTTL(10 * time.Minute)(opts)

		assert.Equal(t, 10*time.Minute, opts.indexTTL)
	})

	t.Run("zero duration ignored", func(t *testing.T) {
		opts := defaultOptions()
		WithIndexTTL(0)(opts)

		assert.Equal(t, 5*time.Minute, opts.indexTTL) // default unchanged
	})
}

func TestWithIndexCacheSize(t *testing.T) {
	t.Run("positive size", func(t *testing.T) {
		opts := defaultOptions()
		WithIndexCacheSize(200)(opts)

		assert.Equal(t, 200, opts.indexCacheSize)
	})

	t.Run("zero size ignored", func(t *testing.T) {
		opts := defaultOptions()
		WithIndexCacheSize(0)(opts)

		assert.Equal(t, 100, opts.indexCacheSize) // default unchanged
	})

	t.Run("negative size ignored", func(t *testing.T) {
		opts := defaultOptions()
		WithIndexCacheSize(-10)(opts)

		assert.Equal(t, 100, opts.indexCacheSize) // default unchanged
	})
}

func TestWithChartCacheSize(t *testing.T) {
	t.Run("positive size", func(t *testing.T) {
		opts := defaultOptions()
		WithChartCacheSize(100)(opts)

		assert.Equal(t, 100, opts.chartCacheSize)
	})

	t.Run("zero size ignored", func(t *testing.T) {
		opts := defaultOptions()
		WithChartCacheSize(0)(opts)

		assert.Equal(t, 50, opts.chartCacheSize) // default unchanged
	})

	t.Run("negative size ignored", func(t *testing.T) {
		opts := defaultOptions()
		WithChartCacheSize(-10)(opts)

		assert.Equal(t, 50, opts.chartCacheSize) // default unchanged
	})
}

func TestWithMaxOutputBytes(t *testing.T) {
	t.Run("positive value", func(t *testing.T) {
		opts := defaultOptions()
		WithMaxOutputBytes(1024 * 1024)(opts)

		assert.Equal(t, 1024*1024, opts.maxOutputBytes)
	})

	t.Run("zero value ignored", func(t *testing.T) {
		opts := defaultOptions()
		WithMaxOutputBytes(0)(opts)

		assert.Equal(t, 2*1024*1024, opts.maxOutputBytes) // default unchanged
	})
}

func TestWithAllowPrivateIPs(t *testing.T) {
	t.Run("enable private IPs", func(t *testing.T) {
		opts := defaultOptions()
		WithAllowPrivateIPs(true)(opts)

		assert.True(t, opts.allowPrivateIPs)
	})

	t.Run("disable private IPs", func(t *testing.T) {
		opts := defaultOptions()
		opts.allowPrivateIPs = true // start enabled
		WithAllowPrivateIPs(false)(opts)

		assert.False(t, opts.allowPrivateIPs)
	})
}

func TestWithAllowedHosts(t *testing.T) {
	t.Run("set allowed hosts", func(t *testing.T) {
		opts := defaultOptions()
		hosts := []string{"charts.bitnami.com", "charts.helm.sh"}
		WithAllowedHosts(hosts)(opts)

		assert.Equal(t, hosts, opts.allowedHosts)
	})

	t.Run("set empty hosts", func(t *testing.T) {
		opts := defaultOptions()
		WithAllowedHosts([]string{})(opts)

		assert.Empty(t, opts.allowedHosts)
	})

	t.Run("set nil hosts", func(t *testing.T) {
		opts := defaultOptions()
		opts.allowedHosts = []string{"existing.com"}
		WithAllowedHosts(nil)(opts)

		assert.Nil(t, opts.allowedHosts)
	})
}

func TestWithDeniedHosts(t *testing.T) {
	t.Run("set denied hosts", func(t *testing.T) {
		opts := defaultOptions()
		hosts := []string{"evil.com", "malware.net"}
		WithDeniedHosts(hosts)(opts)

		assert.Equal(t, hosts, opts.deniedHosts)
	})

	t.Run("set nil hosts", func(t *testing.T) {
		opts := defaultOptions()
		WithDeniedHosts(nil)(opts)

		assert.Nil(t, opts.deniedHosts)
	})
}

func TestWithCacheDir(t *testing.T) {
	t.Run("set cache directory", func(t *testing.T) {
		opts := defaultOptions()
		WithCacheDir("/custom/cache")(opts)

		assert.Equal(t, "/custom/cache", opts.cacheDir)
	})

	t.Run("empty string ignored", func(t *testing.T) {
		opts := defaultOptions()
		WithCacheDir("")(opts)

		assert.Equal(t, filepath.Join(os.TempDir(), "mcp-helm-cache"), opts.cacheDir)
	})
}

func TestWithLogger(t *testing.T) {
	t.Run("set logger", func(t *testing.T) {
		opts := defaultOptions()
		logger := zap.NewExample()
		WithLogger(logger)(opts)

		assert.Equal(t, logger, opts.logger)
	})

	t.Run("nil logger ignored", func(t *testing.T) {
		opts := defaultOptions()
		originalLogger := opts.logger
		WithLogger(nil)(opts)

		assert.Equal(t, originalLogger, opts.logger)
	})
}

func TestOptionsChaining(t *testing.T) {
	// Test that multiple options can be applied in sequence
	opts := defaultOptions()

	options := []Option{
		WithTimeout(45 * time.Second),
		WithIndexTTL(15 * time.Minute),
		WithIndexCacheSize(200),
		WithChartCacheSize(75),
		WithMaxOutputBytes(4 * 1024 * 1024),
		WithAllowPrivateIPs(true),
		WithAllowedHosts([]string{"example.com"}),
		WithDeniedHosts([]string{"blocked.com"}),
		WithCacheDir("/tmp/test-cache"),
	}

	for _, opt := range options {
		opt(opts)
	}

	assert.Equal(t, 45*time.Second, opts.timeout)
	assert.Equal(t, 15*time.Minute, opts.indexTTL)
	assert.Equal(t, 200, opts.indexCacheSize)
	assert.Equal(t, 75, opts.chartCacheSize)
	assert.Equal(t, 4*1024*1024, opts.maxOutputBytes)
	assert.True(t, opts.allowPrivateIPs)
	assert.Equal(t, []string{"example.com"}, opts.allowedHosts)
	assert.Equal(t, []string{"blocked.com"}, opts.deniedHosts)
	assert.Equal(t, "/tmp/test-cache", opts.cacheDir)
}
