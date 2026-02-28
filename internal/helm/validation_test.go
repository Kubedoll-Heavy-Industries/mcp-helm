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

func TestResolveHost_RespectsTimeout(t *testing.T) {
	// A cancelled context should cause resolveHost to return promptly with an error.
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := resolveHost(ctx, "example.com")
	if err == nil {
		t.Fatal("expected error from resolveHost with cancelled context, got nil")
	}
	if !strings.Contains(err.Error(), "canceled") && !strings.Contains(err.Error(), "deadline") {
		t.Errorf("expected context cancellation error, got: %v", err)
	}
}

func TestResolveHost_DNSTimeoutConstant(t *testing.T) {
	// Verify the DNS timeout constant is set to a reasonable value.
	if dnsTimeout != 5*time.Second {
		t.Errorf("dnsTimeout = %v, want 5s", dnsTimeout)
	}
}

func TestValidateRepoURL_DNSTimeoutPropagates(t *testing.T) {
	// When the parent context is already expired, DNS resolution should fail fast.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := ValidateRepoURL(ctx, "https://charts.bitnami.com/bitnami", ValidationOptions{})
	if err == nil {
		t.Fatal("expected error with cancelled context")
	}
	if !strings.Contains(err.Error(), "failed to resolve host") {
		t.Errorf("expected DNS resolution failure, got: %v", err)
	}
}

func TestValidateOCIURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		opts    ValidationOptions
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid OCI URL",
			url:     "oci://ghcr.io/traefik/helm",
			wantErr: false,
		},
		{
			name:    "valid OCI URL with trailing slash",
			url:     "oci://ghcr.io/traefik/helm/",
			wantErr: false,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
			errMsg:  "URL is empty",
		},
		{
			name:    "non-OCI scheme",
			url:     "https://ghcr.io/traefik/helm",
			wantErr: true,
			errMsg:  "OCI URL must use oci:// scheme",
		},
		{
			name:    "missing host",
			url:     "oci:///path/only",
			wantErr: true,
			errMsg:  "URL must include a host",
		},
		{
			name:    "URL with query",
			url:     "oci://ghcr.io/traefik/helm?foo=bar",
			wantErr: true,
			errMsg:  "URL must not contain a query",
		},
		{
			name:    "URL with fragment",
			url:     "oci://ghcr.io/traefik/helm#section",
			wantErr: true,
			errMsg:  "URL must not contain a fragment",
		},
		{
			name:    "URL with userinfo",
			url:     "oci://user:pass@ghcr.io/traefik/helm",
			wantErr: true,
			errMsg:  "URL must not contain userinfo",
		},
		{
			name:    "localhost blocked",
			url:     "oci://localhost/charts",
			wantErr: true,
			errMsg:  "localhost and .local hosts are not allowed",
		},
		{
			name:    ".local domain blocked",
			url:     "oci://myhost.local/charts",
			wantErr: true,
			errMsg:  "localhost and .local hosts are not allowed",
		},
		{
			name: "denied host",
			url:  "oci://ghcr.io/traefik/helm",
			opts: ValidationOptions{
				DeniedHosts: []string{"ghcr.io"},
			},
			wantErr: true,
			errMsg:  "host is in denylist",
		},
		{
			name: "allowlist enforced - not in list",
			url:  "oci://ghcr.io/traefik/helm",
			opts: ValidationOptions{
				AllowedHosts: []string{"registry.example.com"},
			},
			wantErr: true,
			errMsg:  "host is not in allowlist",
		},
		{
			name: "allowlist allows matching host",
			url:  "oci://ghcr.io/traefik/helm",
			opts: ValidationOptions{
				AllowedHosts: []string{"ghcr.io"},
			},
			wantErr: false,
		},
		{
			name:    "valid OCI URL with port",
			url:     "oci://ghcr.io:443/traefik/helm",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			_, err := ValidateOCIURL(ctx, tt.url, tt.opts)

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

func TestValidateOCIURL_NormalizesURL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := ValidateOCIURL(ctx, "oci://ghcr.io/traefik/helm/", ValidationOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "oci://ghcr.io/traefik/helm" {
		t.Errorf("expected trailing slash removed, got %q", result)
	}
}

func TestValidateOCIURL_PreservesPort(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := ValidateOCIURL(ctx, "oci://ghcr.io:443/traefik/helm", ValidationOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "oci://ghcr.io:443/traefik/helm" {
		t.Errorf("expected port preserved, got %q", result)
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
