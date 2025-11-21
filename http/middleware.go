package http

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"
)

// securityHeaders adds security headers to responses
func securityHeaders(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		// Remove server header
		w.Header().Set("Server", "")

		next(w, r)
	}
}

// requestID adds a unique request ID to each request
func requestID(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Generate or use existing request ID
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			bytes := make([]byte, 8)
			rand.Read(bytes)
			reqID = hex.EncodeToString(bytes)
		}

		// Add to response header
		w.Header().Set("X-Request-ID", reqID)

		// Add to request header for logging
		r.Header.Set("X-Request-ID", reqID)

		next(w, r)
	}
}

// loggingMiddleware logs request details
func loggingMiddleware(log *slog.Logger) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create response writer wrapper to capture status code
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Process request
			next(rw, r)

			// Log request
			duration := time.Since(start)
			reqID := r.Header.Get("X-Request-ID")

			log.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.statusCode,
				"duration_ms", duration.Milliseconds(),
				"request_id", reqID,
				"remote_addr", r.RemoteAddr,
			)
		}
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
