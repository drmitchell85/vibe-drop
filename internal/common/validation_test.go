package common

import (
	"testing"
)

func TestValidateFileUpload(t *testing.T) {
	tests := []struct {
		name     string
		req      *FileUploadRequest
		wantErrs int
	}{
		{
			name: "valid file upload",
			req: &FileUploadRequest{
				Filename: "test.pdf",
				Size:     intPtr(1024),
				MimeType: "application/pdf",
			},
			wantErrs: 0,
		},
		{
			name: "empty filename",
			req: &FileUploadRequest{
				Filename: "",
				Size:     intPtr(1024),
			},
			wantErrs: 1,
		},
		{
			name: "file too large",
			req: &FileUploadRequest{
				Filename: "test.pdf",
				Size:     intPtr(MaxFileSize + 1),
			},
			wantErrs: 1,
		},
		{
			name: "invalid filename characters",
			req: &FileUploadRequest{
				Filename: "test<>:.pdf",
				Size:     intPtr(1024),
			},
			wantErrs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateFileUpload(tt.req)
			if len(errors) != tt.wantErrs {
				t.Errorf("ValidateFileUpload() = %d errors, want %d", len(errors), tt.wantErrs)
				for _, err := range errors {
					t.Logf("  Error: %s - %s", err.Field, err.Message)
				}
			}
		})
	}
}

func TestValidateUserRegistration(t *testing.T) {
	tests := []struct {
		name     string
		req      *UserRegistrationRequest
		wantErrs int
	}{
		{
			name: "valid registration",
			req: &UserRegistrationRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "SecurePass123!",
			},
			wantErrs: 0,
		},
		{
			name: "weak password",
			req: &UserRegistrationRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "weak",
			},
			wantErrs: 2, // too short + too weak
		},
		{
			name: "invalid email",
			req: &UserRegistrationRequest{
				Username: "testuser",
				Email:    "invalid-email",
				Password: "SecurePass123!",
			},
			wantErrs: 1,
		},
		{
			name: "username too short",
			req: &UserRegistrationRequest{
				Username: "ab",
				Email:    "test@example.com",
				Password: "SecurePass123!",
			},
			wantErrs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateUserRegistration(tt.req)
			if len(errors) != tt.wantErrs {
				t.Errorf("ValidateUserRegistration() = %d errors, want %d", len(errors), tt.wantErrs)
				for _, err := range errors {
					t.Logf("  Error: %s - %s (%s)", err.Field, err.Message, err.Code)
				}
			}
		})
	}
}

func intPtr(i int64) *int64 {
	return &i
}