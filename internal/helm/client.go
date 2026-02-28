package helm

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"go.uber.org/zap"
	"helm.sh/helm/v4/pkg/chart/loader"
	chartv2 "helm.sh/helm/v4/pkg/chart/v2"
	"helm.sh/helm/v4/pkg/cli"
	"helm.sh/helm/v4/pkg/downloader"
	"helm.sh/helm/v4/pkg/getter"
	"helm.sh/helm/v4/pkg/registry"
	repo "helm.sh/helm/v4/pkg/repo/v1"
)

// Client implements ChartService for interacting with Helm repositories.
type Client struct {
	opts           *clientOptions
	settings       *cli.EnvSettings
	indexCache     *IndexCache
	chartCache     *ChartCache
	registryClient *registry.Client
	logger         *zap.Logger
}

// Ensure Client implements ChartService.
var _ ChartService = (*Client)(nil)

// NewClient creates a new Helm client with the given options.
func NewClient(opts ...Option) *Client {
	o := defaultOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(o)
		}
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(o.cacheDir, 0o755); err != nil {
		o.logger.Warn("failed to create cache directory", zap.Error(err))
	}

	settings := cli.New()
	settings.RepositoryCache = filepath.Join(o.cacheDir, "repository")
	settings.RegistryConfig = filepath.Join(o.cacheDir, "registry.json")
	settings.RepositoryConfig = filepath.Join(o.cacheDir, "repositories.yaml")

	regClient, err := registry.NewClient(
		registry.ClientOptCredentialsFile(settings.RegistryConfig),
		registry.ClientOptEnableCache(true),
	)
	if err != nil {
		o.logger.Warn("failed to create OCI registry client; OCI operations will be unavailable", zap.Error(err))
	}

	return &Client{
		opts:           o,
		settings:       settings,
		indexCache:     NewIndexCache(o.indexCacheSize, o.indexTTL),
		chartCache:     NewChartCache(o.chartCacheSize),
		registryClient: regClient,
		logger:         o.logger,
	}
}

// validationOpts returns ValidationOptions from client configuration.
func (c *Client) validationOpts() ValidationOptions {
	return ValidationOptions{
		AllowPrivateIPs: c.opts.allowPrivateIPs,
		AllowedHosts:    c.opts.allowedHosts,
		DeniedHosts:     c.opts.deniedHosts,
	}
}

// ListCharts returns all chart names available in the repository.
func (c *Client) ListCharts(ctx context.Context, repoURL string) ([]string, error) {
	if registry.IsOCI(repoURL) {
		return nil, &RepositoryError{
			URL:     repoURL,
			Op:      "list",
			Message: "OCI registries do not support listing all charts; use get_values with a specific chart name",
		}
	}

	index, err := c.getIndex(ctx, repoURL, false)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	for _, versions := range index.Entries {
		for _, v := range versions {
			if v != nil && v.Name != "" && !seen[v.Name] {
				seen[v.Name] = true
				break
			}
		}
	}

	charts := make([]string, 0, len(seen))
	for name := range seen {
		charts = append(charts, name)
	}
	sort.Strings(charts)

	return charts, nil
}

// ListVersions returns all versions of a chart with metadata.
func (c *Client) ListVersions(ctx context.Context, repoURL, chart string) ([]ChartVersion, error) {
	if registry.IsOCI(repoURL) {
		return c.ociListVersions(ctx, repoURL, chart)
	}

	index, err := c.getIndex(ctx, repoURL, false)
	if err != nil {
		return nil, err
	}

	entries, ok := index.Entries[chart]
	if !ok {
		return nil, &ChartNotFoundError{Repository: repoURL, Chart: chart}
	}

	versions := make([]ChartVersion, 0, len(entries))
	for _, entry := range entries {
		if entry == nil || entry.Metadata == nil {
			continue
		}
		versions = append(versions, ChartVersion{
			Version:    entry.Version,
			AppVersion: entry.AppVersion,
			Created:    entry.Created,
			Deprecated: entry.Deprecated,
		})
	}

	return versions, nil
}

// GetLatestVersion returns the latest version string for a chart.
func (c *Client) GetLatestVersion(ctx context.Context, repoURL, chart string) (string, error) {
	if registry.IsOCI(repoURL) {
		versions, err := c.ociListVersions(ctx, repoURL, chart)
		if err != nil {
			return "", err
		}
		if len(versions) == 0 {
			return "", &ChartNotFoundError{Repository: repoURL, Chart: chart}
		}
		return versions[0].Version, nil
	}

	index, err := c.getIndex(ctx, repoURL, false)
	if err != nil {
		return "", err
	}

	entries, ok := index.Entries[chart]
	if !ok || len(entries) == 0 {
		return "", &ChartNotFoundError{Repository: repoURL, Chart: chart}
	}

	// Index entries are sorted by version (newest first)
	return entries[0].Version, nil
}

// GetValues returns the values.yaml contents for a chart.
func (c *Client) GetValues(ctx context.Context, repoURL, chartName, version string) ([]byte, error) {
	hc, err := c.loadHelmChart(ctx, repoURL, chartName, version)
	if err != nil {
		return nil, err
	}

	for _, file := range hc.Raw {
		if file.Name == "values.yaml" {
			if c.opts.maxOutputBytes > 0 && len(file.Data) > c.opts.maxOutputBytes {
				return nil, &OutputTooLargeError{Size: len(file.Data), Limit: c.opts.maxOutputBytes}
			}
			return file.Data, nil
		}
	}

	return nil, nil
}

// GetValuesSchema returns the values.schema.json contents if present.
func (c *Client) GetValuesSchema(ctx context.Context, repoURL, chartName, version string) ([]byte, bool, error) {
	hc, err := c.loadHelmChart(ctx, repoURL, chartName, version)
	if err != nil {
		return nil, false, err
	}

	for _, file := range hc.Raw {
		if file.Name == "values.schema.json" {
			if c.opts.maxOutputBytes > 0 && len(file.Data) > c.opts.maxOutputBytes {
				return nil, true, &OutputTooLargeError{Size: len(file.Data), Limit: c.opts.maxOutputBytes}
			}
			return file.Data, true, nil
		}
	}

	return nil, false, nil
}

// GetNotes returns the NOTES.txt contents if present.
func (c *Client) GetNotes(ctx context.Context, repoURL, chartName, version string) ([]byte, bool, error) {
	hc, err := c.loadHelmChart(ctx, repoURL, chartName, version)
	if err != nil {
		return nil, false, err
	}

	// NOTES.txt is in the templates directory
	for _, file := range hc.Templates {
		if file.Name == "templates/NOTES.txt" || file.Name == "NOTES.txt" {
			if c.opts.maxOutputBytes > 0 && len(file.Data) > c.opts.maxOutputBytes {
				return nil, true, &OutputTooLargeError{Size: len(file.Data), Limit: c.opts.maxOutputBytes}
			}
			return file.Data, true, nil
		}
	}

	return nil, false, nil
}

// GetDependencies returns the chart's dependencies.
func (c *Client) GetDependencies(ctx context.Context, repoURL, chartName, version string) ([]Dependency, error) {
	hc, err := c.loadHelmChart(ctx, repoURL, chartName, version)
	if err != nil {
		return nil, err
	}

	return extractDependencies(hc)
}

// getIndex retrieves the repository index, using cache if available.
func (c *Client) getIndex(ctx context.Context, repoURL string, forceRefresh bool) (*repo.IndexFile, error) {
	validatedURL, err := ValidateRepoURL(ctx, repoURL, c.validationOpts())
	if err != nil {
		return nil, err
	}

	// Acquire per-repo lock
	unlock := c.indexCache.LockRepo(validatedURL)
	defer unlock()

	// Check cache first
	if !forceRefresh {
		if index, ok := c.indexCache.Get(validatedURL); ok {
			return index, nil
		}
	}

	// Fetch index
	c.logger.Debug("fetching repository index", zap.String("url", validatedURL))

	chartRepo, err := repo.NewChartRepository(&repo.Entry{
		Name: sanitizeRepoName(validatedURL),
		URL:  validatedURL,
	}, getter.All(c.settings, getter.WithTimeout(c.opts.timeout)))
	if err != nil {
		return nil, &RepositoryError{URL: validatedURL, Op: "create", Message: "failed to create repository", Err: err}
	}
	chartRepo.CachePath = c.settings.RepositoryCache

	res := runWithContext(ctx, func() (string, error) {
		return chartRepo.DownloadIndexFile()
	})
	defer res.Wait()
	if res.Err != nil {
		return nil, &RepositoryError{URL: validatedURL, Op: "fetch", Message: "failed to download index", Err: res.Err}
	}
	indexPath := res.Val

	index, err := repo.LoadIndexFile(indexPath)
	if err != nil {
		return nil, &RepositoryError{URL: validatedURL, Op: "parse", Message: "failed to parse index", Err: err}
	}

	index.SortEntries()
	c.indexCache.Put(validatedURL, index)

	return index, nil
}

// loadHelmChart loads a chart, using cache if available.
func (c *Client) loadHelmChart(ctx context.Context, repoURL, chartName, version string) (*chartv2.Chart, error) {
	if registry.IsOCI(repoURL) {
		return c.ociLoadChart(ctx, repoURL, chartName, version)
	}

	validatedURL, err := ValidateRepoURL(ctx, repoURL, c.validationOpts())
	if err != nil {
		return nil, err
	}

	// Check cache first
	if chart, ok := c.chartCache.Get(validatedURL, chartName, version); ok {
		return chart, nil
	}

	// Get index to find chart URL
	index, err := c.getIndex(ctx, validatedURL, false)
	if err != nil {
		return nil, err
	}

	// Find chart version
	var chartVersion *repo.ChartVersion
	entries, ok := index.Entries[chartName]
	if !ok {
		return nil, &ChartNotFoundError{Repository: validatedURL, Chart: chartName, Version: version}
	}

	for _, entry := range entries {
		if entry.Version == version {
			chartVersion = entry
			break
		}
	}
	if chartVersion == nil {
		return nil, &ChartNotFoundError{Repository: validatedURL, Chart: chartName, Version: version}
	}

	if len(chartVersion.URLs) == 0 {
		return nil, &RepositoryError{URL: validatedURL, Op: "load", Message: "no download URLs for chart"}
	}

	// Resolve chart URL
	chartURL := chartVersion.URLs[0]
	if !strings.HasPrefix(chartURL, "http://") && !strings.HasPrefix(chartURL, "https://") {
		chartURL = strings.TrimSuffix(validatedURL, "/") + "/" + strings.TrimPrefix(chartURL, "/")
	}

	// Validate chart URL
	validatedChartURL, err := ValidateChartURL(ctx, chartURL, c.validationOpts())
	if err != nil {
		return nil, err
	}

	// Download chart
	c.logger.Debug("downloading chart",
		zap.String("chart", chartName),
		zap.String("version", version),
		zap.String("url", validatedChartURL),
	)

	tempDir, err := os.MkdirTemp("", "mcp-helm-chart-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	dl := downloader.ChartDownloader{
		Out:              io.Discard,
		Getters:          getter.All(c.settings),
		Options:          []getter.Option{getter.WithTimeout(c.opts.timeout)},
		RepositoryConfig: c.settings.RepositoryConfig,
		RepositoryCache:  c.settings.RepositoryCache,
		ContentCache:     c.settings.ContentCache,
		Verify:           downloader.VerifyNever,
	}

	res := runWithContext(ctx, func() (string, error) {
		path, _, err := dl.DownloadTo(validatedChartURL, version, tempDir)
		return path, err
	})
	// Wait for the goroutine to finish before removing tempDir,
	// even if context was cancelled, to avoid a directory race.
	defer func() {
		res.Wait()
		if err := os.RemoveAll(tempDir); err != nil {
			c.logger.Warn("failed to clean up temp directory",
				zap.String("path", tempDir),
				zap.Error(err))
		}
	}()
	if res.Err != nil {
		return nil, &RepositoryError{URL: validatedURL, Op: "download", Message: "failed to download chart", Err: res.Err}
	}
	chartPath := res.Val

	// Check chart file size before decompression
	if c.opts.maxChartBytes > 0 {
		fi, err := os.Stat(chartPath)
		if err != nil {
			return nil, fmt.Errorf("failed to stat downloaded chart: %w", err)
		}
		if fi.Size() > c.opts.maxChartBytes {
			return nil, &ChartTooLargeError{Size: fi.Size(), Limit: c.opts.maxChartBytes}
		}
	}

	// Load chart
	loaded, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	chart, ok := loaded.(*chartv2.Chart)
	if !ok {
		return nil, fmt.Errorf("unsupported chart format")
	}

	// Cache and return
	c.chartCache.Put(validatedURL, chartName, version, chart)

	return chart, nil
}

// ociRef builds an OCI reference from a validated OCI URL and chart name.
// e.g. oci://ghcr.io/traefik/helm + traefik → ghcr.io/traefik/helm/traefik
func ociRef(validatedURL, chartName string) string {
	base := strings.TrimPrefix(validatedURL, "oci://")
	base = strings.TrimSuffix(base, "/")
	return base + "/" + chartName
}

// ociRefVersioned builds a versioned OCI reference.
// e.g. oci://ghcr.io/traefik/helm + traefik + 1.0.0 → ghcr.io/traefik/helm/traefik:1.0.0
func ociRefVersioned(validatedURL, chartName, version string) string {
	return ociRef(validatedURL, chartName) + ":" + version
}

// ociListVersions lists chart versions from an OCI registry using Tags().
func (c *Client) ociListVersions(ctx context.Context, repoURL, chartName string) ([]ChartVersion, error) {
	if c.registryClient == nil {
		return nil, &RepositoryError{URL: repoURL, Op: "list_versions", Message: "OCI registry client is not available"}
	}

	validatedURL, err := ValidateOCIURL(ctx, repoURL, c.validationOpts())
	if err != nil {
		return nil, err
	}

	ref := ociRef(validatedURL, chartName)

	c.logger.Debug("listing OCI tags", zap.String("ref", ref))

	res := runWithContext(ctx, func() ([]string, error) {
		return c.registryClient.Tags(ref)
	})
	defer res.Wait()
	if res.Err != nil {
		return nil, &RepositoryError{URL: repoURL, Op: "list_versions", Message: "failed to list OCI tags", Err: res.Err}
	}

	tags := res.Val
	if len(tags) == 0 {
		return nil, &ChartNotFoundError{Repository: repoURL, Chart: chartName}
	}

	// Sort tags by semver (newest first); skip non-semver tags
	type parsed struct {
		ver *semver.Version
		raw string
	}
	var semverTags []parsed
	for _, tag := range tags {
		v, err := semver.NewVersion(tag)
		if err != nil {
			continue // skip non-semver tags
		}
		semverTags = append(semverTags, parsed{ver: v, raw: tag})
	}
	if len(semverTags) == 0 {
		return nil, &ChartNotFoundError{Repository: repoURL, Chart: chartName}
	}
	sort.Slice(semverTags, func(i, j int) bool {
		return semverTags[j].ver.LessThan(semverTags[i].ver) // descending
	})

	versions := make([]ChartVersion, 0, len(semverTags))
	for _, t := range semverTags {
		versions = append(versions, ChartVersion{Version: t.raw})
	}

	return versions, nil
}

// ociLoadChart downloads and loads a chart from an OCI registry.
func (c *Client) ociLoadChart(ctx context.Context, repoURL, chartName, version string) (*chartv2.Chart, error) {
	if c.registryClient == nil {
		return nil, &RepositoryError{URL: repoURL, Op: "load", Message: "OCI registry client is not available"}
	}

	validatedURL, err := ValidateOCIURL(ctx, repoURL, c.validationOpts())
	if err != nil {
		return nil, err
	}

	// Check cache first
	if chart, ok := c.chartCache.Get(validatedURL, chartName, version); ok {
		return chart, nil
	}

	ref := ociRefVersioned(validatedURL, chartName, version)

	c.logger.Debug("pulling OCI chart",
		zap.String("chart", chartName),
		zap.String("version", version),
		zap.String("ref", ref),
	)

	res := runWithContext(ctx, func() (*registry.PullResult, error) {
		return c.registryClient.Pull(ref, registry.PullOptWithChart(true))
	})
	defer res.Wait()
	if res.Err != nil {
		return nil, &RepositoryError{URL: repoURL, Op: "download", Message: "failed to pull OCI chart", Err: res.Err}
	}

	pullResult := res.Val
	if pullResult.Chart == nil || len(pullResult.Chart.Data) == 0 {
		return nil, &RepositoryError{URL: repoURL, Op: "download", Message: "OCI pull returned empty chart data"}
	}

	// Check chart size before decompression
	chartSize := int64(len(pullResult.Chart.Data))
	if c.opts.maxChartBytes > 0 && chartSize > c.opts.maxChartBytes {
		return nil, &ChartTooLargeError{Size: chartSize, Limit: c.opts.maxChartBytes}
	}

	// Write to temp file for loader.Load
	tempDir, err := os.MkdirTemp("", "mcp-helm-oci-chart-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			c.logger.Warn("failed to clean up temp directory",
				zap.String("path", tempDir),
				zap.Error(err))
		}
	}()

	chartPath := filepath.Join(tempDir, filepath.Base(chartName)+"-"+filepath.Base(version)+".tgz")
	if err := os.WriteFile(chartPath, pullResult.Chart.Data, 0o600); err != nil {
		return nil, fmt.Errorf("failed to write chart data: %w", err)
	}

	loaded, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	chart, ok := loaded.(*chartv2.Chart)
	if !ok {
		return nil, fmt.Errorf("unsupported chart format")
	}

	c.chartCache.Put(validatedURL, chartName, version, chart)

	return chart, nil
}

// sanitizeRepoName converts a URL to a valid filename for use as the Helm repo name.
// This is necessary because Helm uses the repo name to create cache filenames,
// and URLs contain characters (like colons) that are invalid in Windows paths.
func sanitizeRepoName(url string) string {
	// Replace characters that are invalid in Windows filenames
	replacer := strings.NewReplacer(
		"://", "_",
		":", "_",
		"/", "_",
		"\\", "_",
		"?", "_",
		"*", "_",
		"<", "_",
		">", "_",
		"|", "_",
		"\"", "_",
	)
	return replacer.Replace(url)
}
