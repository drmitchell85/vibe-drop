package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"vibe-drop/internal/fileservice/storage"
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

func GenerateUploadURLHandler(s3Client *storage.S3Client, dynamoClient *storage.DynamoClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request to get filename
		var req struct {
			Filename string `json:"filename"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Filename == "" {
			http.Error(w, "Filename is required", http.StatusBadRequest)
			return
		}

		// Generate presigned URL
		url, fileID, err := s3Client.GenerateUploadURL(context.Background(), req.Filename)
		if err != nil {
			http.Error(w, "Failed to generate upload URL", http.StatusInternalServerError)
			return
		}

		// Save file metadata to DynamoDB
		metadata := &storage.FileMetadata{
			FileID:      fileID,
			Filename:    req.Filename,
			TotalSize:   0, // Will be updated when upload completes (future enhancement)
			ContentType: "application/octet-stream", // Default, could be inferred from filename
			Status:      "uploading", // Will be "completed" when upload finishes
			UploadType:  "single",
			UploadedAt:  time.Now().Format(time.RFC3339),
			UserID:      "default-user", // TODO: Replace with real user ID from auth
			S3Key:       fileID + "-" + req.Filename, // This matches S3Client.GenerateUploadURL key format
		}

		if err := dynamoClient.SaveFileMetadata(context.Background(), metadata); err != nil {
			log.Printf("Warning: Failed to save file metadata: %v", err)
			// Don't fail the request - S3 URL is still valid
		}

		response := PresignedURLResponse{
			URL:       url,
			ExpiresAt: time.Now().Add(15 * time.Minute),
			FileID:    fileID,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func GenerateDownloadURLHandler(s3Client *storage.S3Client, dynamoClient *storage.DynamoClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fileID := vars["id"]

		// Look up file metadata from DynamoDB to get the correct S3 key
		metadata, err := dynamoClient.GetFileMetadata(context.Background(), fileID)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		// Generate presigned URL using the correct S3 key from metadata
		url, err := s3Client.GenerateDownloadURL(context.Background(), metadata.S3Key)
		if err != nil {
			http.Error(w, "Failed to generate download URL", http.StatusInternalServerError)
			return
		}

		response := PresignedURLResponse{
			URL:       url,
			ExpiresAt: time.Now().Add(15 * time.Minute),
			FileID:    fileID,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func GetFileMetadataHandler(dynamoClient *storage.DynamoClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fileID := vars["id"]

		// Get real file metadata from DynamoDB
		metadata, err := dynamoClient.GetFileMetadata(context.Background(), fileID)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		// Convert to response format (matches existing API)
		response := FileMetadata{
			ID:          metadata.FileID,
			Filename:    metadata.Filename,
			Size:        metadata.TotalSize,
			ContentType: metadata.ContentType,
			UploadedAt:  parseTime(metadata.UploadedAt),
			UserID:      metadata.UserID,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func ListFilesHandler(dynamoClient *storage.DynamoClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get real files from DynamoDB for default user
		// TODO: Replace with real user ID from auth
		metadataList, err := dynamoClient.ListUserFiles(context.Background(), "default-user")
		if err != nil {
			http.Error(w, "Failed to list files", http.StatusInternalServerError)
			return
		}

		// Convert to response format
		files := make([]FileMetadata, len(metadataList))
		for i, metadata := range metadataList {
			files[i] = FileMetadata{
				ID:          metadata.FileID,
				Filename:    metadata.Filename,
				Size:        metadata.TotalSize,
				ContentType: metadata.ContentType,
				UploadedAt:  parseTime(metadata.UploadedAt),
				UserID:      metadata.UserID,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"files": files,
			"count": len(files),
		})
	}
}

func DeleteFileHandler(s3Client *storage.S3Client, dynamoClient *storage.DynamoClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fileID := vars["id"]

		// Get file metadata to find S3 key
		metadata, err := dynamoClient.GetFileMetadata(context.Background(), fileID)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		// Delete from S3 first (fail fast if S3 deletion fails)
		if err := s3Client.DeleteObject(context.Background(), metadata.S3Key); err != nil {
			log.Printf("Failed to delete S3 object %s: %v", metadata.S3Key, err)
			http.Error(w, "Failed to delete file from storage", http.StatusInternalServerError)
			return
		}

		// Delete metadata from DynamoDB (only after S3 deletion succeeds)
		if err := dynamoClient.DeleteFileMetadata(context.Background(), fileID); err != nil {
			log.Printf("Warning: S3 object deleted but DynamoDB cleanup failed for %s: %v", fileID, err)
			http.Error(w, "File deleted but metadata cleanup failed", http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"message": "File deleted successfully",
			"file_id": fileID,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// parseTime converts RFC3339 string to time.Time, with fallback to current time
func parseTime(timeStr string) time.Time {
	if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
		return t
	}
	return time.Now() // Fallback
}