package routes

import (
	"github.com/gorilla/mux"
	"vibe-drop/internal/fileservice/handlers"
)

func SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	// Health check
	r.HandleFunc("/health", handlers.HealthHandler).Methods("GET")

	// File operations
	r.HandleFunc("/files/upload-url", handlers.GenerateUploadURLHandler).Methods("POST")
	r.HandleFunc("/files", handlers.ListFilesHandler).Methods("GET")
	r.HandleFunc("/files/{id}", handlers.GetFileMetadataHandler).Methods("GET")
	r.HandleFunc("/files/{id}/download-url", handlers.GenerateDownloadURLHandler).Methods("GET")
	r.HandleFunc("/files/{id}", handlers.DeleteFileHandler).Methods("DELETE")

	return r
}