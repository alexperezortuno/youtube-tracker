
ifneq (,$(wildcard .env))
	include .env
	export
endif

APP_NAME := youtube-tracker
GO := go

CMD := ./cmd/main.go

BINARY := bin/$(APP_NAME)

ENV_FILE := .env

# PHONY TARGETS

.PHONY: help build run dev clean test tidy fmt lint docker-up docker-down logs db-init


# HELP
help:
	@echo "Available commands:"
	@echo "  make build         - Build binary"
	@echo "  make run           - Run app"
	@echo "  make dev           - Run with auto-reload (air)"
	@echo "  make clean         - Clean binaries"
	@echo "  make test          - Run tests"
	@echo "  make tidy          - go mod tidy"
	@echo "  make fmt           - Format code"
	@echo "  make lint          - Run linter"
	@echo "  make docker-up     - Start infra (redis + db)"
	@echo "  make docker-down   - Stop infra"
	@echo "  make logs          - Tail docker logs"
	@echo "  make db-init       - Init DB schema"


# BUILD
build:
	@echo "Building optimized binary..."
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	$(GO) build -ldflags="-s -w" -o $(BINARY) $(CMD)


# RUN
run:
	@echo "Running..."
	$(GO) run $(CMD)


# DEV (HOT RELOAD)
dev:
	@echo "Running in dev mode (air)..."
	air


# CLEAN
clean:
	@echo "Cleaning..."
	rm -rf bin


# TEST
test:
	@echo "Running tests..."
	$(GO) test ./... -v


# DEPENDENCIES
tidy:
	@echo "Tidying modules..."
	$(GO) mod tidy

fmt:
	@echo "Formatting..."
	$(GO) fmt ./...

lint:
	@echo "Linting..."
	golangci-lint run


# DOCKER INFRA
docker-up:
	@echo "Starting docker services..."
	docker-compose up -d

docker-down:
	@echo "Stopping docker services..."
	docker-compose down

logs:
	docker-compose logs -f


# DB INIT
db-init:
	@echo "Initializing database..."
	docker exec -i youtube-tracker-db-1 psql -U user -d metrics < scripts/init.sql
