package routes

import (
	"github.com/gorilla/mux"
	"vibe-drop/internal/apigateway/handlers"
	"vibe-drop/internal/apigateway/middleware"
)

func SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	// Apply middleware to all routes (order matters!)
	r.Use(middleware.RequestLogging())
	r.Use(middleware.DefaultRateLimit())

	// Health check
	r.HandleFunc("/health", handlers.HealthHandler).Methods("GET")

	// File service routes
	fileRouter := r.PathPrefix("/files").Subrouter()
	fileRouter.HandleFunc("", handlers.ListFilesHandler).Methods("GET")
	fileRouter.HandleFunc("", handlers.UploadFileHandler).Methods("POST")
	fileRouter.HandleFunc("/{id}", handlers.GetFileMetadataHandler).Methods("GET")
	fileRouter.HandleFunc("/{id}/download", handlers.DownloadFileHandler).Methods("GET")
	fileRouter.HandleFunc("/{id}", handlers.DeleteFileHandler).Methods("DELETE")

	// Auth service routes
	authRouter := r.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/login", handlers.LoginHandler).Methods("POST")
	authRouter.HandleFunc("/register", handlers.RegisterHandler).Methods("POST")
	authRouter.HandleFunc("/refresh", handlers.RefreshTokenHandler).Methods("POST")

	// User service routes
	userRouter := r.PathPrefix("/users").Subrouter()
	userRouter.HandleFunc("/me", handlers.GetCurrentUserHandler).Methods("GET")
	userRouter.HandleFunc("/{id}", handlers.GetUserProfileHandler).Methods("GET")
	userRouter.HandleFunc("/{id}", handlers.UpdateUserProfileHandler).Methods("PUT")

	return r
}