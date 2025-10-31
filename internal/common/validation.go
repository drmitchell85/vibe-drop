package common

import (
	"fmt"
	"mime"
	"path/filepath"
	"regexp"
	"strings"
)

// Validation constants
const (
	// File size limits
	MaxFileSize          = 50 * 1024 * 1024 * 1024 // 50GB
	MaxFilenameLength    = 255
	MultipartThreshold   = 5 * 1024 * 1024 * 1024  // 5GB
	
	// User validation limits
	MinUsernameLength    = 3
	MaxUsernameLength    = 50
	MinPasswordLength    = 8
	MaxPasswordLength    = 128
	
	// Request limits
	MaxChunkSize         = 5 * 1024 * 1024 * 1024 // 5GB per chunk
	MaxMultipartParts    = 10000 // AWS S3 limit
)

// File validation error codes
const (
	ErrorCodeFileTooLarge      ErrorCode = "FILE_TOO_LARGE"
	ErrorCodeInvalidFilename   ErrorCode = "INVALID_FILENAME"
	ErrorCodeInvalidFileType   ErrorCode = "INVALID_FILE_TYPE"
	ErrorCodeFilenameRequired  ErrorCode = "FILENAME_REQUIRED"
	ErrorCodeSizeRequired      ErrorCode = "SIZE_REQUIRED"
	ErrorCodeInvalidSize       ErrorCode = "INVALID_SIZE"
	
	// User validation error codes
	ErrorCodeUsernameRequired  ErrorCode = "USERNAME_REQUIRED"
	ErrorCodeUsernameTooShort  ErrorCode = "USERNAME_TOO_SHORT"
	ErrorCodeUsernameTooLong   ErrorCode = "USERNAME_TOO_LONG"
	ErrorCodeInvalidUsername   ErrorCode = "INVALID_USERNAME"
	ErrorCodeEmailRequired     ErrorCode = "EMAIL_REQUIRED"
	ErrorCodeInvalidEmail      ErrorCode = "INVALID_EMAIL"
	ErrorCodePasswordRequired  ErrorCode = "PASSWORD_REQUIRED"
	ErrorCodePasswordTooShort  ErrorCode = "PASSWORD_TOO_SHORT"
	ErrorCodePasswordTooLong   ErrorCode = "PASSWORD_TOO_LONG"
	ErrorCodePasswordTooWeak   ErrorCode = "PASSWORD_TOO_WEAK"
)

// Allowed file types (MIME types)
var AllowedMimeTypes = map[string]bool{
	// Documents
	"application/pdf":                                        true,
	"application/msword":                                     true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel":                               true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
	"application/vnd.ms-powerpoint":                          true,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
	"text/plain":                                             true,
	"text/csv":                                               true,
	"text/rtf":                                               true,
	
	// Images
	"image/jpeg":                                             true,
	"image/png":                                              true,
	"image/gif":                                              true,
	"image/webp":                                             true,
	"image/svg+xml":                                          true,
	"image/bmp":                                              true,
	"image/tiff":                                             true,
	
	// Audio
	"audio/mpeg":                                             true,
	"audio/wav":                                              true,
	"audio/ogg":                                              true,
	"audio/mp4":                                              true,
	"audio/aac":                                              true,
	"audio/flac":                                             true,
	
	// Video
	"video/mp4":                                              true,
	"video/mpeg":                                             true,
	"video/quicktime":                                        true,
	"video/x-msvideo":                                        true,
	"video/webm":                                             true,
	"video/ogg":                                              true,
	
	// Archives
	"application/zip":                                        true,
	"application/x-tar":                                      true,
	"application/gzip":                                       true,
	"application/x-7z-compressed":                            true,
	"application/x-rar-compressed":                           true,
	
	// Code/Text
	"application/json":                                       true,
	"application/xml":                                        true,
	"text/html":                                              true,
	"text/css":                                               true,
	"text/javascript":                                        true,
	"application/javascript":                                 true,
}

// File validation functions

// ValidateFileUpload validates file upload request parameters
type FileUploadRequest struct {
	Filename string `json:"filename"`
	Size     *int64 `json:"size,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
}

func ValidateFileUpload(req *FileUploadRequest) []ValidationError {
	var errors []ValidationError
	
	// Validate filename
	if filenameErrors := ValidateFilename(req.Filename); len(filenameErrors) > 0 {
		errors = append(errors, filenameErrors...)
	}
	
	// Validate size
	if sizeErrors := ValidateFileSize(req.Size); len(sizeErrors) > 0 {
		errors = append(errors, sizeErrors...)
	}
	
	// Validate MIME type if provided
	if req.MimeType != "" {
		if mimeErrors := ValidateMimeType(req.MimeType, req.Filename); len(mimeErrors) > 0 {
			errors = append(errors, mimeErrors...)
		}
	}
	
	return errors
}

func ValidateFilename(filename string) []ValidationError {
	var errors []ValidationError
	
	if filename == "" {
		errors = append(errors, ValidationError{
			Field:   "filename",
			Code:    ErrorCodeFilenameRequired,
			Message: "Filename is required",
		})
		return errors
	}
	
	if len(filename) > MaxFilenameLength {
		errors = append(errors, ValidationError{
			Field:   "filename",
			Code:    ErrorCodeInvalidFilename,
			Message: fmt.Sprintf("Filename must be less than %d characters", MaxFilenameLength),
		})
	}
	
	// Check for invalid characters
	invalidChars := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	if invalidChars.MatchString(filename) {
		errors = append(errors, ValidationError{
			Field:   "filename",
			Code:    ErrorCodeInvalidFilename,
			Message: "Filename contains invalid characters",
		})
	}
	
	// Check for reserved names (Windows)
	reservedNames := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
	baseName := strings.ToUpper(strings.TrimSuffix(filename, filepath.Ext(filename)))
	for _, reserved := range reservedNames {
		if baseName == reserved {
			errors = append(errors, ValidationError{
				Field:   "filename",
				Code:    ErrorCodeInvalidFilename,
				Message: fmt.Sprintf("'%s' is a reserved filename", reserved),
			})
			break
		}
	}
	
	return errors
}

func ValidateFileSize(size *int64) []ValidationError {
	var errors []ValidationError
	
	if size == nil {
		errors = append(errors, ValidationError{
			Field:   "size",
			Code:    ErrorCodeSizeRequired,
			Message: "File size is required",
		})
		return errors
	}
	
	if *size <= 0 {
		errors = append(errors, ValidationError{
			Field:   "size",
			Code:    ErrorCodeInvalidSize,
			Message: "File size must be greater than 0",
		})
	}
	
	if *size > MaxFileSize {
		errors = append(errors, ValidationError{
			Field:   "size",
			Code:    ErrorCodeFileTooLarge,
			Message: fmt.Sprintf("File size cannot exceed %d bytes (%.1f GB)", MaxFileSize, float64(MaxFileSize)/(1024*1024*1024)),
		})
	}
	
	return errors
}

func ValidateMimeType(mimeType, filename string) []ValidationError {
	var errors []ValidationError
	
	// Check if MIME type is allowed
	if !AllowedMimeTypes[mimeType] {
		errors = append(errors, ValidationError{
			Field:   "mime_type",
			Code:    ErrorCodeInvalidFileType,
			Message: fmt.Sprintf("File type '%s' is not allowed", mimeType),
		})
	}
	
	// Validate MIME type matches file extension
	if filename != "" {
		ext := strings.ToLower(filepath.Ext(filename))
		expectedMimeType := mime.TypeByExtension(ext)
		if expectedMimeType != "" && expectedMimeType != mimeType {
			errors = append(errors, ValidationError{
				Field:   "mime_type",
				Code:    ErrorCodeInvalidFileType,
				Message: fmt.Sprintf("MIME type '%s' does not match file extension '%s'", mimeType, ext),
			})
		}
	}
	
	return errors
}

// User validation functions

type UserRegistrationRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func ValidateUserRegistration(req *UserRegistrationRequest) []ValidationError {
	var errors []ValidationError
	
	// Validate username
	if usernameErrors := ValidateUsername(req.Username); len(usernameErrors) > 0 {
		errors = append(errors, usernameErrors...)
	}
	
	// Validate email
	if emailErrors := ValidateEmail(req.Email); len(emailErrors) > 0 {
		errors = append(errors, emailErrors...)
	}
	
	// Validate password
	if passwordErrors := ValidatePassword(req.Password); len(passwordErrors) > 0 {
		errors = append(errors, passwordErrors...)
	}
	
	return errors
}

func ValidateUsername(username string) []ValidationError {
	var errors []ValidationError
	
	username = strings.TrimSpace(username)
	
	if username == "" {
		errors = append(errors, ValidationError{
			Field:   "username",
			Code:    ErrorCodeUsernameRequired,
			Message: "Username is required",
		})
		return errors
	}
	
	if len(username) < MinUsernameLength {
		errors = append(errors, ValidationError{
			Field:   "username",
			Code:    ErrorCodeUsernameTooShort,
			Message: fmt.Sprintf("Username must be at least %d characters long", MinUsernameLength),
		})
	}
	
	if len(username) > MaxUsernameLength {
		errors = append(errors, ValidationError{
			Field:   "username",
			Code:    ErrorCodeUsernameTooLong,
			Message: fmt.Sprintf("Username must be less than %d characters long", MaxUsernameLength),
		})
	}
	
	// Username should contain only alphanumeric characters, underscores, and hyphens
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validUsername.MatchString(username) {
		errors = append(errors, ValidationError{
			Field:   "username",
			Code:    ErrorCodeInvalidUsername,
			Message: "Username can only contain letters, numbers, underscores, and hyphens",
		})
	}
	
	return errors
}

func ValidateEmail(email string) []ValidationError {
	var errors []ValidationError
	
	email = strings.TrimSpace(email)
	
	if email == "" {
		errors = append(errors, ValidationError{
			Field:   "email",
			Code:    ErrorCodeEmailRequired,
			Message: "Email is required",
		})
		return errors
	}
	
	// Comprehensive email regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		errors = append(errors, ValidationError{
			Field:   "email",
			Code:    ErrorCodeInvalidEmail,
			Message: "Invalid email format",
		})
	}
	
	return errors
}

func ValidatePassword(password string) []ValidationError {
	var errors []ValidationError
	
	if password == "" {
		errors = append(errors, ValidationError{
			Field:   "password",
			Code:    ErrorCodePasswordRequired,
			Message: "Password is required",
		})
		return errors
	}
	
	if len(password) < MinPasswordLength {
		errors = append(errors, ValidationError{
			Field:   "password",
			Code:    ErrorCodePasswordTooShort,
			Message: fmt.Sprintf("Password must be at least %d characters long", MinPasswordLength),
		})
	}
	
	if len(password) > MaxPasswordLength {
		errors = append(errors, ValidationError{
			Field:   "password",
			Code:    ErrorCodePasswordTooLong,
			Message: fmt.Sprintf("Password must be less than %d characters long", MaxPasswordLength),
		})
	}
	
	// Check password complexity
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)
	
	complexityCount := 0
	if hasLower { complexityCount++ }
	if hasUpper { complexityCount++ }
	if hasNumber { complexityCount++ }
	if hasSpecial { complexityCount++ }
	
	if complexityCount < 3 {
		errors = append(errors, ValidationError{
			Field:   "password",
			Code:    ErrorCodePasswordTooWeak,
			Message: "Password must contain at least 3 of the following: lowercase letters, uppercase letters, numbers, special characters",
		})
	}
	
	return errors
}


// FormatValidationErrors formats multiple validation errors into a single error response
func FormatValidationErrors(errors []ValidationError) (ErrorCode, string, string) {
	if len(errors) == 0 {
		return ErrorCodeValidation, "Validation failed", ""
	}
	
	if len(errors) == 1 {
		return errors[0].Code, errors[0].Message, fmt.Sprintf("Field: %s", errors[0].Field)
	}
	
	// Multiple errors
	var messages []string
	var details []string
	for _, err := range errors {
		messages = append(messages, err.Message)
		details = append(details, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	
	return ErrorCodeValidation, "Multiple validation errors", strings.Join(details, "; ")
}