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
			w.Write([]byte("this is not valid yaml: [[["))
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
		w.Write([]byte("apiVersion: v1\nentries: {}\n"))
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
		w.Write([]byte("too late"))
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

func (s *FailureSuite) TestHTMLError_ReturnsError() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><body><h1>502 Bad Gateway</h1></body></html>"))
	}))
	defer server.Close()

	client := s.testClient()
	_, err := client.ListCharts(context.Background(), server.URL)

	s.Require().Error(err, "HTML response should return error")
}

func TestFailureSuite(t *testing.T) {
	suite.Run(t, new(FailureSuite))
}
