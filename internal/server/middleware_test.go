package server

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestRateLimiter_GetClientIP(t *testing.T) {
	trustedNetwork := net.IPNet{
		IP:   net.ParseIP("10.0.0.0"),
		Mask: net.CIDRMask(8, 32),
	}

	tests := []struct {
		name           string
		remoteAddr     string
		xff            string
		trustedProxies []net.IPNet
		want           string
	}{
		{
			name:           "direct connection, no proxies",
			remoteAddr:     "192.168.1.100:12345",
			xff:            "",
			trustedProxies: nil,
			want:           "192.168.1.100",
		},
		{
			name:           "xff header ignored without trusted proxies",
			remoteAddr:     "192.168.1.100:12345",
			xff:            "8.8.8.8",
			trustedProxies: nil,
			want:           "192.168.1.100",
		},
		{
			name:           "xff header used when from trusted proxy",
			remoteAddr:     "10.0.0.1:12345",
			xff:            "8.8.8.8",
			trustedProxies: []net.IPNet{trustedNetwork},
			want:           "8.8.8.8",
		},
		{
			name:           "xff chain - rightmost non-trusted IP",
			remoteAddr:     "10.0.0.1:12345",
			xff:            "1.2.3.4, 10.0.0.5, 10.0.0.6",
			trustedProxies: []net.IPNet{trustedNetwork},
			want:           "1.2.3.4",
		},
		{
			name:           "xff chain - all trusted returns first",
			remoteAddr:     "10.0.0.1:12345",
			xff:            "10.0.0.2, 10.0.0.3",
			trustedProxies: []net.IPNet{trustedNetwork},
			want:           "10.0.0.2",
		},
		{
			name:           "untrusted remote addr ignores xff",
			remoteAddr:     "192.168.1.100:12345",
			xff:            "8.8.8.8",
			trustedProxies: []net.IPNet{trustedNetwork},
			want:           "192.168.1.100",
		},
		{
			name:           "empty xff from trusted proxy",
			remoteAddr:     "10.0.0.1:12345",
			xff:            "",
			trustedProxies: []net.IPNet{trustedNetwork},
			want:           "10.0.0.1",
		},
		{
			name:           "remoteAddr without port",
			remoteAddr:     "192.168.1.100",
			xff:            "",
			trustedProxies: nil,
			want:           "192.168.1.100",
		},
		{
			name:           "invalid IP in xff is skipped",
			remoteAddr:     "10.0.0.1:12345",
			xff:            "invalid, 8.8.8.8",
			trustedProxies: []net.IPNet{trustedNetwork},
			want:           "8.8.8.8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := NewRateLimiter(RateLimitConfig{
				RPS:            10,
				Burst:          20,
				TrustedProxies: tt.trustedProxies,
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}

			got := rl.getClientIP(req)
			if got != tt.want {
				t.Errorf("getClientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRateLimiter_Middleware(t *testing.T) {
	t.Run("allows requests under limit", func(t *testing.T) {
		rl := NewRateLimiter(RateLimitConfig{
			RPS:   100, // High limit
			Burst: 100,
		})

		handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		for i := 0; i < 10; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.100:12345"
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("request %d: got status %d, want %d", i, rr.Code, http.StatusOK)
			}
		}
	})

	t.Run("rejects requests over limit", func(t *testing.T) {
		rl := NewRateLimiter(RateLimitConfig{
			RPS:   1, // Very low limit
			Burst: 1,
		})

		handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.100:12345"

		// First request should succeed
		rr1 := httptest.NewRecorder()
		handler.ServeHTTP(rr1, req)
		if rr1.Code != http.StatusOK {
			t.Fatalf("first request: got status %d, want %d", rr1.Code, http.StatusOK)
		}

		// Second request should be rate limited
		rr2 := httptest.NewRecorder()
		handler.ServeHTTP(rr2, req)
		if rr2.Code != http.StatusTooManyRequests {
			t.Errorf("second request: got status %d, want %d", rr2.Code, http.StatusTooManyRequests)
		}

		// Check response body
		var resp map[string]string
		if err := json.NewDecoder(rr2.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp["error"] != "rate limit exceeded" {
			t.Errorf("got error %q, want %q", resp["error"], "rate limit exceeded")
		}

		// Check Retry-After header
		if rr2.Header().Get("Retry-After") != "1" {
			t.Errorf("got Retry-After %q, want %q", rr2.Header().Get("Retry-After"), "1")
		}
	})

	t.Run("different IPs have separate limits", func(t *testing.T) {
		rl := NewRateLimiter(RateLimitConfig{
			RPS:   1,
			Burst: 1,
		})

		handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// First IP exhausts its limit
		req1 := httptest.NewRequest("GET", "/test", nil)
		req1.RemoteAddr = "192.168.1.100:12345"
		rr1 := httptest.NewRecorder()
		handler.ServeHTTP(rr1, req1)

		// Second IP should still work
		req2 := httptest.NewRequest("GET", "/test", nil)
		req2.RemoteAddr = "192.168.1.101:12345"
		rr2 := httptest.NewRecorder()
		handler.ServeHTTP(rr2, req2)

		if rr2.Code != http.StatusOK {
			t.Errorf("different IP: got status %d, want %d", rr2.Code, http.StatusOK)
		}
	})
}

func TestRateLimiter_DefaultConfig(t *testing.T) {
	t.Run("defaults RPS when zero", func(t *testing.T) {
		rl := NewRateLimiter(RateLimitConfig{
			RPS:   0,
			Burst: 20,
		})
		if rl.cfg.RPS != 10 {
			t.Errorf("got RPS %v, want 10", rl.cfg.RPS)
		}
	})

	t.Run("defaults Burst when zero", func(t *testing.T) {
		rl := NewRateLimiter(RateLimitConfig{
			RPS:   10,
			Burst: 0,
		})
		if rl.cfg.Burst != 20 {
			t.Errorf("got Burst %v, want 20", rl.cfg.Burst)
		}
	})
}

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
