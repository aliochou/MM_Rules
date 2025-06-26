.PHONY: build run test clean docker-build docker-run k8s-deploy k8s-clean demo import-dashboard demo-rules load-rules test-rules manage-rules

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

# Run the rules demo script
demo-rules:
	./examples/rules-demo.sh

# Load predefined rule sets
load-rules:
	./scripts/load-rules.sh

# Test rule sets
test-rules:
	./scripts/test-rules.sh

# Manage rules (add, edit, delete)
manage-rules:
	./scripts/manage-rules.sh

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

# Start monitoring stack (Prometheus + Grafana)
monitoring:
	docker-compose up -d prometheus grafana

# Stop monitoring stack
monitoring-stop:
	docker-compose down

# Start full stack (Redis + Prometheus + Grafana)
full-stack:
	docker-compose up -d

# Stop full stack
full-stack-stop:
	docker-compose down

# View monitoring logs
monitoring-logs:
	docker-compose logs -f prometheus grafana

# View all logs
logs:
	docker-compose logs -f

# Check monitoring status
monitoring-status:
	docker-compose ps

# Open Prometheus in browser (macOS)
prometheus-open:
	open http://localhost:9090

# Open Grafana in browser (macOS)
grafana-open:
	open http://localhost:3000

# Import Grafana dashboard
import-dashboard:
	./scripts/import-dashboard.sh

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
	@echo "  demo-rules   - Run the rules demo script"
	@echo "  load-rules   - Load predefined rule sets"
	@echo "  test-rules   - Test rule sets"
	@echo "  manage-rules - Manage rules (add, edit, delete)"
	@echo "  deps         - Install dependencies"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code"
	@echo "  tidy         - Generate go.sum"
	@echo "  redis        - Start Redis for development"
	@echo "  redis-stop   - Stop Redis"
	@echo "  monitoring   - Start monitoring stack (Prometheus + Grafana)"
	@echo "  monitoring-stop - Stop monitoring stack"
	@echo "  full-stack   - Start full stack (Redis + Prometheus + Grafana)"
	@echo "  full-stack-stop - Stop full stack"
	@echo "  monitoring-logs - View monitoring logs"
	@echo "  logs         - View all logs"
	@echo "  monitoring-status - Check monitoring status"
	@echo "  prometheus-open - Open Prometheus in browser"
	@echo "  grafana-open - Open Grafana in browser"
	@echo "  import-dashboard - Import Grafana dashboard"
	@echo "  help         - Show this help" 