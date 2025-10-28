package apigateway

import (
	"context"
	"log"
	"net/http"
	"time"

	"vibe-drop/internal/apigateway/config"
	"vibe-drop/internal/apigateway/routes"
)

var server *http.Server

func Start() {
	cfg := config.Load()
	router := routes.SetupRoutes(cfg)

	server = &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	log.Printf("API Gateway starting on port %s...", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("API Gateway failed to start:", err)
	}
}

func Stop() {
	if server != nil {
		log.Println("Shutting down API Gateway...")
		
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		} else {
			log.Println("API Gateway stopped gracefully")
		}
	}
}