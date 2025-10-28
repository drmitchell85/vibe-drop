package fileservice

import (
	"context"
	"log"
	"net/http"
	"time"

	"vibe-drop/internal/fileservice/config"
	"vibe-drop/internal/fileservice/routes"
	"vibe-drop/internal/fileservice/storage"
)

var server *http.Server

func Start() {
	cfg := config.Load()
	
	// Initialize S3 client
	s3Client, err := storage.NewS3Client(cfg.S3Bucket, cfg.S3Region, cfg.S3Endpoint)
	if err != nil {
		log.Fatalf("Failed to create S3 client: %v", err)
	}

	// Test S3 connection
	if err := s3Client.TestConnection(context.Background()); err != nil {
		log.Printf("Warning: S3 connection test failed: %v", err)
	}

	// Initialize DynamoDB client
	dynamoClient, err := storage.NewDynamoClient(cfg.DynamoRegion, cfg.DynamoEndpoint)
	if err != nil {
		log.Fatalf("Failed to create DynamoDB client: %v", err)
	}

	// Test DynamoDB connection
	if err := dynamoClient.TestConnection(context.Background()); err != nil {
		log.Printf("Warning: DynamoDB connection test failed: %v", err)
	}
	
	router := routes.SetupRoutes(cfg, s3Client)

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