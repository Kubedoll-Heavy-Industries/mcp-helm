// Package helm provides a client for interacting with Helm chart repositories.
package helm

import (
	"errors"
	"fmt"
)

// Sentinel errors for common conditions.
var (
	ErrChartNotFound = errors.New("chart not found")
	ErrRepoNotFound  = errors.New("repository not found")
	ErrInvalidURL    = errors.New("invalid URL")
)

// ChartNotFoundError indicates that a specific chart version was not found.
type ChartNotFoundError struct {
	Repository string
	Chart      string
	Version    string
}

func (e *ChartNotFoundError) Error() string {
	if e.Version == "" {
		return fmt.Sprintf("chart %q not found in repository %q", e.Chart, e.Repository)
	}
	return fmt.Sprintf("chart %q version %q not found in repository %q", e.Chart, e.Version, e.Repository)
}

// Is implements errors.Is for ChartNotFoundError.
func (e *ChartNotFoundError) Is(target error) bool {
	return target == ErrChartNotFound
}

// RepositoryError indicates a problem with a Helm repository.
type RepositoryError struct {
	URL     string
	Op      string // Operation that failed: "fetch", "parse", "resolve"
	Message string
	Err     error
}

func (e *RepositoryError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("repository %q: %s: %s: %v", e.URL, e.Op, e.Message, e.Err)
	}
	return fmt.Sprintf("repository %q: %s: %s", e.URL, e.Op, e.Message)
}

func (e *RepositoryError) Unwrap() error {
	return e.Err
}

// Is implements errors.Is for RepositoryError.
func (e *RepositoryError) Is(target error) bool {
	return target == ErrRepoNotFound
}

// ValidationError indicates invalid input.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Is implements errors.Is for ValidationError.
func (e *ValidationError) Is(target error) bool {
	return target == ErrInvalidURL
}

// URLValidationError indicates a URL failed security validation.
type URLValidationError struct {
	URL    string
	Reason string
}

func (e *URLValidationError) Error() string {
	return fmt.Sprintf("URL %q: %s", e.URL, e.Reason)
}

// Is implements errors.Is for URLValidationError.
func (e *URLValidationError) Is(target error) bool {
	return target == ErrInvalidURL
}

// OutputTooLargeError indicates that output exceeded the configured limit.
type OutputTooLargeError struct {
	Size  int
	Limit int
}

func (e *OutputTooLargeError) Error() string {
	return fmt.Sprintf("output size %d exceeds limit %d", e.Size, e.Limit)
}

// ChartTooLargeError indicates that a downloaded chart file exceeds the size limit.
type ChartTooLargeError struct {
	Size  int64
	Limit int64
}

func (e *ChartTooLargeError) Error() string {
	return fmt.Sprintf("chart file size %d bytes exceeds limit %d bytes", e.Size, e.Limit)
}

// IsChartTooLarge returns true if err wraps a ChartTooLargeError.
func IsChartTooLarge(err error) bool {
	var e *ChartTooLargeError
	return errors.As(err, &e)
}

// IsChartNotFound returns true if err wraps a ChartNotFoundError.
func IsChartNotFound(err error) bool {
	var e *ChartNotFoundError
	return errors.As(err, &e)
}

// IsRepositoryError returns true if err wraps a RepositoryError.
func IsRepositoryError(err error) bool {
	var e *RepositoryError
	return errors.As(err, &e)
}

// IsValidationError returns true if err wraps a ValidationError.
func IsValidationError(err error) bool {
	var e *ValidationError
	return errors.As(err, &e)
}

// IsURLValidationError returns true if err wraps a URLValidationError.
func IsURLValidationError(err error) bool {
	var e *URLValidationError
	return errors.As(err, &e)
}

// IsOutputTooLarge returns true if err wraps an OutputTooLargeError.
func IsOutputTooLarge(err error) bool {
	var e *OutputTooLargeError
	return errors.As(err, &e)
}
