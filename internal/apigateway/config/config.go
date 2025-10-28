package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	FileServiceURL string
}

func Load() *Config {
	// Load .env file if it exists (ignore errors for production)
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or error loading .env file: %v", err)
	}

	return &Config{
		Port:           getEnv("API_GATEWAY_PORT", "8080"),
		FileServiceURL: getEnv("FILE_SERVICE_URL", "http://localhost:8081"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}