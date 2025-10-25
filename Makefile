.PHONY: api-gateway clean test build

# Build targets
build: build-api-gateway

build-api-gateway:
	go build -o bin/api-gateway cmd/apigateway/main.go

# Run targets
api-gateway:
	go run cmd/apigateway/main.go

# Development targets
dev: api-gateway

# Test targets
test:
	go test ./...

# Health check
health:
	curl -s http://localhost:8080/health | jq .

# Clean targets
clean:
	rm -rf bin/