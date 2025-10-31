package common

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// ErrorCode represents specific error types for better client handling
type ErrorCode string

const (
	// Client errors (4xx)
	ErrorCodeBadRequest     ErrorCode = "BAD_REQUEST"
	ErrorCodeUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrorCodeForbidden      ErrorCode = "FORBIDDEN"
	ErrorCodeNotFound       ErrorCode = "NOT_FOUND"
	ErrorCodeConflict       ErrorCode = "CONFLICT"
	ErrorCodeValidation     ErrorCode = "VALIDATION_ERROR"
	ErrorCodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"
	
	// Server errors (5xx)
	ErrorCodeInternalServer ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrorCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrorCodeDatabaseError  ErrorCode = "DATABASE_ERROR"
	ErrorCodeS3Error        ErrorCode = "STORAGE_ERROR"
)

// ErrorResponse represents the standard error response format
type ErrorResponse struct {
	Success   bool      `json:"success"`
	Error     ErrorInfo `json:"error"`
	RequestID string    `json:"request_id"`
	Timestamp string    `json:"timestamp"`
}

// ErrorInfo contains detailed error information
type ErrorInfo struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
}

// SuccessCode represents specific success types for better client handling
type SuccessCode string

const (
	SuccessCodeCreated   SuccessCode = "CREATED"
	SuccessCodeOK        SuccessCode = "OK"
	SuccessCodeAccepted  SuccessCode = "ACCEPTED"
	SuccessCodeNoContent SuccessCode = "NO_CONTENT"
)

// SuccessResponse represents the standard success response format
type SuccessResponse struct {
	Success   bool        `json:"success"`
	Code      SuccessCode `json:"code"`
	Data      interface{} `json:"data"`
	RequestID string      `json:"request_id"`
	Timestamp string      `json:"timestamp"`
}

// WriteErrorResponse sends a standardized error response
func WriteErrorResponse(w http.ResponseWriter, statusCode int, errorCode ErrorCode, message, details string) {
	requestID := generateRequestID()
	
	errorResponse := ErrorResponse{
		Success: false,
		Error: ErrorInfo{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
		RequestID: requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	
	// Log the error for debugging
	log.Printf("[%s] Error %d: %s - %s (Details: %s)", 
		requestID, statusCode, errorCode, message, details)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		log.Printf("[%s] Failed to encode error response: %v", requestID, err)
		// Fallback to plain text if JSON encoding fails
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Internal Server Error"))
	}
}

// WriteSuccessResponse sends a standardized success response
func WriteSuccessResponse(w http.ResponseWriter, statusCode int, successCode SuccessCode, data interface{}) {
	requestID := generateRequestID()
	
	successResponse := SuccessResponse{
		Success:   true,
		Code:      successCode,
		Data:      data,
		RequestID: requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(successResponse); err != nil {
		log.Printf("[%s] Failed to encode success response: %v", requestID, err)
		WriteErrorResponse(w, http.StatusInternalServerError, ErrorCodeInternalServer, 
			"Failed to encode response", "JSON encoding error")
	}
}

// Convenience functions for common success responses

// WriteCreatedResponse sends a 201 Created success response
func WriteCreatedResponse(w http.ResponseWriter, data interface{}) {
	WriteSuccessResponse(w, http.StatusCreated, SuccessCodeCreated, data)
}

// WriteOKResponse sends a 200 OK success response
func WriteOKResponse(w http.ResponseWriter, data interface{}) {
	WriteSuccessResponse(w, http.StatusOK, SuccessCodeOK, data)
}

// WriteAcceptedResponse sends a 202 Accepted success response
func WriteAcceptedResponse(w http.ResponseWriter, data interface{}) {
	WriteSuccessResponse(w, http.StatusAccepted, SuccessCodeAccepted, data)
}

// WriteNoContentResponse sends a 204 No Content success response
func WriteNoContentResponse(w http.ResponseWriter) {
	WriteSuccessResponse(w, http.StatusNoContent, SuccessCodeNoContent, nil)
}

// Convenience functions for common error types

// WriteBadRequestError sends a 400 Bad Request error
func WriteBadRequestError(w http.ResponseWriter, message, details string) {
	WriteErrorResponse(w, http.StatusBadRequest, ErrorCodeBadRequest, message, details)
}

// WriteUnauthorizedError sends a 401 Unauthorized error
func WriteUnauthorizedError(w http.ResponseWriter, message, details string) {
	WriteErrorResponse(w, http.StatusUnauthorized, ErrorCodeUnauthorized, message, details)
}

// WriteForbiddenError sends a 403 Forbidden error
func WriteForbiddenError(w http.ResponseWriter, message, details string) {
	WriteErrorResponse(w, http.StatusForbidden, ErrorCodeForbidden, message, details)
}

// WriteNotFoundError sends a 404 Not Found error
func WriteNotFoundError(w http.ResponseWriter, message, details string) {
	WriteErrorResponse(w, http.StatusNotFound, ErrorCodeNotFound, message, details)
}

// WriteConflictError sends a 409 Conflict error
func WriteConflictError(w http.ResponseWriter, message, details string) {
	WriteErrorResponse(w, http.StatusConflict, ErrorCodeConflict, message, details)
}

// WriteValidationError sends a 400 Bad Request error for validation failures
func WriteValidationError(w http.ResponseWriter, message, details string) {
	WriteErrorResponse(w, http.StatusBadRequest, ErrorCodeValidation, message, details)
}

// WriteInternalServerError sends a 500 Internal Server Error
func WriteInternalServerError(w http.ResponseWriter, message, details string) {
	WriteErrorResponse(w, http.StatusInternalServerError, ErrorCodeInternalServer, message, details)
}

// WriteDatabaseError sends a 500 error for database-related issues
func WriteDatabaseError(w http.ResponseWriter, message, details string) {
	WriteErrorResponse(w, http.StatusInternalServerError, ErrorCodeDatabaseError, message, details)
}

// WriteS3Error sends a 500 error for S3/storage-related issues
func WriteS3Error(w http.ResponseWriter, message, details string) {
	WriteErrorResponse(w, http.StatusInternalServerError, ErrorCodeS3Error, message, details)
}

// generateRequestID creates a unique request ID for tracking
func generateRequestID() string {
	return "req-" + uuid.New().String()[:8]
}

// ValidateJSONRequest validates that request body contains valid JSON
func ValidateJSONRequest(r *http.Request, target interface{}) error {
	if r.Header.Get("Content-Type") != "application/json" {
		return &ValidationError{
			Field:   "Content-Type",
			Message: "Content-Type must be application/json",
		}
	}
	
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Strict JSON parsing
	
	if err := decoder.Decode(target); err != nil {
		return &ValidationError{
			Field:   "request_body",
			Message: "Invalid JSON format: " + err.Error(),
		}
	}
	
	return nil
}

// ValidationError represents a validation error with field-specific information
type ValidationError struct {
	Field   string    `json:"field"`
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) (*ValidationError, bool) {
	if validationErr, ok := err.(*ValidationError); ok {
		return validationErr, true
	}
	return nil, false
}