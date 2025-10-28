package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type PresignedURLResponse struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
	FileID    string    `json:"file_id"`
}

type FileMetadata struct {
	ID          string    `json:"id"`
	Filename    string    `json:"filename"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	UploadedAt  time.Time `json:"uploaded_at"`
	UserID      string    `json:"user_id"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func GenerateUploadURLHandler(w http.ResponseWriter, r *http.Request) {
	// Mock presigned URL for file upload
	response := PresignedURLResponse{
		URL:       "https://mock-s3.amazonaws.com/vibe-drop-bucket/mock-file-id-12345?X-Amz-Signature=mock",
		ExpiresAt: time.Now().Add(15 * time.Minute),
		FileID:    "mock-file-id-12345",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func GenerateDownloadURLHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["id"]

	// Mock presigned URL for file download
	response := PresignedURLResponse{
		URL:       "https://mock-s3.amazonaws.com/vibe-drop-bucket/" + fileID + "?X-Amz-Signature=mock-download",
		ExpiresAt: time.Now().Add(15 * time.Minute),
		FileID:    fileID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func GetFileMetadataHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["id"]

	// Mock file metadata
	response := FileMetadata{
		ID:          fileID,
		Filename:    "example-file.pdf",
		Size:        1024000,
		ContentType: "application/pdf",
		UploadedAt:  time.Now().Add(-24 * time.Hour),
		UserID:      "mock-user-123",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	// Mock file list
	files := []FileMetadata{
		{
			ID:          "file-1",
			Filename:    "document1.pdf",
			Size:        512000,
			ContentType: "application/pdf",
			UploadedAt:  time.Now().Add(-2 * time.Hour),
			UserID:      "mock-user-123",
		},
		{
			ID:          "file-2", 
			Filename:    "image.jpg",
			Size:        256000,
			ContentType: "image/jpeg",
			UploadedAt:  time.Now().Add(-1 * time.Hour),
			UserID:      "mock-user-123",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"files": files,
		"count": len(files),
	})
}

func DeleteFileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["id"]

	// Mock successful deletion
	response := map[string]interface{}{
		"message": "File deleted successfully",
		"file_id": fileID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}