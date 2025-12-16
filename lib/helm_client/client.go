package helm_client

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"helm.sh/helm/v4/pkg/chart/loader"
	chartv2 "helm.sh/helm/v4/pkg/chart/v2"
	"helm.sh/helm/v4/pkg/cli"
	"helm.sh/helm/v4/pkg/downloader"
	"helm.sh/helm/v4/pkg/getter"
	"helm.sh/helm/v4/pkg/repo/v1"

	"github.com/Kubedoll-Heavy-Industries/mcp-helm/lib/helm_parser"
)

var (
	tmpDir = "/tmp/helm_cache"
)

type HelmClient struct {
	settings *cli.EnvSettings

	timeout            time.Duration
	allowPrivateIPs    bool
	allowedHosts       []string
	deniedHosts        []string
	maxToolOutputBytes int

	reposMu sync.Mutex
	repos   map[string]*repo.ChartRepository
}

type Options struct {
	Timeout            time.Duration
	AllowPrivateIPs    bool
	AllowedHosts       []string
	DeniedHosts        []string
	MaxToolOutputBytes int
}

type ChartVersionMetadata struct {
	Version    string
	AppVersion string
	Created    string
	Deprecated bool
}

func NewClient() *HelmClient {
	return NewClientWithOptions(Options{
		Timeout:            30 * time.Second,
		AllowPrivateIPs:    false,
		AllowedHosts:       nil,
		DeniedHosts:        nil,
		MaxToolOutputBytes: 2 * 1024 * 1024,
	})
}

func NewClientWithOptions(opts Options) *HelmClient {
	settings := cli.New()
	settings.RepositoryCache = path.Join(tmpDir, "helm-cache")
	settings.RegistryConfig = path.Join(tmpDir, "helm-registry.conf")
	settings.RepositoryConfig = path.Join(tmpDir, "helm-repository.conf")

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &HelmClient{
		settings: settings,
		timeout:  timeout,

		allowPrivateIPs:    opts.AllowPrivateIPs,
		allowedHosts:       normalizeHosts(opts.AllowedHosts),
		deniedHosts:        normalizeHosts(opts.DeniedHosts),
		maxToolOutputBytes: opts.MaxToolOutputBytes,
	}
}

func normalizeHosts(in []string) []string {
	if len(in) == 0 {
		return nil
	}

	out := make([]string, 0, len(in))
	for _, v := range in {
		vv := strings.ToLower(strings.TrimSpace(v))
		if vv == "" {
			continue
		}
		out = append(out, vv)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func hostMatchesList(host string, patterns []string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return false
	}

	for _, p := range patterns {
		pp := strings.ToLower(strings.TrimSpace(p))
		if pp == "" {
			continue
		}
		if pp == "*" {
			return true
		}
		if strings.HasPrefix(pp, ".") {
			if host == strings.TrimPrefix(pp, ".") || strings.HasSuffix(host, pp) {
				return true
			}
			continue
		}
		if host == pp || strings.HasSuffix(host, "."+pp) {
			return true
		}
	}

	return false
}

func isDisallowedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
		return true
	}
	if ip.IsMulticast() {
		return true
	}
	return false
}

func (c *HelmClient) validateRepositoryURL(raw string) (string, error) {
	repoURL := strings.TrimSpace(raw)
	if repoURL == "" {
		return "", fmt.Errorf("repository URL is empty")
	}

	u, err := url.Parse(repoURL)
	if err != nil {
		return "", fmt.Errorf("invalid repository URL: %v", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("unsupported repository URL scheme: %s", u.Scheme)
	}
	if u.User != nil {
		return "", fmt.Errorf("repository URL must not contain userinfo")
	}
	if u.Hostname() == "" {
		return "", fmt.Errorf("repository URL must include a host")
	}
	if u.Fragment != "" {
		return "", fmt.Errorf("repository URL must not contain a fragment")
	}
	if u.RawQuery != "" {
		return "", fmt.Errorf("repository URL must not contain a query")
	}

	host := strings.ToLower(u.Hostname())
	if host == "localhost" || strings.HasSuffix(host, ".local") {
		return "", fmt.Errorf("repository URL host is not allowed")
	}
	if hostMatchesList(host, c.deniedHosts) {
		return "", fmt.Errorf("repository URL host is denied")
	}
	if len(c.allowedHosts) > 0 && !hostMatchesList(host, c.allowedHosts) {
		return "", fmt.Errorf("repository URL host is not in allowlist")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return "", fmt.Errorf("failed to resolve repository host: %v", err)
	}
	if len(addrs) == 0 {
		return "", fmt.Errorf("repository host resolved to no addresses")
	}
	if !c.allowPrivateIPs {
		for _, a := range addrs {
			if isDisallowedIP(a.IP) {
				return "", fmt.Errorf("repository host resolves to a non-public IP")
			}
		}
	}

	u.Path = strings.TrimSuffix(u.Path, "/")
	return u.String(), nil
}

func (c *HelmClient) validateFetchURL(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", fmt.Errorf("invalid URL: %v", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("unsupported URL scheme: %s", u.Scheme)
	}
	if u.User != nil {
		return "", fmt.Errorf("URL must not contain userinfo")
	}
	if u.Hostname() == "" {
		return "", fmt.Errorf("URL must include a host")
	}

	host := strings.ToLower(u.Hostname())
	if host == "localhost" || strings.HasSuffix(host, ".local") {
		return "", fmt.Errorf("URL host is not allowed")
	}
	if hostMatchesList(host, c.deniedHosts) {
		return "", fmt.Errorf("URL host is denied")
	}
	if len(c.allowedHosts) > 0 && !hostMatchesList(host, c.allowedHosts) {
		return "", fmt.Errorf("URL host is not in allowlist")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return "", fmt.Errorf("failed to resolve host: %v", err)
	}
	if len(addrs) == 0 {
		return "", fmt.Errorf("host resolved to no addresses")
	}
	if !c.allowPrivateIPs {
		for _, a := range addrs {
			if isDisallowedIP(a.IP) {
				return "", fmt.Errorf("host resolves to a non-public IP")
			}
		}
	}

	return u.String(), nil
}

func (c *HelmClient) getRepo(repoURL string) (*repo.ChartRepository, error) {
	c.reposMu.Lock()
	defer c.reposMu.Unlock()

	if c.repos == nil {
		c.repos = make(map[string]*repo.ChartRepository)
	}

	validatedRepoURL, err := c.validateRepositoryURL(repoURL)
	if err != nil {
		return nil, err
	}

	// todo: refresh index periodically based on last update time or a fixed interval
	if v, exists := c.repos[validatedRepoURL]; exists {
		return v, nil
	}

	requestedRepo, err := repo.NewChartRepository(&repo.Entry{
		Name: validatedRepoURL,
		URL:  validatedRepoURL,
	}, getter.All(c.settings, getter.WithTimeout(c.timeout)))
	if err != nil {
		return nil, fmt.Errorf("failed to create chartv2 repository: %v", err)
	}

	indexFileLocation, err := requestedRepo.DownloadIndexFile()
	if err != nil {
		return nil, fmt.Errorf("failed to download repository index: %v", err)
	}

	file, err := repo.LoadIndexFile(indexFileLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to load index file: %v", err)
	}
	requestedRepo.IndexFile = file
	requestedRepo.IndexFile.SortEntries()

	c.repos[validatedRepoURL] = requestedRepo
	return requestedRepo, nil
}

func (c *HelmClient) ListCharts(repoURL string) ([]string, error) {
	// todo: sanitize repoURL url to create a name

	helmRepo, err := c.getRepo(repoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to add repository: %v", err)
	}

	charts := make(map[string]bool)
	for _, entry := range helmRepo.IndexFile.Entries {
		for _, version := range entry {
			if !charts[version.Name] {
				charts[version.Name] = true
				break
			}
		}
	}

	chartsList := make([]string, 0, len(charts))
	for chart := range charts {
		chartsList = append(chartsList, chart)
	}
	sort.Strings(chartsList)

	return chartsList, nil
}

func (c *HelmClient) ListChartVersions(repoURL string, chart string) ([]string, error) {
	helmRepo, err := c.getRepo(repoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to add repository: %v", err)
	}

	versions := make([]string, 0)
	for k, v := range helmRepo.IndexFile.Entries {
		if k != chart {
			continue
		}

		for _, ver := range v {
			versions = append(versions, ver.Version)
		}
	}
	// Do not sort version as those were sorted in original index file

	return versions, nil
}

func (c *HelmClient) ListChartVersionsWithMetadata(repoURL string, chart string) ([]ChartVersionMetadata, error) {
	helmRepo, err := c.getRepo(repoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to add repository: %v", err)
	}

	chartVersions, ok := helmRepo.IndexFile.Entries[chart]
	if !ok {
		return nil, fmt.Errorf("chart %s not found in repository %s", chart, repoURL)
	}

	versions := make([]ChartVersionMetadata, 0, len(chartVersions))
	for _, ver := range chartVersions {
		if ver == nil || ver.Metadata == nil {
			continue
		}

		created := ""
		if !ver.Created.IsZero() {
			created = ver.Created.UTC().Format("2006-01-02T15:04:05Z")
		}

		versions = append(versions, ChartVersionMetadata{
			Version:    ver.Version,
			AppVersion: ver.AppVersion,
			Created:    created,
			Deprecated: ver.Deprecated,
		})
	}

	return versions, nil
}

func (c *HelmClient) GetChartValues(repoURL, chartName, version string) (string, error) {
	loadedChart, err := c.loadChart(repoURL, chartName, version)
	if err != nil {
		return "", fmt.Errorf("failed to load chartv2 %s version %s: %v", chartName, version, err)
	}

	var rawContent []byte
	for _, file := range loadedChart.Raw {
		if file.Name != "values.yaml" {
			continue
		}
		rawContent = file.Data
		break
	}

	values := string(rawContent)
	if c.maxToolOutputBytes > 0 && len(values) > c.maxToolOutputBytes {
		return "", fmt.Errorf("chart values output exceeds max size limit")
	}
	return values, nil
}

func (c *HelmClient) GetChartContents(repoURL, chartName, version string, recursive bool) (string, error) {
	loadedChart, err := c.loadChart(repoURL, chartName, version)
	if err != nil {
		return "", fmt.Errorf("failed to load chartv2 %s version %s: %v", chartName, version, err)
	}

	if loadedChart == nil {
		return "", fmt.Errorf("chartv2 %s version %s not found", chartName, version)
	}

	contents, err := helm_parser.GetChartContents(loadedChart, recursive)
	if err != nil {
		return "", fmt.Errorf("failed to get chartv2 contents for %s version %s: %v", chartName, version, err)
	}
	if c.maxToolOutputBytes > 0 && len(contents) > c.maxToolOutputBytes {
		return "", fmt.Errorf("chart contents output exceeds max size limit")
	}
	return contents, nil
}

func (c *HelmClient) loadChart(repoURL string, chartName string, version string) (*chartv2.Chart, error) {
	// TODO: implement caching for values file
	helmRepo, err := c.getRepo(repoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %v", err)
	}

	var cv *repo.ChartVersion
	for k, v := range helmRepo.IndexFile.Entries {
		if k != chartName {
			continue
		}
		for _, ver := range v {
			if ver.Version != version {
				continue
			}
			cv = ver
			break
		}
		if cv != nil {
			break
		}
	}
	if cv == nil {
		return nil, fmt.Errorf("failed to find chartv2 %s version %s", chartName, version)
	}

	if len(cv.URLs) == 0 {
		return nil, fmt.Errorf("no download URLs found for chartv2 %s version %s", chartName, version)
	}

	chartURL := cv.URLs[0]
	if !strings.HasPrefix(chartURL, "http://") && !strings.HasPrefix(chartURL, "https://") {
		repoBaseURL := strings.TrimSuffix(helmRepo.Config.URL, "/")
		chartURL = fmt.Sprintf("%s/%s", repoBaseURL, strings.TrimPrefix(chartURL, "/"))
	}

	validatedChartURL, err := c.validateFetchURL(chartURL)
	if err != nil {
		return nil, fmt.Errorf("chart download URL is not allowed: %v", err)
	}

	tempDir, err := os.MkdirTemp("", "helm-chartv2-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	chartPath := filepath.Join(tempDir, fmt.Sprintf("%s-%s", chartName, version))
	_ = os.MkdirAll(chartPath, 0755)

	dl := downloader.ChartDownloader{
		Out:     io.Discard,
		Keyring: "",
		Getters: getter.All(c.settings),
		Options: []getter.Option{
			getter.WithURL(helmRepo.Config.URL), // Pass repo URL for context if needed by getters
			getter.WithTimeout(c.timeout),
		},
		RepositoryConfig: c.settings.RepositoryConfig,
		RepositoryCache:  c.settings.RepositoryCache,
		ContentCache:     c.settings.ContentCache,
		Verify:           downloader.VerifyNever,
	}

	chartOutputPath, _, err := dl.DownloadTo(validatedChartURL, version, chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to download chartv2 %s version %s from %s: %v", chartName, version, chartURL, err)
	}

	// Load the downloaded chartv2
	loadedChart, err := loader.Load(chartOutputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chartv2 from %s: %v", chartPath, err)
	}

	v2Chart, ok := loadedChart.(*chartv2.Chart)
	if !ok {
		return nil, fmt.Errorf("charts V3 format is not supported")
	}

	return v2Chart, nil
}

func (c *HelmClient) GetChartLatestVersion(repoURL, chartName string) (string, error) {
	helmRepo, err := c.getRepo(repoURL)
	if err != nil {
		return "", fmt.Errorf("failed to get repository: %v", err)
	}

	chartVersions, ok := helmRepo.IndexFile.Entries[chartName]
	if !ok || len(chartVersions) == 0 {
		return "", fmt.Errorf("chartv2 %s not found in repository %s", chartName, repoURL)
	}

	// IndexFile.SortEntries() sorts versions in descending order, so the first one is the latest.
	latestVersion := chartVersions[0].Version
	return latestVersion, nil
}

func (c *HelmClient) GetChartLatestValues(repoURL, chartName string) (string, error) {
	v, err := c.GetChartLatestVersion(repoURL, chartName)
	if err != nil {
		return "", fmt.Errorf("failed to get chartv2 %s version %s: %v", chartName, v, err)
	}

	return c.GetChartValues(repoURL, chartName, v)
}

func (c *HelmClient) GetChartDependencies(repoURL, chartName, version string) ([]string, error) {
	loadedChart, err := c.loadChart(repoURL, chartName, version)
	if err != nil {
		return nil, fmt.Errorf("failed to load chartv2 %s version %s: %v", chartName, version, err)
	}

	if loadedChart == nil {
		return nil, fmt.Errorf("chartv2 %s version %s not found", chartName, version)
	}

	deps, err := helm_parser.GetChartDependencies(loadedChart)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies for chartv2 %s version %s: %v", chartName, version, err)
	}
	return deps, nil
}
