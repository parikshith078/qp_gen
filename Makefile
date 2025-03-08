# Default variables
DOCKER_COMPOSE = docker compose
DEV_COMPOSE_FILES = -f docker-compose.yaml -f docker-compose.dev.yaml
DOCKER_REGISTRY = pariksh1th
IMAGE_NAME = broker-service
SQLC = docker run --rm -v $(PWD)/broker-service:/src -w /src kjconroy/sqlc

# testing
# Version management
VERSION = $(shell cat VERSION)
BUILD_DATE = $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
COMMIT_SHA = $(shell git rev-parse --short HEAD)

# database
include .env
export

DSN = postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable

# Colors for terminal output
COLOR_RESET = \033[0m
COLOR_GREEN = \033[32m
COLOR_YELLOW = \033[33m

.PHONY: up down build dev clean restart logs help version bump-patch bump-minor bump-major tag migrate-create migrate-up migrate-down db-prod-up db-prod-down db-dev-up db-dev-down sqlc-init sqlc-generate

# Default target when just running 'make'
help:
		@echo "$(COLOR_GREEN)Available commands:$(COLOR_RESET)"
		@echo "$(COLOR_YELLOW)make up$(COLOR_RESET)		   - Start production containers"
		@echo "$(COLOR_YELLOW)make down$(COLOR_RESET)	   - Stop and remove containers"
		@echo "$(COLOR_YELLOW)make build$(COLOR_RESET)	   - Build production containers"
		@echo "$(COLOR_YELLOW)make dev$(COLOR_RESET)	   - Start development environment with hot-reload"
		@echo "$(COLOR_YELLOW)make clean$(COLOR_RESET)	   - Clean up all containers and volumes"
		@echo "$(COLOR_YELLOW)make restart$(COLOR_RESET)   - Restart all containers"
		@echo "$(COLOR_YELLOW)make logs$(COLOR_RESET)	   - View container logs"
		@echo "$(COLOR_YELLOW)make version$(COLOR_RESET)   - Show current version"
		@echo "$(COLOR_YELLOW)make bump-patch$(COLOR_RESET) - Bump patch version (1.0.0 -> 1.0.1)"
		@echo "$(COLOR_YELLOW)make bump-minor$(COLOR_RESET) - Bump minor version (1.0.0 -> 1.1.0)"
		@echo "$(COLOR_YELLOW)make bump-major$(COLOR_RESET) - Bump major version (1.0.0 -> 2.0.0)"
		@echo "$(COLOR_YELLOW)make tag$(COLOR_RESET)		  - Build and tag the current version"
		@echo "$(COLOR_YELLOW)make migrate-create$(COLOR_RESET) - Create a new migration"
		@echo "$(COLOR_YELLOW)make migrate-up$(COLOR_RESET) - Run migrations up"
		@echo "$(COLOR_YELLOW)make migrate-down$(COLOR_RESET) - Run migrations down"

# Start production containers
up:
		@echo "$(COLOR_GREEN)Starting production containers version $(VERSION)...$(COLOR_RESET)"
		@VERSION=$(VERSION) BUILD_DATE=$(BUILD_DATE) COMMIT_SHA=$(COMMIT_SHA) $(DOCKER_COMPOSE) up --build -d

# Start development environment
dev:
		@echo "$(COLOR_GREEN)Starting development environment...$(COLOR_RESET)"
		@$(DOCKER_COMPOSE) $(DEV_COMPOSE_FILES) up --build

# Stop containers
down:
		@echo "$(COLOR_GREEN)Stopping containers...$(COLOR_RESET)"
		@$(DOCKER_COMPOSE) down

# Build containers
build:
		@echo "$(COLOR_GREEN)Building containers version $(VERSION)...$(COLOR_RESET)"
		@VERSION=$(VERSION) BUILD_DATE=$(BUILD_DATE) COMMIT_SHA=$(COMMIT_SHA) $(DOCKER_COMPOSE) build

# Clean up containers, volumes, and build cache
clean:
		@echo "$(COLOR_GREEN)Cleaning up...$(COLOR_RESET)"
		@$(DOCKER_COMPOSE) down -v
		@docker system prune -f
		@rm -rf broker-service/tmp
		@rm -rf tmp

# Restart containers
restart:
	@echo "$(COLOR_GREEN)Restarting containers...$(COLOR_RESET)"
	@$(DOCKER_COMPOSE) restart

# View logs
logs:
	@echo "$(COLOR_GREEN)Viewing logs...$(COLOR_RESET)"
	@$(DOCKER_COMPOSE) logs -f

# Version management commands
version:
	@echo "Current version: $(VERSION)"

bump-patch:
	@echo "$(VERSION)" | awk -F. '{$$NF = $$NF + 1;} 1' | sed 's/ /./g' > VERSION
	@echo "Bumped patch version to $$(cat VERSION)"

bump-minor:
	@echo "$(VERSION)" | awk -F. '{$$2 = $$2 + 1; $$3 = 0;} 1' | sed 's/ /./g' > VERSION
	@echo "Bumped minor version to $$(cat VERSION)"

bump-major:
	@echo "$(VERSION)" | awk -F. '{$$1 = $$1 + 1; $$2 = 0; $$3 = 0;} 1' | sed 's/ /./g' > VERSION
	@echo "Bumped major version to $$(cat VERSION)"

tag: build
	@echo "$(COLOR_GREEN)Tagging version $(VERSION)...$(COLOR_RESET)"
	@git tag -a v$(VERSION) -m "Version $(VERSION)"
	@git push --follow-tags
	@echo "$(COLOR_GREEN)Pushing Docker images version $(VERSION)...$(COLOR_RESET)"
	@docker tag $(IMAGE_NAME):$(VERSION) $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(VERSION)
	@docker tag $(IMAGE_NAME):$(VERSION) $(DOCKER_REGISTRY)/$(IMAGE_NAME):latest
	@docker push $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(VERSION)
	@docker push $(DOCKER_REGISTRY)/$(IMAGE_NAME):latest
	@echo "$(COLOR_GREEN)Successfully pushed version $(VERSION) to Docker registry$(COLOR_RESET)"

migrate-create:
	@read -p "Enter migration name: " name; \
	docker run -v $$(pwd)/broker-service/migrations:/migrations --network host migrate/migrate \
		create -ext sql -dir /migrations -seq $$name

migrate-up:
	@echo "$(COLOR_GREEN)Running migrations up...$(COLOR_RESET)"
	@docker run -v $$(pwd)/broker-service/migrations:/migrations --network host migrate/migrate \
		-path=/migrations/ -database "$(DSN)" up

migrate-down:
	@echo "$(COLOR_GREEN)Running migrations down...$(COLOR_RESET)"
	@docker run -v $$(pwd)/broker-service/migrations:/migrations --network host migrate/migrate \
		-path=/migrations/ -database "$(DSN)" down 1

# Database container management
db-prod-up:
	@echo "$(COLOR_GREEN)Starting production database...$(COLOR_RESET)"
	@$(DOCKER_COMPOSE) up -d postgres

db-prod-down:
	@echo "$(COLOR_GREEN)Stopping production database...$(COLOR_RESET)"
	@$(DOCKER_COMPOSE) stop postgres

db-dev-up:
	@echo "$(COLOR_GREEN)Starting development database...$(COLOR_RESET)"
	@$(DOCKER_COMPOSE) $(DEV_COMPOSE_FILES) up -d postgres

db-dev-down:
	@echo "$(COLOR_GREEN)Stopping development database...$(COLOR_RESET)"
	@$(DOCKER_COMPOSE) $(DEV_COMPOSE_FILES) stop postgres

sqlc-init:
	@echo "$(COLOR_GREEN)Initializing SQLC...$(COLOR_RESET)"
	@$(SQLC) init

sqlc-generate:
	@echo "$(COLOR_GREEN)Generating SQLC code...$(COLOR_RESET)"
	@$(SQLC) generate

