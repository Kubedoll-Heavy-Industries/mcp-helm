package mcputil

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Kubedoll-Heavy-Industries/mcp-helm/internal/helm"
)

func TestTextError(t *testing.T) {
	t.Run("creates error result with message", func(t *testing.T) {
		result := TextError("something went wrong")

		require.NotNil(t, result)
		assert.True(t, result.IsError)
		require.Len(t, result.Content, 1)
	})
}

func TestHandleError(t *testing.T) {
	t.Run("nil error returns nil", func(t *testing.T) {
		result := HandleError(nil)

		assert.Nil(t, result)
	})

	t.Run("ChartNotFoundError", func(t *testing.T) {
		err := &helm.ChartNotFoundError{
			Repository: "https://repo.com",
			Chart:      "nginx",
			Version:    "1.0.0",
		}

		result := HandleError(err)

		require.NotNil(t, result)
		assert.True(t, result.IsError)
		require.Len(t, result.Content, 1)
	})

	t.Run("RepositoryError", func(t *testing.T) {
		err := &helm.RepositoryError{
			URL:     "https://bad.repo",
			Op:      "fetch",
			Message: "connection failed",
		}

		result := HandleError(err)

		require.NotNil(t, result)
		assert.True(t, result.IsError)
		require.Len(t, result.Content, 1)
	})

	t.Run("URLValidationError", func(t *testing.T) {
		err := &helm.URLValidationError{
			URL:    "ftp://invalid.url",
			Reason: "scheme not allowed",
		}

		result := HandleError(err)

		require.NotNil(t, result)
		assert.True(t, result.IsError)
		require.Len(t, result.Content, 1)
	})

	t.Run("OutputTooLargeError", func(t *testing.T) {
		err := &helm.OutputTooLargeError{
			Size:  5000000,
			Limit: 2000000,
		}

		result := HandleError(err)

		require.NotNil(t, result)
		assert.True(t, result.IsError)
		require.Len(t, result.Content, 1)
	})

	t.Run("generic error", func(t *testing.T) {
		err := errors.New("something went wrong")

		result := HandleError(err)

		require.NotNil(t, result)
		assert.True(t, result.IsError)
		require.Len(t, result.Content, 1)
	})
}

func TestOperationError(t *testing.T) {
	baseErr := errors.New("network timeout")

	t.Run("formats with chart and version", func(t *testing.T) {
		err := OpError("get_values", "https://repo.com", "nginx", "1.0.0", baseErr)

		assert.Equal(t, "get_values: nginx@1.0.0: network timeout", err.Error())
	})

	t.Run("formats with chart only", func(t *testing.T) {
		err := OpError("get_values", "https://repo.com", "nginx", "", baseErr)

		assert.Equal(t, "get_values: nginx: network timeout", err.Error())
	})

	t.Run("formats with repo only", func(t *testing.T) {
		err := OpError("list_charts", "https://repo.com", "", "", baseErr)

		assert.Equal(t, "list_charts: https://repo.com: network timeout", err.Error())
	})

	t.Run("formats with op only", func(t *testing.T) {
		err := OpError("init", "", "", "", baseErr)

		assert.Equal(t, "init: network timeout", err.Error())
	})

	t.Run("unwraps to underlying error", func(t *testing.T) {
		err := OpError("get_values", "https://repo.com", "nginx", "1.0.0", baseErr)

		assert.True(t, errors.Is(err, baseErr))
	})
}

func TestHandleOpError(t *testing.T) {
	t.Run("nil error returns nil", func(t *testing.T) {
		result := HandleOpError("get_values", "repo", "chart", "version", nil)

		assert.Nil(t, result)
	})

	t.Run("wraps and handles error", func(t *testing.T) {
		err := errors.New("something failed")

		result := HandleOpError("get_values", "repo", "nginx", "1.0.0", err)

		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}
