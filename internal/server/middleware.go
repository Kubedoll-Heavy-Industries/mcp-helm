package server

import (
	"encoding/json"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// RateLimitConfig configures the rate limiter.
type RateLimitConfig struct {
	RPS            float64
	Burst          int
	TrustedProxies []net.IPNet
}

// RateLimiter implements per-IP rate limiting using token bucket algorithm.
type RateLimiter struct {
	cfg      RateLimitConfig
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
}

// NewRateLimiter creates a new rate limiter with the given configuration.
func NewRateLimiter(cfg RateLimitConfig) *RateLimiter {
	if cfg.RPS <= 0 {
		cfg.RPS = 10
	}
	if cfg.Burst <= 0 {
		cfg.Burst = 20
	}
	return &RateLimiter{
		cfg:      cfg,
		limiters: make(map[string]*rate.Limiter),
	}
}

// getLimiter returns the rate limiter for the given IP, creating one if necessary.
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(rl.cfg.RPS), rl.cfg.Burst)
		rl.limiters[ip] = limiter
	}
	return limiter
}

// Middleware returns an HTTP middleware that enforces rate limiting.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := rl.getClientIP(r)
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "rate limit exceeded",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getClientIP extracts the client IP from the request.
// It only trusts X-Forwarded-For from configured trusted proxies.
func (rl *RateLimiter) getClientIP(r *http.Request) string {
	// Get the direct connection IP
	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteIP = r.RemoteAddr
	}

	// If no trusted proxies configured, use RemoteAddr directly
	if len(rl.cfg.TrustedProxies) == 0 {
		return remoteIP
	}

	// Check if request is from a trusted proxy
	parsedRemote := net.ParseIP(remoteIP)
	if parsedRemote == nil {
		return remoteIP
	}

	isTrusted := false
	for _, network := range rl.cfg.TrustedProxies {
		if network.Contains(parsedRemote) {
			isTrusted = true
			break
		}
	}

	if !isTrusted {
		return remoteIP
	}

	// Request is from trusted proxy, check X-Forwarded-For
	xff := r.Header.Get("X-Forwarded-For")
	if xff == "" {
		return remoteIP
	}

	// X-Forwarded-For format: client, proxy1, proxy2, ...
	// We want the rightmost non-trusted IP
	ips := strings.Split(xff, ",")
	for i := len(ips) - 1; i >= 0; i-- {
		ip := strings.TrimSpace(ips[i])
		parsedIP := net.ParseIP(ip)
		if parsedIP == nil {
			continue
		}

		// Check if this IP is also a trusted proxy
		trusted := false
		for _, network := range rl.cfg.TrustedProxies {
			if network.Contains(parsedIP) {
				trusted = true
				break
			}
		}

		if !trusted {
			return ip
		}
	}

	// All IPs in chain are trusted, use the first one
	if len(ips) > 0 {
		return strings.TrimSpace(ips[0])
	}

	return remoteIP
}

// LoggingMiddleware logs HTTP requests.
func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			logger.Debug("http request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", wrapped.statusCode),
				zap.Duration("duration", time.Since(start)),
				zap.String("remote", r.RemoteAddr),
			)
		})
	}
}

// RecoveryMiddleware recovers from panics and logs them.
func RecoveryMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("panic recovered",
						zap.Any("error", err),
						zap.String("stack", string(debug.Stack())),
						zap.String("path", r.URL.Path),
					)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					_ = json.NewEncoder(w).Encode(map[string]string{
						"error": "internal server error",
					})
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Flush preserves streaming capabilities if the underlying writer supports it.
func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
