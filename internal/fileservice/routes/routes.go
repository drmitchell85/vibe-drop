package routes

import (
	"vibe-drop/internal/fileservice/config"
	"vibe-drop/internal/fileservice/handlers"
	"vibe-drop/internal/fileservice/storage"

	"github.com/gorilla/mux"
)

func SetupRoutes(cfg *config.Config, s3Client *storage.S3Client) *mux.Router {
	// S3 client is now passed in from server.go
	r := mux.NewRouter()

	// Health check
	r.HandleFunc("/health", handlers.HealthHandler).Methods("GET")

	// File operations - pass s3Client to handlers that need it
	r.Handle("/files/upload-url", handlers.GenerateUploadURLHandler(s3Client)).Methods("POST")
	r.HandleFunc("/files", handlers.ListFilesHandler).Methods("GET")
	r.HandleFunc("/files/{id}", handlers.GetFileMetadataHandler).Methods("GET")
	r.Handle("/files/{id}/download-url", handlers.GenerateDownloadURLHandler(s3Client)).Methods("GET")
	r.HandleFunc("/files/{id}", handlers.DeleteFileHandler).Methods("DELETE")

	return r
}
