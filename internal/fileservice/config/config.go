package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	S3Bucket        string
	S3Region        string
	S3Endpoint      string // For LocalStack vs real AWS
	DynamoEndpoint  string // For LocalStack vs real AWS
	DynamoRegion    string
	Environment     string // dev, staging, prod
}

func Load() *Config {
	// Load .env file if it exists (ignore errors for production)
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or error loading .env file: %v", err)
	}

	env := getEnv("ENVIRONMENT", "dev")
	cfg := &Config{
		Port:           getEnv("FILE_SERVICE_PORT", getDefaultPort(env)),
		S3Bucket:       getRequiredEnv("S3_BUCKET"),
		S3Region:       getEnv("S3_REGION", getDefaultRegion(env)),
		S3Endpoint:     getS3Endpoint(env),
		DynamoEndpoint: getDynamoEndpoint(env),
		DynamoRegion:   getEnv("DYNAMO_REGION", getDefaultRegion(env)),
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
		return "8080" // Standard HTTP port in production
	case "staging":
		return "8081"
	default: // dev
		return "8081"
	}
}

func getDefaultRegion(env string) string {
	switch env {
	case "prod":
		return "us-west-2" // Common production region
	case "staging":
		return "us-west-2"
	default: // dev
		return "us-east-1" // LocalStack default
	}
}

func getS3Endpoint(env string) string {
	if endpoint := os.Getenv("S3_ENDPOINT"); endpoint != "" {
		return endpoint
	}
	
	switch env {
	case "prod", "staging":
		return "" // Use default AWS endpoint
	default: // dev
		return "http://localhost:4566" // LocalStack
	}
}

func getDynamoEndpoint(env string) string {
	if endpoint := os.Getenv("DYNAMO_ENDPOINT"); endpoint != "" {
		return endpoint
	}
	
	switch env {
	case "prod", "staging":
		return "" // Use default AWS endpoint
	default: // dev
		return "http://localhost:4566" // LocalStack
	}
}

func validateConfig(cfg *Config) {
	var errors []string
	
	if cfg.S3Bucket == "" {
		errors = append(errors, "S3_BUCKET must be set")
	}
	
	if cfg.Environment != "dev" && cfg.S3Endpoint != "" && strings.Contains(cfg.S3Endpoint, "localhost") {
		errors = append(errors, "S3_ENDPOINT should not use localhost in non-dev environments")
	}
	
	if len(errors) > 0 {
		log.Fatalf("Configuration validation failed:\n%s", strings.Join(errors, "\n"))
	}
}