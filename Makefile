.PHONY: help build run test clean migrate-up migrate-down sqlc proto docker-up docker-down

help:
	@echo "Control Plane Microservice - Available Commands:"
	@echo ""
	@echo "  make build          - Build the application"
	@echo "  make run            - Run the application"
	@echo "  make test           - Run tests"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make migrate-up     - Run database migrations"
	@echo "  make migrate-down   - Rollback database migrations"
	@echo "  make sqlc           - Generate SQLC code"
	@echo "  make proto          - Generate protobuf code"
	@echo "  make docker-up      - Start docker services"
	@echo "  make docker-down    - Stop docker services"
	@echo "  make install-deps   - Install development dependencies"
	@echo "  make lint           - Run linters"

build:
	@echo "Building application..."
	go build -o bin/cp_service cmd/main_new.go

run:
	@echo "Running application..."
	go run cmd/main_new.go

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

migrate-up:
	@echo "Running database migrations..."
	psql "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)" \
		-f internal/adapters/database/migration/001_init.up.sql

migrate-down:
	@echo "Rolling back database migrations..."
	psql "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)" \
		-f internal/adapters/database/migration/001_init.down.sql

sqlc:
	@echo "Generating SQLC code..."
	sqlc generate

proto:
	@echo "Generating protobuf code..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/*.proto

docker-up:
	@echo "Starting Docker services..."
	docker-compose up -d

docker-down:
	@echo "Stopping Docker services..."
	docker-compose down

install-deps:
	@echo "Installing development dependencies..."
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

lint:
	@echo "Running linters..."
	go fmt ./...
	go vet ./...

mod-tidy:
	@echo "Tidying go.mod..."
	go mod tidy

mod-download:
	@echo "Downloading dependencies..."
	go mod download

setup: install-deps mod-download sqlc docker-up migrate-up
	@echo "Setup complete! Run 'make run' to start the service."

dev: docker-up
	@echo "Starting in development mode..."
	go run cmd/main_new.go
