package fileservice

import (
	"context"
	"log"
	"net/http"
	"time"

	"vibe-drop/internal/fileservice/config"
	"vibe-drop/internal/fileservice/routes"
)

var server *http.Server

func Start() {
	cfg := config.Load()
	router := routes.SetupRoutes(cfg)

	server = &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	log.Printf("File Service starting on port %s...", cfg.Port)
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