package config

import (
	"net"
	"strings"
	"testing"
	"time"
)

func TestParseCSV(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
		{
			name:  "whitespace only",
			input: "   ",
			want:  nil,
		},
		{
			name:  "single value",
			input: "example.com",
			want:  []string{"example.com"},
		},
		{
			name:  "multiple values",
			input: "a.com,b.com,c.com",
			want:  []string{"a.com", "b.com", "c.com"},
		},
		{
			name:  "values with spaces",
			input: " a.com , b.com , c.com ",
			want:  []string{"a.com", "b.com", "c.com"},
		},
		{
			name:  "empty items filtered",
			input: "a.com,,b.com,",
			want:  []string{"a.com", "b.com"},
		},
		{
			name:  "only commas",
			input: ",,,",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCSV(tt.input)
			if !slicesEqual(got, tt.want) {
				t.Errorf("parseCSV(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseCIDRs(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLen int
		wantErr bool
	}{
		{
			name:    "empty string",
			input:   "",
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "single IPv4 CIDR",
			input:   "10.0.0.0/8",
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "multiple CIDRs",
			input:   "10.0.0.0/8,172.16.0.0/12,192.168.0.0/16",
			wantLen: 3,
			wantErr: false,
		},
		{
			name:    "IPv6 CIDR",
			input:   "::1/128",
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "mixed IPv4 and IPv6",
			input:   "10.0.0.0/8,::1/128",
			wantLen: 2,
			wantErr: false,
		},
		{
			name:    "invalid CIDR",
			input:   "not-a-cidr",
			wantLen: 0,
			wantErr: true,
		},
		{
			name:    "partially invalid",
			input:   "10.0.0.0/8,invalid",
			wantLen: 0,
			wantErr: true,
		},
		{
			name:    "IP without mask",
			input:   "10.0.0.1",
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCIDRs(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCIDRs(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("parseCIDRs(%q) returned %d CIDRs, want %d", tt.input, len(got), tt.wantLen)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	baseConfig := func() *Config {
		return &Config{
			Transport:      "stdio",
			Listen:         ":8012",
			HelmTimeout:    30 * time.Second,
			CacheSize:      50,
			IndexTTL:       5 * time.Minute,
			MaxOutputBytes: 2 * 1024 * 1024,
			ReadTimeout:    30 * time.Second,
			WriteTimeout:   30 * time.Second,
			LogLevel:       "info",
			LogFormat:      "json",
		}
	}

	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr string
	}{
		{
			name:    "valid stdio config",
			modify:  func(c *Config) {},
			wantErr: "",
		},
		{
			name:    "valid http config",
			modify:  func(c *Config) { c.Transport = "http" },
			wantErr: "",
		},
		{
			name:    "invalid transport",
			modify:  func(c *Config) { c.Transport = "grpc" },
			wantErr: "invalid transport",
		},
		{
			name: "http without listen address",
			modify: func(c *Config) {
				c.Transport = "http"
				c.Listen = ""
			},
			wantErr: "--listen required",
		},
		{
			name:    "zero helm timeout",
			modify:  func(c *Config) { c.HelmTimeout = 0 },
			wantErr: "--helm-timeout must be positive",
		},
		{
			name:    "negative helm timeout",
			modify:  func(c *Config) { c.HelmTimeout = -1 * time.Second },
			wantErr: "--helm-timeout must be positive",
		},
		{
			name:    "zero index TTL",
			modify:  func(c *Config) { c.IndexTTL = 0 },
			wantErr: "--index-ttl must be positive",
		},
		{
			name:    "zero cache size",
			modify:  func(c *Config) { c.CacheSize = 0 },
			wantErr: "--cache-size must be positive",
		},
		{
			name:    "negative cache size",
			modify:  func(c *Config) { c.CacheSize = -1 },
			wantErr: "--cache-size must be positive",
		},
		{
			name:    "zero max output",
			modify:  func(c *Config) { c.MaxOutputBytes = 0 },
			wantErr: "--max-output-size must be positive",
		},
		{
			name:    "zero read timeout",
			modify:  func(c *Config) { c.ReadTimeout = 0 },
			wantErr: "--read-timeout must be positive",
		},
		{
			name:    "zero write timeout",
			modify:  func(c *Config) { c.WriteTimeout = 0 },
			wantErr: "--write-timeout must be positive",
		},
		{
			name: "rate limit with zero RPS",
			modify: func(c *Config) {
				c.RateLimitEnabled = true
				c.RateLimitRPS = 0
			},
			wantErr: "--rate-limit-rps must be positive",
		},
		{
			name: "rate limit with negative RPS",
			modify: func(c *Config) {
				c.RateLimitEnabled = true
				c.RateLimitRPS = -1
			},
			wantErr: "--rate-limit-rps must be positive",
		},
		{
			name: "rate limit with zero burst",
			modify: func(c *Config) {
				c.RateLimitEnabled = true
				c.RateLimitRPS = 10
				c.RateLimitBurst = 0
			},
			wantErr: "--rate-limit-burst must be positive",
		},
		{
			name: "rate limit disabled ignores RPS/burst",
			modify: func(c *Config) {
				c.RateLimitEnabled = false
				c.RateLimitRPS = 0
				c.RateLimitBurst = 0
			},
			wantErr: "", // Should pass - rate limiting is disabled
		},
		{
			name:    "invalid log level",
			modify:  func(c *Config) { c.LogLevel = "trace" },
			wantErr: "invalid log-level",
		},
		{
			name:    "valid debug log level",
			modify:  func(c *Config) { c.LogLevel = "debug" },
			wantErr: "",
		},
		{
			name:    "valid warn log level",
			modify:  func(c *Config) { c.LogLevel = "warn" },
			wantErr: "",
		},
		{
			name:    "valid error log level",
			modify:  func(c *Config) { c.LogLevel = "error" },
			wantErr: "",
		},
		{
			name:    "invalid log format",
			modify:  func(c *Config) { c.LogFormat = "text" },
			wantErr: "invalid log-format",
		},
		{
			name:    "valid console log format",
			modify:  func(c *Config) { c.LogFormat = "console" },
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := baseConfig()
			tt.modify(cfg)

			err := cfg.validate()

			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("validate() unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("validate() expected error containing %q, got nil", tt.wantErr)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("validate() error = %v, want error containing %q", err, tt.wantErr)
				}
			}
		})
	}
}

func TestConfig_Validate_MultipleErrors(t *testing.T) {
	cfg := &Config{
		Transport:        "invalid",
		HelmTimeout:      0,
		IndexTTL:         0,
		CacheSize:        0,
		MaxOutputBytes:   0,
		ReadTimeout:      0,
		WriteTimeout:     0,
		LogLevel:         "invalid",
		LogFormat:        "invalid",
		RateLimitEnabled: true,
		RateLimitRPS:     0,
		RateLimitBurst:   0,
	}

	err := cfg.validate()
	if err == nil {
		t.Fatal("validate() expected error, got nil")
	}

	// Should contain all error messages
	errStr := err.Error()
	expectedErrors := []string{
		"invalid transport",
		"--helm-timeout must be positive",
		"--index-ttl must be positive",
		"--cache-size must be positive",
		"--max-output-size must be positive",
		"--read-timeout must be positive",
		"--write-timeout must be positive",
		"--rate-limit-rps must be positive",
		"--rate-limit-burst must be positive",
		"invalid log-level",
		"invalid log-format",
	}

	for _, expected := range expectedErrors {
		if !strings.Contains(errStr, expected) {
			t.Errorf("validate() error missing %q", expected)
		}
	}
}

func TestConfig_IsTrustedProxy(t *testing.T) {
	_, network1, _ := net.ParseCIDR("10.0.0.0/8")
	_, network2, _ := net.ParseCIDR("192.168.0.0/16")

	cfg := &Config{
		TrustedProxies: []net.IPNet{*network1, *network2},
	}

	tests := []struct {
		name string
		ip   string
		want bool
	}{
		{"in first network", "10.0.0.1", true},
		{"in first network edge", "10.255.255.255", true},
		{"in second network", "192.168.1.1", true},
		{"not in any network", "172.16.0.1", false},
		{"public IP", "8.8.8.8", false},
		{"localhost", "127.0.0.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			got := cfg.IsTrustedProxy(ip)
			if got != tt.want {
				t.Errorf("IsTrustedProxy(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestConfig_IsTrustedProxy_EmptyList(t *testing.T) {
	cfg := &Config{
		TrustedProxies: nil,
	}

	ip := net.ParseIP("10.0.0.1")
	if cfg.IsTrustedProxy(ip) {
		t.Error("IsTrustedProxy() should return false with empty list")
	}
}

// Helper function to compare string slices
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
