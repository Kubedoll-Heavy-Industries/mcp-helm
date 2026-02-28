package helm

import (
	"context"
	"time"
)

// ChartService defines the contract for chart operations.
// This interface is defined here (at the domain layer) and should be
// implemented by Client. Consumers should depend on this interface.
//
// Context handling:
//   - All methods accept a context for cancellation and deadlines
//   - Network operations respect context deadlines via the client's configured timeout
//   - Pass context.Background() for unbounded operations, or use context.WithTimeout()
//   - Cached responses may return faster than the timeout
type ChartService interface {
	// ListCharts returns all chart names available in the repository.
	ListCharts(ctx context.Context, repoURL string) ([]string, error)

	// ListVersions returns all versions of a chart with metadata.
	ListVersions(ctx context.Context, repoURL, chart string) ([]ChartVersion, error)

	// GetLatestVersion returns the latest version string for a chart.
	GetLatestVersion(ctx context.Context, repoURL, chart string) (string, error)

	// GetValues returns the values.yaml contents for a chart.
	GetValues(ctx context.Context, repoURL, chart, version string) ([]byte, error)

	// GetValuesSchema returns the values.schema.json contents if present.
	// The boolean indicates whether the schema exists.
	GetValuesSchema(ctx context.Context, repoURL, chart, version string) ([]byte, bool, error)

	// GetNotes returns the NOTES.txt contents if present.
	// The boolean indicates whether the file exists.
	GetNotes(ctx context.Context, repoURL, chart, version string) ([]byte, bool, error)

	// GetDependencies returns the chart's dependencies.
	GetDependencies(ctx context.Context, repoURL, chart, version string) ([]Dependency, error)
}

// ChartVersion represents metadata about a chart version.
type ChartVersion struct {
	Version    string
	AppVersion string
	Created    time.Time
	Deprecated bool
}

// Dependency represents a chart dependency.
type Dependency struct {
	Name       string `json:"name" yaml:"name"`
	Version    string `json:"version" yaml:"version"`
	Repository string `json:"repository,omitempty" yaml:"repository"`
	Condition  string `json:"condition,omitempty" yaml:"condition"`
	Alias      string `json:"alias,omitempty" yaml:"alias"`
}
