package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/Kubedoll-Heavy-Industries/mcp-helm/internal/helm"
	"github.com/Kubedoll-Heavy-Industries/mcp-helm/internal/helm/mocks"
)

func TestNew(t *testing.T) {
	t.Run("with nil logger uses nop", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		h := New(mockSvc, nil)

		assert.NotNil(t, h)
		assert.NotNil(t, h.logger)
	})

	t.Run("with provided logger", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		logger := zap.NewNop()
		h := New(mockSvc, logger)

		assert.NotNil(t, h)
		assert.Equal(t, logger, h.logger)
	})
}

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		name    string
		fields  map[string]string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "all fields present",
			fields:  map[string]string{"repo": "https://example.com", "chart": "nginx"},
			wantErr: false,
		},
		{
			name:    "empty field",
			fields:  map[string]string{"repo": "", "chart": "nginx"},
			wantErr: true,
			errMsg:  "repo is required",
		},
		{
			name:    "whitespace only",
			fields:  map[string]string{"repo": "   ", "chart": "nginx"},
			wantErr: true,
			errMsg:  "repo is required",
		},
		{
			name:    "empty map",
			fields:  map[string]string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateRequired(tt.fields)

			if tt.wantErr {
				assert.NotNil(t, result)
				assert.True(t, result.IsError)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestResolveVersion(t *testing.T) {
	ctx := context.Background()

	t.Run("explicit version returned as-is", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		h := New(mockSvc, zap.NewNop())

		version, err := h.resolveVersion(ctx, "https://repo.com", "nginx", "1.2.3")

		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
		mockSvc.AssertNotCalled(t, "GetLatestVersion")
	})

	t.Run("whitespace version trimmed", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		h := New(mockSvc, zap.NewNop())

		version, err := h.resolveVersion(ctx, "https://repo.com", "nginx", "  1.2.3  ")

		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
	})

	t.Run("empty version fetches latest", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		mockSvc.On("GetLatestVersion", ctx, "https://repo.com", "nginx").
			Return("2.0.0", nil)

		h := New(mockSvc, zap.NewNop())

		version, err := h.resolveVersion(ctx, "https://repo.com", "nginx", "")

		assert.NoError(t, err)
		assert.Equal(t, "2.0.0", version)
		mockSvc.AssertExpectations(t)
	})

	t.Run("whitespace-only version fetches latest", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		mockSvc.On("GetLatestVersion", ctx, "https://repo.com", "nginx").
			Return("2.0.0", nil)

		h := New(mockSvc, zap.NewNop())

		version, err := h.resolveVersion(ctx, "https://repo.com", "nginx", "   ")

		assert.NoError(t, err)
		assert.Equal(t, "2.0.0", version)
	})

	t.Run("error from GetLatestVersion propagated", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		mockSvc.On("GetLatestVersion", ctx, "https://repo.com", "nginx").
			Return("", errors.New("chart not found"))

		h := New(mockSvc, zap.NewNop())

		_, err := h.resolveVersion(ctx, "https://repo.com", "nginx", "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "chart not found")
	})
}

func TestListRepositoryCharts(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		mockSvc.On("ListCharts", ctx, "https://charts.bitnami.com/bitnami").
			Return([]string{"nginx", "redis", "postgresql"}, nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.listRepositoryCharts()

		result, output, err := handler(ctx, nil, listRepositoryChartsInput{
			RepositoryURL: "https://charts.bitnami.com/bitnami",
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Equal(t, []string{"nginx", "redis", "postgresql"}, output.Charts)
		assert.Equal(t, 3, output.Total)
		assert.Equal(t, defaultChartListLimit, output.Limit)
		mockSvc.AssertExpectations(t)
	})

	t.Run("empty repository", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		mockSvc.On("ListCharts", ctx, "https://empty.repo").
			Return([]string{}, nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.listRepositoryCharts()

		result, output, err := handler(ctx, nil, listRepositoryChartsInput{
			RepositoryURL: "https://empty.repo",
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Empty(t, output.Charts)
	})

	t.Run("missing repository_url", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		h := New(mockSvc, zap.NewNop())
		handler := h.listRepositoryCharts()

		result, _, err := handler(ctx, nil, listRepositoryChartsInput{
			RepositoryURL: "",
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsError)
		mockSvc.AssertNotCalled(t, "ListCharts")
	})

	t.Run("service error", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		mockSvc.On("ListCharts", ctx, "https://bad.repo").
			Return(nil, errors.New("network error"))

		h := New(mockSvc, zap.NewNop())
		handler := h.listRepositoryCharts()

		result, _, err := handler(ctx, nil, listRepositoryChartsInput{
			RepositoryURL: "https://bad.repo",
		})

		assert.NoError(t, err) // Handler errors are in result, not err
		assert.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("trims whitespace from URL", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		mockSvc.On("ListCharts", ctx, "https://charts.bitnami.com/bitnami").
			Return([]string{"nginx"}, nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.listRepositoryCharts()

		result, output, err := handler(ctx, nil, listRepositoryChartsInput{
			RepositoryURL: "  https://charts.bitnami.com/bitnami  ",
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Equal(t, []string{"nginx"}, output.Charts)
	})

	t.Run("with pagination", func(t *testing.T) {
		charts := []string{"a", "b", "c", "d", "e"}
		mockSvc := new(mocks.ChartService)
		mockSvc.On("ListCharts", ctx, "https://repo.com").
			Return(charts, nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.listRepositoryCharts()

		result, output, err := handler(ctx, nil, listRepositoryChartsInput{
			RepositoryURL: "https://repo.com",
			Limit:         2,
			Offset:        1,
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Equal(t, []string{"b", "c"}, output.Charts)
		assert.Equal(t, 5, output.Total)
		assert.Equal(t, 2, output.Limit)
		assert.Equal(t, 1, output.Offset)
	})

	t.Run("with search filter", func(t *testing.T) {
		charts := []string{"nginx", "nginx-ingress", "redis", "redis-cluster"}
		mockSvc := new(mocks.ChartService)
		mockSvc.On("ListCharts", ctx, "https://repo.com").
			Return(charts, nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.listRepositoryCharts()

		result, output, err := handler(ctx, nil, listRepositoryChartsInput{
			RepositoryURL: "https://repo.com",
			Search:        "redis",
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Equal(t, []string{"redis", "redis-cluster"}, output.Charts)
		assert.Equal(t, 2, output.Total)
	})
}

func TestListChartVersions(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		versions := []helm.ChartVersion{
			{Version: "2.0.0", AppVersion: "1.25.0", Created: time.Now(), Deprecated: false},
			{Version: "1.0.0", AppVersion: "1.24.0", Created: time.Now(), Deprecated: true},
		}

		mockSvc := new(mocks.ChartService)
		mockSvc.On("ListVersions", ctx, "https://repo.com", "nginx").
			Return(versions, nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.listChartVersions()

		result, output, err := handler(ctx, nil, listChartVersionsInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "nginx",
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Len(t, output.Versions, 2)
		assert.Equal(t, "2.0.0", output.Versions[0].Version)
		assert.Equal(t, 2, output.Total)
	})

	t.Run("with pagination", func(t *testing.T) {
		versions := []helm.ChartVersion{
			{Version: "3.0.0"},
			{Version: "2.0.0"},
			{Version: "1.0.0"},
		}

		mockSvc := new(mocks.ChartService)
		mockSvc.On("ListVersions", ctx, "https://repo.com", "nginx").
			Return(versions, nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.listChartVersions()

		result, output, err := handler(ctx, nil, listChartVersionsInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "nginx",
			Limit:         2,
			Offset:        1,
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Len(t, output.Versions, 2)
		assert.Equal(t, "2.0.0", output.Versions[0].Version)
		assert.Equal(t, "1.0.0", output.Versions[1].Version)
		assert.Equal(t, 3, output.Total)
		assert.Equal(t, 2, output.Limit)
		assert.Equal(t, 1, output.Offset)
	})

	t.Run("negative limit rejected", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		h := New(mockSvc, zap.NewNop())
		handler := h.listChartVersions()

		result, _, err := handler(ctx, nil, listChartVersionsInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "nginx",
			Limit:         -1,
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("negative offset rejected", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		h := New(mockSvc, zap.NewNop())
		handler := h.listChartVersions()

		result, _, err := handler(ctx, nil, listChartVersionsInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "nginx",
			Offset:        -1,
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("missing required fields", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		h := New(mockSvc, zap.NewNop())
		handler := h.listChartVersions()

		result, _, err := handler(ctx, nil, listChartVersionsInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "",
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("default limit applied", func(t *testing.T) {
		// Create 25 versions
		versions := make([]helm.ChartVersion, 25)
		for i := 0; i < 25; i++ {
			versions[i] = helm.ChartVersion{Version: "1.0." + string(rune('0'+i))}
		}

		mockSvc := new(mocks.ChartService)
		mockSvc.On("ListVersions", ctx, "https://repo.com", "nginx").
			Return(versions, nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.listChartVersions()

		result, output, err := handler(ctx, nil, listChartVersionsInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "nginx",
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Len(t, output.Versions, defaultVersionListLimit)
		assert.Equal(t, 25, output.Total)
		assert.Equal(t, defaultVersionListLimit, output.Limit)
	})
}

func TestGetChartLatestVersion(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		mockSvc.On("GetLatestVersion", ctx, "https://repo.com", "nginx").
			Return("2.0.0", nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.getChartLatestVersion()

		result, output, err := handler(ctx, nil, getChartLatestVersionInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "nginx",
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "2.0.0", output.Version)
	})

	t.Run("chart not found", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		mockSvc.On("GetLatestVersion", ctx, "https://repo.com", "nonexistent").
			Return("", &helm.ChartNotFoundError{Chart: "nonexistent"})

		h := New(mockSvc, zap.NewNop())
		handler := h.getChartLatestVersion()

		result, _, err := handler(ctx, nil, getChartLatestVersionInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "nonexistent",
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

func TestGetChartValues(t *testing.T) {
	ctx := context.Background()

	t.Run("success with explicit version", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		mockSvc.On("GetValues", ctx, "https://repo.com", "nginx", "1.0.0").
			Return([]byte("replicaCount: 1\nimage: nginx"), nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.getChartValues()

		result, output, err := handler(ctx, nil, getChartValuesInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "nginx",
			ChartVersion:  "1.0.0",
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "replicaCount: 1\nimage: nginx", output.Values)
	})

	t.Run("resolves latest version", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		mockSvc.On("GetLatestVersion", ctx, "https://repo.com", "nginx").
			Return("2.0.0", nil)
		mockSvc.On("GetValues", ctx, "https://repo.com", "nginx", "2.0.0").
			Return([]byte("replicaCount: 2"), nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.getChartValues()

		result, output, err := handler(ctx, nil, getChartValuesInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "nginx",
			ChartVersion:  "", // Should resolve to latest
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "replicaCount: 2", output.Values)
		mockSvc.AssertExpectations(t)
	})

	t.Run("missing required fields", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		h := New(mockSvc, zap.NewNop())
		handler := h.getChartValues()

		result, _, err := handler(ctx, nil, getChartValuesInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "", // Missing
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("with path extraction", func(t *testing.T) {
		yamlContent := `server:
  port: 8080
  host: localhost
client:
  timeout: 30`
		mockSvc := new(mocks.ChartService)
		mockSvc.On("GetValues", ctx, "https://repo.com", "app", "1.0.0").
			Return([]byte(yamlContent), nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.getChartValues()

		result, output, err := handler(ctx, nil, getChartValuesInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "app",
			ChartVersion:  "1.0.0",
			Path:          ".server",
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Contains(t, output.Values, "port: 8080")
		assert.Contains(t, output.Values, "host: localhost")
	})

	t.Run("with max_lines truncation", func(t *testing.T) {
		// Create content with many lines
		lines := ""
		for i := 0; i < 300; i++ {
			lines += "line" + string(rune('0'+i%10)) + ": value\n"
		}
		mockSvc := new(mocks.ChartService)
		mockSvc.On("GetValues", ctx, "https://repo.com", "nginx", "1.0.0").
			Return([]byte(lines), nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.getChartValues()

		result, output, err := handler(ctx, nil, getChartValuesInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "nginx",
			ChartVersion:  "1.0.0",
			MaxLines:      50,
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.True(t, output.Truncated)
		assert.Contains(t, output.Values, "... truncated")
	})
}

func TestGetChartValuesSchema(t *testing.T) {
	ctx := context.Background()

	t.Run("schema present", func(t *testing.T) {
		schema := []byte(`{"type": "object"}`)
		mockSvc := new(mocks.ChartService)
		mockSvc.On("GetValuesSchema", ctx, "https://repo.com", "nginx", "1.0.0").
			Return(schema, true, nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.getValuesSchema()

		result, output, err := handler(ctx, nil, getValuesSchemaInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "nginx",
			ChartVersion:  "1.0.0",
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Equal(t, `{"type": "object"}`, output.Schema)
		assert.True(t, output.Present)
	})

	t.Run("schema absent", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		mockSvc.On("GetValuesSchema", ctx, "https://repo.com", "nginx", "1.0.0").
			Return(nil, false, nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.getValuesSchema()

		result, output, err := handler(ctx, nil, getValuesSchemaInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "nginx",
			ChartVersion:  "1.0.0",
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Empty(t, output.Schema)
		assert.False(t, output.Present)
	})
}

func TestListChartContents(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		files := []helm.FileInfo{
			{Path: "Chart.yaml", Size: 100},
			{Path: "values.yaml", Size: 500},
			{Path: "templates/deployment.yaml", Size: 1000},
		}
		mockSvc := new(mocks.ChartService)
		mockSvc.On("ListFiles", ctx, "https://repo.com", "nginx", "1.0.0").
			Return(files, nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.listChartContents()

		result, output, err := handler(ctx, nil, listChartContentsInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "nginx",
			ChartVersion:  "1.0.0",
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Len(t, output.Files, 3)
		assert.Equal(t, "Chart.yaml", output.Files[0].Path)
		assert.Equal(t, 3, output.Total)
	})

	t.Run("with glob pattern", func(t *testing.T) {
		files := []helm.FileInfo{
			{Path: "Chart.yaml", Size: 100},
			{Path: "values.yaml", Size: 500},
			{Path: "templates/deployment.yaml", Size: 1000},
			{Path: "templates/service.yaml", Size: 800},
		}
		mockSvc := new(mocks.ChartService)
		mockSvc.On("ListFiles", ctx, "https://repo.com", "nginx", "1.0.0").
			Return(files, nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.listChartContents()

		result, output, err := handler(ctx, nil, listChartContentsInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "nginx",
			ChartVersion:  "1.0.0",
			Pattern:       "templates/*.yaml",
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Len(t, output.Files, 2)
		assert.Equal(t, "templates/deployment.yaml", output.Files[0].Path)
	})

	t.Run("resolves latest version", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		mockSvc.On("GetLatestVersion", ctx, "https://repo.com", "nginx").
			Return("2.0.0", nil)
		mockSvc.On("ListFiles", ctx, "https://repo.com", "nginx", "2.0.0").
			Return([]helm.FileInfo{{Path: "Chart.yaml", Size: 100}}, nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.listChartContents()

		result, output, err := handler(ctx, nil, listChartContentsInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "nginx",
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Len(t, output.Files, 1)
		mockSvc.AssertExpectations(t)
	})
}

func TestGetChartContent(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		mockSvc.On("GetFile", ctx, "https://repo.com", "nginx", "1.0.0", "Chart.yaml").
			Return([]byte("name: nginx\nversion: 1.0.0"), nil)
		mockSvc.On("GetFile", ctx, "https://repo.com", "nginx", "1.0.0", "values.yaml").
			Return([]byte("replicaCount: 1"), nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.getChartContent()

		result, output, err := handler(ctx, nil, getChartContentInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "nginx",
			ChartVersion:  "1.0.0",
			Files:         []string{"Chart.yaml", "values.yaml"},
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Len(t, output.Files, 2)
		assert.Equal(t, "Chart.yaml", output.Files[0].Path)
		assert.Contains(t, output.Files[0].Content, "name: nginx")
	})

	t.Run("file not found", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		mockSvc.On("GetFile", ctx, "https://repo.com", "nginx", "1.0.0", "nonexistent.yaml").
			Return(nil, errors.New("file not found: nonexistent.yaml"))

		h := New(mockSvc, zap.NewNop())
		handler := h.getChartContent()

		result, output, err := handler(ctx, nil, getChartContentInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "nginx",
			ChartVersion:  "1.0.0",
			Files:         []string{"nonexistent.yaml"},
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Len(t, output.Files, 1)
		assert.Contains(t, output.Files[0].Content, "not found")
	})

	t.Run("missing files parameter", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		h := New(mockSvc, zap.NewNop())
		handler := h.getChartContent()

		result, _, err := handler(ctx, nil, getChartContentInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "nginx",
			ChartVersion:  "1.0.0",
			Files:         []string{},
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

func TestGetChartDependencies(t *testing.T) {
	ctx := context.Background()

	t.Run("success with dependencies", func(t *testing.T) {
		deps := []helm.Dependency{
			{Name: "redis", Version: "17.x", Repository: "https://charts.bitnami.com/bitnami"},
			{Name: "postgresql", Version: "12.x", Repository: "https://charts.bitnami.com/bitnami"},
		}

		mockSvc := new(mocks.ChartService)
		mockSvc.On("GetDependencies", ctx, "https://repo.com", "app", "1.0.0").
			Return(deps, nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.getChartDependencies()

		result, output, err := handler(ctx, nil, getChartDependenciesInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "app",
			ChartVersion:  "1.0.0",
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Len(t, output.Dependencies, 2)
		assert.Equal(t, "redis", output.Dependencies[0].Name)
		assert.Equal(t, "postgresql", output.Dependencies[1].Name)
	})

	t.Run("no dependencies", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		mockSvc.On("GetDependencies", ctx, "https://repo.com", "simple", "1.0.0").
			Return([]helm.Dependency{}, nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.getChartDependencies()

		result, output, err := handler(ctx, nil, getChartDependenciesInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "simple",
			ChartVersion:  "1.0.0",
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Empty(t, output.Dependencies)
	})
}

func TestRefreshIndex(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		mockSvc.On("RefreshIndex", ctx, "https://repo.com").
			Return(nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.refreshIndex()

		result, output, err := handler(ctx, nil, refreshIndexInput{
			RepositoryURL: "https://repo.com",
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.True(t, output.Success)
		assert.Equal(t, "Repository index refreshed successfully", output.Message)
	})

	t.Run("error", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		mockSvc.On("RefreshIndex", ctx, "https://bad.repo").
			Return(errors.New("network error"))

		h := New(mockSvc, zap.NewNop())
		handler := h.refreshIndex()

		result, _, err := handler(ctx, nil, refreshIndexInput{
			RepositoryURL: "https://bad.repo",
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("missing repository_url", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		h := New(mockSvc, zap.NewNop())
		handler := h.refreshIndex()

		result, _, err := handler(ctx, nil, refreshIndexInput{
			RepositoryURL: "",
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsError)
		mockSvc.AssertNotCalled(t, "RefreshIndex", mock.Anything, mock.Anything)
	})
}
