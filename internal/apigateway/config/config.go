package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	FileServiceURL string
	Environment    string // dev, staging, prod
}

func Load() *Config {
	// Load .env file if it exists (ignore errors for production)
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or error loading .env file: %v", err)
	}

	env := getEnv("ENVIRONMENT", "dev")
	cfg := &Config{
		Port:           getEnv("API_GATEWAY_PORT", getDefaultPort(env)),
		FileServiceURL: getRequiredEnv("FILE_SERVICE_URL"),
		Environment:    env,
	}

	validateConfig(cfg)
	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getRequiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return value
}

func getDefaultPort(env string) string {
	switch env {
	case "prod":
		return "80"   // Standard HTTP port
	case "staging":
		return "8080"
	default: // dev
		return "8080"
	}
}

func validateConfig(cfg *Config) {
	var errors []string
	
	if cfg.FileServiceURL == "" {
		errors = append(errors, "FILE_SERVICE_URL must be set")
	}
	
	if cfg.Environment != "dev" && strings.Contains(cfg.FileServiceURL, "localhost") {
		errors = append(errors, "FILE_SERVICE_URL should not use localhost in non-dev environments")
	}
	
	if len(errors) > 0 {
		log.Fatalf("Configuration validation failed:\n%s", strings.Join(errors, "\n"))
	}
}