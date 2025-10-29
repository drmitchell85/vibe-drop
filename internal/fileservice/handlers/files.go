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
	URL        string    `json:"url,omitempty"`        // For single uploads
	ExpiresAt  time.Time `json:"expires_at,omitempty"` // For single uploads  
	FileID     string    `json:"file_id"`
	UploadType string    `json:"upload_type"`          // "single" or "multipart"
	Chunks     []ChunkURL `json:"chunks,omitempty"`    // For multipart uploads
}

type ChunkURL struct {
	ChunkNumber int       `json:"chunk_number"`
	URL         string    `json:"url"`
	ExpiresAt   time.Time `json:"expires_at"`
	Size        int64     `json:"size"` // Expected chunk size
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
		// Parse request to get filename and optional size
		var req struct {
			Filename string `json:"filename"`
			Size     *int64 `json:"size,omitempty"` // Optional: file size in bytes for multipart decision
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Filename == "" {
			http.Error(w, "Filename is required", http.StatusBadRequest)
			return
		}

		// Determine upload type based on file size
		const multipartThreshold = 5 * 1024 * 1024 * 1024 // 5GB in bytes
		useMultipart := req.Size != nil && *req.Size >= multipartThreshold

		var response PresignedURLResponse
		var fileID string
		var s3Key string

		if useMultipart {
			// Handle multipart upload
			uploadInfo, err := s3Client.InitiateMultipartUpload(context.Background(), req.Filename)
			if err != nil {
				http.Error(w, "Failed to initiate multipart upload", http.StatusInternalServerError)
				return
			}

			// Extract fileID from S3 key (format: uuid-filename)
			// UUID format: 8-4-4-4-12 characters, so we need to find where filename starts
			// Look for the pattern: uuid (36 chars) + dash + filename
			if len(uploadInfo.Key) < 37 { // UUID(36) + dash(1) = 37 minimum
				http.Error(w, "Invalid S3 key format", http.StatusInternalServerError)
				return
			}
			fileID = uploadInfo.Key[:36] // Extract the full UUID (36 characters)
			s3Key = uploadInfo.Key

			// Calculate chunk details
			chunkSize := int64(5 * 1024 * 1024 * 1024) // 5GB per chunk
			totalChunks := int((*req.Size + chunkSize - 1) / chunkSize) // Ceiling division

			// Generate presigned URLs for each chunk
			chunks := make([]ChunkURL, totalChunks)
			for i := 0; i < totalChunks; i++ {
				partNumber := i + 1 // S3 part numbers are 1-indexed
				chunkURL, err := s3Client.GenerateMultipartUploadURL(context.Background(), uploadInfo, partNumber)
				if err != nil {
					http.Error(w, "Failed to generate chunk upload URL", http.StatusInternalServerError)
					return
				}

				// Calculate this chunk's size (last chunk may be smaller)
				currentChunkSize := chunkSize
				if i == totalChunks-1 {
					currentChunkSize = *req.Size - int64(i)*chunkSize
				}

				chunks[i] = ChunkURL{
					ChunkNumber: partNumber,
					URL:         chunkURL,
					ExpiresAt:   time.Now().Add(15 * time.Minute),
					Size:        currentChunkSize,
				}
			}

			response = PresignedURLResponse{
				FileID:     fileID,
				UploadType: "multipart",
				Chunks:     chunks,
			}

			// Save multipart metadata
			totalChunksInt := totalChunks
			chunkSizeInt := chunkSize
			uploadID := uploadInfo.UploadID
			metadata := &storage.FileMetadata{
				FileID:      fileID,
				Filename:    req.Filename,
				TotalSize:   *req.Size,
				ContentType: "application/octet-stream",
				Status:      "uploading",
				UploadType:  "multipart",
				UploadedAt:  time.Now().Format(time.RFC3339),
				UserID:      "default-user",
				S3Key:       s3Key,
				S3UploadID:  &uploadID,
				ChunkSize:   &chunkSizeInt,
				TotalChunks: &totalChunksInt,
			}

			if err := dynamoClient.SaveFileMetadata(context.Background(), metadata); err != nil {
				log.Printf("Warning: Failed to save multipart metadata: %v", err)
			}

		} else {
			// Handle single upload (existing logic)
			url, singleFileID, err := s3Client.GenerateUploadURL(context.Background(), req.Filename)
			if err != nil {
				http.Error(w, "Failed to generate upload URL", http.StatusInternalServerError)
				return
			}

			fileID = singleFileID
			s3Key = fileID + "-" + req.Filename

			response = PresignedURLResponse{
				URL:        url,
				ExpiresAt:  time.Now().Add(15 * time.Minute),
				FileID:     fileID,
				UploadType: "single",
			}

			// Save single upload metadata
			totalSize := int64(0)
			if req.Size != nil {
				totalSize = *req.Size
			}
			metadata := &storage.FileMetadata{
				FileID:      fileID,
				Filename:    req.Filename,
				TotalSize:   totalSize,
				ContentType: "application/octet-stream",
				Status:      "uploading",
				UploadType:  "single",
				UploadedAt:  time.Now().Format(time.RFC3339),
				UserID:      "default-user",
				S3Key:       s3Key,
			}

			if err := dynamoClient.SaveFileMetadata(context.Background(), metadata); err != nil {
				log.Printf("Warning: Failed to save file metadata: %v", err)
			}
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