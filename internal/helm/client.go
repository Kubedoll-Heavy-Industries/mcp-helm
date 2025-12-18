package helm

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.uber.org/zap"
	"helm.sh/helm/v4/pkg/chart/loader"
	chartv2 "helm.sh/helm/v4/pkg/chart/v2"
	"helm.sh/helm/v4/pkg/cli"
	"helm.sh/helm/v4/pkg/downloader"
	"helm.sh/helm/v4/pkg/getter"
	repo "helm.sh/helm/v4/pkg/repo/v1"
)

// Client implements ChartService for interacting with Helm repositories.
type Client struct {
	opts       *clientOptions
	settings   *cli.EnvSettings
	indexCache *IndexCache
	chartCache *ChartCache
	logger     *zap.Logger
}

// Ensure Client implements ChartService.
var _ ChartService = (*Client)(nil)

// NewClient creates a new Helm client with the given options.
func NewClient(opts ...Option) *Client {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(o.cacheDir, 0o755); err != nil {
		o.logger.Warn("failed to create cache directory", zap.Error(err))
	}

	settings := cli.New()
	settings.RepositoryCache = filepath.Join(o.cacheDir, "repository")
	settings.RegistryConfig = filepath.Join(o.cacheDir, "registry.json")
	settings.RepositoryConfig = filepath.Join(o.cacheDir, "repositories.yaml")

	return &Client{
		opts:       o,
		settings:   settings,
		indexCache: NewIndexCache(o.indexTTL),
		chartCache: NewChartCache(o.cacheSize),
		logger:     o.logger,
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

// LoadChart loads a chart from the repository.
func (c *Client) LoadChart(ctx context.Context, repoURL, chartName, version string) (*Chart, error) {
	hc, err := c.loadHelmChart(ctx, repoURL, chartName, version)
	if err != nil {
		return nil, err
	}
	return convertToChart(hc), nil
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

// GetContents returns a formatted string of all chart files.
func (c *Client) GetContents(ctx context.Context, repoURL, chartName, version string, recursive bool) (string, error) {
	hc, err := c.loadHelmChart(ctx, repoURL, chartName, version)
	if err != nil {
		return "", err
	}

	contents := formatContents(hc, recursive)
	if c.opts.maxOutputBytes > 0 && len(contents) > c.opts.maxOutputBytes {
		return "", &OutputTooLargeError{Size: len(contents), Limit: c.opts.maxOutputBytes}
	}

	return contents, nil
}

// GetDependencies returns the chart's dependencies.
func (c *Client) GetDependencies(ctx context.Context, repoURL, chartName, version string) ([]Dependency, error) {
	hc, err := c.loadHelmChart(ctx, repoURL, chartName, version)
	if err != nil {
		return nil, err
	}

	return extractDependencies(hc)
}

// ListFiles returns metadata about all files in a chart.
func (c *Client) ListFiles(ctx context.Context, repoURL, chartName, version string) ([]FileInfo, error) {
	hc, err := c.loadHelmChart(ctx, repoURL, chartName, version)
	if err != nil {
		return nil, err
	}

	var files []FileInfo

	for _, f := range hc.Raw {
		files = append(files, FileInfo{Path: f.Name, Size: len(f.Data)})
	}
	for _, f := range hc.Templates {
		files = append(files, FileInfo{Path: f.Name, Size: len(f.Data)})
	}
	for _, f := range hc.Files {
		files = append(files, FileInfo{Path: f.Name, Size: len(f.Data)})
	}

	return files, nil
}

// GetFile returns the contents of a specific file in a chart.
func (c *Client) GetFile(ctx context.Context, repoURL, chartName, version, path string) ([]byte, error) {
	hc, err := c.loadHelmChart(ctx, repoURL, chartName, version)
	if err != nil {
		return nil, err
	}

	data, found := findChartFile(hc, path)
	if !found {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	if c.opts.maxOutputBytes > 0 && len(data) > c.opts.maxOutputBytes {
		return nil, &OutputTooLargeError{Size: len(data), Limit: c.opts.maxOutputBytes}
	}

	return data, nil
}

// RefreshIndex forces a refresh of the repository index cache.
func (c *Client) RefreshIndex(ctx context.Context, repoURL string) error {
	_, err := c.getIndex(ctx, repoURL, true)
	return err
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
		Name: validatedURL,
		URL:  validatedURL,
	}, getter.All(c.settings, getter.WithTimeout(c.opts.timeout)))
	if err != nil {
		return nil, &RepositoryError{URL: validatedURL, Op: "create", Message: "failed to create repository", Err: err}
	}

	indexPath, err := chartRepo.DownloadIndexFile()
	if err != nil {
		return nil, &RepositoryError{URL: validatedURL, Op: "fetch", Message: "failed to download index", Err: err}
	}

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
	defer os.RemoveAll(tempDir)

	dl := downloader.ChartDownloader{
		Out:              io.Discard,
		Getters:          getter.All(c.settings),
		Options:          []getter.Option{getter.WithTimeout(c.opts.timeout)},
		RepositoryConfig: c.settings.RepositoryConfig,
		RepositoryCache:  c.settings.RepositoryCache,
		ContentCache:     c.settings.ContentCache,
		Verify:           downloader.VerifyNever,
	}

	chartPath, _, err := dl.DownloadTo(validatedChartURL, version, tempDir)
	if err != nil {
		return nil, &RepositoryError{URL: validatedURL, Op: "download", Message: "failed to download chart", Err: err}
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
