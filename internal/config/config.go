// Package config provides configuration loading and validation for mcp-helm.
package config

import (
	"errors"
	"fmt"
	"net"
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
	TrustedProxies  []net.IPNet

	// Rate limiting
	RateLimitEnabled bool
	RateLimitRPS     float64
	RateLimitBurst   int

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

// Load parses command-line flags and environment variables, validates the
// configuration, and returns an immutable Config. Errors are returned for
// invalid configurations.
func Load() (*Config, error) {
	cfg := &Config{}

	fs := pflag.NewFlagSet("mcp-helm", pflag.ContinueOnError)

	// Transport flags
	fs.StringVar(&cfg.Transport, "transport", "stdio", "Transport mode: stdio, http")
	fs.StringVar(&cfg.Listen, "listen", ":8012", "Listen address for HTTP mode")

	// Helm flags
	fs.DurationVar(&cfg.HelmTimeout, "helm-timeout", 30*time.Second, "Timeout for Helm operations")
	fs.IntVar(&cfg.CacheSize, "cache-size", 50, "Max charts to cache")
	fs.DurationVar(&cfg.IndexTTL, "index-ttl", 5*time.Minute, "Repository index cache TTL")
	fs.IntVar(&cfg.MaxOutputBytes, "max-output-size", 2*1024*1024, "Max tool output bytes")

	// Security flags
	fs.BoolVar(&cfg.AllowPrivateIPs, "allow-private-ips", false, "Allow URLs resolving to private IPs")
	var allowedHosts, deniedHosts, trustedProxies string
	fs.StringVar(&allowedHosts, "allowed-hosts", "", "Comma-separated allowlist of hostnames")
	fs.StringVar(&deniedHosts, "denied-hosts", "", "Comma-separated denylist of hostnames")
	fs.StringVar(&trustedProxies, "trusted-proxies", "", "CIDR ranges of trusted proxies (comma-separated)")

	// Rate limiting flags
	fs.BoolVar(&cfg.RateLimitEnabled, "rate-limit", false, "Enable rate limiting (HTTP mode only)")
	fs.Float64Var(&cfg.RateLimitRPS, "rate-limit-rps", 10, "Requests per second per client")
	fs.IntVar(&cfg.RateLimitBurst, "rate-limit-burst", 20, "Burst capacity")

	// Server flags
	fs.DurationVar(&cfg.ReadTimeout, "read-timeout", 30*time.Second, "HTTP read timeout")
	fs.DurationVar(&cfg.WriteTimeout, "write-timeout", 30*time.Second, "HTTP write timeout")

	// Logging flags
	fs.StringVar(&cfg.LogLevel, "log-level", "info", "Log level: debug, info, warn, error")
	fs.StringVar(&cfg.LogFormat, "log-format", "json", "Log format: json, console")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, fmt.Errorf("parsing flags: %w", err)
	}

	// Parse comma-separated lists
	cfg.AllowedHosts = parseCSV(allowedHosts)
	cfg.DeniedHosts = parseCSV(deniedHosts)

	// Parse trusted proxies as CIDRs
	if trustedProxies != "" {
		cidrs, err := parseCIDRs(trustedProxies)
		if err != nil {
			return nil, fmt.Errorf("parsing trusted-proxies: %w", err)
		}
		cfg.TrustedProxies = cidrs
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
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

	// Rate limit validation
	if c.RateLimitEnabled {
		if c.RateLimitRPS <= 0 {
			errs = append(errs, errors.New("--rate-limit-rps must be positive"))
		}
		if c.RateLimitBurst <= 0 {
			errs = append(errs, errors.New("--rate-limit-burst must be positive"))
		}
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

// parseCIDRs parses a comma-separated list of CIDR ranges.
func parseCIDRs(s string) ([]net.IPNet, error) {
	parts := parseCSV(s)
	if len(parts) == 0 {
		return nil, nil
	}

	result := make([]net.IPNet, 0, len(parts))
	for _, p := range parts {
		_, network, err := net.ParseCIDR(p)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR %q: %w", p, err)
		}
		result = append(result, *network)
	}
	return result, nil
}

// IsTrustedProxy checks if the given IP is in the trusted proxies list.
func (c *Config) IsTrustedProxy(ip net.IP) bool {
	for _, network := range c.TrustedProxies {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}
