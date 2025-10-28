package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port     string
	S3Bucket string
	S3Region string
}

func Load() *Config {
	// Load .env file if it exists (ignore errors for production)
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or error loading .env file: %v", err)
	}

	return &Config{
		Port:     getEnv("FILE_SERVICE_PORT", "8081"),
		S3Bucket: getEnv("S3_BUCKET", "vibe-drop-bucket"),
		S3Region: getEnv("S3_REGION", "us-east-1"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}