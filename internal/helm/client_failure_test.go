package helm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

// FailureSuite tests client behavior under adverse conditions.
// Uses httptest to simulate various failure modes.
type FailureSuite struct {
	suite.Suite
}

// testClient creates a client configured for httptest servers (allows private IPs)
func (s *FailureSuite) testClient(opts ...Option) ChartService {
	defaults := []Option{
		WithTimeout(5 * time.Second),
		WithAllowPrivateIPs(true), // Required for httptest servers on localhost
		WithLogger(zap.NewNop()),
	}
	return NewClient(append(defaults, opts...)...)
}

// =============================================================================
// Malformed Response Tests
// =============================================================================

func (s *FailureSuite) TestMalformedYAML_ReturnsError() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "index.yaml") {
			w.Header().Set("Content-Type", "application/x-yaml")
			_, _ = w.Write([]byte("this is not valid yaml: [[["))
		}
	}))
	defer server.Close()

	client := s.testClient()
	_, err := client.ListCharts(context.Background(), server.URL)

	s.Require().Error(err, "Malformed YAML should return error")
}

func (s *FailureSuite) TestEmptyIndex_ReturnsEmptyList() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-yaml")
		_, _ = w.Write([]byte("apiVersion: v1\nentries: {}\n"))
	}))
	defer server.Close()

	client := s.testClient()
	charts, err := client.ListCharts(context.Background(), server.URL)

	s.Require().NoError(err, "Empty index is valid")
	s.Empty(charts, "Should return empty list for empty index")
}

// =============================================================================
// HTTP Error Tests
// =============================================================================

func (s *FailureSuite) TestHTTP404_ReturnsRepositoryError() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := s.testClient()
	_, err := client.ListCharts(context.Background(), server.URL)

	s.Require().Error(err, "404 should return error")
	s.True(IsRepositoryError(err), "Should be RepositoryError, got: %T", err)
}

func (s *FailureSuite) TestHTTP500_ReturnsRepositoryError() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := s.testClient()
	_, err := client.ListCharts(context.Background(), server.URL)

	s.Require().Error(err, "500 should return error")
	s.True(IsRepositoryError(err), "Should be RepositoryError, got: %T", err)
}

func (s *FailureSuite) TestHTTP503_ReturnsRepositoryError() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := s.testClient()
	_, err := client.ListCharts(context.Background(), server.URL)

	s.Require().Error(err, "503 should return error")
}

// =============================================================================
// Timeout Tests
// =============================================================================

func (s *FailureSuite) TestSlowServer_TimesOut() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		_, _ = w.Write([]byte("too late"))
	}))
	defer server.Close()

	client := s.testClient(WithTimeout(100 * time.Millisecond))

	start := time.Now()
	_, err := client.ListCharts(context.Background(), server.URL)
	elapsed := time.Since(start)

	s.Require().Error(err, "Slow server should timeout")
	s.Less(elapsed, 2*time.Second, "Should timeout quickly, not wait for server")
}

// =============================================================================
// Connection Tests
// =============================================================================

func (s *FailureSuite) TestConnectionRefused_ReturnsError() {
	client := s.testClient()
	_, err := client.ListCharts(context.Background(), "http://127.0.0.1:1")

	s.Require().Error(err, "Connection refused should return error")
}

// =============================================================================
// Content-Type Tests
// =============================================================================

// =============================================================================
// Chart Size Validation Tests
// =============================================================================

func (s *FailureSuite) TestOversizedChart_ReturnsError() {
	// Serve a valid index that points to a chart, then serve a large file as the chart.
	// The client should reject the file before attempting to load/decompress it.
	largePayload := make([]byte, 1024) // 1 KB "chart" file
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "index.yaml") {
			w.Header().Set("Content-Type", "application/x-yaml")
			_, _ = w.Write([]byte(`apiVersion: v1
entries:
  testchart:
    - name: testchart
      version: "1.0.0"
      urls:
        - testchart-1.0.0.tgz
`))
			return
		}
		// Serve any other request as the "chart" file
		w.Header().Set("Content-Type", "application/gzip")
		_, _ = w.Write(largePayload)
	}))
	defer server.Close()

	// Set maxChartBytes to 100 bytes so our 1KB payload exceeds it
	client := s.testClient(WithMaxChartBytes(100))
	_, err := client.GetValues(context.Background(), server.URL, "testchart", "1.0.0")

	s.Require().Error(err, "Oversized chart should return error")
	s.True(IsChartTooLarge(err), "Should be ChartTooLargeError, got: %T: %v", err, err)
}

func (s *FailureSuite) TestChartWithinSizeLimit_Proceeds() {
	// A chart file under the limit should not be rejected by the size check.
	// It may fail later (e.g., invalid tgz), but should not fail the size check.
	smallPayload := make([]byte, 50)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "index.yaml") {
			w.Header().Set("Content-Type", "application/x-yaml")
			_, _ = w.Write([]byte(`apiVersion: v1
entries:
  testchart:
    - name: testchart
      version: "1.0.0"
      urls:
        - testchart-1.0.0.tgz
`))
			return
		}
		w.Header().Set("Content-Type", "application/gzip")
		_, _ = w.Write(smallPayload)
	}))
	defer server.Close()

	// Set a generous limit that our small payload fits within
	client := s.testClient(WithMaxChartBytes(1024 * 1024))
	_, err := client.GetValues(context.Background(), server.URL, "testchart", "1.0.0")

	// The error should NOT be a ChartTooLargeError (it may fail for other reasons
	// like invalid tgz format, which is fine).
	if err != nil {
		s.False(IsChartTooLarge(err), "Should not be ChartTooLargeError for small chart, got: %v", err)
	}
}

// =============================================================================
// Content-Type Tests
// =============================================================================

func (s *FailureSuite) TestHTMLError_ReturnsError() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body><h1>502 Bad Gateway</h1></body></html>"))
	}))
	defer server.Close()

	client := s.testClient()
	_, err := client.ListCharts(context.Background(), server.URL)

	s.Require().Error(err, "HTML response should return error")
}

// =============================================================================
// OCI Registry Tests
// =============================================================================

func (s *FailureSuite) TestOCI_ListCharts_ReturnsError() {
	client := s.testClient()
	_, err := client.ListCharts(context.Background(), "oci://ghcr.io/traefik/helm")

	s.Require().Error(err, "OCI ListCharts should return error")
	s.True(IsRepositoryError(err), "Should be RepositoryError, got: %T", err)
	s.Contains(err.Error(), "OCI registries do not support listing all charts")
}

func (s *FailureSuite) TestOCI_NilRegistryClient_ListVersions_ReturnsError() {
	// Create a client and nil out the registry client to simulate init failure
	c := NewClient(
		WithAllowPrivateIPs(true),
		WithLogger(zap.NewNop()),
	)
	c.registryClient = nil

	_, err := c.ListVersions(context.Background(), "oci://ghcr.io/traefik/helm", "traefik")

	s.Require().Error(err, "Nil registry client should return error")
	s.True(IsRepositoryError(err), "Should be RepositoryError, got: %T", err)
	s.Contains(err.Error(), "OCI registry client is not available")
}

func (s *FailureSuite) TestOCI_NilRegistryClient_GetValues_ReturnsError() {
	c := NewClient(
		WithAllowPrivateIPs(true),
		WithLogger(zap.NewNop()),
	)
	c.registryClient = nil

	_, err := c.GetValues(context.Background(), "oci://ghcr.io/traefik/helm", "traefik", "1.0.0")

	s.Require().Error(err, "Nil registry client should return error for GetValues")
	s.True(IsRepositoryError(err), "Should be RepositoryError, got: %T", err)
	s.Contains(err.Error(), "OCI registry client is not available")
}

func (s *FailureSuite) TestOCI_NilRegistryClient_GetLatestVersion_ReturnsError() {
	c := NewClient(
		WithAllowPrivateIPs(true),
		WithLogger(zap.NewNop()),
	)
	c.registryClient = nil

	_, err := c.GetLatestVersion(context.Background(), "oci://ghcr.io/traefik/helm", "traefik")

	s.Require().Error(err, "Nil registry client should return error for GetLatestVersion")
	s.True(IsRepositoryError(err), "Should be RepositoryError, got: %T", err)
	s.Contains(err.Error(), "OCI registry client is not available")
}

func (s *FailureSuite) TestOCI_NilRegistryClient_GetDependencies_ReturnsError() {
	c := NewClient(
		WithAllowPrivateIPs(true),
		WithLogger(zap.NewNop()),
	)
	c.registryClient = nil

	_, err := c.GetDependencies(context.Background(), "oci://ghcr.io/traefik/helm", "traefik", "1.0.0")

	s.Require().Error(err, "Nil registry client should return error for GetDependencies")
	s.True(IsRepositoryError(err), "Should be RepositoryError, got: %T", err)
	s.Contains(err.Error(), "OCI registry client is not available")
}

func (s *FailureSuite) TestOCI_NilRegistryClient_GetNotes_ReturnsError() {
	c := NewClient(
		WithAllowPrivateIPs(true),
		WithLogger(zap.NewNop()),
	)
	c.registryClient = nil

	_, _, err := c.GetNotes(context.Background(), "oci://ghcr.io/traefik/helm", "traefik", "1.0.0")

	s.Require().Error(err, "Nil registry client should return error for GetNotes")
	s.True(IsRepositoryError(err), "Should be RepositoryError, got: %T", err)
	s.Contains(err.Error(), "OCI registry client is not available")
}

func (s *FailureSuite) TestOCI_NilRegistryClient_GetValuesSchema_ReturnsError() {
	c := NewClient(
		WithAllowPrivateIPs(true),
		WithLogger(zap.NewNop()),
	)
	c.registryClient = nil

	_, _, err := c.GetValuesSchema(context.Background(), "oci://ghcr.io/traefik/helm", "traefik", "1.0.0")

	s.Require().Error(err, "Nil registry client should return error for GetValuesSchema")
	s.True(IsRepositoryError(err), "Should be RepositoryError, got: %T", err)
	s.Contains(err.Error(), "OCI registry client is not available")
}

// =============================================================================
// OCI Reference Construction Tests
// =============================================================================

func (s *FailureSuite) TestOCIRef_Construction() {
	tests := []struct {
		name      string
		url       string
		chartName string
		want      string
	}{
		{
			name:      "standard ref",
			url:       "oci://ghcr.io/traefik/helm",
			chartName: "traefik",
			want:      "ghcr.io/traefik/helm/traefik",
		},
		{
			name:      "with trailing slash",
			url:       "oci://ghcr.io/traefik/helm/",
			chartName: "traefik",
			want:      "ghcr.io/traefik/helm/traefik",
		},
		{
			name:      "single path segment",
			url:       "oci://registry.example.com/charts",
			chartName: "nginx",
			want:      "registry.example.com/charts/nginx",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := ociRef(tt.url, tt.chartName)
			s.Equal(tt.want, got)
		})
	}
}

func (s *FailureSuite) TestOCIRefVersioned_Construction() {
	got := ociRefVersioned("oci://ghcr.io/traefik/helm", "traefik", "1.2.3")
	s.Equal("ghcr.io/traefik/helm/traefik:1.2.3", got)
}

func TestFailureSuite(t *testing.T) {
	suite.Run(t, new(FailureSuite))
}
