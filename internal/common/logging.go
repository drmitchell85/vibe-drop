package common

import (
	"fmt"
	"log"
	"runtime"
	"time"
)

// LogLevel represents different log levels
type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
)

// StructuredLogger provides structured logging with request context
type StructuredLogger struct {
	requestID string
	userID    string
	service   string
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(requestID, userID, service string) *StructuredLogger {
	return &StructuredLogger{
		requestID: requestID,
		userID:    userID,
		service:   service,
	}
}

// logMessage formats and logs a structured message
func (sl *StructuredLogger) logMessage(level LogLevel, message string, fields map[string]interface{}) {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	
	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	caller := "unknown"
	if ok {
		// Extract just the filename, not the full path
		for i := len(file) - 1; i >= 0; i-- {
			if file[i] == '/' {
				caller = fmt.Sprintf("%s:%d", file[i+1:], line)
				break
			}
		}
	}
	
	// Build log entry
	logEntry := fmt.Sprintf("[%s] %s [%s] [%s]", timestamp, level, sl.service, caller)
	
	if sl.requestID != "" {
		logEntry += fmt.Sprintf(" [req:%s]", sl.requestID)
	}
	
	if sl.userID != "" {
		logEntry += fmt.Sprintf(" [user:%s]", sl.userID)
	}
	
	logEntry += fmt.Sprintf(" %s", message)
	
	// Add additional fields
	if len(fields) > 0 {
		logEntry += " |"
		for key, value := range fields {
			logEntry += fmt.Sprintf(" %s=%v", key, value)
		}
	}
	
	log.Println(logEntry)
}

// Debug logs a debug message
func (sl *StructuredLogger) Debug(message string, fields ...map[string]interface{}) {
	f := make(map[string]interface{})
	if len(fields) > 0 {
		f = fields[0]
	}
	sl.logMessage(LogLevelDebug, message, f)
}

// Info logs an info message
func (sl *StructuredLogger) Info(message string, fields ...map[string]interface{}) {
	f := make(map[string]interface{})
	if len(fields) > 0 {
		f = fields[0]
	}
	sl.logMessage(LogLevelInfo, message, f)
}

// Warn logs a warning message
func (sl *StructuredLogger) Warn(message string, fields ...map[string]interface{}) {
	f := make(map[string]interface{})
	if len(fields) > 0 {
		f = fields[0]
	}
	sl.logMessage(LogLevelWarn, message, f)
}

// Error logs an error message
func (sl *StructuredLogger) Error(message string, fields ...map[string]interface{}) {
	f := make(map[string]interface{})
	if len(fields) > 0 {
		f = fields[0]
	}
	sl.logMessage(LogLevelError, message, f)
}

// LogValidationError logs validation errors with structured format
func (sl *StructuredLogger) LogValidationError(endpoint string, errors []ValidationError) {
	fields := map[string]interface{}{
		"endpoint":     endpoint,
		"error_count":  len(errors),
		"error_fields": make([]string, len(errors)),
	}
	
	errorFields := make([]string, len(errors))
	for i, err := range errors {
		errorFields[i] = err.Field
	}
	fields["error_fields"] = errorFields
	
	sl.Warn("Validation failed", fields)
}

// LogAuthenticationAttempt logs authentication attempts
func (sl *StructuredLogger) LogAuthenticationAttempt(email string, success bool, reason string) {
	fields := map[string]interface{}{
		"email":   email,
		"success": success,
	}
	
	if reason != "" {
		fields["reason"] = reason
	}
	
	if success {
		sl.Info("Authentication successful", fields)
	} else {
		sl.Warn("Authentication failed", fields)
	}
}

// LogFileOperation logs file operations
func (sl *StructuredLogger) LogFileOperation(operation, filename, fileID string, size *int64) {
	fields := map[string]interface{}{
		"operation": operation,
		"filename":  filename,
		"file_id":   fileID,
	}
	
	if size != nil {
		fields["file_size"] = *size
	}
	
	sl.Info("File operation", fields)
}

// LogError logs errors with additional context
func (sl *StructuredLogger) LogError(operation string, err error, fields ...map[string]interface{}) {
	f := map[string]interface{}{
		"operation": operation,
		"error":     err.Error(),
	}
	
	if len(fields) > 0 {
		for key, value := range fields[0] {
			f[key] = value
		}
	}
	
	sl.Error("Operation failed", f)
}