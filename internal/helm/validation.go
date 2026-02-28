package helm

import (
	"context"
	"net"
	"net/url"
	"strings"
	"time"
)

// MaxURLLength is the maximum allowed length for URLs to prevent resource exhaustion.
const MaxURLLength = 4096

// ValidationOptions configures URL validation behavior.
type ValidationOptions struct {
	AllowPrivateIPs bool
	AllowedHosts    []string
	DeniedHosts     []string
}

// urlValidationFlags controls optional behavior in the shared URL validation helper.
type urlValidationFlags struct {
	rejectQuery    bool
	rejectFragment bool
	normalizePath  bool
}

// validateHost performs host-level security checks shared by all URL validators.
// It checks localhost blocking, deny/allow lists, DNS resolution, and SSRF protections.
func validateHost(ctx context.Context, rawURL, host string, opts ValidationOptions) error {
	// Localhost and .local are never allowed
	hostLower := strings.ToLower(host)
	if hostLower == "localhost" || strings.HasSuffix(hostLower, ".local") {
		return &URLValidationError{URL: rawURL, Reason: "localhost and .local hosts are not allowed"}
	}

	// Check denylist
	if matchesHostList(hostLower, opts.DeniedHosts) {
		return &URLValidationError{URL: rawURL, Reason: "host is in denylist"}
	}

	// Check allowlist (if configured)
	if len(opts.AllowedHosts) > 0 && !matchesHostList(hostLower, opts.AllowedHosts) {
		return &URLValidationError{URL: rawURL, Reason: "host is not in allowlist"}
	}

	// DNS resolution check
	addrs, err := resolveHost(ctx, host)
	if err != nil {
		return &URLValidationError{URL: rawURL, Reason: "failed to resolve host: " + err.Error()}
	}
	if len(addrs) == 0 {
		return &URLValidationError{URL: rawURL, Reason: "host resolved to no addresses"}
	}

	// SSRF protection: check for private IPs
	if !opts.AllowPrivateIPs {
		for _, addr := range addrs {
			if isPrivateIP(addr) {
				return &URLValidationError{URL: rawURL, Reason: "host resolves to a private IP address"}
			}
		}
	}

	return nil
}

// validateURL is the shared implementation for HTTP/HTTPS URL validation.
// It checks the URL scheme, resolves the hostname, and applies SSRF protections.
func validateURL(ctx context.Context, rawURL string, opts ValidationOptions, flags urlValidationFlags) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", &URLValidationError{URL: rawURL, Reason: "URL is empty"}
	}
	if len(rawURL) > MaxURLLength {
		return "", &URLValidationError{URL: rawURL[:100] + "...", Reason: "URL exceeds maximum length"}
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", &URLValidationError{URL: rawURL, Reason: "invalid URL format"}
	}

	// Scheme validation
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", &URLValidationError{URL: rawURL, Reason: "scheme must be http or https"}
	}

	// No userinfo allowed
	if u.User != nil {
		return "", &URLValidationError{URL: rawURL, Reason: "URL must not contain userinfo"}
	}

	// Must have a host
	host := u.Hostname()
	if host == "" {
		return "", &URLValidationError{URL: rawURL, Reason: "URL must include a host"}
	}

	// Optional fragment/query rejection
	if flags.rejectFragment && u.Fragment != "" {
		return "", &URLValidationError{URL: rawURL, Reason: "URL must not contain a fragment"}
	}
	if flags.rejectQuery && u.RawQuery != "" {
		return "", &URLValidationError{URL: rawURL, Reason: "URL must not contain a query"}
	}

	if err := validateHost(ctx, rawURL, host, opts); err != nil {
		return "", err
	}

	// Optional path normalization
	if flags.normalizePath {
		u.Path = strings.TrimSuffix(u.Path, "/")
	}

	return u.String(), nil
}

// ValidateRepoURL validates a repository URL for security concerns.
// It checks the URL scheme, resolves the hostname, and applies SSRF protections.
// Fragments and query strings are rejected for repository URLs.
func ValidateRepoURL(ctx context.Context, rawURL string, opts ValidationOptions) (string, error) {
	return validateURL(ctx, rawURL, opts, urlValidationFlags{
		rejectQuery:    true,
		rejectFragment: true,
		normalizePath:  true,
	})
}

// ValidateChartURL validates a chart download URL.
// Similar to ValidateRepoURL but allows query strings and fragments (for signed URLs).
func ValidateChartURL(ctx context.Context, rawURL string, opts ValidationOptions) (string, error) {
	return validateURL(ctx, rawURL, opts, urlValidationFlags{})
}

// ValidateOCIURL validates an OCI registry URL (oci://host/path).
// It checks the URL scheme, resolves the hostname, and applies SSRF protections.
// Fragments and query strings are rejected.
func ValidateOCIURL(ctx context.Context, rawURL string, opts ValidationOptions) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", &URLValidationError{URL: rawURL, Reason: "URL is empty"}
	}
	if len(rawURL) > MaxURLLength {
		return "", &URLValidationError{URL: rawURL[:100] + "...", Reason: "URL exceeds maximum length"}
	}

	if !strings.HasPrefix(rawURL, "oci://") {
		return "", &URLValidationError{URL: rawURL, Reason: "OCI URL must use oci:// scheme"}
	}

	// Parse the remainder after oci:// as a URL with https scheme for validation
	httpsURL := "https://" + strings.TrimPrefix(rawURL, "oci://")
	u, err := url.Parse(httpsURL)
	if err != nil {
		return "", &URLValidationError{URL: rawURL, Reason: "invalid URL format"}
	}

	if u.User != nil {
		return "", &URLValidationError{URL: rawURL, Reason: "URL must not contain userinfo"}
	}

	host := u.Hostname()
	if host == "" {
		return "", &URLValidationError{URL: rawURL, Reason: "URL must include a host"}
	}

	if u.Fragment != "" {
		return "", &URLValidationError{URL: rawURL, Reason: "URL must not contain a fragment"}
	}
	if u.RawQuery != "" {
		return "", &URLValidationError{URL: rawURL, Reason: "URL must not contain a query"}
	}

	if err := validateHost(ctx, rawURL, host, opts); err != nil {
		return "", err
	}

	// Reconstruct normalized OCI URL (u.Host preserves port if present)
	path := strings.TrimSuffix(u.Path, "/")
	return "oci://" + u.Host + path, nil
}

// matchesHostList checks if a host matches any pattern in the list.
// Patterns can be:
//   - Exact match: "example.com"
//   - Wildcard subdomain: ".example.com" (matches example.com and *.example.com)
//   - Wildcard all: "*"
//
// Matching is case-insensitive.
func matchesHostList(host string, patterns []string) bool {
	host = strings.ToLower(host)
	if host == "" {
		return false
	}
	for _, pattern := range patterns {
		pattern = strings.ToLower(strings.TrimSpace(pattern))
		if pattern == "" {
			continue
		}
		if pattern == "*" {
			return true
		}
		if strings.HasPrefix(pattern, ".") {
			// Matches the domain itself and any subdomain
			suffix := pattern
			if host == strings.TrimPrefix(pattern, ".") || strings.HasSuffix(host, suffix) {
				return true
			}
			continue
		}
		// Exact match or subdomain match
		if host == pattern || strings.HasSuffix(host, "."+pattern) {
			return true
		}
	}
	return false
}

// dnsTimeout is the maximum time allowed for DNS resolution.
const dnsTimeout = 5 * time.Second

// resolveHost resolves a hostname to IP addresses with a dedicated DNS timeout.
// The timeout ensures an unresponsive DNS server cannot hang the request indefinitely.
func resolveHost(ctx context.Context, host string) ([]net.IP, error) {
	dnsCtx, cancel := context.WithTimeout(ctx, dnsTimeout)
	defer cancel()

	addrs, err := net.DefaultResolver.LookupIPAddr(dnsCtx, host)
	if err != nil {
		return nil, err
	}
	ips := make([]net.IP, len(addrs))
	for i, addr := range addrs {
		ips[i] = addr.IP
	}
	return ips, nil
}

// isPrivateIP returns true if the IP is private, loopback, link-local, or otherwise non-public.
func isPrivateIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() {
		return true
	}
	if ip.IsPrivate() {
		return true
	}
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	if ip.IsUnspecified() {
		return true
	}
	if ip.IsMulticast() {
		return true
	}
	return false
}
