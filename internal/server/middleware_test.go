package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestLoggingMiddleware(t *testing.T) {
	// Use a no-op logger for testing
	logger := zap.NewNop()

	handler := LoggingMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("test"))
	}))

	req := httptest.NewRequest("GET", "/test-path", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Main verification is that it doesn't panic and passes through correctly
	if rr.Code != http.StatusCreated {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusCreated)
	}
	if rr.Body.String() != "test" {
		t.Errorf("got body %q, want %q", rr.Body.String(), "test")
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	logger := zap.NewNop()

	t.Run("recovers from panic", func(t *testing.T) {
		handler := RecoveryMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		// Should not panic
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("got status %d, want %d", rr.Code, http.StatusInternalServerError)
		}

		var resp map[string]string
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp["error"] != "internal server error" {
			t.Errorf("got error %q, want %q", resp["error"], "internal server error")
		}
	})

	t.Run("passes through normal requests", func(t *testing.T) {
		handler := RecoveryMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", rr.Code, http.StatusOK)
		}
		if rr.Body.String() != "ok" {
			t.Errorf("got body %q, want %q", rr.Body.String(), "ok")
		}
	})
}

func TestResponseWriter(t *testing.T) {
	t.Run("captures status code", func(t *testing.T) {
		underlying := httptest.NewRecorder()
		rw := &responseWriter{ResponseWriter: underlying, statusCode: http.StatusOK}

		rw.WriteHeader(http.StatusNotFound)

		if rw.statusCode != http.StatusNotFound {
			t.Errorf("got status %d, want %d", rw.statusCode, http.StatusNotFound)
		}
		if underlying.Code != http.StatusNotFound {
			t.Errorf("underlying got status %d, want %d", underlying.Code, http.StatusNotFound)
		}
	})

	t.Run("writes through to underlying writer", func(t *testing.T) {
		underlying := httptest.NewRecorder()
		rw := &responseWriter{ResponseWriter: underlying, statusCode: http.StatusOK}

		_, err := rw.Write([]byte("test body"))
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}

		if underlying.Body.String() != "test body" {
			t.Errorf("got body %q, want %q", underlying.Body.String(), "test body")
		}
	})
}
