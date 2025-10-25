package apigateway

import (
	"context"
	"log"
	"net/http"
	"time"

	"vibe-drop/internal/apigateway/routes"
)

var server *http.Server

func Start() {
	router := routes.SetupRoutes()

	server = &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	log.Println("API Gateway starting on port 8080...")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server failed to start:", err)
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