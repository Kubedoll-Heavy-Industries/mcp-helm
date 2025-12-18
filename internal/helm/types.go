package helm

import (
	"context"
	"time"
)

// ChartService defines the contract for chart operations.
// This interface is defined here (at the domain layer) and should be
// implemented by Client. Consumers should depend on this interface.
type ChartService interface {
	// ListCharts returns all chart names available in the repository.
	ListCharts(ctx context.Context, repoURL string) ([]string, error)

	// ListVersions returns all versions of a chart with metadata.
	ListVersions(ctx context.Context, repoURL, chart string) ([]ChartVersion, error)

	// GetLatestVersion returns the latest version string for a chart.
	GetLatestVersion(ctx context.Context, repoURL, chart string) (string, error)

	// LoadChart loads a chart from the repository.
	LoadChart(ctx context.Context, repoURL, chart, version string) (*Chart, error)

	// GetValues returns the values.yaml contents for a chart.
	GetValues(ctx context.Context, repoURL, chart, version string) ([]byte, error)

	// GetValuesSchema returns the values.schema.json contents if present.
	// The boolean indicates whether the schema exists.
	GetValuesSchema(ctx context.Context, repoURL, chart, version string) ([]byte, bool, error)

	// GetContents returns a formatted string of all chart files.
	GetContents(ctx context.Context, repoURL, chart, version string, recursive bool) (string, error)

	// GetDependencies returns the chart's dependencies.
	GetDependencies(ctx context.Context, repoURL, chart, version string) ([]Dependency, error)

	// ListFiles returns metadata about all files in a chart.
	ListFiles(ctx context.Context, repoURL, chart, version string) ([]FileInfo, error)

	// GetFile returns the contents of a specific file in a chart.
	GetFile(ctx context.Context, repoURL, chart, version, path string) ([]byte, error)

	// RefreshIndex forces a refresh of the repository index cache.
	RefreshIndex(ctx context.Context, repoURL string) error
}

// ChartVersion represents metadata about a chart version.
type ChartVersion struct {
	Version    string
	AppVersion string
	Created    time.Time
	Deprecated bool
}

// File represents a file within a chart.
type File struct {
	Name string
	Data []byte
}

// Chart represents a loaded Helm chart.
type Chart struct {
	Name      string
	Version   string
	Raw       []File // Chart.yaml, values.yaml, etc.
	Templates []File // Template files
	Files     []File // Extra files
}

// Dependency represents a chart dependency.
type Dependency struct {
	Name       string `json:"name" yaml:"name"`
	Version    string `json:"version" yaml:"version"`
	Repository string `json:"repository,omitempty" yaml:"repository"`
	Condition  string `json:"condition,omitempty" yaml:"condition"`
	Alias      string `json:"alias,omitempty" yaml:"alias"`
}

// FileInfo represents metadata about a file in a chart.
type FileInfo struct {
	Path string
	Size int
}

// ChartReference is a validated reference to a specific chart version.
type ChartReference struct {
	Repository string
	Name       string
	Version    string // Empty means latest
}

// IsLatest returns true if this reference points to the latest version.
func (r ChartReference) IsLatest() bool {
	return r.Version == ""
}

// String returns a human-readable representation.
func (r ChartReference) String() string {
	if r.IsLatest() {
		return r.Repository + "/" + r.Name + ":latest"
	}
	return r.Repository + "/" + r.Name + ":" + r.Version
}
