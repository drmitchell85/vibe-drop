package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime"
	"runtime/debug"
)

type RecoveryResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

func getRequestID(r *http.Request) string {
	if id := r.Context().Value("request_id"); id != nil {
		if requestID, ok := id.(string); ok {
			return requestID
		}
	}
	return ""
}

func Recovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					requestID := getRequestID(r)
					
					// Get stack trace
					stack := debug.Stack()
					
					// Log the panic with context
					log.Printf("[%s] PANIC RECOVERED: %v\nStack trace:\n%s", 
						requestID, err, string(stack))
					
					// Prepare error response
					response := RecoveryResponse{
						Error:     "internal_server_error",
						Message:   "An unexpected error occurred. Please try again later.",
						RequestID: requestID,
					}
					
					// Set headers for error response
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					
					// Send JSON error response
					if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
						// If JSON encoding fails, fall back to plain text
						log.Printf("[%s] Failed to encode recovery response: %v", requestID, encodeErr)
						w.Header().Set("Content-Type", "text/plain")
						w.Write([]byte("Internal Server Error"))
					}
					
					// Log memory stats if panic might be memory-related
					var m runtime.MemStats
					runtime.ReadMemStats(&m)
					log.Printf("[%s] Memory stats after panic - Alloc: %d KB, TotalAlloc: %d KB, Sys: %d KB",
						requestID, m.Alloc/1024, m.TotalAlloc/1024, m.Sys/1024)
				}
			}()
			
			next.ServeHTTP(w, r)
		})
	}
}