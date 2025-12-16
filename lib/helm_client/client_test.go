package helm_client

import (
	"archive/tar"
	"compress/gzip"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createTestHelmRepo(t *testing.T) (repoURL string, chartName string, chartVersion string) {
	t.Helper()

	chartName = "test-chart"
	chartVersion = "1.2.3"
	previousChartVersion := "1.2.2"

	rootDir := t.TempDir()

	chartDir := filepath.Join(rootDir, chartName)
	if err := os.MkdirAll(filepath.Join(chartDir, "templates"), 0o755); err != nil {
		t.Fatalf("failed to create chart directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(chartDir, "files"), 0o755); err != nil {
		t.Fatalf("failed to create chart files directory: %v", err)
	}

	writeChartYAML := func(version string, appVersion string, deprecated bool) {
		t.Helper()
		deprecatedYAML := "false"
		if deprecated {
			deprecatedYAML = "true"
		}
		src := strings.TrimSpace(`
apiVersion: v2
name: test-chart
version: `+version+`
appVersion: `+appVersion+`
deprecated: `+deprecatedYAML+`
description: test chart for mcp-helm
`) + "\n"
		if err := os.WriteFile(filepath.Join(chartDir, "Chart.yaml"), []byte(src), 0o644); err != nil {
			t.Fatalf("failed to write Chart.yaml: %v", err)
		}
	}

	writeChartYAML(chartVersion, "2.0.0", false)

	if err := os.WriteFile(filepath.Join(chartDir, "values.yaml"), []byte(strings.TrimSpace(`
replicaCount: 1
image:
  repository: nginx
  tag: latest
`)+"\n"), 0o644); err != nil {
		t.Fatalf("failed to write values.yaml: %v", err)
	}

	if err := os.WriteFile(filepath.Join(chartDir, "templates", "deployment.yaml"), []byte(strings.TrimSpace(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Release.Name }}
`)+"\n"), 0o644); err != nil {
		t.Fatalf("failed to write template: %v", err)
	}

	// Helm loader populates chart.Files from the chart "files/" directory (not "templates/").
	if err := os.WriteFile(filepath.Join(chartDir, "files", "extra.txt"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("failed to write chart file: %v", err)
	}

	chartArchiveName := chartName + "-" + chartVersion + ".tgz"
	chartArchivePath := filepath.Join(rootDir, chartArchiveName)
	if err := createTgz(chartArchivePath, chartDir, chartName); err != nil {
		t.Fatalf("failed to create chart archive: %v", err)
	}

	writeChartYAML(previousChartVersion, "1.9.0", true)
	previousArchiveName := chartName + "-" + previousChartVersion + ".tgz"
	previousArchivePath := filepath.Join(rootDir, previousArchiveName)
	if err := createTgz(previousArchivePath, chartDir, chartName); err != nil {
		t.Fatalf("failed to create previous chart archive: %v", err)
	}

	indexYAML := strings.TrimSpace(`
apiVersion: v1
entries:
  test-chart:
    - apiVersion: v2
      name: test-chart
      version: `+chartVersion+`
      appVersion: 2.0.0
      created: "2025-01-02T03:04:05Z"
      deprecated: false
      urls:
        - `+chartArchiveName+`
    - apiVersion: v2
      name: test-chart
      version: `+previousChartVersion+`
      appVersion: 1.9.0
      created: "2025-01-01T03:04:05Z"
      deprecated: true
      urls:
        - `+previousArchiveName+`
generated: "2025-01-03T00:00:00Z"
`) + "\n"
	if err := os.WriteFile(filepath.Join(rootDir, "index.yaml"), []byte(indexYAML), 0o644); err != nil {
		t.Fatalf("failed to write index.yaml: %v", err)
	}

	srv := httptest.NewServer(http.FileServer(http.Dir(rootDir)))
	t.Cleanup(srv.Close)

	return srv.URL, chartName, chartVersion
}

func createTgz(outputPath string, inputDir string, archiveRootName string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	gz := gzip.NewWriter(file)
	defer func() { _ = gz.Close() }()

	tw := tar.NewWriter(gz)
	defer func() { _ = tw.Close() }()

	return filepath.WalkDir(inputDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		realRel, err := filepath.Rel(inputDir, path)
		if err != nil {
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = filepath.ToSlash(filepath.Join(archiveRootName, realRel))

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		_, err = tw.Write(b)
		return err
	})
}

func TestNewClient(t *testing.T) {
	tmpDir = t.TempDir()

	client := NewClient()
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
	if client.settings == nil {
		t.Fatal("client.settings is nil")
	}
}

func newTestClient() *HelmClient {
	return NewClientWithOptions(Options{
		AllowPrivateIPs:    true,
		MaxToolOutputBytes: 2 * 1024 * 1024,
	})
}

func TestListCharts(t *testing.T) {
	tmpDir = t.TempDir()
	repoURL, chartName, _ := createTestHelmRepo(t)

	client := newTestClient()
	charts, err := client.ListCharts(repoURL)
	if err != nil {
		t.Fatalf("ListCharts() error = %v", err)
	}
	if len(charts) == 0 {
		t.Fatal("ListCharts() returned empty list")
	}

	// Check if expected chart is in the list
	found := false
	for _, chart := range charts {
		if chart == chartName {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("ListCharts() did not return the expected chart %s", chartName)
	}
}

func TestListChartVersions(t *testing.T) {
	tmpDir = t.TempDir()
	repoURL, chartName, chartVersion := createTestHelmRepo(t)

	client := newTestClient()
	versions, err := client.ListChartVersions(repoURL, chartName)
	if err != nil {
		t.Fatalf("ListChartVersions() error = %v", err)
	}
	if len(versions) == 0 {
		t.Fatal("ListChartVersions() returned empty list")
	}
	if versions[0] != chartVersion {
		t.Fatalf("ListChartVersions()[0] = %s, want %s", versions[0], chartVersion)
	}
}

func TestListChartVersionsWithMetadata(t *testing.T) {
	tmpDir = t.TempDir()
	repoURL, chartName, chartVersion := createTestHelmRepo(t)

	client := newTestClient()
	versions, err := client.ListChartVersionsWithMetadata(repoURL, chartName)
	if err != nil {
		t.Fatalf("ListChartVersionsWithMetadata() error = %v", err)
	}
	if len(versions) < 2 {
		t.Fatalf("ListChartVersionsWithMetadata() returned %d versions, want at least 2", len(versions))
	}
	if versions[0].Version != chartVersion {
		t.Fatalf("ListChartVersionsWithMetadata()[0].Version = %s, want %s", versions[0].Version, chartVersion)
	}
	if versions[0].Created == "" {
		t.Fatal("ListChartVersionsWithMetadata()[0].Created is empty")
	}
}

func TestGetChartLatestVersion(t *testing.T) {
	tmpDir = t.TempDir()
	repoURL, chartName, chartVersion := createTestHelmRepo(t)

	client := newTestClient()
	version, err := client.GetChartLatestVersion(repoURL, chartName)
	if err != nil {
		t.Fatalf("GetChartLatestVersion() error = %v", err)
	}
	if version == "" {
		t.Fatal("GetChartLatestVersion() returned empty version")
	}
	if version != chartVersion {
		t.Fatalf("GetChartLatestVersion() = %s, want %s", version, chartVersion)
	}
}

func TestGetChartValues(t *testing.T) {
	tmpDir = t.TempDir()
	repoURL, chartName, _ := createTestHelmRepo(t)

	client := newTestClient()

	// Get the latest version first
	version, err := client.GetChartLatestVersion(repoURL, chartName)
	if err != nil {
		t.Fatalf("GetChartLatestVersion() error = %v", err)
	}

	values, err := client.GetChartValues(repoURL, chartName, version)
	if err != nil {
		t.Fatalf("GetChartValues() error = %v", err)
	}
	if values == "" {
		t.Fatal("GetChartValues() returned empty values")
	}

	// Check if the values contain some YAML structure (any key-value pair)
	if !strings.Contains(values, ":") {
		t.Fatal("GetChartValues() did not return expected YAML structure")
	}
}

func TestGetChartLatestValues(t *testing.T) {
	tmpDir = t.TempDir()
	repoURL, chartName, _ := createTestHelmRepo(t)

	client := newTestClient()
	values, err := client.GetChartLatestValues(repoURL, chartName)
	if err != nil {
		t.Fatalf("GetChartLatestValues() error = %v", err)
	}
	if values == "" {
		t.Fatal("GetChartLatestValues() returned empty values")
	}
}

func TestGetChartContents(t *testing.T) {
	tmpDir = t.TempDir()
	repoURL, chartName, _ := createTestHelmRepo(t)

	client := newTestClient()

	// Get the latest version first
	version, err := client.GetChartLatestVersion(repoURL, chartName)
	if err != nil {
		t.Fatalf("GetChartLatestVersion() error = %v", err)
	}

	// Test without recursion
	contents, err := client.GetChartContents(repoURL, chartName, version, false)
	if err != nil {
		t.Fatalf("GetChartContents(recursive=false) error = %v", err)
	}
	if contents == "" {
		t.Fatal("GetChartContents(recursive=false) returned empty contents")
	}

	// Test with recursion
	contentsRecursive, err := client.GetChartContents(repoURL, chartName, version, true)
	if err != nil {
		t.Fatalf("GetChartContents(recursive=true) error = %v", err)
	}
	if contentsRecursive == "" {
		t.Fatal("GetChartContents(recursive=true) returned empty contents")
	}

	// Recursive contents should be longer or equal to non-recursive contents
	if len(contentsRecursive) < len(contents) {
		t.Fatal("Recursive contents should be longer or equal to non-recursive contents")
	}
}

func TestGetChartDependencies(t *testing.T) {
	tmpDir = t.TempDir()
	repoURL, chartName, _ := createTestHelmRepo(t)

	client := newTestClient()

	// Get the latest version first
	version, err := client.GetChartLatestVersion(repoURL, chartName)
	if err != nil {
		t.Fatalf("GetChartLatestVersion() error = %v", err)
	}

	deps, err := client.GetChartDependencies(repoURL, chartName, version)
	if err != nil {
		t.Fatalf("GetChartDependencies() error = %v", err)
	}

	// Note: The test chart may or may not have dependencies.
	// A nil slice is a valid representation of "no dependencies".
	_ = deps
}
