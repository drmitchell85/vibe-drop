package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func generateRequestID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}
	
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}
	
	return r.RemoteAddr
}

func RequestLogging() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := generateRequestID()
			
			// Add request ID to context and response headers
			ctx := context.WithValue(r.Context(), "request_id", requestID)
			r = r.WithContext(ctx)
			w.Header().Set("X-Request-ID", requestID)
			
			// Wrap response writer to capture status code and size
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     200, // default status code
			}
			
			// Log incoming request
			log.Printf("[%s] %s %s %s - Started", requestID, getClientIP(r), r.Method, r.URL.Path)
			
			// Process request
			next.ServeHTTP(wrapped, r)
			
			// Log completed request
			duration := time.Since(start)
			log.Printf("[%s] %s %s %s - Completed %d %d bytes in %v", 
				requestID, 
				getClientIP(r), 
				r.Method, 
				r.URL.Path, 
				wrapped.statusCode, 
				wrapped.size, 
				duration,
			)
		})
	}
}