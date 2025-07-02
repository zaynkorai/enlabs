.PHONY: build run test clean docker-build docker-up docker-down

APP_NAME=enlabs-api
BUILD_DIR=bin
MIGRATION_DIR=migrations

build:
	@echo "Building Go application..."
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server/main.go

run: build
	@echo "Running application..."
	./$(BUILD_DIR)/$(APP_NAME)

test:
	@echo "Running tests..."
	go test ./...

clean:
	@echo "Cleaning up build artifacts..."
	rm -rf $(BUILD_DIR)

docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):latest .

docker-up:
	@echo "Starting Docker Compose services..."
	docker compose up --build -d

docker-down:
	@echo "Stopping Docker Compose services..."
	docker compose down --remove-orphans

docker-down-volumes:
	@echo "Stopping Docker Compose services and removing volumes..."
	docker compose down -v --remove-orphans