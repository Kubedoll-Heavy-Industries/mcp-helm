package config

import (
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
		Transport:      "invalid",
		HelmTimeout:    0,
		IndexTTL:       0,
		CacheSize:      0,
		MaxOutputBytes: 0,
		ReadTimeout:    0,
		WriteTimeout:   0,
		LogLevel:       "invalid",
		LogFormat:      "invalid",
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
		"invalid log-level",
		"invalid log-format",
	}

	for _, expected := range expectedErrors {
		if !strings.Contains(errStr, expected) {
			t.Errorf("validate() error missing %q", expected)
		}
	}
}

func TestFlagEnvName(t *testing.T) {
	tests := []struct {
		flag string
		want string
	}{
		{"transport", "MCP_HELM_TRANSPORT"},
		{"listen", "MCP_HELM_LISTEN"},
		{"helm-timeout", "MCP_HELM_HELM_TIMEOUT"},
		{"cache-size", "MCP_HELM_CACHE_SIZE"},
		{"index-ttl", "MCP_HELM_INDEX_TTL"},
		{"max-output-size", "MCP_HELM_MAX_OUTPUT_SIZE"},
		{"allow-private-ips", "MCP_HELM_ALLOW_PRIVATE_IPS"},
		{"allowed-hosts", "MCP_HELM_ALLOWED_HOSTS"},
		{"denied-hosts", "MCP_HELM_DENIED_HOSTS"},
		{"read-timeout", "MCP_HELM_READ_TIMEOUT"},
		{"write-timeout", "MCP_HELM_WRITE_TIMEOUT"},
		{"log-level", "MCP_HELM_LOG_LEVEL"},
		{"log-format", "MCP_HELM_LOG_FORMAT"},
	}

	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			got := flagEnvName(tt.flag)
			if got != tt.want {
				t.Errorf("flagEnvName(%q) = %q, want %q", tt.flag, got, tt.want)
			}
		})
	}
}

func TestLoad_EnvVarFallback(t *testing.T) {
	// Helper that creates a lookupEnv from a map
	envFrom := func(m map[string]string) func(string) (string, bool) {
		return func(key string) (string, bool) {
			v, ok := m[key]
			return v, ok
		}
	}
	noEnv := envFrom(map[string]string{})

	t.Run("defaults with no args and no env", func(t *testing.T) {
		cfg, err := load(nil, noEnv)
		if err != nil {
			t.Fatalf("load() error: %v", err)
		}
		if cfg.Transport != "stdio" {
			t.Errorf("Transport = %q, want %q", cfg.Transport, "stdio")
		}
		if cfg.CacheSize != 50 {
			t.Errorf("CacheSize = %d, want %d", cfg.CacheSize, 50)
		}
	})

	t.Run("env var sets transport", func(t *testing.T) {
		cfg, err := load(nil, envFrom(map[string]string{
			"MCP_HELM_TRANSPORT": "http",
		}))
		if err != nil {
			t.Fatalf("load() error: %v", err)
		}
		if cfg.Transport != "http" {
			t.Errorf("Transport = %q, want %q", cfg.Transport, "http")
		}
	})

	t.Run("CLI flag overrides env var", func(t *testing.T) {
		cfg, err := load(
			[]string{"--transport", "stdio"},
			envFrom(map[string]string{
				"MCP_HELM_TRANSPORT": "http",
			}),
		)
		if err != nil {
			t.Fatalf("load() error: %v", err)
		}
		if cfg.Transport != "stdio" {
			t.Errorf("Transport = %q, want %q (CLI should override env)", cfg.Transport, "stdio")
		}
	})

	t.Run("env var sets duration", func(t *testing.T) {
		cfg, err := load(nil, envFrom(map[string]string{
			"MCP_HELM_HELM_TIMEOUT": "60s",
		}))
		if err != nil {
			t.Fatalf("load() error: %v", err)
		}
		if cfg.HelmTimeout != 60*time.Second {
			t.Errorf("HelmTimeout = %v, want %v", cfg.HelmTimeout, 60*time.Second)
		}
	})

	t.Run("env var sets integer", func(t *testing.T) {
		cfg, err := load(nil, envFrom(map[string]string{
			"MCP_HELM_CACHE_SIZE": "100",
		}))
		if err != nil {
			t.Fatalf("load() error: %v", err)
		}
		if cfg.CacheSize != 100 {
			t.Errorf("CacheSize = %d, want %d", cfg.CacheSize, 100)
		}
	})

	t.Run("env var sets boolean", func(t *testing.T) {
		cfg, err := load(nil, envFrom(map[string]string{
			"MCP_HELM_ALLOW_PRIVATE_IPS": "true",
		}))
		if err != nil {
			t.Fatalf("load() error: %v", err)
		}
		if !cfg.AllowPrivateIPs {
			t.Error("AllowPrivateIPs = false, want true")
		}
	})

	t.Run("env var sets CSV hosts", func(t *testing.T) {
		cfg, err := load(nil, envFrom(map[string]string{
			"MCP_HELM_ALLOWED_HOSTS": "a.com,b.com",
		}))
		if err != nil {
			t.Fatalf("load() error: %v", err)
		}
		if !slicesEqual(cfg.AllowedHosts, []string{"a.com", "b.com"}) {
			t.Errorf("AllowedHosts = %v, want [a.com b.com]", cfg.AllowedHosts)
		}
	})

	t.Run("env var sets log level", func(t *testing.T) {
		cfg, err := load(nil, envFrom(map[string]string{
			"MCP_HELM_LOG_LEVEL": "debug",
		}))
		if err != nil {
			t.Fatalf("load() error: %v", err)
		}
		if cfg.LogLevel != "debug" {
			t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "debug")
		}
	})

	t.Run("empty env var ignored", func(t *testing.T) {
		cfg, err := load(nil, envFrom(map[string]string{
			"MCP_HELM_TRANSPORT": "",
		}))
		if err != nil {
			t.Fatalf("load() error: %v", err)
		}
		if cfg.Transport != "stdio" {
			t.Errorf("Transport = %q, want %q (empty env should be ignored)", cfg.Transport, "stdio")
		}
	})

	t.Run("multiple env vars", func(t *testing.T) {
		cfg, err := load(nil, envFrom(map[string]string{
			"MCP_HELM_TRANSPORT":  "http",
			"MCP_HELM_LOG_LEVEL":  "warn",
			"MCP_HELM_CACHE_SIZE": "200",
		}))
		if err != nil {
			t.Fatalf("load() error: %v", err)
		}
		if cfg.Transport != "http" {
			t.Errorf("Transport = %q, want %q", cfg.Transport, "http")
		}
		if cfg.LogLevel != "warn" {
			t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "warn")
		}
		if cfg.CacheSize != 200 {
			t.Errorf("CacheSize = %d, want %d", cfg.CacheSize, 200)
		}
	})
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
