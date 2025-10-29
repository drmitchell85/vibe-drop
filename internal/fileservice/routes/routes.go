package routes

import (
	"vibe-drop/internal/fileservice/config"
	"vibe-drop/internal/fileservice/handlers"
	"vibe-drop/internal/fileservice/storage"

	"github.com/gorilla/mux"
)

func SetupRoutes(cfg *config.Config, s3Client *storage.S3Client, dynamoClient *storage.DynamoClient) *mux.Router {
	// S3 client is now passed in from server.go
	r := mux.NewRouter()

	// Health check
	r.HandleFunc("/health", handlers.HealthHandler).Methods("GET")

	// File operations - pass clients to handlers that need them
	r.Handle("/files/upload-url", handlers.GenerateUploadURLHandler(s3Client, dynamoClient)).Methods("POST")
	r.Handle("/files", handlers.ListFilesHandler(dynamoClient)).Methods("GET")
	r.Handle("/files/{id}", handlers.GetFileMetadataHandler(dynamoClient)).Methods("GET")
	r.Handle("/files/{id}/download-url", handlers.GenerateDownloadURLHandler(s3Client, dynamoClient)).Methods("GET")
	r.Handle("/files/{id}", handlers.DeleteFileHandler(s3Client, dynamoClient)).Methods("DELETE")

	return r
}
