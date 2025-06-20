.PHONY: build run test clean docker-build docker-run k8s-deploy k8s-clean demo

# Build the application
build:
	go build -o bin/server cmd/server/main.go

# Run the application locally
run:
	go run cmd/server/main.go

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out

# Build Docker image
docker-build:
	docker build -t mm-rules-matchmaking .

# Run Docker container
docker-run:
	docker run -p 8080:8080 \
		-e MM_RULES_REDIS_ADDR=host.docker.internal:6379 \
		mm-rules-matchmaking

# Deploy to Kubernetes
k8s-deploy:
	kubectl apply -f k8s/

# Clean Kubernetes deployment
k8s-clean:
	kubectl delete -f k8s/

# Run the demo script
demo:
	./examples/demo.sh

# Install dependencies
deps:
	go mod download

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Generate go.sum
tidy:
	go mod tidy

# Start Redis for development
redis:
	docker run -d --name mm-rules-redis -p 6379:6379 redis:7-alpine

# Stop Redis
redis-stop:
	docker stop mm-rules-redis
	docker rm mm-rules-redis

# Show help
help:
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application locally"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  k8s-deploy   - Deploy to Kubernetes"
	@echo "  k8s-clean    - Clean Kubernetes deployment"
	@echo "  demo         - Run the demo script"
	@echo "  deps         - Install dependencies"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code"
	@echo "  tidy         - Generate go.sum"
	@echo "  redis        - Start Redis for development"
	@echo "  redis-stop   - Stop Redis"
	@echo "  help         - Show this help" 