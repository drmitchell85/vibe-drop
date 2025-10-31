package routes

import (
	"time"
	"vibe-drop/internal/auth"
	"vibe-drop/internal/fileservice/config"
	"vibe-drop/internal/fileservice/handlers"
	"vibe-drop/internal/fileservice/storage"

	"github.com/gorilla/mux"
)

func SetupRoutes(cfg *config.Config, s3Client *storage.S3Client, dynamoClient *storage.DynamoClient) *mux.Router {
	// S3 client is now passed in from server.go
	r := mux.NewRouter()

	// Create auth services
	jwtService := auth.NewJWTService("your-jwt-secret-key-change-in-production", time.Hour)
	passwordService := auth.NewPasswordService()
	authServices := &handlers.AuthServices{
		JWTService:      jwtService,
		PasswordService: passwordService,
		DynamoClient:    dynamoClient,
	}

	// Health check (no auth needed)
	r.HandleFunc("/health", handlers.HealthHandler).Methods("GET")

	// Authentication endpoints (no auth needed)
	r.Handle("/auth/register", handlers.RegisterHandler(authServices)).Methods("POST")
	r.Handle("/auth/login", handlers.LoginHandler(authServices)).Methods("POST")

	// File operations - pass clients to handlers that need them
	r.Handle("/files/upload-url", handlers.GenerateUploadURLHandler(s3Client, dynamoClient)).Methods("POST")
	r.Handle("/files", handlers.ListFilesHandler(dynamoClient)).Methods("GET")
	r.Handle("/files/{id}", handlers.GetFileMetadataHandler(dynamoClient)).Methods("GET")
	r.Handle("/files/{id}/download-url", handlers.GenerateDownloadURLHandler(s3Client, dynamoClient)).Methods("GET")
	r.Handle("/files/{id}", handlers.DeleteFileHandler(s3Client, dynamoClient)).Methods("DELETE")
	
	// Chunk completion for multipart uploads
	r.Handle("/files/{fileId}/chunks/{chunkNumber}/complete", handlers.ChunkCompletionHandler(dynamoClient)).Methods("POST")
	
	// Complete multipart upload
	r.Handle("/files/{fileId}/complete", handlers.CompleteMultipartUploadHandler(s3Client, dynamoClient)).Methods("POST")

	return r
}
