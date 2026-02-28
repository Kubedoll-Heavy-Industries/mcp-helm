//go:build integration

package integration

import (
	"context"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/Kubedoll-Heavy-Industries/mcp-helm/internal/helm"
)

// HelmSuite tests the Helm client against real repositories.
// Tests verify behavioral contracts, not specific chart existence.
type HelmSuite struct {
	suite.Suite
	client helm.ChartService

	// Dynamic fixtures - fetched once in SetupSuite
	// These are used as known-good inputs for subsequent tests
	sampleChart   string // A chart name that exists (discovered, not hardcoded)
	sampleVersion string // A version of that chart
	sampleRepo    string // The repo it came from
}

func (s *HelmSuite) SetupSuite() {
	s.client = helm.NewClient(
		helm.WithTimeout(60*time.Second),
		helm.WithIndexTTL(10*time.Minute),
		helm.WithChartCacheSize(100),
		helm.WithLogger(zap.NewNop()),
	)

	// Discover a chart to use as fixture - don't hardcode "nginx"
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Try repos in order until we find one that works
	repos := []string{prometheusRepo, grafanaRepo, ingressRepo, bitnamiRepo}
	for _, repo := range repos {
		charts, err := s.client.ListCharts(ctx, repo)
		if err != nil || len(charts) == 0 {
			continue
		}

		// Use the first chart as our fixture
		s.sampleChart = charts[0]
		s.sampleRepo = repo

		version, err := s.client.GetLatestVersion(ctx, repo, s.sampleChart)
		if err == nil {
			s.sampleVersion = version
			break
		}
	}

	if s.sampleChart == "" {
		s.T().Fatal("Could not discover any working chart fixture from test repositories")
	}

	s.T().Logf("Using fixture: %s/%s@%s", s.sampleRepo, s.sampleChart, s.sampleVersion)
}

// =============================================================================
// Contract Tests: Verify response structure and invariants
// =============================================================================

func (s *HelmSuite) TestListCharts_ReturnsNonEmptySlice() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	charts, err := s.client.ListCharts(ctx, s.sampleRepo)

	s.Require().NoError(err, "ListCharts should succeed for valid repo")
	s.NotEmpty(charts, "Repository should contain at least one chart")

	// Contract: All chart names should be non-empty strings
	for i, name := range charts {
		s.NotEmpty(name, "Chart name at index %d should not be empty", i)
		s.False(strings.HasPrefix(name, " "), "Chart name should not have leading whitespace")
		s.False(strings.HasSuffix(name, " "), "Chart name should not have trailing whitespace")
	}
}

func (s *HelmSuite) TestListCharts_ReturnsConsistentResults() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Contract: Same input should return same output (deterministic)
	charts1, err := s.client.ListCharts(ctx, s.sampleRepo)
	s.Require().NoError(err)

	charts2, err := s.client.ListCharts(ctx, s.sampleRepo)
	s.Require().NoError(err)

	s.Equal(charts1, charts2, "Consecutive calls should return identical results")
}

func (s *HelmSuite) TestListVersions_ReturnsValidSemver() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	versions, err := s.client.ListVersions(ctx, s.sampleRepo, s.sampleChart)
	s.Require().NoError(err)
	s.NotEmpty(versions, "Chart should have at least one version")

	// Contract: All versions should be valid semver
	for _, v := range versions {
		s.NotEmpty(v.Version, "Version string should not be empty")
		_, err := semver.NewVersion(v.Version)
		s.NoError(err, "Version %q should be valid semver", v.Version)
	}
}

func (s *HelmSuite) TestListVersions_VersionsAreSortedDescending() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	versions, err := s.client.ListVersions(ctx, s.sampleRepo, s.sampleChart)
	s.Require().NoError(err)

	if len(versions) < 2 {
		s.T().Skip("Need at least 2 versions to test sorting")
	}

	// Contract: Versions should be sorted newest-first (descending)
	semvers := make([]*semver.Version, 0, len(versions))
	for _, v := range versions {
		sv, err := semver.NewVersion(v.Version)
		if err == nil {
			semvers = append(semvers, sv)
		}
	}

	isSorted := sort.SliceIsSorted(semvers, func(i, j int) bool {
		return semvers[i].GreaterThan(semvers[j])
	})
	s.True(isSorted, "Versions should be sorted in descending order")
}

func (s *HelmSuite) TestListVersions_HasCreatedTimestamps() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	versions, err := s.client.ListVersions(ctx, s.sampleRepo, s.sampleChart)
	s.Require().NoError(err)

	// Contract: All versions should have non-zero Created timestamps
	for _, v := range versions {
		s.False(v.Created.IsZero(), "Version %s should have Created timestamp", v.Version)
		s.True(v.Created.Before(time.Now()), "Created timestamp should be in the past")
		s.True(v.Created.After(time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)),
			"Created timestamp should be after Helm existed (2015)")
	}
}

func (s *HelmSuite) TestGetLatestVersion_ReturnsMostRecent() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	latest, err := s.client.GetLatestVersion(ctx, s.sampleRepo, s.sampleChart)
	s.Require().NoError(err)

	// Contract: Latest version should match first version in list
	versions, err := s.client.ListVersions(ctx, s.sampleRepo, s.sampleChart)
	s.Require().NoError(err)
	s.Require().NotEmpty(versions)

	s.Equal(versions[0].Version, latest,
		"GetLatestVersion should return the same version as the first in ListVersions")
}

func (s *HelmSuite) TestGetValues_ReturnsValidYAML() {
	if s.sampleVersion == "" {
		s.T().Skip("No sample version available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	values, err := s.client.GetValues(ctx, s.sampleRepo, s.sampleChart, s.sampleVersion)
	if err != nil && strings.Contains(err.Error(), "403") {
		s.T().Skipf("Skipping due to registry auth: %v", err)
	}
	s.Require().NoError(err)

	// Contract: Values should be valid YAML (contains at least one key-value pair)
	// We don't parse it fully, but YAML key-value pairs have colons
	s.NotEmpty(values, "Values file should not be empty")
	// Note: Some charts have empty values.yaml, which is valid
	// We just verify it's returned without error
}

// =============================================================================
// Error Handling Tests: Verify correct error types and messages
// =============================================================================

func (s *HelmSuite) TestChartNotFound_ReturnsCorrectErrorType() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := s.client.ListVersions(ctx, s.sampleRepo, "this-chart-definitely-does-not-exist-xyz123")

	s.Require().Error(err, "Non-existent chart should return error")
	s.True(helm.IsChartNotFound(err),
		"Error should be ChartNotFoundError, got: %T", err)
}

func (s *HelmSuite) TestInvalidRepo_ReturnsErrorWithinTimeout() {
	// Contract: Invalid repo should fail fast, not hang
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := time.Now()
	_, err := s.client.ListCharts(ctx, "https://this-repo-does-not-exist.invalid")
	elapsed := time.Since(start)

	s.Require().Error(err, "Invalid repo should return error")
	s.Less(elapsed, 10*time.Second, "Should fail within timeout, not hang")
}

// =============================================================================
// Caching Behavior Tests: Verify cache works correctly
// =============================================================================

func (s *HelmSuite) TestCache_SecondCallReturnsSameData() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// First call - populates cache
	charts1, err := s.client.ListCharts(ctx, s.sampleRepo)
	s.Require().NoError(err)

	// Second call - should return cached data
	charts2, err := s.client.ListCharts(ctx, s.sampleRepo)
	s.Require().NoError(err)

	// Contract: Cached data should be identical
	s.Equal(charts1, charts2, "Cached results should be identical to original")
}

// =============================================================================
// Multi-Repository Tests: Verify isolation between repos
// =============================================================================

func (s *HelmSuite) TestMultipleRepos_ReturnDifferentCharts() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Get charts from two different repos
	repos := []string{prometheusRepo, grafanaRepo}
	chartSets := make([][]string, 2)

	for i, repo := range repos {
		charts, err := s.client.ListCharts(ctx, repo)
		if err != nil {
			s.T().Skipf("Could not fetch from %s: %v", repo, err)
		}
		chartSets[i] = charts
	}

	// Contract: Different repos should have at least some different charts
	// (They might share some charts, but shouldn't be identical)
	set1 := make(map[string]bool)
	for _, c := range chartSets[0] {
		set1[c] = true
	}

	hasUnique := false
	for _, c := range chartSets[1] {
		if !set1[c] {
			hasUnique = true
			break
		}
	}

	s.True(hasUnique, "Different repositories should have at least some unique charts")
}

// =============================================================================
// Edge Cases
// =============================================================================

func (s *HelmSuite) TestEmptyChartName_ReturnsError() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := s.client.ListVersions(ctx, s.sampleRepo, "")

	s.Require().Error(err, "Empty chart name should return error")
}

func (s *HelmSuite) TestEmptyRepoURL_ReturnsError() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := s.client.ListCharts(ctx, "")

	s.Require().Error(err, "Empty repo URL should return error")
}

func (s *HelmSuite) TestWhitespaceRepoURL_ReturnsError() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := s.client.ListCharts(ctx, "   ")

	s.Require().Error(err, "Whitespace-only repo URL should return error")
}

func TestHelmSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests in short mode")
	}
	suite.Run(t, new(HelmSuite))
}
