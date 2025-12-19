//go:build integration

package integration

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/Kubedoll-Heavy-Industries/mcp-helm/internal/helm"
)

// HelmSuite tests the Helm client with a shared client instance.
// The client is created once in SetupSuite, warming the cache for all tests.
type HelmSuite struct {
	suite.Suite
	client helm.ChartService

	// Dynamic fixtures - fetched once in SetupSuite
	bitnamiNginxVersion string
	prometheusVersion   string
	grafanaVersion      string
	ingressVersion      string
}

func (s *HelmSuite) SetupSuite() {
	// Create shared client with generous cache settings for test run
	s.client = helm.NewClient(
		helm.WithTimeout(60*time.Second),
		helm.WithIndexTTL(10*time.Minute), // Longer TTL for test run
		helm.WithCacheSize(100),           // Room for multiple repos
		helm.WithLogger(zap.NewNop()),
	)

	// Warm cache by fetching latest versions (dynamic fixtures)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// These calls warm the index cache for each repo
	var err error
	s.bitnamiNginxVersion, err = s.client.GetLatestVersion(ctx, bitnamiRepo, "nginx")
	if err != nil {
		s.T().Logf("Warning: could not fetch bitnami nginx version: %v", err)
	}

	s.prometheusVersion, err = s.client.GetLatestVersion(ctx, prometheusRepo, "prometheus")
	if err != nil {
		s.T().Logf("Warning: could not fetch prometheus version: %v", err)
	}

	s.grafanaVersion, err = s.client.GetLatestVersion(ctx, grafanaRepo, "grafana")
	if err != nil {
		s.T().Logf("Warning: could not fetch grafana version: %v", err)
	}

	s.ingressVersion, err = s.client.GetLatestVersion(ctx, ingressRepo, "ingress-nginx")
	if err != nil {
		s.T().Logf("Warning: could not fetch ingress-nginx version: %v", err)
	}

	s.T().Logf("Suite setup complete. Dynamic fixtures: nginx=%s, prometheus=%s, grafana=%s, ingress=%s",
		s.bitnamiNginxVersion, s.prometheusVersion, s.grafanaVersion, s.ingressVersion)
}

// TestListCharts verifies chart listing from a repository
func (s *HelmSuite) TestListCharts() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	charts, err := s.client.ListCharts(ctx, bitnamiRepo)
	s.Require().NoError(err)
	s.NotEmpty(charts)

	// Bitnami should have common charts
	s.Contains(charts, "nginx")
	s.Contains(charts, "redis")
	s.Contains(charts, "postgresql")
}

// TestListVersions verifies version listing for a chart
func (s *HelmSuite) TestListVersions() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	versions, err := s.client.ListVersions(ctx, bitnamiRepo, "nginx")
	s.Require().NoError(err)
	s.NotEmpty(versions)

	// Check version structure
	for _, v := range versions {
		s.NotEmpty(v.Version)
		s.False(v.Created.IsZero())
	}
}

// TestGetLatestVersion verifies fetching the latest version
func (s *HelmSuite) TestGetLatestVersion() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	version, err := s.client.GetLatestVersion(ctx, bitnamiRepo, "nginx")
	s.Require().NoError(err)
	s.NotEmpty(version)
	s.Regexp(`^\d+\.\d+\.\d+`, version)
}

// TestLoadChart verifies chart loading using dynamic fixture
func (s *HelmSuite) TestLoadChart() {
	if s.bitnamiNginxVersion == "" {
		s.T().Skip("bitnami nginx version not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	chart, err := s.client.LoadChart(ctx, bitnamiRepo, "nginx", s.bitnamiNginxVersion)
	if err != nil && strings.Contains(err.Error(), "403") {
		s.T().Skipf("skipping due to external auth issue: %v", err)
	}
	s.Require().NoError(err)

	s.Equal("nginx", chart.Name)
	s.Equal(s.bitnamiNginxVersion, chart.Version)
	s.NotEmpty(chart.Raw)
	s.NotEmpty(chart.Templates)
}

// TestGetValues verifies values file retrieval
func (s *HelmSuite) TestGetValues() {
	if s.bitnamiNginxVersion == "" {
		s.T().Skip("bitnami nginx version not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	values, err := s.client.GetValues(ctx, bitnamiRepo, "nginx", s.bitnamiNginxVersion)
	if err != nil && strings.Contains(err.Error(), "403") {
		s.T().Skipf("skipping due to external auth issue: %v", err)
	}
	s.Require().NoError(err)

	s.NotEmpty(values)
	s.Contains(string(values), ":")
}

// TestGetContents verifies chart contents listing
func (s *HelmSuite) TestGetContents() {
	if s.bitnamiNginxVersion == "" {
		s.T().Skip("bitnami nginx version not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	contents, err := s.client.GetContents(ctx, bitnamiRepo, "nginx", s.bitnamiNginxVersion, false)
	if err != nil && strings.Contains(err.Error(), "403") {
		s.T().Skipf("skipping due to external auth issue: %v", err)
	}
	s.Require().NoError(err)

	s.NotEmpty(contents)
	s.Contains(contents, "Chart.yaml")
}

// TestGetDependencies verifies dependency listing
func (s *HelmSuite) TestGetDependencies() {
	if s.bitnamiNginxVersion == "" {
		s.T().Skip("bitnami nginx version not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	deps, err := s.client.GetDependencies(ctx, bitnamiRepo, "nginx", s.bitnamiNginxVersion)
	if err != nil && strings.Contains(err.Error(), "403") {
		s.T().Skipf("skipping due to external auth issue: %v", err)
	}
	s.Require().NoError(err)

	// nginx may or may not have dependencies - just verify call succeeds
	_ = deps
}

// TestRefreshIndex verifies index refresh
func (s *HelmSuite) TestRefreshIndex() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err := s.client.RefreshIndex(ctx, bitnamiRepo)
	s.Require().NoError(err)

	// After refresh, listing should still work
	charts, err := s.client.ListCharts(ctx, bitnamiRepo)
	s.Require().NoError(err)
	s.NotEmpty(charts)
}

// TestChartNotFound verifies error handling for missing chart
func (s *HelmSuite) TestChartNotFound() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := s.client.ListVersions(ctx, bitnamiRepo, "nonexistent-chart-xyz")
	s.Require().Error(err)
	s.True(helm.IsChartNotFound(err))
}

// TestInvalidRepo verifies error handling for invalid repository
func (s *HelmSuite) TestInvalidRepo() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err := s.client.ListCharts(ctx, "https://nonexistent-helm-repo.example.com")
	s.Require().Error(err)
}

// TestCacheHit verifies that cache improves performance
func (s *HelmSuite) TestCacheHit() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// First call - may hit existing cache from SetupSuite
	start1 := time.Now()
	charts1, err := s.client.ListCharts(ctx, bitnamiRepo)
	s.Require().NoError(err)
	duration1 := time.Since(start1)

	// Second call - definitely from cache
	start2 := time.Now()
	charts2, err := s.client.ListCharts(ctx, bitnamiRepo)
	s.Require().NoError(err)
	duration2 := time.Since(start2)

	s.Equal(charts1, charts2)
	s.T().Logf("First call: %v, Second call (cached): %v", duration1, duration2)
}

// TestMultiRepo tests charts from different repositories
func (s *HelmSuite) TestMultiRepo() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Test Prometheus repo
	s.Run("Prometheus", func() {
		charts, err := s.client.ListCharts(ctx, prometheusRepo)
		s.Require().NoError(err)
		s.Contains(charts, "prometheus")
		s.Contains(charts, "kube-prometheus-stack")
	})

	// Test Grafana repo
	s.Run("Grafana", func() {
		charts, err := s.client.ListCharts(ctx, grafanaRepo)
		s.Require().NoError(err)
		s.Contains(charts, "grafana")
		s.Contains(charts, "loki")
	})

	// Test Ingress-nginx repo
	s.Run("IngressNginx", func() {
		charts, err := s.client.ListCharts(ctx, ingressRepo)
		s.Require().NoError(err)
		s.Contains(charts, "ingress-nginx")
	})
}

// TestChartWithDependencies tests a chart known to have dependencies
func (s *HelmSuite) TestChartWithDependencies() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	version, err := s.client.GetLatestVersion(ctx, prometheusRepo, "kube-prometheus-stack")
	s.Require().NoError(err)

	deps, err := s.client.GetDependencies(ctx, prometheusRepo, "kube-prometheus-stack", version)
	if err != nil && strings.Contains(err.Error(), "403") {
		s.T().Skipf("skipping due to external auth issue: %v", err)
	}
	s.Require().NoError(err)

	// kube-prometheus-stack should have dependencies
	s.Require().NotEmpty(deps, "kube-prometheus-stack should have dependencies")

	s.T().Logf("Found %d dependencies:", len(deps))
	for _, d := range deps {
		s.T().Logf("  - %s: %s (repo: %s)", d.Name, d.Version, d.Repository)
	}

	// Should have grafana as a dependency
	depNames := make(map[string]bool)
	for _, d := range deps {
		depNames[d.Name] = true
	}
	s.True(depNames["grafana"], "should have grafana dependency")
}

// TestVersionListingAcrossRepos tests version listing across multiple repos
func (s *HelmSuite) TestVersionListingAcrossRepos() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	testCases := []struct {
		name  string
		repo  string
		chart string
	}{
		{"Bitnami nginx", bitnamiRepo, "nginx"},
		{"Prometheus", prometheusRepo, "prometheus"},
		{"Grafana", grafanaRepo, "grafana"},
		{"Ingress-nginx", ingressRepo, "ingress-nginx"},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			versions, err := s.client.ListVersions(ctx, tc.repo, tc.chart)
			s.Require().NoError(err)
			s.NotEmpty(versions, "should have at least one version")

			for _, v := range versions {
				s.NotEmpty(v.Version, "version should not be empty")
			}

			latest, err := s.client.GetLatestVersion(ctx, tc.repo, tc.chart)
			s.Require().NoError(err)
			s.NotEmpty(latest)
			s.T().Logf("%s: found %d versions, latest is %s", tc.name, len(versions), latest)
		})
	}
}

func TestHelmSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests in short mode")
	}
	suite.Run(t, new(HelmSuite))
}
