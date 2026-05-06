ifneq (,$(wildcard .env))
	include .env
	export
endif

APP_NAME := youtube-tracker
GO := go

CMD := ./main.go
BINARY := bin/$(APP_NAME)

# FLAGS (puedes sobreescribir)
LOG_LEVEL ?= info

# =========================
# PHONY
# =========================
.PHONY: help build run discover metrics daily clean test tidy fmt lint docker-up docker-down logs


# =========================
# HELP
# =========================
help:
	@echo "Available commands:"
	@echo "  make build"
	@echo "  make run CMD='discover --interval=60'"
	@echo "  make discover"
	@echo "  make metrics"
	@echo "  make daily"
	@echo "  make docker-up"
	@echo "  make docker-down"


# =========================
# BUILD
# =========================
build:
	@echo "Building binary..."
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	$(GO) build -ldflags="-s -w" -o $(BINARY) $(CMD)


# =========================
# RUN GENÉRICO
# =========================
run:
	@echo "Running $(CMD)..."
	$(GO) run $(CMD) $(CMD_ARGS)


# =========================
# SUBCOMMANDS
# =========================

discover:
	@echo "Running discovery..."
	$(GO) run $(CMD) discover --log-level=$(LOG_LEVEL)

metrics:
	@echo "Running metrics..."
	$(GO) run $(CMD) metrics --log-level=$(LOG_LEVEL)

daily:
	@echo "Running daily..."
	$(GO) run $(CMD) daily --log-level=$(LOG_LEVEL)


# =========================
# BINARY EXEC
# =========================

run-bin:
	@echo "Running binary..."
	$(BINARY) $(CMD_ARGS)


# =========================
# DEV
# =========================
dev:
	@echo "Running dev..."
	air


# =========================
# CLEAN
# =========================
clean:
	rm -rf bin


# =========================
# TEST
# =========================
test:
	$(GO) test ./... -v


# =========================
# TOOLS
# =========================
tidy:
	$(GO) mod tidy

fmt:
	$(GO) fmt ./...

lint:
	golangci-lint run


# =========================
# DOCKER
# =========================
docker-up:
	docker compose up -d

docker-down:
	docker compose down

logs:
	docker compose logs -f
