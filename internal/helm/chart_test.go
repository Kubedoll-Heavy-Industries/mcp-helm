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
		assert.Contains(t, err.Error(), "chart.yaml not found")
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
