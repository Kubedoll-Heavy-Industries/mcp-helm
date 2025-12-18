//go:build integration

package integration

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/Kubedoll-Heavy-Industries/mcp-helm/internal/helm"
)

const (
	// HTTP repositories
	bitnamiRepo      = "https://charts.bitnami.com/bitnami"                 // Popular vendor, many charts
	prometheusRepo   = "https://prometheus-community.github.io/helm-charts" // CNCF ecosystem
	grafanaRepo      = "https://grafana.github.io/helm-charts"              // Popular observability vendor
	ingressNginxRepo = "https://kubernetes.github.io/ingress-nginx"         // Kubernetes official

	// Test charts from different repos
	testChart       = "nginx"         // Simple Bitnami chart
	prometheusChart = "prometheus"    // Has dependencies (alertmanager, etc.)
	grafanaChart    = "grafana"       // Popular standalone chart
	ingressChart    = "ingress-nginx" // Kubernetes official chart
)

func newTestClient(t *testing.T) helm.ChartService {
	t.Helper()
	return helm.NewClient(
		helm.WithTimeout(60*time.Second),
		helm.WithIndexTTL(5*time.Minute),
		helm.WithCacheSize(10),
		helm.WithLogger(zap.NewNop()),
	)
}

func TestIntegration_ListCharts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client := newTestClient(t)
	charts, err := client.ListCharts(ctx, bitnamiRepo)

	require.NoError(t, err)
	assert.NotEmpty(t, charts)

	// Bitnami should have common charts
	assert.Contains(t, charts, "nginx")
	assert.Contains(t, charts, "redis")
	assert.Contains(t, charts, "postgresql")
}

func TestIntegration_ListVersions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client := newTestClient(t)
	versions, err := client.ListVersions(ctx, bitnamiRepo, testChart)

	require.NoError(t, err)
	assert.NotEmpty(t, versions)

	// Check version structure
	for _, v := range versions {
		assert.NotEmpty(t, v.Version)
		assert.False(t, v.Created.IsZero())
	}
}

func TestIntegration_GetLatestVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client := newTestClient(t)
	version, err := client.GetLatestVersion(ctx, bitnamiRepo, testChart)

	require.NoError(t, err)
	assert.NotEmpty(t, version)
	// Should be a semver-like string
	assert.Regexp(t, `^\d+\.\d+\.\d+`, version)
}

func TestIntegration_LoadChart(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client := newTestClient(t)

	// First get the latest version
	version, err := client.GetLatestVersion(ctx, bitnamiRepo, testChart)
	require.NoError(t, err)

	// Load the chart
	chart, err := client.LoadChart(ctx, bitnamiRepo, testChart, version)
	if err != nil {
		// Skip if we get a 403 (Bitnami OCI registry auth issue)
		if strings.Contains(err.Error(), "403") {
			t.Skipf("skipping due to external auth issue: %v", err)
		}
		require.NoError(t, err)
	}

	assert.Equal(t, testChart, chart.Name)
	assert.Equal(t, version, chart.Version)
	assert.NotEmpty(t, chart.Raw)
	assert.NotEmpty(t, chart.Templates)
}

func TestIntegration_GetValues(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client := newTestClient(t)

	// First get the latest version
	version, err := client.GetLatestVersion(ctx, bitnamiRepo, testChart)
	require.NoError(t, err)

	// Get values
	values, err := client.GetValues(ctx, bitnamiRepo, testChart, version)
	if err != nil {
		if strings.Contains(err.Error(), "403") {
			t.Skipf("skipping due to external auth issue: %v", err)
		}
		require.NoError(t, err)
	}

	assert.NotEmpty(t, values)
	// Should be valid YAML starting with comments or config
	assert.Contains(t, string(values), ":")
}

func TestIntegration_GetContents(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client := newTestClient(t)

	// First get the latest version
	version, err := client.GetLatestVersion(ctx, bitnamiRepo, testChart)
	require.NoError(t, err)

	// Get contents (non-recursive)
	contents, err := client.GetContents(ctx, bitnamiRepo, testChart, version, false)
	if err != nil {
		if strings.Contains(err.Error(), "403") {
			t.Skipf("skipping due to external auth issue: %v", err)
		}
		require.NoError(t, err)
	}

	assert.NotEmpty(t, contents)
	// Should contain Chart.yaml content
	assert.Contains(t, contents, "Chart.yaml")
}

func TestIntegration_GetDependencies(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client := newTestClient(t)

	// First get the latest version
	version, err := client.GetLatestVersion(ctx, bitnamiRepo, testChart)
	require.NoError(t, err)

	// Get dependencies - nginx may or may not have deps
	deps, err := client.GetDependencies(ctx, bitnamiRepo, testChart, version)
	if err != nil {
		if strings.Contains(err.Error(), "403") {
			t.Skipf("skipping due to external auth issue: %v", err)
		}
		require.NoError(t, err)
	}

	// Just verify the call succeeds - deps may be nil
	_ = deps
}

func TestIntegration_RefreshIndex(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client := newTestClient(t)
	err := client.RefreshIndex(ctx, bitnamiRepo)

	require.NoError(t, err)

	// After refresh, listing should still work
	charts, err := client.ListCharts(ctx, bitnamiRepo)
	require.NoError(t, err)
	assert.NotEmpty(t, charts)
}

func TestIntegration_ChartNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client := newTestClient(t)
	_, err := client.ListVersions(ctx, bitnamiRepo, "nonexistent-chart-xyz")

	require.Error(t, err)
	assert.True(t, helm.IsChartNotFound(err))
}

func TestIntegration_InvalidRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := newTestClient(t)
	_, err := client.ListCharts(ctx, "https://nonexistent-helm-repo.example.com")

	require.Error(t, err)
}

func TestIntegration_CacheHit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client := newTestClient(t)

	// First call - populates cache
	start1 := time.Now()
	charts1, err := client.ListCharts(ctx, bitnamiRepo)
	require.NoError(t, err)
	duration1 := time.Since(start1)

	// Second call - should hit cache and be faster
	start2 := time.Now()
	charts2, err := client.ListCharts(ctx, bitnamiRepo)
	require.NoError(t, err)
	duration2 := time.Since(start2)

	// Results should be identical
	assert.Equal(t, charts1, charts2)

	// Second call should be significantly faster (cache hit)
	// Being generous with the comparison due to network variability
	t.Logf("First call: %v, Second call (cached): %v", duration1, duration2)
	assert.Less(t, duration2, duration1, "cached call should be faster")
}

// Multi-repository tests - verify we can work with different repo types

func TestIntegration_MultiRepo_PrometheusCharts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client := newTestClient(t)

	// List charts from Prometheus community repo
	charts, err := client.ListCharts(ctx, prometheusRepo)
	require.NoError(t, err)
	assert.NotEmpty(t, charts)

	// Should have prometheus chart
	assert.Contains(t, charts, "prometheus")
	// Should also have kube-prometheus-stack (popular, has dependencies)
	assert.Contains(t, charts, "kube-prometheus-stack")
}

func TestIntegration_MultiRepo_GrafanaCharts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client := newTestClient(t)

	// List charts from Grafana repo
	charts, err := client.ListCharts(ctx, grafanaRepo)
	require.NoError(t, err)
	assert.NotEmpty(t, charts)

	// Should have grafana and loki
	assert.Contains(t, charts, "grafana")
	assert.Contains(t, charts, "loki")
}

func TestIntegration_MultiRepo_IngressNginx(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client := newTestClient(t)

	// List charts from Kubernetes official ingress-nginx repo
	charts, err := client.ListCharts(ctx, ingressNginxRepo)
	require.NoError(t, err)
	assert.NotEmpty(t, charts)

	// Should have ingress-nginx
	assert.Contains(t, charts, ingressChart)
}

func TestIntegration_ChartWithDependencies(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client := newTestClient(t)

	// kube-prometheus-stack has multiple dependencies including bundled ones
	version, err := client.GetLatestVersion(ctx, prometheusRepo, "kube-prometheus-stack")
	require.NoError(t, err)

	deps, err := client.GetDependencies(ctx, prometheusRepo, "kube-prometheus-stack", version)
	if err != nil {
		// Skip only for external access issues (403), not validation errors
		if strings.Contains(err.Error(), "403") {
			t.Skipf("skipping due to external auth issue: %v", err)
		}
		require.NoError(t, err)
	}

	// kube-prometheus-stack should have dependencies
	require.NotEmpty(t, deps, "kube-prometheus-stack should have dependencies")

	// Log what we found
	t.Logf("Found %d dependencies:", len(deps))
	for _, d := range deps {
		t.Logf("  - %s: %s (repo: %s)", d.Name, d.Version, d.Repository)
	}

	// Verify some known dependencies exist
	depNames := make(map[string]bool)
	for _, d := range deps {
		depNames[d.Name] = true
	}

	// Should have grafana as a dependency
	assert.True(t, depNames["grafana"], "should have grafana dependency")
}

func TestIntegration_VersionListingAcrossRepos(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	client := newTestClient(t)

	testCases := []struct {
		name  string
		repo  string
		chart string
	}{
		{"Bitnami nginx", bitnamiRepo, testChart},
		{"Prometheus community", prometheusRepo, prometheusChart},
		{"Grafana", grafanaRepo, grafanaChart},
		{"Ingress-nginx", ingressNginxRepo, ingressChart},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			versions, err := client.ListVersions(ctx, tc.repo, tc.chart)
			require.NoError(t, err)
			assert.NotEmpty(t, versions, "should have at least one version")

			// Verify version structure
			for _, v := range versions {
				assert.NotEmpty(t, v.Version, "version should not be empty")
			}

			// Get latest version
			latest, err := client.GetLatestVersion(ctx, tc.repo, tc.chart)
			require.NoError(t, err)
			assert.NotEmpty(t, latest)
			t.Logf("%s: found %d versions, latest is %s", tc.name, len(versions), latest)
		})
	}
}
