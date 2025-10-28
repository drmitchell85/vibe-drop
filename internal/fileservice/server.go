package fileservice

import (
	"context"
	"log"
	"net/http"
	"time"

	"vibe-drop/internal/fileservice/routes"
)

var server *http.Server

func Start() {
	router := routes.SetupRoutes()

	server = &http.Server{
		Addr:    ":8081",
		Handler: router,
	}

	log.Println("File Service starting on port 8081...")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("File Service failed to start:", err)
	}
}

func Stop() {
	if server != nil {
		log.Println("Shutting down File Service...")
		
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("File Service shutdown error: %v", err)
		} else {
			log.Println("File Service stopped gracefully")
		}
	}
}