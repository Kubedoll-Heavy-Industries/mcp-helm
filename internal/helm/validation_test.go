package helm

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"
)

func TestValidateRepoURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		opts    ValidationOptions
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid https URL",
			url:     "https://charts.bitnami.com/bitnami",
			wantErr: false,
		},
		{
			name:    "valid http URL",
			url:     "http://charts.helm.sh/stable",
			wantErr: false,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
			errMsg:  "URL is empty",
		},
		{
			name:    "whitespace URL",
			url:     "   ",
			wantErr: true,
			errMsg:  "URL is empty",
		},
		{
			name:    "invalid scheme",
			url:     "ftp://charts.example.com",
			wantErr: true,
			errMsg:  "scheme must be http or https",
		},
		{
			name:    "file scheme",
			url:     "file:///etc/passwd",
			wantErr: true,
			errMsg:  "scheme must be http or https",
		},
		{
			name:    "URL with userinfo",
			url:     "https://user:pass@charts.example.com",
			wantErr: true,
			errMsg:  "URL must not contain userinfo",
		},
		{
			name:    "URL without host",
			url:     "https:///path",
			wantErr: true,
			errMsg:  "URL must include a host",
		},
		{
			name:    "URL with fragment",
			url:     "https://charts.bitnami.com#section",
			wantErr: true,
			errMsg:  "URL must not contain a fragment",
		},
		{
			name:    "URL with query",
			url:     "https://charts.bitnami.com?foo=bar",
			wantErr: true,
			errMsg:  "URL must not contain a query",
		},
		{
			name:    "localhost blocked",
			url:     "https://localhost/charts",
			wantErr: true,
			errMsg:  "localhost and .local hosts are not allowed",
		},
		{
			name:    "LOCALHOST case insensitive",
			url:     "https://LOCALHOST/charts",
			wantErr: true,
			errMsg:  "localhost and .local hosts are not allowed",
		},
		{
			name:    ".local domain blocked",
			url:     "https://myhost.local/charts",
			wantErr: true,
			errMsg:  "localhost and .local hosts are not allowed",
		},
		{
			name: "denied host",
			url:  "https://charts.bitnami.com/bitnami",
			opts: ValidationOptions{
				DeniedHosts: []string{"charts.bitnami.com"},
			},
			wantErr: true,
			errMsg:  "host is in denylist",
		},
		{
			name: "denied host subdomain",
			url:  "https://sub.blocked.com/path",
			opts: ValidationOptions{
				DeniedHosts: []string{"blocked.com"},
			},
			wantErr: true,
			errMsg:  "host is in denylist",
		},
		{
			name: "allowlist enforced - not in list",
			url:  "https://charts.bitnami.com/bitnami",
			opts: ValidationOptions{
				AllowedHosts: []string{"trusted.com"},
			},
			wantErr: true,
			errMsg:  "host is not in allowlist",
		},
		{
			name: "allowlist allows matching host",
			url:  "https://charts.bitnami.com/bitnami",
			opts: ValidationOptions{
				AllowedHosts: []string{"charts.bitnami.com"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			_, err := ValidateRepoURL(ctx, tt.url, tt.opts)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestMatchesHostList(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		patterns []string
		want     bool
	}{
		{"empty patterns", "example.com", nil, false},
		{"exact match", "example.com", []string{"example.com"}, true},
		{"no match", "other.com", []string{"different.com"}, false},
		{"case insensitive", "EXAMPLE.COM", []string{"example.com"}, true},
		{"wildcard all", "anything.com", []string{"*"}, true},
		{"subdomain match", "sub.example.com", []string{"example.com"}, true},
		{"deep subdomain", "a.b.c.example.com", []string{"example.com"}, true},
		{"prefix wildcard", "sub.example.com", []string{".example.com"}, true},
		{"prefix wildcard exact", "example.com", []string{".example.com"}, true},
		{"no partial match", "notexample.com", []string{"example.com"}, false},
		{"empty host", "", []string{"example.com"}, false},
		{"whitespace pattern ignored", "example.com", []string{"  ", "example.com"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesHostList(tt.host, tt.patterns)
			if got != tt.want {
				t.Errorf("matchesHostList(%q, %v) = %v, want %v", tt.host, tt.patterns, got, tt.want)
			}
		})
	}
}

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want bool
	}{
		{"nil", "", true},
		{"loopback v4", "127.0.0.1", true},
		{"loopback v6", "::1", true},
		{"private 10.x", "10.0.0.1", true},
		{"private 172.16.x", "172.16.0.1", true},
		{"private 192.168.x", "192.168.1.1", true},
		{"link local v4", "169.254.1.1", true},
		{"link local v6", "fe80::1", true},
		{"multicast v4", "224.0.0.1", true},
		{"multicast v6", "ff02::1", true},
		{"unspecified v4", "0.0.0.0", true},
		{"unspecified v6", "::", true},
		{"public v4", "8.8.8.8", false},
		{"public v6", "2001:4860:4860::8888", false},
		{"cloudflare dns", "1.1.1.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ip net.IP
			if tt.ip != "" {
				ip = net.ParseIP(tt.ip)
			}
			got := isPrivateIP(ip)
			if got != tt.want {
				t.Errorf("isPrivateIP(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}
