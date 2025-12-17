package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v4/pkg/chart/common"
	chartv2 "helm.sh/helm/v4/pkg/chart/v2"
)

func makeTestChart(name, version string, rawFiles, templates, extraFiles []*common.File) *chartv2.Chart {
	return &chartv2.Chart{
		Metadata:  &chartv2.Metadata{Name: name, Version: version},
		Raw:       rawFiles,
		Templates: templates,
		Files:     extraFiles,
	}
}

func TestExtractDependencies(t *testing.T) {
	t.Run("chart with dependencies", func(t *testing.T) {
		chartYAML := `
name: myapp
version: 1.0.0
dependencies:
  - name: redis
    version: "17.x"
    repository: "https://charts.bitnami.com/bitnami"
  - name: postgresql
    version: "12.x"
    repository: "https://charts.bitnami.com/bitnami"
`
		chart := makeTestChart("myapp", "1.0.0",
			[]*common.File{{Name: "Chart.yaml", Data: []byte(chartYAML)}},
			nil, nil)

		deps, err := extractDependencies(chart)

		require.NoError(t, err)
		assert.Len(t, deps, 2)
		assert.Equal(t, "redis", deps[0].Name)
		assert.Equal(t, "17.x", deps[0].Version)
		assert.Equal(t, "https://charts.bitnami.com/bitnami", deps[0].Repository)
	})

	t.Run("chart without dependencies", func(t *testing.T) {
		chartYAML := `
name: simple
version: 1.0.0
`
		chart := makeTestChart("simple", "1.0.0",
			[]*common.File{{Name: "Chart.yaml", Data: []byte(chartYAML)}},
			nil, nil)

		deps, err := extractDependencies(chart)

		require.NoError(t, err)
		assert.Nil(t, deps)
	})

	t.Run("missing Chart.yaml", func(t *testing.T) {
		chart := makeTestChart("nofile", "1.0.0", nil, nil, nil)

		_, err := extractDependencies(chart)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Chart.yaml not found")
	})

	t.Run("invalid YAML", func(t *testing.T) {
		chart := makeTestChart("invalid", "1.0.0",
			[]*common.File{{Name: "Chart.yaml", Data: []byte("invalid: [yaml")}},
			nil, nil)

		_, err := extractDependencies(chart)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse")
	})

	t.Run("dependency missing name", func(t *testing.T) {
		chartYAML := `
name: myapp
version: 1.0.0
dependencies:
  - name: ""
    version: "1.0.0"
    repository: "https://example.com"
`
		chart := makeTestChart("myapp", "1.0.0",
			[]*common.File{{Name: "Chart.yaml", Data: []byte(chartYAML)}},
			nil, nil)

		_, err := extractDependencies(chart)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required field: name")
	})

	t.Run("dependency missing version", func(t *testing.T) {
		chartYAML := `
name: myapp
version: 1.0.0
dependencies:
  - name: redis
    version: ""
    repository: "https://example.com"
`
		chart := makeTestChart("myapp", "1.0.0",
			[]*common.File{{Name: "Chart.yaml", Data: []byte(chartYAML)}},
			nil, nil)

		_, err := extractDependencies(chart)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required field: version")
	})

	t.Run("dependency with empty repository is valid (bundled)", func(t *testing.T) {
		chartYAML := `
name: myapp
version: 1.0.0
dependencies:
  - name: crds
    version: "1.0.0"
    repository: ""
`
		chart := makeTestChart("myapp", "1.0.0",
			[]*common.File{{Name: "Chart.yaml", Data: []byte(chartYAML)}},
			nil, nil)

		deps, err := extractDependencies(chart)

		require.NoError(t, err)
		assert.Len(t, deps, 1)
		assert.Equal(t, "crds", deps[0].Name)
		assert.Equal(t, "", deps[0].Repository) // empty is valid
	})
}

func TestFormatDependenciesJSON(t *testing.T) {
	t.Run("formats dependencies as JSON", func(t *testing.T) {
		deps := []Dependency{
			{Name: "redis", Version: "17.x", Repository: "https://bitnami.com"},
			{Name: "postgresql", Version: "12.x", Repository: "https://bitnami.com"},
		}

		result, err := formatDependenciesJSON(deps)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Contains(t, result[0], `"name":"redis"`)
		assert.Contains(t, result[1], `"name":"postgresql"`)
	})

	t.Run("empty dependencies", func(t *testing.T) {
		result, err := formatDependenciesJSON(nil)

		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestFormatContents(t *testing.T) {
	t.Run("includes all file types", func(t *testing.T) {
		chart := makeTestChart("myapp", "1.0.0",
			[]*common.File{{Name: "Chart.yaml", Data: []byte("name: myapp")}},
			[]*common.File{{Name: "templates/deployment.yaml", Data: []byte("kind: Deployment")}},
			[]*common.File{{Name: "README.md", Data: []byte("# README")}},
		)

		result := formatContents(chart, false)

		assert.Contains(t, result, "# file: myapp/Chart.yaml")
		assert.Contains(t, result, "name: myapp")
		assert.Contains(t, result, "# file: myapp/templates/deployment.yaml")
		assert.Contains(t, result, "kind: Deployment")
		assert.Contains(t, result, "# file: myapp/README.md")
		assert.Contains(t, result, "# README")
	})

	t.Run("non-recursive skips sub-charts", func(t *testing.T) {
		subChart := makeTestChart("redis", "17.0.0",
			[]*common.File{{Name: "Chart.yaml", Data: []byte("name: redis")}},
			nil, nil)

		parent := makeTestChart("myapp", "1.0.0",
			[]*common.File{{Name: "Chart.yaml", Data: []byte("name: myapp")}},
			nil, nil)
		parent.SetDependencies(subChart)

		result := formatContents(parent, false)

		assert.Contains(t, result, "myapp/Chart.yaml")
		assert.NotContains(t, result, "Subchart: redis")
		assert.NotContains(t, result, "redis/Chart.yaml")
	})

	t.Run("recursive includes sub-charts", func(t *testing.T) {
		subChart := makeTestChart("redis", "17.0.0",
			[]*common.File{{Name: "Chart.yaml", Data: []byte("name: redis")}},
			nil, nil)

		parent := makeTestChart("myapp", "1.0.0",
			[]*common.File{{Name: "Chart.yaml", Data: []byte("name: myapp")}},
			nil, nil)
		parent.SetDependencies(subChart)

		result := formatContents(parent, true)

		assert.Contains(t, result, "myapp/Chart.yaml")
		assert.Contains(t, result, "# Subchart: redis")
		assert.Contains(t, result, "redis/Chart.yaml")
	})
}

func TestListChartFiles(t *testing.T) {
	t.Run("lists all files", func(t *testing.T) {
		chart := makeTestChart("myapp", "1.0.0",
			[]*common.File{{Name: "Chart.yaml"}, {Name: "values.yaml"}},
			[]*common.File{{Name: "templates/deployment.yaml"}, {Name: "templates/service.yaml"}},
			[]*common.File{{Name: "README.md"}},
		)

		files := listChartFiles(chart)

		assert.Len(t, files, 5)
		assert.Contains(t, files, "Chart.yaml")
		assert.Contains(t, files, "values.yaml")
		assert.Contains(t, files, "templates/deployment.yaml")
		assert.Contains(t, files, "templates/service.yaml")
		assert.Contains(t, files, "README.md")
	})

	t.Run("empty chart", func(t *testing.T) {
		chart := makeTestChart("empty", "1.0.0", nil, nil, nil)

		files := listChartFiles(chart)

		assert.Empty(t, files)
	})
}

func TestFindChartFile(t *testing.T) {
	chart := makeTestChart("myapp", "1.0.0",
		[]*common.File{{Name: "Chart.yaml", Data: []byte("name: myapp")}},
		[]*common.File{{Name: "templates/deployment.yaml", Data: []byte("kind: Deployment")}},
		[]*common.File{{Name: "README.md", Data: []byte("# README")}},
	)

	t.Run("finds raw file", func(t *testing.T) {
		data, found := findChartFile(chart, "Chart.yaml")

		assert.True(t, found)
		assert.Equal(t, []byte("name: myapp"), data)
	})

	t.Run("finds template with full path", func(t *testing.T) {
		data, found := findChartFile(chart, "templates/deployment.yaml")

		assert.True(t, found)
		assert.Equal(t, []byte("kind: Deployment"), data)
	})

	t.Run("finds extra file", func(t *testing.T) {
		data, found := findChartFile(chart, "README.md")

		assert.True(t, found)
		assert.Equal(t, []byte("# README"), data)
	})

	t.Run("returns false for missing file", func(t *testing.T) {
		_, found := findChartFile(chart, "nonexistent.yaml")

		assert.False(t, found)
	})
}

func TestConvertToChart(t *testing.T) {
	t.Run("converts all fields", func(t *testing.T) {
		hc := makeTestChart("myapp", "1.2.3",
			[]*common.File{
				{Name: "Chart.yaml", Data: []byte("name: myapp")},
				{Name: "values.yaml", Data: []byte("key: value")},
			},
			[]*common.File{
				{Name: "templates/deployment.yaml", Data: []byte("kind: Deployment")},
			},
			[]*common.File{
				{Name: "README.md", Data: []byte("# README")},
			},
		)

		result := convertToChart(hc)

		assert.Equal(t, "myapp", result.Name)
		assert.Equal(t, "1.2.3", result.Version)
		assert.Len(t, result.Raw, 2)
		assert.Len(t, result.Templates, 1)
		assert.Len(t, result.Files, 1)

		// Verify file contents preserved
		assert.Equal(t, "Chart.yaml", result.Raw[0].Name)
		assert.Equal(t, []byte("name: myapp"), result.Raw[0].Data)
	})

	t.Run("handles empty chart", func(t *testing.T) {
		hc := makeTestChart("empty", "0.0.1", nil, nil, nil)

		result := convertToChart(hc)

		assert.Equal(t, "empty", result.Name)
		assert.Equal(t, "0.0.1", result.Version)
		assert.Empty(t, result.Raw)
		assert.Empty(t, result.Templates)
		assert.Empty(t, result.Files)
	})
}
