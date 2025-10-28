.PHONY: api-gateway file-service clean test build

# Build targets
build: build-api-gateway build-file-service

build-api-gateway:
	go build -o bin/api-gateway cmd/apigateway/main.go

build-file-service:
	go build -o bin/file-service cmd/fileservice/main.go

# Run targets
api-gateway:
	go run cmd/apigateway/main.go

file-service:
	go run cmd/fileservice/main.go

# Development targets
dev: api-gateway

# Test targets
test:
	go test ./...

# Health check
health:
	curl -s http://localhost:8080/health | jq .

health-file-service:
	curl -s http://localhost:8081/health | jq .

# Clean targets
clean:
	rm -rf bin/