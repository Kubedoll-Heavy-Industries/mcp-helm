// Package mocks provides test doubles for the helm package.
package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/Kubedoll-Heavy-Industries/mcp-helm/internal/helm"
)

// ChartService is a mock implementation of helm.ChartService.
type ChartService struct {
	mock.Mock
}

// Ensure ChartService implements helm.ChartService.
var _ helm.ChartService = (*ChartService)(nil)

// ListCharts mocks the ListCharts method.
func (m *ChartService) ListCharts(ctx context.Context, repoURL string) ([]string, error) {
	args := m.Called(ctx, repoURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// ListVersions mocks the ListVersions method.
func (m *ChartService) ListVersions(ctx context.Context, repoURL, chart string) ([]helm.ChartVersion, error) {
	args := m.Called(ctx, repoURL, chart)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]helm.ChartVersion), args.Error(1)
}

// GetLatestVersion mocks the GetLatestVersion method.
func (m *ChartService) GetLatestVersion(ctx context.Context, repoURL, chart string) (string, error) {
	args := m.Called(ctx, repoURL, chart)
	return args.String(0), args.Error(1)
}

// GetValues mocks the GetValues method.
func (m *ChartService) GetValues(ctx context.Context, repoURL, chart, version string) ([]byte, error) {
	args := m.Called(ctx, repoURL, chart, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// GetValuesSchema mocks the GetValuesSchema method.
func (m *ChartService) GetValuesSchema(ctx context.Context, repoURL, chart, version string) ([]byte, bool, error) {
	args := m.Called(ctx, repoURL, chart, version)
	if args.Get(0) == nil {
		return nil, args.Bool(1), args.Error(2)
	}
	return args.Get(0).([]byte), args.Bool(1), args.Error(2)
}

// GetNotes mocks the GetNotes method.
func (m *ChartService) GetNotes(ctx context.Context, repoURL, chart, version string) ([]byte, bool, error) {
	args := m.Called(ctx, repoURL, chart, version)
	if args.Get(0) == nil {
		return nil, args.Bool(1), args.Error(2)
	}
	return args.Get(0).([]byte), args.Bool(1), args.Error(2)
}

// GetDependencies mocks the GetDependencies method.
func (m *ChartService) GetDependencies(ctx context.Context, repoURL, chart, version string) ([]helm.Dependency, error) {
	args := m.Called(ctx, repoURL, chart, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]helm.Dependency), args.Error(1)
}
