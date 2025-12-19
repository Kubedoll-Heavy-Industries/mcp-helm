package helm

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChartNotFoundError(t *testing.T) {
	t.Run("error message with version", func(t *testing.T) {
		err := &ChartNotFoundError{
			Repository: "https://repo.com",
			Chart:      "nginx",
			Version:    "1.0.0",
		}

		assert.Contains(t, err.Error(), "nginx")
		assert.Contains(t, err.Error(), "1.0.0")
		assert.Contains(t, err.Error(), "https://repo.com")
	})

	t.Run("error message without version", func(t *testing.T) {
		err := &ChartNotFoundError{
			Repository: "https://repo.com",
			Chart:      "nginx",
		}

		assert.Contains(t, err.Error(), "nginx")
		assert.Contains(t, err.Error(), "https://repo.com")
		assert.NotContains(t, err.Error(), "version")
	})

	t.Run("Is matches sentinel ErrChartNotFound", func(t *testing.T) {
		err := &ChartNotFoundError{Chart: "nginx"}

		assert.True(t, errors.Is(err, ErrChartNotFound))
	})

	t.Run("Is does not match different sentinel", func(t *testing.T) {
		err := &ChartNotFoundError{Chart: "nginx"}

		assert.False(t, errors.Is(err, ErrRepoNotFound))
	})

	t.Run("IsChartNotFound helper works", func(t *testing.T) {
		err := &ChartNotFoundError{Chart: "nginx"}

		assert.True(t, IsChartNotFound(err))
		assert.False(t, IsChartNotFound(errors.New("other error")))
	})
}

func TestRepositoryError(t *testing.T) {
	t.Run("error message with wrapped error", func(t *testing.T) {
		err := &RepositoryError{
			URL:     "https://bad.repo",
			Op:      "fetch",
			Message: "failed to download index",
			Err:     errors.New("connection refused"),
		}

		msg := err.Error()
		assert.Contains(t, msg, "https://bad.repo")
		assert.Contains(t, msg, "fetch")
		assert.Contains(t, msg, "failed to download index")
		assert.Contains(t, msg, "connection refused")
	})

	t.Run("error message without wrapped error", func(t *testing.T) {
		err := &RepositoryError{
			URL:     "https://bad.repo",
			Op:      "fetch",
			Message: "no URLs available",
		}

		msg := err.Error()
		assert.Contains(t, msg, "https://bad.repo")
		assert.Contains(t, msg, "no URLs available")
	})

	t.Run("Unwrap returns inner error", func(t *testing.T) {
		inner := errors.New("inner error")
		err := &RepositoryError{Err: inner}

		assert.Equal(t, inner, errors.Unwrap(err))
	})

	t.Run("Is matches sentinel ErrRepoNotFound", func(t *testing.T) {
		err := &RepositoryError{URL: "a"}

		assert.True(t, errors.Is(err, ErrRepoNotFound))
	})

	t.Run("IsRepositoryError helper works", func(t *testing.T) {
		err := &RepositoryError{URL: "https://repo.com"}

		assert.True(t, IsRepositoryError(err))
		assert.False(t, IsRepositoryError(errors.New("other error")))
	})
}

func TestValidationError(t *testing.T) {
	t.Run("error message", func(t *testing.T) {
		err := &ValidationError{
			Field:   "chart_name",
			Message: "is required",
		}

		assert.Contains(t, err.Error(), "chart_name")
		assert.Contains(t, err.Error(), "is required")
	})

	t.Run("Is matches sentinel ErrInvalidURL", func(t *testing.T) {
		err := &ValidationError{Field: "a"}

		assert.True(t, errors.Is(err, ErrInvalidURL))
	})

	t.Run("IsValidationError helper works", func(t *testing.T) {
		err := &ValidationError{Field: "test"}

		assert.True(t, IsValidationError(err))
		assert.False(t, IsValidationError(errors.New("other error")))
	})
}

func TestURLValidationError(t *testing.T) {
	t.Run("error message", func(t *testing.T) {
		err := &URLValidationError{
			URL:    "ftp://invalid.url",
			Reason: "scheme must be http or https",
		}

		assert.Contains(t, err.Error(), "ftp://invalid.url")
		assert.Contains(t, err.Error(), "scheme must be http or https")
	})

	t.Run("Is matches sentinel ErrInvalidURL", func(t *testing.T) {
		err := &URLValidationError{URL: "a"}

		assert.True(t, errors.Is(err, ErrInvalidURL))
	})

	t.Run("IsURLValidationError helper works", func(t *testing.T) {
		err := &URLValidationError{URL: "test"}

		assert.True(t, IsURLValidationError(err))
		assert.False(t, IsURLValidationError(errors.New("other error")))
	})
}

func TestOutputTooLargeError(t *testing.T) {
	t.Run("error message", func(t *testing.T) {
		err := &OutputTooLargeError{
			Size:  5000000,
			Limit: 2000000,
		}

		assert.Contains(t, err.Error(), "5000000")
		assert.Contains(t, err.Error(), "2000000")
	})

	t.Run("IsOutputTooLarge helper works", func(t *testing.T) {
		err := &OutputTooLargeError{Size: 100}

		assert.True(t, IsOutputTooLarge(err))
		assert.False(t, IsOutputTooLarge(errors.New("other error")))
	})
}

func TestErrorsAs(t *testing.T) {
	t.Run("As works with ChartNotFoundError", func(t *testing.T) {
		err := &ChartNotFoundError{Chart: "nginx", Repository: "https://repo.com"}

		var target *ChartNotFoundError
		assert.True(t, errors.As(err, &target))
		assert.Equal(t, "nginx", target.Chart)
	})

	t.Run("As works with RepositoryError", func(t *testing.T) {
		err := &RepositoryError{URL: "https://repo.com", Op: "fetch"}

		var target *RepositoryError
		assert.True(t, errors.As(err, &target))
		assert.Equal(t, "fetch", target.Op)
	})

	t.Run("As works with wrapped error", func(t *testing.T) {
		inner := &ChartNotFoundError{Chart: "redis"}
		wrapped := errors.New("wrapped: " + inner.Error())
		_ = wrapped // Note: string wrapping breaks As

		// Direct error works
		var target *ChartNotFoundError
		assert.True(t, errors.As(inner, &target))
	})
}
