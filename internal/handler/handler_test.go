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

func TestListCharts(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		mockSvc.On("ListCharts", ctx, "https://charts.bitnami.com/bitnami").
			Return([]string{"nginx", "redis", "postgresql"}, nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.listCharts()

		result, output, err := handler(ctx, nil, listChartsInput{
			RepositoryURL: "https://charts.bitnami.com/bitnami",
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Equal(t, []string{"nginx", "redis", "postgresql"}, output.Charts)
		mockSvc.AssertExpectations(t)
	})

	t.Run("empty repository", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		mockSvc.On("ListCharts", ctx, "https://empty.repo").
			Return([]string{}, nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.listCharts()

		result, output, err := handler(ctx, nil, listChartsInput{
			RepositoryURL: "https://empty.repo",
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Empty(t, output.Charts)
	})

	t.Run("missing repository_url", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		h := New(mockSvc, zap.NewNop())
		handler := h.listCharts()

		result, _, err := handler(ctx, nil, listChartsInput{
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
		handler := h.listCharts()

		result, _, err := handler(ctx, nil, listChartsInput{
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
		handler := h.listCharts()

		result, output, err := handler(ctx, nil, listChartsInput{
			RepositoryURL: "  https://charts.bitnami.com/bitnami  ",
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Equal(t, []string{"nginx"}, output.Charts)
	})
}

func TestListVersions(t *testing.T) {
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
		handler := h.listVersions()

		result, output, err := handler(ctx, nil, listVersionsInput{
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
		handler := h.listVersions()

		result, output, err := handler(ctx, nil, listVersionsInput{
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
		handler := h.listVersions()

		result, _, err := handler(ctx, nil, listVersionsInput{
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
		handler := h.listVersions()

		result, _, err := handler(ctx, nil, listVersionsInput{
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
		handler := h.listVersions()

		result, _, err := handler(ctx, nil, listVersionsInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "",
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

func TestGetLatestVersion(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		mockSvc.On("GetLatestVersion", ctx, "https://repo.com", "nginx").
			Return("2.0.0", nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.getLatestVersion()

		result, output, err := handler(ctx, nil, getLatestVersionInput{
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
		handler := h.getLatestVersion()

		result, _, err := handler(ctx, nil, getLatestVersionInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "nonexistent",
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

func TestGetValues(t *testing.T) {
	ctx := context.Background()

	t.Run("success with explicit version", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		mockSvc.On("GetValues", ctx, "https://repo.com", "nginx", "1.0.0").
			Return([]byte("replicaCount: 1\nimage: nginx"), nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.getValues()

		result, output, err := handler(ctx, nil, getValuesInput{
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
		handler := h.getValues()

		result, output, err := handler(ctx, nil, getValuesInput{
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
		handler := h.getValues()

		result, _, err := handler(ctx, nil, getValuesInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "", // Missing
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

func TestGetValuesSchema(t *testing.T) {
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

func TestGetContents(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		mockSvc.On("GetContents", ctx, "https://repo.com", "nginx", "1.0.0", false).
			Return("# Chart.yaml\n...", nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.getContents()

		result, output, err := handler(ctx, nil, getContentsInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "nginx",
			ChartVersion:  "1.0.0",
			Recursive:     false,
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "# Chart.yaml\n...", output.Contents)
	})

	t.Run("recursive flag passed", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		mockSvc.On("GetContents", ctx, "https://repo.com", "nginx", "1.0.0", true).
			Return("# Full contents...", nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.getContents()

		result, output, err := handler(ctx, nil, getContentsInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "nginx",
			ChartVersion:  "1.0.0",
			Recursive:     true,
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "# Full contents...", output.Contents)
		mockSvc.AssertExpectations(t)
	})
}

func TestGetDependencies(t *testing.T) {
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
		handler := h.getDependencies()

		result, output, err := handler(ctx, nil, getDependenciesInput{
			RepositoryURL: "https://repo.com",
			ChartName:     "app",
			ChartVersion:  "1.0.0",
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Len(t, output.Dependencies, 2)
		// Dependencies are JSON-encoded strings
		assert.Contains(t, output.Dependencies[0], "redis")
		assert.Contains(t, output.Dependencies[1], "postgresql")
	})

	t.Run("no dependencies", func(t *testing.T) {
		mockSvc := new(mocks.ChartService)
		mockSvc.On("GetDependencies", ctx, "https://repo.com", "simple", "1.0.0").
			Return([]helm.Dependency{}, nil)

		h := New(mockSvc, zap.NewNop())
		handler := h.getDependencies()

		result, output, err := handler(ctx, nil, getDependenciesInput{
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
