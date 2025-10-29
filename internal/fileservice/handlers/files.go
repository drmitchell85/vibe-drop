package handlers

import (
	"context"
	"encoding/json"
	"fmt"
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

type uploadRequest struct {
	Filename string `json:"filename"`
	Size     *int64 `json:"size,omitempty"`
}

func parseUploadRequest(r *http.Request) (*uploadRequest, error) {
	var req uploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, fmt.Errorf("invalid request body")
	}
	if req.Filename == "" {
		return nil, fmt.Errorf("filename is required")
	}
	return &req, nil
}

func shouldUseMultipart(size *int64) bool {
	const multipartThreshold = 5 * 1024 * 1024 * 1024 // 5GB in bytes
	return size != nil && *size >= multipartThreshold
}

func handleMultipartUpload(s3Client *storage.S3Client, dynamoClient *storage.DynamoClient, req *uploadRequest) (PresignedURLResponse, error) {
	uploadInfo, err := s3Client.InitiateMultipartUpload(context.Background(), req.Filename)
	if err != nil {
		return PresignedURLResponse{}, fmt.Errorf("failed to initiate multipart upload: %w", err)
	}

	// Extract fileID from S3 key (format: uuid-filename)
	if len(uploadInfo.Key) < 37 { // UUID(36) + dash(1) = 37 minimum
		return PresignedURLResponse{}, fmt.Errorf("invalid S3 key format")
	}
	fileID := uploadInfo.Key[:36] // Extract the full UUID (36 characters)
	s3Key := uploadInfo.Key

	// Calculate chunk details
	chunkSize := int64(5 * 1024 * 1024 * 1024) // 5GB per chunk
	totalChunks := int((*req.Size + chunkSize - 1) / chunkSize) // Ceiling division

	// Generate presigned URLs for each chunk and create chunk records
	chunks, err := createChunksAndRecords(s3Client, dynamoClient, uploadInfo, fileID, totalChunks, chunkSize, *req.Size)
	if err != nil {
		return PresignedURLResponse{}, err
	}

	response := PresignedURLResponse{
		FileID:     fileID,
		UploadType: "multipart",
		Chunks:     chunks,
	}

	// Save multipart metadata
	if err := saveMultipartMetadata(dynamoClient, fileID, req.Filename, *req.Size, s3Key, uploadInfo.UploadID, chunkSize, totalChunks); err != nil {
		log.Printf("Warning: Failed to save multipart metadata: %v", err)
	}

	return response, nil
}

func createChunksAndRecords(s3Client *storage.S3Client, dynamoClient *storage.DynamoClient, uploadInfo *storage.MultipartUploadInfo, fileID string, totalChunks int, chunkSize int64, totalSize int64) ([]ChunkURL, error) {
	chunks := make([]ChunkURL, totalChunks)
	for i := 0; i < totalChunks; i++ {
		partNumber := i + 1 // S3 part numbers are 1-indexed
		chunkURL, err := s3Client.GenerateMultipartUploadURL(context.Background(), uploadInfo, partNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to generate chunk upload URL: %w", err)
		}

		// Calculate this chunk's size (last chunk may be smaller)
		currentChunkSize := chunkSize
		if i == totalChunks-1 {
			currentChunkSize = totalSize - int64(i)*chunkSize
		}

		chunks[i] = ChunkURL{
			ChunkNumber: partNumber,
			URL:         chunkURL,
			ExpiresAt:   time.Now().Add(15 * time.Minute),
			Size:        currentChunkSize,
		}

		// Create chunk record in DynamoDB
		chunkRecord := &storage.FileChunk{
			FileID:       fileID,
			ChunkNumber:  partNumber,
			Size:         currentChunkSize,
			Status:       "pending",
			S3PartNumber: partNumber,
		}
		if err := dynamoClient.SaveFileChunk(context.Background(), chunkRecord); err != nil {
			log.Printf("Warning: Failed to save chunk record: %v", err)
		}
	}
	return chunks, nil
}

func saveMultipartMetadata(dynamoClient *storage.DynamoClient, fileID, filename string, totalSize int64, s3Key, uploadID string, chunkSize int64, totalChunks int) error {
	totalChunksInt := totalChunks
	chunkSizeInt := chunkSize
	metadata := &storage.FileMetadata{
		FileID:      fileID,
		Filename:    filename,
		TotalSize:   totalSize,
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
	return dynamoClient.SaveFileMetadata(context.Background(), metadata)
}

func handleSingleUpload(s3Client *storage.S3Client, dynamoClient *storage.DynamoClient, req *uploadRequest) (PresignedURLResponse, error) {
	url, fileID, err := s3Client.GenerateUploadURL(context.Background(), req.Filename)
	if err != nil {
		return PresignedURLResponse{}, fmt.Errorf("failed to generate upload URL: %w", err)
	}

	s3Key := fileID + "-" + req.Filename
	response := PresignedURLResponse{
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

	return response, nil
}

func GenerateUploadURLHandler(s3Client *storage.S3Client, dynamoClient *storage.DynamoClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := parseUploadRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var response PresignedURLResponse
		if shouldUseMultipart(req.Size) {
			response, err = handleMultipartUpload(s3Client, dynamoClient, req)
		} else {
			response, err = handleSingleUpload(s3Client, dynamoClient, req)
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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

// CompleteMultipartUploadHandler handles completion of multipart uploads
func CompleteMultipartUploadHandler(s3Client *storage.S3Client, dynamoClient *storage.DynamoClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fileID := vars["fileId"]

		// Get file metadata to retrieve upload info
		metadata, err := dynamoClient.GetFileMetadata(context.Background(), fileID)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		// Verify this is a multipart upload
		if metadata.UploadType != "multipart" || metadata.S3UploadID == nil {
			http.Error(w, "Not a multipart upload", http.StatusBadRequest)
			return
		}

		// Check that all chunks are uploaded
		complete, chunks, err := dynamoClient.CheckUploadComplete(context.Background(), fileID)
		if err != nil {
			http.Error(w, "Failed to check upload status", http.StatusInternalServerError)
			return
		}

		if !complete {
			http.Error(w, "Not all chunks are uploaded yet", http.StatusBadRequest)
			return
		}

		// Prepare parts for S3 completion
		parts := make([]storage.CompletedPart, len(chunks))
		for i, chunk := range chunks {
			parts[i] = storage.CompletedPart{
				PartNumber: chunk.S3PartNumber,
				ETag:       chunk.ETag,
			}
		}

		// Complete the multipart upload in S3
		uploadInfo := &storage.MultipartUploadInfo{
			UploadID: *metadata.S3UploadID,
			Key:      metadata.S3Key,
		}

		if err := s3Client.CompleteMultipartUpload(context.Background(), uploadInfo, parts); err != nil {
			log.Printf("Failed to complete multipart upload: %v", err)
			http.Error(w, "Failed to complete upload", http.StatusInternalServerError)
			return
		}

		// Update file metadata status to "completed"
		metadata.Status = "completed"
		metadata.CompletedAt = &[]string{time.Now().Format(time.RFC3339)}[0]
		if err := dynamoClient.SaveFileMetadata(context.Background(), metadata); err != nil {
			log.Printf("Warning: Failed to update file status: %v", err)
		}

		response := map[string]interface{}{
			"message":       "Multipart upload completed successfully",
			"file_id":       fileID,
			"total_chunks":  len(chunks),
			"completed_at":  time.Now().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// ChunkCompletionHandler handles chunk upload completion notifications
func ChunkCompletionHandler(dynamoClient *storage.DynamoClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fileID := vars["fileId"]
		chunkNumberStr := vars["chunkNumber"]

		// Parse chunk number
		var chunkNumber int
		if _, err := fmt.Sscanf(chunkNumberStr, "%d", &chunkNumber); err != nil {
			http.Error(w, "Invalid chunk number", http.StatusBadRequest)
			return
		}

		// Parse request body for ETag
		var req struct {
			ETag   string `json:"etag"`
			Status string `json:"status"` // "uploaded" or "failed"
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate status
		if req.Status != "uploaded" && req.Status != "failed" {
			http.Error(w, "Status must be 'uploaded' or 'failed'", http.StatusBadRequest)
			return
		}

		// Update chunk status
		if err := dynamoClient.UpdateChunkStatus(context.Background(), fileID, chunkNumber, req.Status, req.ETag); err != nil {
			log.Printf("Failed to update chunk status: %v", err)
			http.Error(w, "Failed to update chunk status", http.StatusInternalServerError)
			return
		}

		// Check if upload is complete
		complete, chunks, err := dynamoClient.CheckUploadComplete(context.Background(), fileID)
		if err != nil {
			log.Printf("Failed to check upload completion: %v", err)
			http.Error(w, "Failed to check upload status", http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"chunk_number": chunkNumber,
			"status":       req.Status,
			"upload_complete": complete,
		}

		// If upload is complete, include completion details
		if complete {
			response["total_chunks"] = len(chunks)
			response["message"] = "All chunks uploaded successfully - ready for completion"
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}