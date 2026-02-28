// Package config provides configuration loading and validation for mcp-helm.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

// Config holds all configuration for the mcp-helm server.
// It is immutable after construction via Load().
type Config struct {
	// Transport settings
	Transport string
	Listen    string

	// Helm settings
	HelmTimeout    time.Duration
	CacheSize      int
	IndexTTL       time.Duration
	MaxOutputBytes int

	// Security settings
	AllowPrivateIPs bool
	AllowedHosts    []string
	DeniedHosts     []string

	// HTTP server settings
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	// Logging
	LogLevel  string
	LogFormat string

	// Build info (set at runtime)
	Version string
	Commit  string
	Date    string
}

// envPrefix is the prefix for all environment variables.
const envPrefix = "MCP_HELM_"

// Load parses command-line flags and environment variables, validates the
// configuration, and returns an immutable Config. CLI flags take precedence
// over environment variables. Environment variables use the MCP_HELM_ prefix
// with uppercase flag names and hyphens replaced by underscores, e.g.
// --helm-timeout -> MCP_HELM_HELM_TIMEOUT.
func Load() (*Config, error) {
	return load(os.Args[1:], os.LookupEnv)
}

// load is the internal implementation that accepts args and an env lookup
// function for testability.
func load(args []string, lookupEnv func(string) (string, bool)) (*Config, error) {
	cfg := &Config{}

	fs := pflag.NewFlagSet("mcp-helm", pflag.ContinueOnError)

	// Transport flags
	fs.StringVar(&cfg.Transport, "transport", "stdio", "Transport mode: stdio, http (env: MCP_HELM_TRANSPORT)")
	fs.StringVar(&cfg.Listen, "listen", ":8012", "Listen address for HTTP mode (env: MCP_HELM_LISTEN)")

	// Helm flags
	fs.DurationVar(&cfg.HelmTimeout, "helm-timeout", 30*time.Second, "Timeout for Helm operations (env: MCP_HELM_HELM_TIMEOUT)")
	fs.IntVar(&cfg.CacheSize, "cache-size", 50, "Max charts to cache (env: MCP_HELM_CACHE_SIZE)")
	fs.DurationVar(&cfg.IndexTTL, "index-ttl", 5*time.Minute, "Repository index cache TTL (env: MCP_HELM_INDEX_TTL)")
	fs.IntVar(&cfg.MaxOutputBytes, "max-output-size", 2*1024*1024, "Max tool output bytes (env: MCP_HELM_MAX_OUTPUT_SIZE)")

	// Security flags
	fs.BoolVar(&cfg.AllowPrivateIPs, "allow-private-ips", false, "Allow URLs resolving to private IPs (env: MCP_HELM_ALLOW_PRIVATE_IPS)")
	var allowedHosts, deniedHosts string
	fs.StringVar(&allowedHosts, "allowed-hosts", "", "Comma-separated allowlist of hostnames (env: MCP_HELM_ALLOWED_HOSTS)")
	fs.StringVar(&deniedHosts, "denied-hosts", "", "Comma-separated denylist of hostnames (env: MCP_HELM_DENIED_HOSTS)")

	// Server flags
	fs.DurationVar(&cfg.ReadTimeout, "read-timeout", 30*time.Second, "HTTP read timeout (env: MCP_HELM_READ_TIMEOUT)")
	fs.DurationVar(&cfg.WriteTimeout, "write-timeout", 30*time.Second, "HTTP write timeout (env: MCP_HELM_WRITE_TIMEOUT)")

	// Logging flags
	fs.StringVar(&cfg.LogLevel, "log-level", "info", "Log level: debug, info, warn, error (env: MCP_HELM_LOG_LEVEL)")
	fs.StringVar(&cfg.LogFormat, "log-format", "json", "Log format: json, console (env: MCP_HELM_LOG_FORMAT)")

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("parsing flags: %w", err)
	}

	// Apply environment variable fallbacks for flags not explicitly set via CLI.
	applyEnvOverrides(fs, lookupEnv)

	// Parse comma-separated lists (may have been set by env var)
	if f := fs.Lookup("allowed-hosts"); f != nil {
		allowedHosts = f.Value.String()
	}
	if f := fs.Lookup("denied-hosts"); f != nil {
		deniedHosts = f.Value.String()
	}
	cfg.AllowedHosts = parseCSV(allowedHosts)
	cfg.DeniedHosts = parseCSV(deniedHosts)

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// flagEnvName converts a flag name to its environment variable name.
// e.g. "helm-timeout" -> "MCP_HELM_HELM_TIMEOUT"
func flagEnvName(flagName string) string {
	return envPrefix + strings.ToUpper(strings.ReplaceAll(flagName, "-", "_"))
}

// applyEnvOverrides checks each flag that was not explicitly set via CLI
// and applies the value from the corresponding environment variable if present.
func applyEnvOverrides(fs *pflag.FlagSet, lookupEnv func(string) (string, bool)) {
	fs.VisitAll(func(f *pflag.Flag) {
		if f.Changed {
			return // CLI flag was explicitly set, takes precedence
		}
		envName := flagEnvName(f.Name)
		val, ok := lookupEnv(envName)
		if !ok || val == "" {
			return
		}
		_ = fs.Set(f.Name, val)
	})
}

// validate checks that the configuration is valid.
func (c *Config) validate() error {
	var errs []error

	// Transport validation
	switch c.Transport {
	case "stdio", "http":
		// valid
	default:
		errs = append(errs, fmt.Errorf("invalid transport %q: must be stdio or http", c.Transport))
	}

	// Listen address required for HTTP
	if c.Transport == "http" && c.Listen == "" {
		errs = append(errs, errors.New("--listen required for http transport"))
	}

	// Positive durations
	if c.HelmTimeout <= 0 {
		errs = append(errs, errors.New("--helm-timeout must be positive"))
	}
	if c.IndexTTL <= 0 {
		errs = append(errs, errors.New("--index-ttl must be positive"))
	}
	if c.ReadTimeout <= 0 {
		errs = append(errs, errors.New("--read-timeout must be positive"))
	}
	if c.WriteTimeout <= 0 {
		errs = append(errs, errors.New("--write-timeout must be positive"))
	}

	// Positive integers
	if c.CacheSize <= 0 {
		errs = append(errs, errors.New("--cache-size must be positive"))
	}
	if c.MaxOutputBytes <= 0 {
		errs = append(errs, errors.New("--max-output-size must be positive"))
	}

	// Log level validation
	switch c.LogLevel {
	case "debug", "info", "warn", "error":
		// valid
	default:
		errs = append(errs, fmt.Errorf("invalid log-level %q: must be debug, info, warn, or error", c.LogLevel))
	}

	// Log format validation
	switch c.LogFormat {
	case "json", "console":
		// valid
	default:
		errs = append(errs, fmt.Errorf("invalid log-format %q: must be json or console", c.LogFormat))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// parseCSV splits a comma-separated string into trimmed, non-empty parts.
func parseCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
