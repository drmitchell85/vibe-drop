package common

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// Request context keys
type ContextKey string

const (
	RequestIDKey ContextKey = "request_id"
	UserIDKey    ContextKey = "user_id"
	UsernameKey  ContextKey = "username"
)

// RequestValidationMiddleware provides common request validation
func RequestValidationMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add request ID if not present
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = "req-" + uuid.New().String()[:8]
			}
			
			// Add request ID to context
			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
			r = r.WithContext(ctx)
			
			// Add request ID to response header
			w.Header().Set("X-Request-ID", requestID)
			
			// Validate Content-Type for POST, PUT, PATCH requests with body
			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
				if r.ContentLength > 0 {
					contentType := r.Header.Get("Content-Type")
					if contentType == "" {
						WriteErrorResponse(w, http.StatusBadRequest, ErrorCodeBadRequest, 
							"Content-Type header is required", "Content-Type must be specified for requests with body")
						return
					}
					
					// For JSON endpoints, ensure Content-Type is application/json
					if strings.Contains(r.URL.Path, "/auth/") || strings.Contains(r.URL.Path, "/files/") {
						if !strings.HasPrefix(contentType, "application/json") {
							WriteErrorResponse(w, http.StatusBadRequest, ErrorCodeBadRequest, 
								"Invalid Content-Type", "Content-Type must be application/json")
							return
						}
					}
				}
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// JSONValidationMiddleware validates JSON request bodies
func JSONValidationMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only validate JSON for POST, PUT, PATCH requests
			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
				if r.ContentLength > 0 {
					contentType := r.Header.Get("Content-Type")
					if strings.HasPrefix(contentType, "application/json") {
						// Read body to validate JSON
						body, err := io.ReadAll(r.Body)
						if err != nil {
							WriteErrorResponse(w, http.StatusBadRequest, ErrorCodeBadRequest, 
								"Failed to read request body", err.Error())
							return
						}
						r.Body.Close()
						
						// Validate JSON
						var js json.RawMessage
						if err := json.Unmarshal(body, &js); err != nil {
							WriteValidationError(w, "Invalid JSON format", err.Error())
							return
						}
						
						// Restore body for next handler
						r.Body = io.NopCloser(strings.NewReader(string(body)))
					}
				}
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// FileSizeValidationMiddleware validates Content-Length for file uploads
func FileSizeValidationMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check file upload endpoints
			if strings.Contains(r.URL.Path, "/files/upload") {
				if r.ContentLength > MaxFileSize {
					WriteErrorResponse(w, http.StatusRequestEntityTooLarge, ErrorCodeFileTooLarge, 
						"Request entity too large", 
						fmt.Sprintf("Maximum file size is %d bytes (%.1f GB)", MaxFileSize, float64(MaxFileSize)/(1024*1024*1024)))
					return
				}
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			
			// Don't cache sensitive endpoints
			if strings.Contains(r.URL.Path, "/auth/") || strings.Contains(r.URL.Path, "/files/") {
				w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
				w.Header().Set("Pragma", "no-cache")
				w.Header().Set("Expires", "0")
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// GetRequestIDFromContext extracts request ID from context
func GetRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}